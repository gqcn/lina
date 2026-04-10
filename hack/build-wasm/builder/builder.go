// Package builder implements a standalone dynamic wasm packer for plugin
// source trees. It intentionally lives outside lina-core so development-time
// packaging does not depend on the host service module.
package builder

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	pluginTypeDynamic                = "dynamic"
	pluginDynamicKindWasm            = "wasm"
	pluginDynamicSupportedABIVersion = "v1"

	pluginDynamicWasmSectionManifest     = "lina.plugin.manifest"
	pluginDynamicWasmSectionDynamic      = "lina.plugin.dynamic"
	pluginDynamicWasmSectionFrontend     = "lina.plugin.frontend.assets"
	pluginDynamicWasmSectionInstallSQL   = "lina.plugin.install.sql"
	pluginDynamicWasmSectionUninstallSQL = "lina.plugin.uninstall.sql"
	pluginDynamicWasmSectionBackendHooks = "lina.plugin.backend.hooks"
	pluginDynamicWasmSectionBackendRes   = "lina.plugin.backend.resources"
)

var (
	pluginManifestIDPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	pluginManifestSemverPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$`)
	safeIdentifierPattern       = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

// RuntimeBuildOutput contains the generated dynamic artifact bytes and output path.
type RuntimeBuildOutput struct {
	ArtifactPath string
	Content      []byte
}

type pluginManifest struct {
	ID          string      `yaml:"id"`
	Name        string      `yaml:"name"`
	Version     string      `yaml:"version"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Menus       []*menuSpec `yaml:"menus"`
}

type dynamicArtifactManifest struct {
	ID          string      `json:"id" yaml:"id"`
	Name        string      `json:"name" yaml:"name"`
	Version     string      `json:"version" yaml:"version"`
	Type        string      `json:"type" yaml:"type"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Menus       []*menuSpec `json:"menus,omitempty" yaml:"menus,omitempty"`
}

type dynamicArtifactMetadata struct {
	RuntimeKind        string `json:"runtimeKind" yaml:"runtimeKind"`
	ABIVersion         string `json:"abiVersion" yaml:"abiVersion"`
	FrontendAssetCount int    `json:"frontendAssetCount,omitempty" yaml:"frontendAssetCount,omitempty"`
	SQLAssetCount      int    `json:"sqlAssetCount,omitempty" yaml:"sqlAssetCount,omitempty"`
}

type frontendAsset struct {
	Path          string `json:"path" yaml:"path"`
	ContentBase64 string `json:"contentBase64" yaml:"contentBase64"`
	ContentType   string `json:"contentType,omitempty" yaml:"contentType,omitempty"`
}

type sqlAsset struct {
	Key     string `json:"key" yaml:"key"`
	Content string `json:"content" yaml:"content"`
}

type menuSpec struct {
	Key        string                 `json:"key" yaml:"key"`
	ParentKey  string                 `json:"parent_key,omitempty" yaml:"parent_key,omitempty"`
	Name       string                 `json:"name" yaml:"name"`
	Path       string                 `json:"path,omitempty" yaml:"path,omitempty"`
	Component  string                 `json:"component,omitempty" yaml:"component,omitempty"`
	Perms      string                 `json:"perms,omitempty" yaml:"perms,omitempty"`
	Icon       string                 `json:"icon,omitempty" yaml:"icon,omitempty"`
	Type       string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Sort       int                    `json:"sort,omitempty" yaml:"sort,omitempty"`
	Visible    *int                   `json:"visible,omitempty" yaml:"visible,omitempty"`
	Status     *int                   `json:"status,omitempty" yaml:"status,omitempty"`
	IsFrame    *int                   `json:"is_frame,omitempty" yaml:"is_frame,omitempty"`
	IsCache    *int                   `json:"is_cache,omitempty" yaml:"is_cache,omitempty"`
	Query      map[string]interface{} `json:"query,omitempty" yaml:"query,omitempty"`
	QueryParam string                 `json:"query_param,omitempty" yaml:"query_param,omitempty"`
	Remark     string                 `json:"remark,omitempty" yaml:"remark,omitempty"`
}

type hookExtensionPoint string
type hookAction string
type callbackExecutionMode string
type resourceSpecType string
type resourceFilterOperator string
type resourceOrderDirection string

const (
	callbackExecutionModeBlocking callbackExecutionMode = "blocking"
	callbackExecutionModeAsync    callbackExecutionMode = "async"

	hookActionInsert hookAction = "insert"
	hookActionSleep  hookAction = "sleep"
	hookActionError  hookAction = "error"

	resourceSpecTypeTableList resourceSpecType = "table-list"

	resourceFilterOperatorEQ      resourceFilterOperator = "eq"
	resourceFilterOperatorLike    resourceFilterOperator = "like"
	resourceFilterOperatorGTEDate resourceFilterOperator = "gte-date"
	resourceFilterOperatorLTEDate resourceFilterOperator = "lte-date"

	resourceOrderDirectionASC  resourceOrderDirection = "asc"
	resourceOrderDirectionDESC resourceOrderDirection = "desc"

	extensionPointAuthLoginSucceeded  hookExtensionPoint = "auth.login.succeeded"
	extensionPointAuthLoginFailed     hookExtensionPoint = "auth.login.failed"
	extensionPointAuthLogoutSucceeded hookExtensionPoint = "auth.logout.succeeded"
	extensionPointPluginInstalled     hookExtensionPoint = "plugin.installed"
	extensionPointPluginEnabled       hookExtensionPoint = "plugin.enabled"
	extensionPointPluginDisabled      hookExtensionPoint = "plugin.disabled"
	extensionPointPluginUninstalled   hookExtensionPoint = "plugin.uninstalled"
	extensionPointSystemStarted       hookExtensionPoint = "system.started"
)

type hookSpec struct {
	Event        hookExtensionPoint    `json:"event" yaml:"event"`
	Action       hookAction            `json:"action" yaml:"action"`
	Mode         callbackExecutionMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	Table        string                `json:"table,omitempty" yaml:"table,omitempty"`
	Fields       map[string]string     `json:"fields,omitempty" yaml:"fields,omitempty"`
	TimeoutMs    int                   `json:"timeoutMs,omitempty" yaml:"timeoutMs,omitempty"`
	SleepMs      int                   `json:"sleepMs,omitempty" yaml:"sleepMs,omitempty"`
	ErrorMessage string                `json:"errorMessage,omitempty" yaml:"errorMessage,omitempty"`
}

type resourceSpec struct {
	Key       string                 `json:"key" yaml:"key"`
	Type      string                 `json:"type" yaml:"type"`
	Table     string                 `json:"table" yaml:"table"`
	Fields    []*resourceField       `json:"fields" yaml:"fields"`
	Filters   []*resourceQuery       `json:"filters" yaml:"filters"`
	OrderBy   resourceOrderBySpec    `json:"orderBy" yaml:"orderBy"`
	DataScope *resourceDataScopeSpec `json:"dataScope,omitempty" yaml:"dataScope,omitempty"`
}

type resourceField struct {
	Name   string `json:"name" yaml:"name"`
	Column string `json:"column" yaml:"column"`
}

type resourceQuery struct {
	Param    string `json:"param" yaml:"param"`
	Column   string `json:"column" yaml:"column"`
	Operator string `json:"operator" yaml:"operator"`
}

type resourceOrderBySpec struct {
	Column    string `json:"column" yaml:"column"`
	Direction string `json:"direction" yaml:"direction"`
}

type resourceDataScopeSpec struct {
	UserColumn string `json:"userColumn,omitempty" yaml:"userColumn,omitempty"`
	DeptColumn string `json:"deptColumn,omitempty" yaml:"deptColumn,omitempty"`
}

var publishedHookPoints = map[hookExtensionPoint]callbackExecutionMode{
	extensionPointAuthLoginSucceeded:  callbackExecutionModeBlocking,
	extensionPointAuthLoginFailed:     callbackExecutionModeBlocking,
	extensionPointAuthLogoutSucceeded: callbackExecutionModeBlocking,
	extensionPointPluginInstalled:     callbackExecutionModeBlocking,
	extensionPointPluginEnabled:       callbackExecutionModeBlocking,
	extensionPointPluginDisabled:      callbackExecutionModeBlocking,
	extensionPointPluginUninstalled:   callbackExecutionModeBlocking,
	extensionPointSystemStarted:       callbackExecutionModeBlocking,
}

var supportedHookModes = map[hookExtensionPoint]map[callbackExecutionMode]struct{}{
	extensionPointAuthLoginSucceeded:  {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointAuthLoginFailed:     {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointAuthLogoutSucceeded: {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointPluginInstalled:     {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointPluginEnabled:       {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointPluginDisabled:      {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointPluginUninstalled:   {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
	extensionPointSystemStarted:       {callbackExecutionModeBlocking: {}, callbackExecutionModeAsync: {}},
}

// BuildRuntimeWasmArtifactFromSource builds one dynamic wasm artifact from a clear-text plugin directory.
func BuildRuntimeWasmArtifactFromSource(pluginDir string) (*RuntimeBuildOutput, error) {
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")

	manifest := &pluginManifest{}
	if err := loadYAMLFile(manifestPath, manifest); err != nil {
		return nil, fmt.Errorf("failed to load dynamic plugin manifest: %w", err)
	}
	if err := validateRuntimeBuildManifest(manifest, manifestPath); err != nil {
		return nil, err
	}

	frontendAssets, err := collectFrontendAssets(pluginDir)
	if err != nil {
		return nil, err
	}
	installSQLAssets, err := collectSQLAssets(pluginDir, false)
	if err != nil {
		return nil, err
	}
	uninstallSQLAssets, err := collectSQLAssets(pluginDir, true)
	if err != nil {
		return nil, err
	}
	hookSpecs, err := collectHookSpecs(pluginDir, manifest.ID)
	if err != nil {
		return nil, err
	}
	resourceSpecs, err := collectResourceSpecs(pluginDir, manifest.ID)
	if err != nil {
		return nil, err
	}

	content, err := buildRuntimeArtifactContent(
		manifest,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		hookSpecs,
		resourceSpecs,
	)
	if err != nil {
		return nil, err
	}

	return &RuntimeBuildOutput{
		ArtifactPath: filepath.Join(pluginDir, buildRuntimeBuildOutputRelativePath(manifest.ID)),
		Content:      content,
	}, nil
}

// WriteRuntimeWasmArtifactFromSource builds and writes one dynamic artifact into
// the requested output directory. When outputDir is empty, temp/<plugin-id>.wasm
// under the plugin source tree is used for backward compatibility.
func WriteRuntimeWasmArtifactFromSource(pluginDir string, outputDir string) (*RuntimeBuildOutput, error) {
	out, err := BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(outputDir) != "" {
		out.ArtifactPath = filepath.Join(filepath.Clean(outputDir), filepath.Base(out.ArtifactPath))
	}
	if err = os.MkdirAll(filepath.Dir(out.ArtifactPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create dynamic artifact directory: %w", err)
	}
	if err = os.WriteFile(out.ArtifactPath, out.Content, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write dynamic artifact: %w", err)
	}
	return out, nil
}

func buildRuntimeArtifactFileName(pluginID string) string {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return "plugin.wasm"
	}
	return normalizedID + ".wasm"
}

func buildRuntimeArtifactRelativePath(pluginID string) string {
	return filepath.Join("runtime", buildRuntimeArtifactFileName(pluginID))
}

func buildRuntimeBuildOutputRelativePath(pluginID string) string {
	return filepath.Join("temp", buildRuntimeArtifactFileName(pluginID))
}

func validateRuntimeBuildManifest(manifest *pluginManifest, manifestPath string) error {
	if manifest == nil {
		return fmt.Errorf("dynamic plugin manifest cannot be nil")
	}
	if strings.TrimSpace(manifest.ID) == "" {
		return fmt.Errorf("dynamic plugin manifest missing id: %s", manifestPath)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("dynamic plugin manifest missing name: %s", manifestPath)
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("dynamic plugin manifest missing version: %s", manifestPath)
	}
	manifest.Type = strings.ToLower(strings.TrimSpace(manifest.Type))
	if manifest.Type != pluginTypeDynamic {
		return fmt.Errorf("dynamic sample manifest type must be dynamic: %s", manifestPath)
	}
	if !pluginManifestIDPattern.MatchString(manifest.ID) {
		return fmt.Errorf("dynamic plugin id must use kebab-case: %s", manifest.ID)
	}
	if err := validateSemanticVersion(manifest.Version); err != nil {
		return fmt.Errorf("dynamic plugin version is invalid: %w", err)
	}
	return nil
}

func collectFrontendAssets(pluginDir string) ([]*frontendAsset, error) {
	frontendDir := filepath.Join(pluginDir, "frontend", "pages")
	info, err := os.Stat(frontendDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*frontendAsset{}, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("runtime frontend pages path is not a directory: %s", frontendDir)
	}

	paths := make([]string, 0)
	if err = filepath.WalkDir(frontendDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		return nil, err
	}

	sort.Strings(paths)
	assets := make([]*frontendAsset, 0, len(paths))
	for _, filePath := range paths {
		relativePath, err := filepath.Rel(frontendDir, filePath)
		if err != nil {
			return nil, err
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		assets = append(assets, &frontendAsset{
			Path:          filepath.ToSlash(relativePath),
			ContentBase64: base64.StdEncoding.EncodeToString(content),
			ContentType:   contentType,
		})
	}
	return assets, nil
}

func collectSQLAssets(pluginDir string, uninstall bool) ([]*sqlAsset, error) {
	searchDir := filepath.Join(pluginDir, "manifest", "sql")
	if uninstall {
		searchDir = filepath.Join(pluginDir, "manifest", "sql", "uninstall")
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*sqlAsset{}, nil
		}
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sort.Strings(fileNames)

	assets := make([]*sqlAsset, 0, len(fileNames))
	for _, name := range fileNames {
		sqlPath := filepath.Join(searchDir, name)
		content, err := os.ReadFile(sqlPath)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &sqlAsset{
			Key:     name,
			Content: strings.TrimSpace(string(content)),
		})
	}
	return assets, nil
}

func collectHookSpecs(pluginDir string, pluginID string) ([]*hookSpec, error) {
	hookDir := filepath.Join(pluginDir, "backend", "hooks")
	entries, err := os.ReadDir(hookDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*hookSpec{}, nil
		}
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sort.Strings(fileNames)

	items := make([]*hookSpec, 0, len(fileNames))
	for _, name := range fileNames {
		filePath := filepath.Join(hookDir, name)
		spec := &hookSpec{}
		if err = loadYAMLFile(filePath, spec); err != nil {
			return nil, err
		}
		if err = validateHookSpec(pluginID, spec, filePath); err != nil {
			return nil, err
		}
		items = append(items, spec)
	}
	return items, nil
}

func collectResourceSpecs(pluginDir string, pluginID string) ([]*resourceSpec, error) {
	resourceDir := filepath.Join(pluginDir, "backend", "resources")
	entries, err := os.ReadDir(resourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*resourceSpec{}, nil
		}
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sort.Strings(fileNames)

	items := make([]*resourceSpec, 0, len(fileNames))
	for _, name := range fileNames {
		filePath := filepath.Join(resourceDir, name)
		spec := &resourceSpec{}
		if err = loadYAMLFile(filePath, spec); err != nil {
			return nil, err
		}
		if err = validateResourceSpec(pluginID, spec, filePath); err != nil {
			return nil, err
		}
		items = append(items, spec)
	}
	return items, nil
}

func loadYAMLFile(filePath string, target interface{}) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if len(content) == 0 {
		return fmt.Errorf("yaml file is empty: %s", filePath)
	}
	if err = yaml.Unmarshal(content, target); err != nil {
		return fmt.Errorf("failed to parse yaml file %s: %w", filePath, err)
	}
	return nil
}

func validateSemanticVersion(value string) error {
	match := pluginManifestSemverPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) < 4 {
		return fmt.Errorf("version must use semver format: %s", value)
	}

	for _, raw := range match[1:4] {
		if _, err := strconv.Atoi(raw); err != nil {
			return err
		}
	}
	return nil
}

func validateHookSpec(pluginID string, spec *hookSpec, filePath string) error {
	if spec == nil {
		return fmt.Errorf("plugin hook cannot be nil: %s", filePath)
	}
	if strings.TrimSpace(string(spec.Event)) == "" {
		return fmt.Errorf("plugin hook missing event: %s", filePath)
	}
	if !isHookExtensionPoint(spec.Event) {
		return fmt.Errorf("plugin hook event is not published by host: %s", filePath)
	}
	if spec.Action == "" {
		spec.Action = hookActionInsert
	}
	if !isSupportedHookAction(spec.Action) {
		return fmt.Errorf("plugin hook action is not supported: %s", filePath)
	}
	if spec.Mode == "" {
		spec.Mode = defaultCallbackExecutionMode(spec.Event)
	}
	if !isExtensionPointExecutionModeSupported(spec.Event, spec.Mode) {
		return fmt.Errorf("plugin hook execution mode is not supported: %s", filePath)
	}
	if spec.TimeoutMs < 0 {
		return fmt.Errorf("plugin hook timeoutMs cannot be negative: %s", filePath)
	}

	switch spec.Action {
	case hookActionInsert:
		if err := validateIdentifier(spec.Table); err != nil {
			return fmt.Errorf("plugin %s hook table is invalid: %s: %w", pluginID, filePath, err)
		}
		if len(spec.Fields) == 0 {
			return fmt.Errorf("plugin hook missing fields: %s", filePath)
		}
		for column := range spec.Fields {
			if err := validateIdentifier(column); err != nil {
				return fmt.Errorf("plugin %s hook field is invalid: %s: %w", pluginID, filePath, err)
			}
		}
	case hookActionSleep:
		if spec.SleepMs <= 0 {
			return fmt.Errorf("plugin hook sleep action requires sleepMs > 0: %s", filePath)
		}
	case hookActionError:
		if strings.TrimSpace(spec.ErrorMessage) == "" {
			return fmt.Errorf("plugin hook error action requires non-empty errorMessage: %s", filePath)
		}
	}

	return nil
}

func validateResourceSpec(pluginID string, spec *resourceSpec, filePath string) error {
	if spec == nil {
		return fmt.Errorf("plugin resource cannot be nil: %s", filePath)
	}
	if strings.TrimSpace(spec.Key) == "" {
		return fmt.Errorf("plugin resource missing key: %s", filePath)
	}
	if spec.Type == "" {
		spec.Type = string(resourceSpecTypeTableList)
	}
	if normalizeResourceSpecType(spec.Type) != resourceSpecTypeTableList {
		return fmt.Errorf("plugin resource type only supports table-list: %s", filePath)
	}
	if err := validateIdentifier(spec.Table); err != nil {
		return fmt.Errorf("plugin %s resource table is invalid: %s: %w", pluginID, filePath, err)
	}
	if len(spec.Fields) == 0 {
		return fmt.Errorf("plugin resource missing fields: %s", filePath)
	}
	for _, field := range spec.Fields {
		if field == nil {
			return fmt.Errorf("plugin resource field cannot be nil: %s", filePath)
		}
		if err := validateIdentifier(field.Name); err != nil {
			return fmt.Errorf("plugin %s resource field name is invalid: %s: %w", pluginID, filePath, err)
		}
		if err := validateIdentifier(field.Column); err != nil {
			return fmt.Errorf("plugin %s resource column is invalid: %s: %w", pluginID, filePath, err)
		}
	}
	for _, filter := range spec.Filters {
		if filter == nil {
			return fmt.Errorf("plugin resource filter cannot be nil: %s", filePath)
		}
		if strings.TrimSpace(filter.Param) == "" {
			return fmt.Errorf("plugin resource filter missing param: %s", filePath)
		}
		if err := validateIdentifier(filter.Column); err != nil {
			return fmt.Errorf("plugin %s resource filter column is invalid: %s: %w", pluginID, filePath, err)
		}
		if normalizeResourceFilterOperator(filter.Operator) == "" {
			return fmt.Errorf("plugin resource filter operator is not supported: %s", filePath)
		}
	}
	if err := validateIdentifier(spec.OrderBy.Column); err != nil {
		return fmt.Errorf("plugin %s resource orderBy column is invalid: %s: %w", pluginID, filePath, err)
	}
	if spec.OrderBy.Direction == "" {
		spec.OrderBy.Direction = string(resourceOrderDirectionASC)
	}
	if normalizeResourceOrderDirection(spec.OrderBy.Direction) == "" {
		return fmt.Errorf("plugin resource order direction only supports asc/desc: %s", filePath)
	}
	if spec.DataScope != nil {
		if spec.DataScope.UserColumn != "" {
			if err := validateIdentifier(spec.DataScope.UserColumn); err != nil {
				return fmt.Errorf("plugin %s resource dataScope userColumn is invalid: %s: %w", pluginID, filePath, err)
			}
		}
		if spec.DataScope.DeptColumn != "" {
			if err := validateIdentifier(spec.DataScope.DeptColumn); err != nil {
				return fmt.Errorf("plugin %s resource dataScope deptColumn is invalid: %s: %w", pluginID, filePath, err)
			}
		}
		if spec.DataScope.UserColumn == "" && spec.DataScope.DeptColumn == "" {
			return fmt.Errorf("plugin resource dataScope requires userColumn or deptColumn: %s", filePath)
		}
	}
	return nil
}

func validateIdentifier(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if !safeIdentifierPattern.MatchString(value) {
		return fmt.Errorf("identifier is invalid: %s", value)
	}
	return nil
}

func defaultCallbackExecutionMode(point hookExtensionPoint) callbackExecutionMode {
	return publishedHookPoints[point]
}

func isHookExtensionPoint(point hookExtensionPoint) bool {
	_, ok := publishedHookPoints[point]
	return ok
}

func isSupportedHookAction(action hookAction) bool {
	switch action {
	case hookActionInsert, hookActionSleep, hookActionError:
		return true
	default:
		return false
	}
}

func isExtensionPointExecutionModeSupported(point hookExtensionPoint, mode callbackExecutionMode) bool {
	modes, ok := supportedHookModes[point]
	if !ok {
		return false
	}
	_, ok = modes[mode]
	return ok
}

func normalizeResourceSpecType(value string) resourceSpecType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceSpecTypeTableList):
		return resourceSpecTypeTableList
	default:
		return ""
	}
}

func normalizeResourceFilterOperator(value string) resourceFilterOperator {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceFilterOperatorEQ):
		return resourceFilterOperatorEQ
	case string(resourceFilterOperatorLike):
		return resourceFilterOperatorLike
	case string(resourceFilterOperatorGTEDate):
		return resourceFilterOperatorGTEDate
	case string(resourceFilterOperatorLTEDate):
		return resourceFilterOperatorLTEDate
	default:
		return ""
	}
}

func normalizeResourceOrderDirection(value string) resourceOrderDirection {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceOrderDirectionASC):
		return resourceOrderDirectionASC
	case string(resourceOrderDirectionDESC):
		return resourceOrderDirectionDESC
	default:
		return ""
	}
}

func buildRuntimeArtifactContent(
	manifest *pluginManifest,
	frontendAssets []*frontendAsset,
	installSQLAssets []*sqlAsset,
	uninstallSQLAssets []*sqlAsset,
	hookSpecs []*hookSpec,
	resourceSpecs []*resourceSpec,
) ([]byte, error) {
	manifestPayload, err := json.Marshal(&dynamicArtifactManifest{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Type:        pluginTypeDynamic,
		Description: manifest.Description,
		Menus:       manifest.Menus,
	})
	if err != nil {
		return nil, err
	}
	runtimePayload, err := json.Marshal(&dynamicArtifactMetadata{
		RuntimeKind:        pluginDynamicKindWasm,
		ABIVersion:         pluginDynamicSupportedABIVersion,
		FrontendAssetCount: len(frontendAssets),
		SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
	})
	if err != nil {
		return nil, err
	}

	content := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	content = appendWasmCustomSection(content, pluginDynamicWasmSectionManifest, manifestPayload)
	content = appendWasmCustomSection(content, pluginDynamicWasmSectionDynamic, runtimePayload)

	if len(frontendAssets) > 0 {
		payload, err := json.Marshal(frontendAssets)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionFrontend, payload)
	}
	if len(installSQLAssets) > 0 {
		payload, err := json.Marshal(installSQLAssets)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionInstallSQL, payload)
	}
	if len(uninstallSQLAssets) > 0 {
		payload, err := json.Marshal(uninstallSQLAssets)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionUninstallSQL, payload)
	}
	if len(hookSpecs) > 0 {
		payload, err := json.Marshal(hookSpecs)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionBackendHooks, payload)
	}
	if len(resourceSpecs) > 0 {
		payload, err := json.Marshal(resourceSpecs)
		if err != nil {
			return nil, err
		}
		content = appendWasmCustomSection(content, pluginDynamicWasmSectionBackendRes, payload)
	}
	return content, nil
}

func appendWasmCustomSection(content []byte, name string, payload []byte) []byte {
	section := make([]byte, 0, len(name)+len(payload)+8)
	section = appendULEB128(section, uint32(len(name)))
	section = append(section, []byte(name)...)
	section = append(section, payload...)

	content = append(content, 0x00)
	content = appendULEB128(content, uint32(len(section)))
	content = append(content, section...)
	return content
}

func appendULEB128(content []byte, value uint32) []byte {
	current := value
	for {
		part := byte(current & 0x7f)
		current >>= 7
		if current != 0 {
			part |= 0x80
		}
		content = append(content, part)
		if current == 0 {
			return content
		}
	}
}
