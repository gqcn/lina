// This file builds governance summary projections from release, migration,
// resource-reference, and node-state records for the plugin management UI.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// pluginGovernanceSnapshot aggregates the review-oriented governance data shown in the plugin management UI.
type pluginGovernanceSnapshot struct {
	ReleaseVersion string
	LifecycleState string
	NodeState      string
	ResourceCount  int
	MigrationState string
}

// buildPluginGovernanceSnapshot loads the current governance projection for one plugin version.
func (s *Service) buildPluginGovernanceSnapshot(
	ctx context.Context,
	pluginID string,
	version string,
	pluginType string,
	installed int,
	enabled int,
) (*pluginGovernanceSnapshot, error) {
	snapshot := &pluginGovernanceSnapshot{
		ReleaseVersion: version,
		LifecycleState: derivePluginLifecycleState(pluginType, installed, enabled),
		NodeState:      derivePluginNodeState(installed, enabled),
		MigrationState: pluginMigrationStateNone.String(),
	}

	release, err := s.getPluginRelease(ctx, pluginID, version)
	if err != nil {
		return nil, err
	}
	if release == nil {
		release, err = s.getPluginActiveRelease(ctx, pluginID)
		if err != nil {
			return nil, err
		}
	}
	if release != nil && strings.TrimSpace(release.ReleaseVersion) != "" {
		snapshot.ReleaseVersion = release.ReleaseVersion
	}

	if release != nil {
		// Resource references already store abstract review records, so a simple row
		// count is enough for the governance summary shown in the management UI.
		resourceCount, countErr := dao.SysPluginResourceRef.Ctx(ctx).
			Where(do.SysPluginResourceRef{
				PluginId:  pluginID,
				ReleaseId: release.Id,
			}).
			Count()
		if countErr != nil {
			return nil, countErr
		}
		snapshot.ResourceCount = resourceCount
	}

	nodeState, err := s.getPluginNodeState(ctx, pluginID, s.currentNodeID())
	if err != nil {
		return nil, err
	}
	if nodeState != nil && strings.TrimSpace(nodeState.CurrentState) != "" {
		snapshot.NodeState = nodeState.CurrentState
	}

	if release != nil {
		latestMigration, migrationErr := s.getLatestPluginMigration(ctx, pluginID, release.Id)
		if migrationErr != nil {
			return nil, migrationErr
		}
		if latestMigration != nil {
			snapshot.MigrationState = derivePluginMigrationState(latestMigration)
		}
	}

	return snapshot, nil
}

// getPluginActiveRelease returns the currently active release row for one plugin.
func (s *Service) getPluginActiveRelease(ctx context.Context, pluginID string) (*entity.SysPluginRelease, error) {
	var release *entity.SysPluginRelease
	err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{
			PluginId: pluginID,
			Status:   pluginReleaseStatusActive.String(),
		}).
		OrderDesc(dao.SysPluginRelease.Columns().Id).
		Scan(&release)
	return release, err
}

// getLatestPluginMigration returns the newest migration record of one plugin release.
func (s *Service) getLatestPluginMigration(ctx context.Context, pluginID string, releaseID int) (*entity.SysPluginMigration, error) {
	if releaseID <= 0 {
		return nil, nil
	}

	var migration *entity.SysPluginMigration
	err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:  pluginID,
			ReleaseId: releaseID,
		}).
		OrderDesc(dao.SysPluginMigration.Columns().Id).
		Scan(&migration)
	return migration, err
}

// derivePluginMigrationState converts the latest migration row into the review-friendly state key.
func derivePluginMigrationState(migration *entity.SysPluginMigration) string {
	if migration == nil {
		return pluginMigrationStateNone.String()
	}
	if migration.Status == pluginMigrationExecutionStatusSucceeded.String() {
		return pluginMigrationStateSucceeded.String()
	}
	return pluginMigrationStateFailed.String()
}
