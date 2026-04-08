// This file loads plugin backend declarations, converts published pluginhost
// resource contracts, and dispatches generic plugin hook and resource queries.

package plugin

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/pluginhost"
)

var safePluginIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// pluginHookSpec defines a plugin-owned hook handler declaration.
type pluginHookSpec struct {
	Event  pluginhost.ExtensionPoint `yaml:"event"`
	Action pluginhost.HookAction     `yaml:"action"`
	Table  string                    `yaml:"table"`
	Fields map[string]string         `yaml:"fields"`
}

// pluginResourceSpec defines a plugin-owned backend resource declaration.
type pluginResourceSpec struct {
	Key     string                 `yaml:"key"`
	Type    string                 `yaml:"type"`
	Table   string                 `yaml:"table"`
	Fields  []*pluginResourceField `yaml:"fields"`
	Filters []*pluginResourceQuery `yaml:"filters"`
	OrderBy pluginOrderBySpec      `yaml:"orderBy"`
}

// pluginResourceField defines one selected output field for a plugin resource.
type pluginResourceField struct {
	Name   string `yaml:"name"`
	Column string `yaml:"column"`
}

// pluginResourceQuery defines one query filter for a plugin resource.
type pluginResourceQuery struct {
	Param    string `yaml:"param"`
	Column   string `yaml:"column"`
	Operator string `yaml:"operator"`
}

// pluginOrderBySpec defines the order-by configuration for a plugin resource.
type pluginOrderBySpec struct {
	Column    string `yaml:"column"`
	Direction string `yaml:"direction"`
}

func normalizePluginResourceSpecType(value string) pluginResourceSpecType {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case pluginResourceSpecTypeTableList.String():
		return pluginResourceSpecTypeTableList
	default:
		return pluginResourceSpecType("")
	}
}

func normalizePluginResourceFilterOperator(value string) pluginResourceFilterOperator {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case pluginResourceFilterOperatorEQ.String():
		return pluginResourceFilterOperatorEQ
	case pluginResourceFilterOperatorLike.String():
		return pluginResourceFilterOperatorLike
	case pluginResourceFilterOperatorGTEDate.String():
		return pluginResourceFilterOperatorGTEDate
	case pluginResourceFilterOperatorLTEDate.String():
		return pluginResourceFilterOperatorLTEDate
	default:
		return pluginResourceFilterOperator("")
	}
}

func normalizePluginResourceOrderDirection(value string) pluginResourceOrderDirection {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case pluginResourceOrderDirectionASC.String():
		return pluginResourceOrderDirectionASC
	case pluginResourceOrderDirectionDESC.String():
		return pluginResourceOrderDirectionDESC
	default:
		return pluginResourceOrderDirection("")
	}
}

// ResourceListInput defines input for querying a plugin-owned backend resource.
type ResourceListInput struct {
	PluginID   string            // PluginID is the plugin identifier.
	ResourceID string            // ResourceID is the plugin-declared resource key.
	Filters    map[string]string // Filters contains query-string filters.
	PageNum    int               // PageNum is the requested page number.
	PageSize   int               // PageSize is the requested page size.
}

// ResourceListOutput defines output for querying a plugin-owned backend resource.
type ResourceListOutput struct {
	List  []map[string]interface{} // List contains the queried resource rows.
	Total int                      // Total is the total row count.
}

// loadPluginBackendConfig loads plugin-owned hook and resource declarations from plugin directory.
func (s *Service) loadPluginBackendConfig(manifest *pluginManifest) error {
	manifest.Hooks = make([]*pluginHookSpec, 0)
	manifest.BackendResources = make(map[string]*pluginResourceSpec)

	if sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID); ok {
		manifest.BackendResources = s.convertSourcePluginResources(sourcePlugin.GetResources())
		return nil
	}

	hookFiles, err := gfile.ScanDirFile(filepath.Join(manifest.RootDir, "backend", "hooks"), "*.yaml", false)
	if err != nil && !gfile.Exists(filepath.Join(manifest.RootDir, "backend", "hooks")) {
		err = nil
	}
	if err != nil {
		return err
	}
	for _, hookFile := range hookFiles {
		spec := &pluginHookSpec{}
		if err = s.loadPluginYAMLFile(hookFile, spec); err != nil {
			return err
		}
		if err = s.validatePluginHookSpec(manifest.ID, spec, hookFile); err != nil {
			return err
		}
		manifest.Hooks = append(manifest.Hooks, spec)
	}

	resourceFiles, err := gfile.ScanDirFile(filepath.Join(manifest.RootDir, "backend", "resources"), "*.yaml", false)
	if err != nil && !gfile.Exists(filepath.Join(manifest.RootDir, "backend", "resources")) {
		err = nil
	}
	if err != nil {
		return err
	}
	for _, resourceFile := range resourceFiles {
		spec := &pluginResourceSpec{}
		if err = s.loadPluginYAMLFile(resourceFile, spec); err != nil {
			return err
		}
		if err = s.validatePluginResourceSpec(manifest.ID, spec, resourceFile); err != nil {
			return err
		}
		manifest.BackendResources[spec.Key] = spec
	}
	return nil
}

// convertSourcePluginResources converts public source plugin resource declarations to internal runtime specs.
func (s *Service) convertSourcePluginResources(resources []*pluginhost.ResourceSpec) map[string]*pluginResourceSpec {
	items := make(map[string]*pluginResourceSpec, len(resources))
	for _, resource := range resources {
		if resource == nil {
			continue
		}

		// Convert the published pluginhost contract into the internal resource shape
		// used by generic query execution.
		fields := make([]*pluginResourceField, 0, len(resource.Fields))
		for _, field := range resource.Fields {
			if field == nil {
				continue
			}
			fields = append(fields, &pluginResourceField{
				Name:   field.Name,
				Column: field.Column,
			})
		}

		filters := make([]*pluginResourceQuery, 0, len(resource.Filters))
		for _, filter := range resource.Filters {
			if filter == nil {
				continue
			}
			filters = append(filters, &pluginResourceQuery{
				Param:    filter.Param,
				Column:   filter.Column,
				Operator: filter.Operator,
			})
		}

		orderBy := pluginOrderBySpec{}
		if resource.OrderBy != nil {
			orderBy = pluginOrderBySpec{
				Column:    resource.OrderBy.Column,
				Direction: resource.OrderBy.Direction,
			}
		}

		items[resource.Key] = &pluginResourceSpec{
			Key:     resource.Key,
			Type:    resource.Type,
			Table:   resource.Table,
			Fields:  fields,
			Filters: filters,
			OrderBy: orderBy,
		}
	}
	return items
}

// ListResourceRecords queries plugin-owned backend resource rows using the generic source-plugin contract.
func (s *Service) ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error) {
	manifest, err := s.getPluginManifestByID(in.PluginID)
	if err != nil {
		return nil, err
	}
	if !s.IsEnabled(ctx, in.PluginID) {
		return nil, gerror.New("插件未启用")
	}

	resource, ok := manifest.BackendResources[in.ResourceID]
	if !ok {
		return nil, gerror.New("插件资源不存在")
	}
	if in.PageNum <= 0 {
		in.PageNum = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}

	m := g.DB().Model(resource.Table).Safe().Ctx(ctx)
	for _, filter := range resource.Filters {
		value := strings.TrimSpace(in.Filters[filter.Param])
		if value == "" {
			continue
		}
		switch normalizePluginResourceFilterOperator(filter.Operator) {
		case pluginResourceFilterOperatorEQ:
			m = m.Where(filter.Column, value)
		case pluginResourceFilterOperatorLike:
			m = m.WhereLike(filter.Column, "%"+value+"%")
		case pluginResourceFilterOperatorGTEDate:
			m = m.WhereGTE(filter.Column, value+" 00:00:00")
		case pluginResourceFilterOperatorLTEDate:
			m = m.WhereLTE(filter.Column, value+" 23:59:59")
		default:
			return nil, gerror.Newf("插件资源过滤操作符不支持: %s", filter.Operator)
		}
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	fields := make([]string, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		fields = append(fields, fmt.Sprintf("%s AS %s", field.Column, field.Name))
	}
	fieldArgs := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		fieldArgs = append(fieldArgs, field)
	}

	orderBy := resource.OrderBy.Column
	if normalizePluginResourceOrderDirection(resource.OrderBy.Direction) == pluginResourceOrderDirectionDESC {
		orderBy += " DESC"
	} else {
		orderBy += " ASC"
	}
	records, err := m.Fields(fieldArgs...).Page(in.PageNum, in.PageSize).Order(orderBy).All()
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		recordMap := record.Map()
		row := make(map[string]interface{}, len(resource.Fields))
		for _, field := range resource.Fields {
			row[field.Name] = s.normalizePluginResourceValue(recordMap[field.Name])
		}
		items = append(items, row)
	}
	return &ResourceListOutput{List: items, Total: total}, nil
}

// DispatchHookEvent dispatches one named hook event to all enabled source plugins.
func (s *Service) DispatchHookEvent(
	ctx context.Context,
	eventName pluginhost.ExtensionPoint,
	payload map[string]interface{},
) error {
	manifests, err := s.scanPluginManifests()
	if err != nil {
		return err
	}
	targetPluginID, _ := payload["pluginId"].(string)
	for _, manifest := range manifests {
		if !s.shouldDispatchHookToPlugin(ctx, manifest.ID, eventName, targetPluginID) {
			continue
		}
		for _, hook := range manifest.Hooks {
			if hook.Event != eventName || hook.Action != pluginhost.HookActionInsert {
				continue
			}
			startedAt := gtime.Now()
			if err = s.executePluginInsertHook(ctx, manifest.ID, hook, payload); err != nil {
				g.Log().Warningf(ctx, "plugin hook failed plugin=%s event=%s cost=%s err=%v", manifest.ID, eventName, gtime.Now().Sub(startedAt), err)
				continue
			}
			g.Log().Infof(ctx, "plugin hook succeeded plugin=%s event=%s cost=%s", manifest.ID, eventName, gtime.Now().Sub(startedAt))
		}
		s.executeSourcePluginHookHandlers(ctx, manifest.ID, eventName, payload)
	}
	return nil
}

// executeSourcePluginHookHandlers executes registered callback-style hook handlers for one source plugin.
func (s *Service) executeSourcePluginHookHandlers(
	ctx context.Context,
	pluginID string,
	eventName pluginhost.ExtensionPoint,
	payload map[string]interface{},
) {
	sourcePlugin, ok := pluginhost.GetSourcePlugin(pluginID)
	if !ok {
		return
	}
	for _, item := range sourcePlugin.GetHookHandlers() {
		if item == nil || item.Point != eventName || item.Handler == nil {
			continue
		}
		s.executeSourcePluginHookHandler(ctx, pluginID, item, payload)
	}
}

func (s *Service) executeSourcePluginHookHandler(
	ctx context.Context,
	pluginID string,
	item *pluginhost.HookHandlerRegistration,
	payload map[string]interface{},
) {
	if item == nil || item.Handler == nil {
		return
	}

	execute := func(executeCtx context.Context, values map[string]interface{}, async bool) {
		startedAt := gtime.Now()
		if err := item.Handler(executeCtx, pluginhost.NewHookPayload(item.Point, values)); err != nil {
			if async {
				g.Log().Warningf(executeCtx, "plugin async callback hook failed plugin=%s event=%s cost=%s err=%v", pluginID, item.Point, gtime.Now().Sub(startedAt), err)
				return
			}
			g.Log().Warningf(executeCtx, "plugin callback hook failed plugin=%s event=%s cost=%s err=%v", pluginID, item.Point, gtime.Now().Sub(startedAt), err)
			return
		}
		if async {
			g.Log().Infof(executeCtx, "plugin async callback hook succeeded plugin=%s event=%s cost=%s", pluginID, item.Point, gtime.Now().Sub(startedAt))
			return
		}
		g.Log().Infof(executeCtx, "plugin callback hook succeeded plugin=%s event=%s cost=%s", pluginID, item.Point, gtime.Now().Sub(startedAt))
	}

	values := cloneHookPayloadValues(payload)
	if item.Mode == pluginhost.CallbackExecutionModeAsync {
		// Clone payload values before spawning a goroutine so plugin callbacks never
		// race on the mutable request-scoped map owned by the caller.
		go execute(context.WithoutCancel(ctx), values, true)
		return
	}
	execute(ctx, values, false)
}

func cloneHookPayloadValues(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return map[string]interface{}{}
	}
	cloned := make(map[string]interface{}, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

// shouldDispatchHookToPlugin determines whether the hook event should be delivered to the target plugin.
func (s *Service) shouldDispatchHookToPlugin(
	ctx context.Context,
	pluginID string,
	eventName pluginhost.ExtensionPoint,
	targetPluginID string,
) bool {
	switch eventName {
	case pluginhost.ExtensionPointPluginInstalled,
		pluginhost.ExtensionPointPluginEnabled,
		pluginhost.ExtensionPointPluginDisabled,
		pluginhost.ExtensionPointPluginUninstalled:
		return pluginID == targetPluginID
	default:
		return s.IsEnabled(ctx, pluginID)
	}
}

// HandleAuthLoginSucceeded handles login succeeded hooks declared by source plugins.
func (s *Service) HandleAuthLoginSucceeded(ctx context.Context, input AuthLoginSucceededInput) error {
	if input.ClientType == "" {
		input.ClientType = "web"
	}
	if input.Message == "" {
		input.Message = "登录成功"
	}
	return s.DispatchHookEvent(ctx, pluginhost.ExtensionPointAuthLoginSucceeded, map[string]interface{}{
		"userName":   input.UserName,
		"status":     input.Status,
		"ip":         input.Ip,
		"clientType": input.ClientType,
		"browser":    input.Browser,
		"os":         input.Os,
		"message":    input.Message,
	})
}

// HandleAuthLoginFailed handles login failed hooks declared by source plugins.
func (s *Service) HandleAuthLoginFailed(ctx context.Context, input AuthLoginSucceededInput) error {
	if input.ClientType == "" {
		input.ClientType = "web"
	}
	if input.Message == "" {
		input.Message = "登录失败"
	}
	return s.DispatchHookEvent(ctx, pluginhost.ExtensionPointAuthLoginFailed, map[string]interface{}{
		"userName":   input.UserName,
		"status":     input.Status,
		"ip":         input.Ip,
		"clientType": input.ClientType,
		"browser":    input.Browser,
		"os":         input.Os,
		"message":    input.Message,
	})
}

// HandleAuthLogoutSucceeded handles logout succeeded hooks declared by source plugins.
func (s *Service) HandleAuthLogoutSucceeded(ctx context.Context, input AuthLoginSucceededInput) error {
	if input.ClientType == "" {
		input.ClientType = "web"
	}
	if input.Message == "" {
		input.Message = "登出成功"
	}
	return s.DispatchHookEvent(ctx, pluginhost.ExtensionPointAuthLogoutSucceeded, map[string]interface{}{
		"userName":   input.UserName,
		"status":     input.Status,
		"ip":         input.Ip,
		"clientType": input.ClientType,
		"browser":    input.Browser,
		"os":         input.Os,
		"message":    input.Message,
	})
}

// getPluginManifestByID returns one plugin manifest by plugin ID.
func (s *Service) getPluginManifestByID(pluginID string) (*pluginManifest, error) {
	if pluginID == "" {
		return nil, gerror.New("插件ID不能为空")
	}
	manifests, err := s.scanPluginManifests()
	if err != nil {
		return nil, err
	}
	for _, manifest := range manifests {
		if manifest.ID == pluginID {
			return manifest, nil
		}
	}
	return nil, gerror.New("插件不存在")
}

// loadPluginYAMLFile loads a YAML file into the target structure.
func (s *Service) loadPluginYAMLFile(filePath string, target interface{}) error {
	content := gfile.GetBytes(filePath)
	if len(content) == 0 {
		return gerror.Newf("插件配置文件为空: %s", filePath)
	}
	if err := yaml.Unmarshal(content, target); err != nil {
		return gerror.Wrapf(err, "解析插件配置文件失败: %s", filePath)
	}
	return nil
}

// validatePluginHookSpec validates one plugin-owned hook declaration.
func (s *Service) validatePluginHookSpec(pluginID string, spec *pluginHookSpec, filePath string) error {
	if spec.Event == "" {
		return gerror.Newf("插件Hook缺少event: %s", filePath)
	}
	if !pluginhost.IsHookExtensionPoint(spec.Event) {
		return gerror.Newf("插件Hook插槽未发布: %s", filePath)
	}
	if !pluginhost.IsSupportedHookAction(spec.Action) {
		return gerror.Newf("插件Hook动作仅支持insert: %s", filePath)
	}
	if err := s.validatePluginIdentifier(spec.Table); err != nil {
		return gerror.Wrapf(err, "插件%s的Hook表名非法: %s", pluginID, filePath)
	}
	if len(spec.Fields) == 0 {
		return gerror.Newf("插件Hook缺少fields映射: %s", filePath)
	}
	for column := range spec.Fields {
		if err := s.validatePluginIdentifier(column); err != nil {
			return gerror.Wrapf(err, "插件%s的Hook字段非法: %s", pluginID, filePath)
		}
	}
	return nil
}

// validatePluginResourceSpec validates one plugin-owned backend resource declaration.
func (s *Service) validatePluginResourceSpec(pluginID string, spec *pluginResourceSpec, filePath string) error {
	if spec.Key == "" {
		return gerror.Newf("插件资源缺少key: %s", filePath)
	}
	if spec.Type == "" {
		spec.Type = pluginResourceSpecTypeTableList.String()
	}
	if normalizePluginResourceSpecType(spec.Type) != pluginResourceSpecTypeTableList {
		return gerror.Newf("插件资源类型仅支持table-list: %s", filePath)
	}
	if err := s.validatePluginIdentifier(spec.Table); err != nil {
		return gerror.Wrapf(err, "插件%s资源表名非法: %s", pluginID, filePath)
	}
	if len(spec.Fields) == 0 {
		return gerror.Newf("插件资源缺少fields定义: %s", filePath)
	}
	for _, field := range spec.Fields {
		if field == nil {
			return gerror.Newf("插件资源字段不能为空: %s", filePath)
		}
		if err := s.validatePluginIdentifier(field.Name); err != nil {
			return gerror.Wrapf(err, "插件%s资源字段名称非法: %s", pluginID, filePath)
		}
		if err := s.validatePluginIdentifier(field.Column); err != nil {
			return gerror.Wrapf(err, "插件%s资源列名非法: %s", pluginID, filePath)
		}
	}
	for _, filter := range spec.Filters {
		if filter == nil {
			return gerror.Newf("插件资源过滤器不能为空: %s", filePath)
		}
		if filter.Param == "" {
			return gerror.Newf("插件资源过滤器缺少param: %s", filePath)
		}
		if err := s.validatePluginIdentifier(filter.Column); err != nil {
			return gerror.Wrapf(err, "插件%s资源过滤列非法: %s", pluginID, filePath)
		}
		if normalizePluginResourceFilterOperator(filter.Operator) == "" {
			return gerror.Newf("插件资源过滤操作符不支持: %s", filePath)
		}
	}
	if err := s.validatePluginIdentifier(spec.OrderBy.Column); err != nil {
		return gerror.Wrapf(err, "插件%s资源排序列非法: %s", pluginID, filePath)
	}
	if spec.OrderBy.Direction == "" {
		spec.OrderBy.Direction = pluginResourceOrderDirectionASC.String()
	}
	if normalizePluginResourceOrderDirection(spec.OrderBy.Direction) == "" {
		return gerror.Newf("插件资源排序方向仅支持 asc/desc: %s", filePath)
	}
	return nil
}

// validatePluginIdentifier validates table and column names from plugin-owned declarations.
func (s *Service) validatePluginIdentifier(value string) error {
	if value == "" {
		return gerror.New("插件标识不能为空")
	}
	if !safePluginIdentifierPattern.MatchString(value) {
		return gerror.Newf("插件标识非法: %s", value)
	}
	return nil
}

// executePluginInsertHook executes a generic insert hook declared by a source plugin.
func (s *Service) executePluginInsertHook(ctx context.Context, pluginID string, hook *pluginHookSpec, payload map[string]interface{}) error {
	columns := make([]string, 0, len(hook.Fields))
	for column := range hook.Fields {
		columns = append(columns, column)
	}
	sort.Strings(columns)

	values := make([]interface{}, 0, len(columns))
	placeholders := make([]string, 0, len(columns))
	for _, column := range columns {
		expr := hook.Fields[column]
		value, err := s.resolvePluginHookValue(expr, payload)
		if err != nil {
			return gerror.Wrapf(err, "解析插件%s的Hook字段失败: %s", pluginID, column)
		}
		values = append(values, value)
		placeholders = append(placeholders, "?")
	}

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		hook.Table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	_, err := g.DB().Exec(ctx, sql, values...)
	return err
}

// resolvePluginHookValue resolves one hook field expression.
func (s *Service) resolvePluginHookValue(expr string, payload map[string]interface{}) (interface{}, error) {
	if expr == "now" {
		return gtime.Now(), nil
	}
	if strings.HasPrefix(expr, "event.") {
		fieldName := strings.TrimPrefix(expr, "event.")
		if value, ok := payload[fieldName]; ok {
			return value, nil
		}
		return nil, gerror.Newf("Hook事件字段不存在: %s", fieldName)
	}
	return nil, gerror.Newf("不支持的Hook字段表达式: %s", expr)
}

// normalizePluginResourceValue converts GoFrame time values to JSON-safe strings.
func (s *Service) normalizePluginResourceValue(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case *gtime.Time:
		if typedValue == nil {
			return ""
		}
		return typedValue.String()
	case gtime.Time:
		return typedValue.String()
	default:
		return value
	}
}
