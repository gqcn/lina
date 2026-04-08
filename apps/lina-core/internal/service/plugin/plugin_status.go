package plugin

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginhost"
)

// syncPluginManifest creates or updates a source plugin registry row.
func (s *Service) syncPluginManifest(ctx context.Context, manifest *pluginManifest) (*entity.SysPlugin, error) {
	installedState := pluginInstalledNo
	if manifest.Type == "source" {
		installedState = pluginInstalledYes
	}

	existing, err := s.getPluginRegistry(ctx, manifest.ID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		data := do.SysPlugin{
			PluginId:     manifest.ID,
			Name:         manifest.Name,
			Version:      manifest.Version,
			Type:         manifest.Type,
			Installed:    installedState,
			Status:       pluginStatusDisabled,
			ManifestPath: manifest.ManifestPath,
			Remark:       manifest.Description,
		}
		if manifest.Type == "source" {
			data.Status = pluginStatusEnabled
			data.InstalledAt = gtime.Now()
			data.EnabledAt = gtime.Now()
		}

		_, err = dao.SysPlugin.Ctx(ctx).Data(data).Insert()
		if err != nil {
			return nil, err
		}
		return s.getPluginRegistry(ctx, manifest.ID)
	}

	data := do.SysPlugin{
		Name:         manifest.Name,
		Version:      manifest.Version,
		Type:         manifest.Type,
		ManifestPath: manifest.ManifestPath,
		Remark:       manifest.Description,
	}
	if manifest.Type == "source" {
		data.Installed = installedState
		if existing.InstalledAt == nil {
			data.InstalledAt = gtime.Now()
		}
		if shouldAutoEnableSourcePlugin(existing) {
			data.Status = pluginStatusEnabled
			data.EnabledAt = gtime.Now()
		}
	}

	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: manifest.ID}).
		Data(data).
		Update()
	if err != nil {
		return nil, err
	}

	return s.getPluginRegistry(ctx, manifest.ID)
}

func shouldAutoEnableSourcePlugin(plugin *entity.SysPlugin) bool {
	if plugin == nil {
		return false
	}
	if plugin.Status == pluginStatusEnabled {
		return false
	}
	return plugin.EnabledAt == nil && plugin.DisabledAt == nil
}

// setPluginStatus updates plugin enabled status in sys_plugin.
func (s *Service) setPluginStatus(ctx context.Context, pluginID string, enabled int) error {
	data := do.SysPlugin{
		Status: enabled,
	}
	if enabled == pluginStatusEnabled {
		data.EnabledAt = gtime.Now()
	} else {
		data.DisabledAt = gtime.Now()
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(data).
		Update()
	if err != nil {
		return err
	}

	eventName := pluginhost.ExtensionPointPluginDisabled
	if enabled == pluginStatusEnabled {
		eventName = pluginhost.ExtensionPointPluginEnabled
	}
	return s.DispatchHookEvent(ctx, eventName, map[string]interface{}{
		"pluginId": pluginID,
		"status":   enabled,
	})
}

// getPluginRegistry queries plugin registry by plugin ID.
func (s *Service) getPluginRegistry(ctx context.Context, pluginID string) (*entity.SysPlugin, error) {
	var plugin *entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Scan(&plugin)
	return plugin, err
}

// buildPluginStatusKey builds display key for plugin status.
func (s *Service) buildPluginStatusKey(pluginID string) string {
	return "sys_plugin.status:" + pluginID
}
