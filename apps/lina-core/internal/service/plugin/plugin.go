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
	List  []*PluginItem // plugin list
	Total int           // total count
}

// PluginItem represents plugin metadata with status.
type PluginItem struct {
	Id          string // plugin id
	Name        string // plugin name
	Version     string // plugin version
	Type        string // plugin type
	Entry       string // plugin entry
	Description string // plugin description
	Installed   int    // installed status: 1=installed, 0=not installed
	InstalledAt string // install time or source integration time
	Enabled     int    // enabled status: 1=enabled, 0=disabled
	StatusKey   string // plugin status config key
	UpdatedAt   string // registry updated time
}

// ListInput defines input for plugin list query.
type ListInput struct {
	ID        string // plugin id filter
	Name      string // plugin name filter
	Type      string // plugin type filter: source/runtime/package/wasm
	Status    *int   // plugin status filter
	Installed *int   // install status filter
}

// AuthLoginSucceededInput defines input for login succeeded hook.
type AuthLoginSucceededInput struct {
	UserName   string // login user name
	Status     int    // login status
	Ip         string // login ip
	ClientType string // client type
	Browser    string // browser information
	Os         string // operating system information
	Message    string // audit message
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
	actual = strings.TrimSpace(strings.ToLower(actual))
	expected = strings.TrimSpace(strings.ToLower(expected))
	if expected == "" {
		return true
	}
	if expected == "runtime" {
		return actual == "package" || actual == "wasm"
	}
	return actual == expected
}

// UpdateStatus updates plugin status, where status is 1=enabled and 0=disabled.
func (s *Service) UpdateStatus(ctx context.Context, pluginID string, status int) error {
	if status != 0 && status != 1 {
		return gerror.New("插件状态仅支持0或1")
	}
	if err := s.checkPluginExists(pluginID); err != nil {
		return err
	}
	if err := s.SyncSourcePlugins(ctx); err != nil {
		return err
	}
	if !s.IsInstalled(ctx, pluginID) {
		return gerror.New("插件未安装")
	}
	return s.setPluginStatus(ctx, pluginID, status)
}

// IsInstalled returns whether plugin is installed.
func (s *Service) IsInstalled(ctx context.Context, pluginID string) bool {
	plugin, err := s.getPluginRegistry(ctx, pluginID)
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
	return plugin.Installed == pluginInstalledYes && plugin.Status == pluginStatusEnabled
}

// FilterMenus filters disabled plugin menus by menu_key prefix `plugin:<plugin-id>`.
func (s *Service) FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	if len(menus) == 0 {
		return menus
	}

	filtered := make([]*entity.SysMenu, 0, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		pluginID := s.parsePluginIDFromMenu(menu)
		if pluginID != "" && !s.IsEnabled(ctx, pluginID) {
			continue
		}
		if s.shouldKeepMenu(ctx, menu) {
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

	items := make([]*PluginItem, 0, len(manifests))
	for _, manifest := range manifests {
		registry, err := s.syncPluginManifest(ctx, manifest)
		if err != nil {
			return nil, err
		}
		installedAt := ""
		if registry.InstalledAt != nil {
			installedAt = registry.InstalledAt.String()
		}
		updatedAt := ""
		if registry.UpdatedAt != nil {
			updatedAt = registry.UpdatedAt.String()
		}

		items = append(items, &PluginItem{
			Id:          manifest.ID,
			Name:        manifest.Name,
			Version:     manifest.Version,
			Type:        manifest.Type,
			Entry:       manifest.Entry,
			Description: manifest.Description,
			Installed:   registry.Installed,
			InstalledAt: installedAt,
			Enabled:     registry.Status,
			StatusKey:   s.buildPluginStatusKey(manifest.ID),
			UpdatedAt:   updatedAt,
		})
	}

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
