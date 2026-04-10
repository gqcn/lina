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
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

var safePluginIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// pluginHookSpec defines a plugin-owned hook handler declaration.
type pluginHookSpec struct {
	Event        pluginhost.ExtensionPoint        `json:"event" yaml:"event"`
	Action       pluginhost.HookAction            `json:"action" yaml:"action"`
	Mode         pluginhost.CallbackExecutionMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	Table        string                           `json:"table,omitempty" yaml:"table,omitempty"`
	Fields       map[string]string                `json:"fields,omitempty" yaml:"fields,omitempty"`
	TimeoutMs    int                              `json:"timeoutMs,omitempty" yaml:"timeoutMs,omitempty"`
	SleepMs      int                              `json:"sleepMs,omitempty" yaml:"sleepMs,omitempty"`
	ErrorMessage string                           `json:"errorMessage,omitempty" yaml:"errorMessage,omitempty"`
}

// pluginResourceSpec defines a plugin-owned backend resource declaration.
type pluginResourceSpec struct {
	Key       string                       `json:"key" yaml:"key"`
	Type      string                       `json:"type" yaml:"type"`
	Table     string                       `json:"table" yaml:"table"`
	Fields    []*pluginResourceField       `json:"fields" yaml:"fields"`
	Filters   []*pluginResourceQuery       `json:"filters" yaml:"filters"`
	OrderBy   pluginOrderBySpec            `json:"orderBy" yaml:"orderBy"`
	DataScope *pluginResourceDataScopeSpec `json:"dataScope,omitempty" yaml:"dataScope,omitempty"`
}

// pluginResourceField defines one selected output field for a plugin resource.
type pluginResourceField struct {
	Name   string `json:"name" yaml:"name"`
	Column string `json:"column" yaml:"column"`
}

// pluginResourceQuery defines one query filter for a plugin resource.
type pluginResourceQuery struct {
	Param    string `json:"param" yaml:"param"`
	Column   string `json:"column" yaml:"column"`
	Operator string `json:"operator" yaml:"operator"`
}

// pluginOrderBySpec defines the order-by configuration for a plugin resource.
type pluginOrderBySpec struct {
	Column    string `json:"column" yaml:"column"`
	Direction string `json:"direction" yaml:"direction"`
}

// pluginResourceDataScopeSpec defines how one plugin resource binds to host role data scopes.
type pluginResourceDataScopeSpec struct {
	UserColumn string `json:"userColumn,omitempty" yaml:"userColumn,omitempty"`
	DeptColumn string `json:"deptColumn,omitempty" yaml:"deptColumn,omitempty"`
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

const pluginHookEventFieldExprPrefix = "event."

// loadPluginBackendConfig loads plugin-owned hook and resource declarations from plugin directory.
func (s *Service) loadPluginBackendConfig(manifest *pluginManifest) error {
	manifest.Hooks = make([]*pluginHookSpec, 0)
	manifest.BackendResources = make(map[string]*pluginResourceSpec)

	if sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID); ok {
		manifest.BackendResources = s.convertSourcePluginResources(sourcePlugin.GetResources())
		return nil
	}

	if manifest.RuntimeArtifact != nil {
		manifest.Hooks = clonePluginHookSpecs(manifest.RuntimeArtifact.HookSpecs)
		manifest.BackendResources = clonePluginResourceSpecsToMap(manifest.RuntimeArtifact.ResourceSpecs)
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

func clonePluginHookSpecs(items []*pluginHookSpec) []*pluginHookSpec {
	if len(items) == 0 {
		return []*pluginHookSpec{}
	}

	cloned := make([]*pluginHookSpec, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := *item
		if len(item.Fields) > 0 {
			next.Fields = make(map[string]string, len(item.Fields))
			for key, value := range item.Fields {
				next.Fields[key] = value
			}
		}
		cloned = append(cloned, &next)
	}
	return cloned
}

func clonePluginResourceSpecsToMap(items []*pluginResourceSpec) map[string]*pluginResourceSpec {
	if len(items) == 0 {
		return map[string]*pluginResourceSpec{}
	}

	cloned := make(map[string]*pluginResourceSpec, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := clonePluginResourceSpec(item)
		cloned[next.Key] = next
	}
	return cloned
}

func clonePluginResourceSpec(item *pluginResourceSpec) *pluginResourceSpec {
	if item == nil {
		return nil
	}

	next := *item
	if len(item.Fields) > 0 {
		next.Fields = make([]*pluginResourceField, 0, len(item.Fields))
		for _, field := range item.Fields {
			if field == nil {
				continue
			}
			fieldCopy := *field
			next.Fields = append(next.Fields, &fieldCopy)
		}
	}
	if len(item.Filters) > 0 {
		next.Filters = make([]*pluginResourceQuery, 0, len(item.Filters))
		for _, filter := range item.Filters {
			if filter == nil {
				continue
			}
			filterCopy := *filter
			next.Filters = append(next.Filters, &filterCopy)
		}
	}
	if item.DataScope != nil {
		dataScopeCopy := *item.DataScope
		next.DataScope = &dataScopeCopy
	}
	return &next
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
	m, err = s.applyPluginResourceDataScope(ctx, m, resource)
	if err != nil {
		return nil, err
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

	runtime, runtimeErr := s.buildFilterRuntimeFromManifests(ctx, manifests)
	if runtimeErr != nil {
		logger.Warningf(ctx, "load plugin enablement runtime for hook dispatch failed: %v", runtimeErr)
	}

	targetPluginID := pluginhost.HookPayloadStringValue(payload, pluginhost.HookPayloadKeyPluginID)
	for _, manifest := range manifests {
		if !s.shouldDispatchHookToPlugin(ctx, runtime, manifest.ID, eventName, targetPluginID) {
			continue
		}
		for _, hook := range manifest.Hooks {
			if hook == nil || hook.Event != eventName {
				continue
			}
			s.executePluginDeclaredHook(ctx, manifest.ID, hook, payload)
		}
		s.executeSourcePluginHookHandlers(ctx, manifest.ID, eventName, payload)
	}
	return nil
}

func (s *Service) executePluginDeclaredHook(
	ctx context.Context,
	pluginID string,
	hook *pluginHookSpec,
	payload map[string]interface{},
) {
	if hook == nil {
		return
	}

	execute := func(executeCtx context.Context, hookPayload map[string]interface{}, async bool) {
		var (
			timeoutCtx context.Context
			cancel     context.CancelFunc
		)
		timeoutCtx, cancel = s.buildPluginHookTimeoutContext(executeCtx, hook)
		defer cancel()

		startedAt := gtime.Now()
		err := s.runPluginDeclaredHook(timeoutCtx, pluginID, hook, hookPayload)
		if err != nil {
			if async {
				logger.Warningf(timeoutCtx, "plugin async declared hook failed plugin=%s event=%s action=%s cost=%s err=%v", pluginID, hook.Event, hook.Action, gtime.Now().Sub(startedAt), err)
				return
			}
			logger.Warningf(timeoutCtx, "plugin declared hook failed plugin=%s event=%s action=%s cost=%s err=%v", pluginID, hook.Event, hook.Action, gtime.Now().Sub(startedAt), err)
			return
		}
		if async {
			logger.Infof(timeoutCtx, "plugin async declared hook succeeded plugin=%s event=%s action=%s cost=%s", pluginID, hook.Event, hook.Action, gtime.Now().Sub(startedAt))
			return
		}
		logger.Infof(timeoutCtx, "plugin declared hook succeeded plugin=%s event=%s action=%s cost=%s", pluginID, hook.Event, hook.Action, gtime.Now().Sub(startedAt))
	}

	values := pluginhost.CloneHookPayloadValues(payload)
	mode := s.normalizePluginHookMode(hook)
	if mode == pluginhost.CallbackExecutionModeAsync {
		go execute(context.WithoutCancel(ctx), values, true)
		return
	}
	execute(ctx, values, false)
}

func (s *Service) buildPluginHookTimeoutContext(
	ctx context.Context,
	hook *pluginHookSpec,
) (context.Context, context.CancelFunc) {
	timeout := 3 * time.Second
	if hook != nil && hook.TimeoutMs > 0 {
		timeout = time.Duration(hook.TimeoutMs) * time.Millisecond
	}
	return context.WithTimeout(ctx, timeout)
}

func (s *Service) normalizePluginHookMode(hook *pluginHookSpec) pluginhost.CallbackExecutionMode {
	if hook == nil {
		return pluginhost.CallbackExecutionModeBlocking
	}
	mode := hook.Mode
	if mode == "" {
		mode = pluginhost.DefaultCallbackExecutionMode(hook.Event)
	}
	if !pluginhost.IsExtensionPointExecutionModeSupported(hook.Event, mode) {
		return pluginhost.DefaultCallbackExecutionMode(hook.Event)
	}
	return mode
}

func (s *Service) runPluginDeclaredHook(
	ctx context.Context,
	pluginID string,
	hook *pluginHookSpec,
	payload map[string]interface{},
) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = gerror.Newf("plugin declared hook panicked: %v", recovered)
		}
	}()

	if hook == nil {
		return nil
	}

	switch hook.Action {
	case pluginhost.HookActionInsert:
		err = s.executePluginInsertHook(ctx, pluginID, hook, payload)
	case pluginhost.HookActionSleep:
		err = s.executePluginSleepHook(ctx, hook)
	case pluginhost.HookActionError:
		err = s.executePluginErrorHook(hook)
	default:
		err = gerror.Newf("插件 Hook 动作不支持: %s", hook.Action)
	}
	if err != nil {
		if ctx.Err() != nil {
			return gerror.Wrapf(ctx.Err(), "plugin hook execution exceeded timeout for %s", hook.Event)
		}
		return err
	}
	if ctx.Err() != nil {
		return gerror.Wrapf(ctx.Err(), "plugin hook execution exceeded timeout for %s", hook.Event)
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
				logger.Warningf(executeCtx, "plugin async callback hook failed plugin=%s event=%s cost=%s err=%v", pluginID, item.Point, gtime.Now().Sub(startedAt), err)
				return
			}
			logger.Warningf(executeCtx, "plugin callback hook failed plugin=%s event=%s cost=%s err=%v", pluginID, item.Point, gtime.Now().Sub(startedAt), err)
			return
		}
		if async {
			logger.Infof(executeCtx, "plugin async callback hook succeeded plugin=%s event=%s cost=%s", pluginID, item.Point, gtime.Now().Sub(startedAt))
			return
		}
		logger.Infof(executeCtx, "plugin callback hook succeeded plugin=%s event=%s cost=%s", pluginID, item.Point, gtime.Now().Sub(startedAt))
	}

	values := pluginhost.CloneHookPayloadValues(payload)
	if item.Mode == pluginhost.CallbackExecutionModeAsync {
		// Clone payload values before spawning a goroutine so plugin callbacks never
		// race on the mutable request-scoped map owned by the caller.
		go execute(context.WithoutCancel(ctx), values, true)
		return
	}
	execute(ctx, values, false)
}

// shouldDispatchHookToPlugin determines whether the hook event should be delivered to the target plugin.
func (s *Service) shouldDispatchHookToPlugin(
	ctx context.Context,
	runtime *pluginFilterRuntime,
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
		if runtime != nil {
			return runtime.isEnabled(pluginID)
		}
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
	return s.DispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointAuthLoginSucceeded,
		pluginhost.BuildAuthHookPayloadValues(pluginhost.AuthHookPayloadInput{
			UserName:   input.UserName,
			Status:     input.Status,
			IP:         input.Ip,
			ClientType: input.ClientType,
			Browser:    input.Browser,
			OS:         input.Os,
			Message:    input.Message,
		}),
	)
}

// HandleAuthLoginFailed handles login failed hooks declared by source plugins.
func (s *Service) HandleAuthLoginFailed(ctx context.Context, input AuthLoginSucceededInput) error {
	if input.ClientType == "" {
		input.ClientType = "web"
	}
	if input.Message == "" {
		input.Message = "登录失败"
	}
	return s.DispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointAuthLoginFailed,
		pluginhost.BuildAuthHookPayloadValues(pluginhost.AuthHookPayloadInput{
			UserName:   input.UserName,
			Status:     input.Status,
			IP:         input.Ip,
			ClientType: input.ClientType,
			Browser:    input.Browser,
			OS:         input.Os,
			Message:    input.Message,
		}),
	)
}

// HandleAuthLogoutSucceeded handles logout succeeded hooks declared by source plugins.
func (s *Service) HandleAuthLogoutSucceeded(ctx context.Context, input AuthLoginSucceededInput) error {
	if input.ClientType == "" {
		input.ClientType = "web"
	}
	if input.Message == "" {
		input.Message = "登出成功"
	}
	return s.DispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointAuthLogoutSucceeded,
		pluginhost.BuildAuthHookPayloadValues(pluginhost.AuthHookPayloadInput{
			UserName:   input.UserName,
			Status:     input.Status,
			IP:         input.Ip,
			ClientType: input.ClientType,
			Browser:    input.Browser,
			OS:         input.Os,
			Message:    input.Message,
		}),
	)
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
	if spec == nil {
		return gerror.Newf("插件Hook不能为空: %s", filePath)
	}
	if spec.Event == "" {
		return gerror.Newf("插件Hook缺少event: %s", filePath)
	}
	if !pluginhost.IsHookExtensionPoint(spec.Event) {
		return gerror.Newf("插件Hook插槽未发布: %s", filePath)
	}
	if spec.Action == "" {
		spec.Action = pluginhost.HookActionInsert
	}
	if !pluginhost.IsSupportedHookAction(spec.Action) {
		return gerror.Newf("插件Hook动作不受宿主支持: %s", filePath)
	}
	if spec.Mode == "" {
		spec.Mode = pluginhost.DefaultCallbackExecutionMode(spec.Event)
	}
	if !pluginhost.IsExtensionPointExecutionModeSupported(spec.Event, spec.Mode) {
		return gerror.Newf("插件Hook执行模式不受当前插槽支持: %s", filePath)
	}
	if spec.TimeoutMs < 0 {
		return gerror.Newf("插件Hook timeoutMs 不能小于0: %s", filePath)
	}
	switch spec.Action {
	case pluginhost.HookActionInsert:
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
	case pluginhost.HookActionSleep:
		if spec.SleepMs <= 0 {
			return gerror.Newf("插件Hook sleep 动作要求 sleepMs > 0: %s", filePath)
		}
	case pluginhost.HookActionError:
		if strings.TrimSpace(spec.ErrorMessage) == "" {
			return gerror.Newf("插件Hook error 动作要求 errorMessage 非空: %s", filePath)
		}
	}
	return nil
}

// validatePluginResourceSpec validates one plugin-owned backend resource declaration.
func (s *Service) validatePluginResourceSpec(pluginID string, spec *pluginResourceSpec, filePath string) error {
	if spec == nil {
		return gerror.Newf("插件资源不能为空: %s", filePath)
	}
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
	if spec.DataScope != nil {
		if spec.DataScope.UserColumn != "" {
			if err := s.validatePluginIdentifier(spec.DataScope.UserColumn); err != nil {
				return gerror.Wrapf(err, "插件%s资源数据权限 userColumn 非法: %s", pluginID, filePath)
			}
		}
		if spec.DataScope.DeptColumn != "" {
			if err := s.validatePluginIdentifier(spec.DataScope.DeptColumn); err != nil {
				return gerror.Wrapf(err, "插件%s资源数据权限 deptColumn 非法: %s", pluginID, filePath)
			}
		}
		if spec.DataScope.UserColumn == "" && spec.DataScope.DeptColumn == "" {
			return gerror.Newf("插件资源 dataScope 至少需要声明 userColumn 或 deptColumn: %s", filePath)
		}
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

func (s *Service) executePluginSleepHook(ctx context.Context, hook *pluginHookSpec) error {
	if hook == nil || hook.SleepMs <= 0 {
		return nil
	}

	timer := time.NewTimer(time.Duration(hook.SleepMs) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (s *Service) executePluginErrorHook(hook *pluginHookSpec) error {
	if hook == nil {
		return nil
	}
	return gerror.New(strings.TrimSpace(hook.ErrorMessage))
}

// resolvePluginHookValue resolves one hook field expression.
func (s *Service) resolvePluginHookValue(expr string, payload map[string]interface{}) (interface{}, error) {
	if expr == "now" {
		return gtime.Now(), nil
	}
	if strings.HasPrefix(expr, pluginHookEventFieldExprPrefix) {
		fieldName := strings.TrimPrefix(expr, pluginHookEventFieldExprPrefix)
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
