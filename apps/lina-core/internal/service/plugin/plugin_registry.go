package plugin

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

func (s *Service) getPluginRegistry(ctx context.Context, pluginID string) (*entity.SysPlugin, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, nil
	}

	var plugin *entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: normalizedPluginID}).
		Scan(&plugin)
	return plugin, err
}
