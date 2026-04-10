// Package plugin implements plugin manifest discovery, lifecycle orchestration,
// governance metadata synchronization, and host integration for Lina plugins.
package plugin

import (
	"context"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/model/entity"
)

const (
	pluginMenuKeyPrefix    = "plugin:" // plugin menu key prefix in sys_menu.menu_key
	pluginMenuRemarkPrefix = "plugin:" // legacy plugin menu mark prefix in sys_menu.remark
	pluginStatusDisabled   = 0         // disabled plugin status
	pluginStatusEnabled    = 1         // enabled plugin status
	pluginInstalledNo      = 0         // plugin is not installed
	pluginInstalledYes     = 1         // plugin is installed
)

var (
	primaryNodeCheckerMu sync.RWMutex
	primaryNodeChecker   func() bool
)

// Service provides plugin management operations.
type Service struct{}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{}
}

// SetPrimaryNodeChecker registers the host callback used by plugin cron integrations to identify the primary node.
func SetPrimaryNodeChecker(checker func() bool) {
	primaryNodeCheckerMu.Lock()
	defer primaryNodeCheckerMu.Unlock()

	primaryNodeChecker = checker
}

func getPrimaryNodeChecker() func() bool {
	primaryNodeCheckerMu.RLock()
	defer primaryNodeCheckerMu.RUnlock()

	return primaryNodeChecker
}

// ListOutput defines output for plugin list query.
type ListOutput struct {
	List  []*PluginItem // List contains the filtered plugin list.
	Total int           // Total is the number of returned plugins.
}

// PluginItem represents plugin metadata with status.
type PluginItem struct {
	Id          string // Id is the stable plugin identifier.
	Name        string // Name is the display name shown in governance screens.
	Version     string // Version is the version declared by plugin.yaml.
	Type        string // Type is the normalized top-level plugin type.
	Description string // Description is the plugin summary declared by the manifest.
	Installed   int    // Installed reports whether the plugin is installed or integrated.
	InstalledAt string // InstalledAt is the install or source-integration timestamp.
	Enabled     int    // Enabled reports whether the plugin is currently enabled.
	StatusKey   string // StatusKey is the host config key used to persist plugin status.
	UpdatedAt   string // UpdatedAt is the last registry update time.
}

// ListInput defines input for plugin list query.
type ListInput struct {
	ID        string // ID filters by plugin identifier.
	Name      string // Name filters by plugin display name.
	Type      string // Type filters by normalized plugin type.
	Status    *int   // Status filters by enabled flag.
	Installed *int   // Installed filters by installed flag.
}

// AuthLoginSucceededInput defines input for login succeeded hook.
type AuthLoginSucceededInput struct {
	UserName   string // UserName is the authenticated username.
	Status     int    // Status is the login status code.
	Ip         string // Ip is the client IP address.
	ClientType string // ClientType identifies the login client type.
	Browser    string // Browser is the detected browser description.
	Os         string // Os is the detected operating-system description.
	Message    string // Message is the audit message delivered to plugins.
}

// SyncSourcePlugins scans source plugin manifests and synchronizes default status.
func (s *Service) SyncSourcePlugins(ctx context.Context) error {
	_, err := s.SyncAndList(ctx)
	return err
}

// List returns plugin list after synchronization.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	out, err := s.SyncAndList(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]*PluginItem, 0, len(out.List))
	for _, item := range out.List {
		// Apply in-memory filtering after synchronization so source plugin discovery
		// remains the single source of truth for the returned list.
		if in.ID != "" && !strings.Contains(item.Id, in.ID) {
			continue
		}
		if in.Name != "" && !strings.Contains(item.Name, in.Name) {
			continue
		}
		if in.Type != "" && !matchPluginType(item.Type, in.Type) {
			continue
		}
		if in.Status != nil && item.Enabled != *in.Status {
			continue
		}
		if in.Installed != nil && item.Installed != *in.Installed {
			continue
		}
		filtered = append(filtered, item)
	}
	return &ListOutput{List: filtered, Total: len(filtered)}, nil
}

func matchPluginType(actual string, expected string) bool {
	actualType := normalizePluginType(actual)
	expectedType := normalizePluginType(expected)
	if expectedType == "" {
		return true
	}
	return actualType == expectedType
}

// UpdateStatus updates plugin status, where status is 1=enabled and 0=disabled.
func (s *Service) UpdateStatus(ctx context.Context, pluginID string, status int) error {
	if status != 0 && status != 1 {
		return gerror.New("插件状态仅支持0或1")
	}
	manifest, err := s.getPluginManifestByID(pluginID)
	if err != nil {
		return err
	}
	if status == pluginStatusEnabled && normalizePluginType(manifest.Type) == pluginTypeDynamic {
		if err = s.ensureRuntimePluginArtifactAvailable(manifest, "启用"); err != nil {
			return err
		}
	}
	if err := s.SyncSourcePlugins(ctx); err != nil {
		return err
	}
	if !s.IsInstalled(ctx, pluginID) {
		return gerror.New("插件未安装")
	}
	if status == pluginStatusEnabled && normalizePluginType(manifest.Type) == pluginTypeDynamic {
		if err = s.ValidateRuntimeFrontendMenuBindings(ctx, manifest); err != nil {
			return err
		}
		if _, err = s.ensureRuntimeFrontendBundle(ctx, manifest); err != nil {
			return err
		}
	}
	if err = s.setPluginStatus(ctx, pluginID, status); err != nil {
		return err
	}
	if status == pluginStatusDisabled && normalizePluginType(manifest.Type) == pluginTypeDynamic {
		s.invalidateRuntimeFrontendBundle(ctx, pluginID, "plugin_disabled")
	}
	return nil
}

// IsInstalled returns whether plugin is installed.
func (s *Service) IsInstalled(ctx context.Context, pluginID string) bool {
	plugin, err := s.getPluginRegistry(ctx, pluginID)
	if err != nil || plugin == nil {
		return false
	}
	plugin, err = s.reconcileRuntimeRegistryArtifactState(ctx, plugin)
	if err != nil || plugin == nil {
		return false
	}
	return plugin.Installed == pluginInstalledYes
}

// IsEnabled returns whether plugin is enabled.
func (s *Service) IsEnabled(ctx context.Context, pluginID string) bool {
	plugin, err := s.getPluginRegistry(ctx, pluginID)
	if err != nil || plugin == nil {
		return false
	}
	plugin, err = s.reconcileRuntimeRegistryArtifactState(ctx, plugin)
	if err != nil || plugin == nil {
		return false
	}
	return plugin.Installed == pluginInstalledYes && plugin.Status == pluginStatusEnabled
}

// FilterMenus filters disabled plugin menus by menu_key prefix `plugin:<plugin-id>`.
func (s *Service) FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	if len(menus) == 0 {
		return menus
	}

	runtime, err := s.buildFilterRuntime(ctx)
	if err != nil {
		return s.filterMenusSlow(ctx, menus)
	}
	return s.filterMenusWithRuntime(ctx, menus, runtime)
}

func (s *Service) filterMenusWithRuntime(
	ctx context.Context,
	menus []*entity.SysMenu,
	runtime *pluginFilterRuntime,
) []*entity.SysMenu {
	filtered := make([]*entity.SysMenu, 0, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		pluginID := s.parsePluginIDFromMenu(menu)
		if pluginID != "" && !runtime.isEnabled(pluginID) {
			continue
		}
		if s.shouldKeepMenuWithRuntime(ctx, menu, runtime) {
			filtered = append(filtered, menu)
		}
	}
	return filtered
}

func (s *Service) filterMenusSlow(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	filtered := make([]*entity.SysMenu, 0, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		pluginID := s.parsePluginIDFromMenu(menu)
		if pluginID != "" && !s.IsEnabled(ctx, pluginID) {
			continue
		}
		if s.shouldKeepMenuSlow(ctx, menu) {
			filtered = append(filtered, menu)
		}
	}
	return filtered
}

// SyncAndList scans plugin manifests and synchronizes plugin registry rows.
func (s *Service) SyncAndList(ctx context.Context) (*ListOutput, error) {
	manifests, err := s.scanPluginManifests()
	if err != nil {
		return nil, err
	}

	manifestMap := make(map[string]*pluginManifest, len(manifests))
	items := make([]*PluginItem, 0, len(manifests))
	for _, manifest := range manifests {
		manifestMap[manifest.ID] = manifest
		registry, err := s.syncPluginManifest(ctx, manifest)
		if err != nil {
			return nil, err
		}
		items = append(items, s.buildPluginItem(manifest, registry))
	}

	runtimeRegistries, err := s.listRuntimeRegistries(ctx)
	if err != nil {
		return nil, err
	}
	for _, registry := range runtimeRegistries {
		if registry == nil || manifestMap[registry.PluginId] != nil {
			continue
		}
		registry, err = s.reconcileRuntimeRegistryArtifactState(ctx, registry)
		if err != nil {
			return nil, err
		}
		items = append(items, s.buildPluginItem(nil, registry))
	}
	sortPluginItems(items)
	return &ListOutput{List: items, Total: len(items)}, nil
}

// Enable enables the specified plugin.
func (s *Service) Enable(ctx context.Context, pluginID string) error {
	return s.UpdateStatus(ctx, pluginID, 1)
}

// Disable disables the specified plugin.
func (s *Service) Disable(ctx context.Context, pluginID string) error {
	return s.UpdateStatus(ctx, pluginID, 0)
}

// checkPluginExists validates plugin manifest existence by plugin ID.
func (s *Service) checkPluginExists(pluginID string) error {
	if pluginID == "" {
		return gerror.New("插件ID不能为空")
	}

	_, err := s.getPluginManifestByID(pluginID)
	return err
}

// parsePluginIDFromMenu parses plugin id from menu metadata.
func (s *Service) parsePluginIDFromMenu(menu *entity.SysMenu) string {
	if pluginID := s.parsePluginIDFromTaggedValue(menu.MenuKey, pluginMenuKeyPrefix); pluginID != "" {
		return pluginID
	}
	return s.parsePluginIDFromTaggedValue(menu.Remark, pluginMenuRemarkPrefix)
}

// parsePluginIDFromTaggedValue parses plugin id from a tagged value.
func (s *Service) parsePluginIDFromTaggedValue(value string, prefix string) string {
	taggedValue := strings.TrimSpace(value)
	if !strings.HasPrefix(taggedValue, prefix) {
		return ""
	}

	suffix := strings.TrimPrefix(taggedValue, prefix)
	end := len(suffix)
	for _, separator := range []string{":", " "} {
		if index := strings.Index(suffix, separator); index >= 0 && index < end {
			end = index
		}
	}
	return strings.TrimSpace(suffix[:end])
}
