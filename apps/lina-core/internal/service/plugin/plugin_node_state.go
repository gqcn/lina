// This file persists current-node lifecycle projections so the host can track
// the observed plugin state of each release on the local node.

package plugin

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

type pluginNodeProjectionInput struct {
	PluginID     string
	ReleaseID    int
	DesiredState string
	CurrentState string
	Generation   int64
	Message      string
}

// syncPluginNodeState updates the current node projection of one plugin lifecycle state.
func (s *Service) syncPluginNodeState(
	ctx context.Context,
	pluginID string,
	version string,
	installed int,
	enabled int,
	message string,
) error {
	registry, err := s.getPluginRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		_, releaseErr := s.getPluginRelease(ctx, pluginID, version)
		if releaseErr != nil {
			return releaseErr
		}
		return s.syncPluginNodeProjection(ctx, pluginNodeProjectionInput{
			PluginID:     pluginID,
			ReleaseID:    0,
			DesiredState: derivePluginNodeState(installed, enabled),
			CurrentState: derivePluginNodeState(installed, enabled),
			Generation:   int64(1),
			Message:      message,
		})
	}

	desiredState := strings.TrimSpace(registry.DesiredState)
	if desiredState == "" {
		desiredState = derivePluginNodeState(installed, enabled)
	}
	currentState := strings.TrimSpace(registry.CurrentState)
	if currentState == "" {
		currentState = derivePluginNodeState(installed, enabled)
	}
	generation := registry.Generation
	if generation <= 0 {
		generation = 1
	}
	return s.syncPluginNodeProjection(ctx, pluginNodeProjectionInput{
		PluginID:     registry.PluginId,
		ReleaseID:    registry.ReleaseId,
		DesiredState: desiredState,
		CurrentState: currentState,
		Generation:   generation,
		Message:      message,
	})
}

func (s *Service) syncPluginNodeProjection(ctx context.Context, in pluginNodeProjectionInput) error {
	pluginID := strings.TrimSpace(in.PluginID)
	nodeKey := s.getCurrentNodeName()
	desiredState := strings.TrimSpace(in.DesiredState)
	if desiredState == "" {
		desiredState = pluginNodeStateUninstalled.String()
	}
	currentState := strings.TrimSpace(in.CurrentState)
	if currentState == "" {
		currentState = desiredState
	}
	generation := in.Generation
	if generation <= 0 {
		generation = 1
	}

	data := do.SysPluginNodeState{
		PluginId:        pluginID,
		ReleaseId:       in.ReleaseID,
		NodeKey:         nodeKey,
		DesiredState:    desiredState,
		CurrentState:    currentState,
		Generation:      generation,
		LastHeartbeatAt: gtime.Now(),
		ErrorMessage:    strings.TrimSpace(in.Message),
	}

	existing, err := s.getPluginNodeState(ctx, pluginID, nodeKey)
	if err != nil {
		return err
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
