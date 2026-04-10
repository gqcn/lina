package plugin

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

const (
	pluginRegistryQueryCacheTTL        = time.Second
	pluginRegistryQueryCacheNamePrefix = "sys_plugin:plugin_id:"
)

func buildPluginRegistryQueryCacheName(pluginID string) string {
	return pluginRegistryQueryCacheNamePrefix + strings.TrimSpace(pluginID)
}

func withPluginRegistryQueryCache(model *gdb.Model, pluginID string, duration time.Duration) *gdb.Model {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if model == nil || normalizedPluginID == "" {
		return model
	}
	option := gdb.CacheOption{
		Duration: duration,
		Name:     buildPluginRegistryQueryCacheName(normalizedPluginID),
	}
	if duration >= 0 {
		option.Force = true
	}
	return model.Cache(option)
}

func (s *Service) getPluginRegistry(ctx context.Context, pluginID string) (*entity.SysPlugin, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, nil
	}

	var plugin *entity.SysPlugin
	err := withPluginRegistryQueryCache(
		dao.SysPlugin.Ctx(ctx),
		normalizedPluginID,
		pluginRegistryQueryCacheTTL,
	).
		Where(do.SysPlugin{PluginId: normalizedPluginID}).
		Scan(&plugin)
	return plugin, err
}
