// This file persists current-node lifecycle projections so the host can track
// the observed plugin state of each release on the local node.

package plugin

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// syncPluginNodeState updates the current node projection of one plugin lifecycle state.
func (s *Service) syncPluginNodeState(
	ctx context.Context,
	pluginID string,
	version string,
	installed int,
	enabled int,
	message string,
) error {
	nodeKey := s.getCurrentNodeName()
	state := derivePluginNodeState(installed, enabled)
	release, err := s.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		return err
	}

	existing, err := s.getPluginNodeState(ctx, pluginID, nodeKey)
	if err != nil {
		return err
	}

	generation := int64(1)
	if existing != nil && existing.Generation > 0 {
		generation = existing.Generation
	}
	// Use the release identifier as the lower bound of the generation so the
	// persisted node projection can evolve into a multi-generation model later.
	if release != nil && generation < int64(release.Id) {
		generation = int64(release.Id)
	}

	data := do.SysPluginNodeState{
		PluginId:        pluginID,
		ReleaseId:       0,
		NodeKey:         nodeKey,
		DesiredState:    state,
		CurrentState:    state,
		Generation:      generation,
		LastHeartbeatAt: gtime.Now(),
		ErrorMessage:    message,
	}
	if release != nil {
		data.ReleaseId = release.Id
	}

	if existing == nil {
		_, err = dao.SysPluginNodeState.Ctx(ctx).Data(data).Insert()
		return err
	}

	_, err = dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{Id: existing.Id}).
		Data(data).
		Update()
	return err
}

// getPluginNodeState returns the latest node projection row for one plugin/node pair.
func (s *Service) getPluginNodeState(ctx context.Context, pluginID string, nodeKey string) (*entity.SysPluginNodeState, error) {
	var state *entity.SysPluginNodeState
	err := dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{
			PluginId: pluginID,
			NodeKey:  nodeKey,
		}).
		Scan(&state)
	return state, err
}

// getCurrentNodeName resolves the current host node key used by plugin node-state projections.
func (s *Service) getCurrentNodeName() string {
	hostName, err := os.Hostname()
	if err != nil || hostName == "" {
		return "local-node"
	}
	return hostName
}
