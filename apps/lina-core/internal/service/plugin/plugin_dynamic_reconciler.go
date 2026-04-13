// This file implements the leader-aware dynamic-plugin reconciler. Management
// APIs persist the desired host state, while the primary node archives the
// staged artifact, performs migrations and menu switches, advances generation,
// and updates per-node convergence rows.

package plugin

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

const runtimePluginReconcilerInterval = 2 * time.Second

var (
	runtimePluginReconcilerOnce sync.Once
	runtimePluginReconcileMu    sync.Mutex
)

// StartRuntimeReconciler starts the background loop that keeps dynamic-plugin
// desired state, active release, and current-node projection converged.
func (s *Service) StartRuntimeReconciler(ctx context.Context) {
	if !s.isClusterModeEnabled() {
		return
	}
	runtimePluginReconcilerOnce.Do(func() {
		go s.runRuntimePluginReconciler(context.WithoutCancel(ctx))
	})
}

func (s *Service) runRuntimePluginReconciler(ctx context.Context) {
	ticker := time.NewTicker(runtimePluginReconcilerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.ReconcileRuntimePlugins(ctx); err != nil {
				logger.Warningf(ctx, "dynamic plugin reconciler tick failed: %v", err)
			}
		}
	}
}

// ReconcileRuntimePlugins runs one convergence pass. It is safe to call from
// both the background loop and synchronous management flows.
func (s *Service) ReconcileRuntimePlugins(ctx context.Context) error {
	runtimePluginReconcileMu.Lock()
	defer runtimePluginReconcileMu.Unlock()

	registries, err := s.listRuntimeRegistries(ctx)
	if err != nil {
		return err
	}

	isPrimary := s.isPrimaryNode()

	var firstErr error
	for _, registry := range registries {
		if registry == nil {
			continue
		}

		registry, err = s.reconcileRuntimeRegistryArtifactState(ctx, registry)
		if err != nil {
			logger.Warningf(ctx, "reconcile runtime registry artifact state failed plugin=%s err=%v", registry.PluginId, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if registry == nil {
			continue
		}

		if isPrimary {
			if err = s.reconcileDynamicPluginIfNeeded(ctx, registry); err != nil {
				logger.Warningf(ctx, "reconcile dynamic plugin failed plugin=%s err=%v", registry.PluginId, err)
				if firstErr == nil {
					firstErr = err
				}
			}
			registry, _ = s.getPluginRegistry(ctx, registry.PluginId)
		}
		if registry == nil {
			continue
		}
		if err = s.reconcileCurrentNodeProjection(ctx, registry); err != nil {
			logger.Warningf(ctx, "reconcile current node projection failed plugin=%s err=%v", registry.PluginId, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (s *Service) reconcileDynamicPluginRequest(
	ctx context.Context,
	pluginID string,
	desiredState pluginHostStateValue,
) error {
	if err := s.updatePluginRegistryDesiredState(ctx, pluginID, desiredState); err != nil {
		return err
	}
	if !s.isPrimaryNode() {
		return nil
	}
	return s.ReconcileRuntimePlugins(ctx)
}

// reconcileDynamicPluginIfNeeded selects the smallest convergence action for
// the current registry row: install, upgrade, same-version refresh, state
// toggle, or uninstall.
func (s *Service) reconcileDynamicPluginIfNeeded(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil || normalizePluginType(registry.Type) != pluginTypeDynamic {
		return nil
	}

	desiredState := strings.TrimSpace(registry.DesiredState)
	if desiredState == "" {
		desiredState = buildStablePluginHostState(registry)
	}
	stableState := buildStablePluginHostState(registry)
	if desiredState == pluginHostStateUninstalled.String() {
		if registry.Installed != pluginInstalledYes {
			return nil
		}
		return s.applyDynamicUninstall(ctx, registry)
	}

	desiredManifest, err := s.getDesiredPluginManifestByID(registry.PluginId)
	if err != nil {
		return err
	}
	if desiredManifest == nil || normalizePluginType(desiredManifest.Type) != pluginTypeDynamic {
		return gerror.New("动态插件目标清单不存在")
	}

	if registry.Installed != pluginInstalledYes {
		return s.applyDynamicInstall(ctx, registry, desiredManifest, desiredState)
	}
	if strings.TrimSpace(desiredManifest.Version) != strings.TrimSpace(registry.Version) {
		// Version drift means upgrade semantics, including upgrade SQL and release switch.
		return s.applyDynamicUpgrade(ctx, registry, desiredManifest, desiredState)
	}
	if s.shouldRefreshInstalledDynamicRelease(ctx, registry, desiredManifest) {
		// Same semantic version can still require refresh when the staged artifact,
		// archive bytes, or synthesized checksum changed after a rebuild.
		return s.applyDynamicRefresh(ctx, registry, desiredManifest, desiredState)
	}
	if desiredState != stableState {
		return s.applyDynamicStateToggle(ctx, registry, desiredManifest, desiredState)
	}
	return nil
}

func (s *Service) reconcileCurrentNodeProjection(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil || normalizePluginType(registry.Type) != pluginTypeDynamic {
		return nil
	}

	if registry.Installed == pluginInstalledYes && registry.Status == pluginStatusEnabled && registry.ReleaseId > 0 {
		manifest, err := s.loadActiveDynamicPluginManifest(ctx, registry)
		if err != nil {
			return s.syncPluginNodeProjection(ctx, pluginNodeProjectionInput{
				PluginID:     registry.PluginId,
				ReleaseID:    registry.ReleaseId,
				DesiredState: registry.DesiredState,
				CurrentState: pluginNodeStateFailed.String(),
				Generation:   registry.Generation,
				Message:      err.Error(),
			})
		}
		if hasRuntimeFrontendAssets(manifest) {
			if _, err = s.ensureRuntimeFrontendBundle(ctx, manifest); err != nil {
				return s.syncPluginNodeProjection(ctx, pluginNodeProjectionInput{
					PluginID:     registry.PluginId,
					ReleaseID:    registry.ReleaseId,
					DesiredState: registry.DesiredState,
					CurrentState: pluginNodeStateFailed.String(),
					Generation:   registry.Generation,
					Message:      err.Error(),
				})
			}
		}
	}

	return s.syncPluginNodeProjection(ctx, pluginNodeProjectionInput{
		PluginID:     registry.PluginId,
		ReleaseID:    registry.ReleaseId,
		DesiredState: registry.DesiredState,
		CurrentState: registry.CurrentState,
		Generation:   registry.Generation,
		Message:      "Current node converged to host plugin generation.",
	})
}

// applyDynamicInstall performs the first activation of a discovered dynamic
// plugin, including artifact archive, SQL install, permission/menu projection,
// optional frontend bundle preparation, and registry finalization.
func (s *Service) applyDynamicInstall(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *pluginManifest,
	desiredState string,
) error {
	release, err := s.getPluginRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("插件发布记录不存在: %s@%s", manifest.ID, manifest.Version)
	}
	if err = s.markPluginRegistryReconciling(ctx, registry, pluginHostStateValue(desiredState)); err != nil {
		return err
	}

	archivedPath, err := s.archiveRuntimePluginReleaseArtifact(ctx, manifest)
	if err != nil {
		return s.rollbackDynamicReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.executeManifestSQLFiles(ctx, manifest, pluginMigrationDirectionInstall); err != nil {
		return s.rollbackDynamicInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
	}
	if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return s.rollbackDynamicInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
	}
	if desiredState == pluginHostStateEnabled.String() {
		if err = s.ValidateRuntimeFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackDynamicInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
		}
		if hasRuntimeFrontendAssets(manifest) {
			if _, err = s.ensureRuntimeFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackDynamicInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
			}
		}
	}

	enabled := pluginStatusDisabled
	if desiredState == pluginHostStateEnabled.String() {
		enabled = pluginStatusEnabled
	}
	registry, err = s.finalizePluginRegistryState(ctx, registry, manifest, release, pluginInstalledYes, enabled)
	if err != nil {
		return s.rollbackDynamicInstallOrUpgrade(ctx, registry, nil, manifest, release.Id, err)
	}
	if err = s.updatePluginReleaseState(ctx, release.Id, s.buildPluginReleaseStatus(pluginInstalledYes, enabled), archivedPath); err != nil {
		return err
	}
	if err = s.syncPluginMetadata(ctx, manifest, registry, "Dynamic plugin release installed on primary node."); err != nil {
		return err
	}
	if err = s.DispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginInstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
		}),
	); err != nil {
		return err
	}
	if enabled == pluginStatusEnabled {
		return s.DispatchHookEvent(
			ctx,
			pluginhost.ExtensionPointPluginEnabled,
			pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
				PluginID: manifest.ID,
				Name:     manifest.Name,
				Version:  manifest.Version,
				Status:   &enabled,
			}),
		)
	}
	return nil
}

// applyDynamicUpgrade moves an installed plugin to a new semantic version.
// Unlike refresh, this path runs upgrade SQL and may replace the active release.
func (s *Service) applyDynamicUpgrade(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *pluginManifest,
	desiredState string,
) error {
	activeManifest, err := s.loadActiveDynamicPluginManifest(ctx, registry)
	if err != nil {
		return err
	}
	// Invalidate the Wasm module cache for the previous active artifact before
	// replacing it so subsequent requests compile from the new artifact.
	if activeManifest != nil && activeManifest.RuntimeArtifact != nil {
		InvalidateWasmModuleCache(activeManifest.RuntimeArtifact.Path)
	}
	release, err := s.getPluginRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("插件发布记录不存在: %s@%s", manifest.ID, manifest.Version)
	}

	if err = s.markPluginRegistryReconciling(ctx, registry, pluginHostStateValue(desiredState)); err != nil {
		return err
	}
	archivedPath, err := s.archiveRuntimePluginReleaseArtifact(ctx, manifest)
	if err != nil {
		return s.rollbackDynamicReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.executeManifestSQLFiles(ctx, manifest, pluginMigrationDirectionUpgrade); err != nil {
		return s.rollbackDynamicInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
	}
	if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return s.rollbackDynamicInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
	}
	if desiredState == pluginHostStateEnabled.String() {
		if err = s.ValidateRuntimeFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackDynamicInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
		}
		if hasRuntimeFrontendAssets(manifest) {
			if _, err = s.ensureRuntimeFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackDynamicInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
			}
		}
	}

	enabled := pluginStatusDisabled
	if desiredState == pluginHostStateEnabled.String() {
		enabled = pluginStatusEnabled
	}
	previousReleaseID := registry.ReleaseId
	registry, err = s.finalizePluginRegistryState(ctx, registry, manifest, release, pluginInstalledYes, enabled)
	if err != nil {
		return s.rollbackDynamicInstallOrUpgrade(ctx, registry, activeManifest, manifest, release.Id, err)
	}
	if previousReleaseID > 0 && previousReleaseID != release.Id {
		if err = s.updatePluginReleaseState(ctx, previousReleaseID, pluginReleaseStatusInstalled, ""); err != nil {
			return err
		}
	}
	if err = s.updatePluginReleaseState(ctx, release.Id, s.buildPluginReleaseStatus(pluginInstalledYes, enabled), archivedPath); err != nil {
		return err
	}
	return s.syncPluginMetadata(ctx, manifest, registry, "Dynamic plugin release upgraded on primary node.")
}

// applyDynamicStateToggle flips enable/disable status for the current active
// release without changing the installed version or artifact archive.
func (s *Service) applyDynamicStateToggle(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *pluginManifest,
	desiredState string,
) error {
	release, err := s.getPluginRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	if err = s.markPluginRegistryReconciling(ctx, registry, pluginHostStateValue(desiredState)); err != nil {
		return err
	}

	enabled := pluginStatusDisabled
	eventName := pluginhost.ExtensionPointPluginDisabled
	if desiredState == pluginHostStateEnabled.String() {
		enabled = pluginStatusEnabled
		eventName = pluginhost.ExtensionPointPluginEnabled
		if err = s.ValidateRuntimeFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackDynamicReleaseFailure(ctx, registry, 0, err)
		}
		if hasRuntimeFrontendAssets(manifest) {
			if _, err = s.ensureRuntimeFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackDynamicReleaseFailure(ctx, registry, 0, err)
			}
		}
	}

	registry, err = s.finalizePluginRegistryState(ctx, registry, manifest, release, pluginInstalledYes, enabled)
	if err != nil {
		return s.rollbackDynamicReleaseFailure(ctx, registry, 0, err)
	}
	if release != nil {
		if err = s.updatePluginReleaseState(ctx, release.Id, s.buildPluginReleaseStatus(pluginInstalledYes, enabled), ""); err != nil {
			return err
		}
	}
	if enabled == pluginStatusDisabled {
		s.invalidateRuntimeFrontendBundle(ctx, manifest.ID, "plugin_disabled")
	}
	if err = s.syncPluginMetadata(ctx, manifest, registry, "Dynamic plugin status converged on primary node."); err != nil {
		return err
	}
	return s.DispatchHookEvent(
		ctx,
		eventName,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
			Status:   &enabled,
		}),
	)
}

// applyDynamicRefresh reapplies host projections for the same semantic version
// when the artifact checksum or archived bytes changed. It intentionally skips
// upgrade SQL because the version contract did not advance.
func (s *Service) applyDynamicRefresh(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *pluginManifest,
	desiredState string,
) error {
	release, err := s.getPluginRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("插件发布记录不存在: %s@%s", manifest.ID, manifest.Version)
	}
	if err = s.markPluginRegistryReconciling(ctx, registry, pluginHostStateValue(desiredState)); err != nil {
		return err
	}

	// Invalidate any previously cached compiled module so the refreshed artifact
	// is recompiled on next bridge invocation.
	if manifest.RuntimeArtifact != nil {
		InvalidateWasmModuleCache(manifest.RuntimeArtifact.Path)
	}
	archivedPath, err := s.archiveRuntimePluginReleaseArtifact(ctx, manifest)
	if err != nil {
		return s.rollbackDynamicReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
		return s.rollbackDynamicReleaseFailure(ctx, registry, release.Id, err)
	}

	enabled := pluginStatusDisabled
	if desiredState == pluginHostStateEnabled.String() {
		enabled = pluginStatusEnabled
		if err = s.ValidateRuntimeFrontendMenuBindings(ctx, manifest); err != nil {
			return s.rollbackDynamicReleaseFailure(ctx, registry, release.Id, err)
		}
		if hasRuntimeFrontendAssets(manifest) {
			if _, err = s.ensureRuntimeFrontendBundle(ctx, manifest); err != nil {
				return s.rollbackDynamicReleaseFailure(ctx, registry, release.Id, err)
			}
		}
	}

	registry, err = s.finalizePluginRegistryState(ctx, registry, manifest, release, pluginInstalledYes, enabled)
	if err != nil {
		return s.rollbackDynamicReleaseFailure(ctx, registry, release.Id, err)
	}
	if err = s.updatePluginReleaseState(ctx, release.Id, s.buildPluginReleaseStatus(pluginInstalledYes, enabled), archivedPath); err != nil {
		return err
	}
	return s.syncPluginMetadata(ctx, manifest, registry, "Dynamic plugin release refreshed on primary node.")
}

func (s *Service) applyDynamicUninstall(ctx context.Context, registry *entity.SysPlugin) error {
	manifest, err := s.loadActiveDynamicPluginManifest(ctx, registry)
	if err != nil {
		return err
	}
	if manifest != nil && manifest.RuntimeArtifact != nil {
		InvalidateWasmModuleCache(manifest.RuntimeArtifact.Path)
	}
	release, err := s.getPluginRegistryRelease(ctx, registry)
	if err != nil {
		return err
	}

	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(do.SysPlugin{
			Status:       pluginStatusDisabled,
			DesiredState: pluginHostStateUninstalled.String(),
			CurrentState: pluginHostStateReconciling.String(),
		}).
		Update()
	if err != nil {
		return err
	}
	if err = s.executeManifestSQLFiles(ctx, manifest, pluginMigrationDirectionUninstall); err != nil {
		return s.rollbackDynamicReleaseFailure(ctx, registry, 0, err)
	}
	if err = s.deletePluginMenusByManifest(ctx, manifest); err != nil {
		return s.rollbackDynamicReleaseFailure(ctx, registry, 0, err)
	}
	registry, err = s.finalizePluginRegistryState(ctx, registry, manifest, nil, pluginInstalledNo, pluginStatusDisabled)
	if err != nil {
		return err
	}
	if release != nil {
		if err = s.updatePluginReleaseState(ctx, release.Id, pluginReleaseStatusUninstalled, ""); err != nil {
			return err
		}
	}
	s.invalidateRuntimeFrontendBundle(ctx, manifest.ID, "plugin_uninstalled")
	if _, err = dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginResourceRef{PluginId: manifest.ID}).
		Delete(); err != nil {
		return err
	}
	if err = s.syncPluginNodeProjection(ctx, pluginNodeProjectionInput{
		PluginID:     registry.PluginId,
		ReleaseID:    0,
		DesiredState: registry.DesiredState,
		CurrentState: registry.CurrentState,
		Generation:   registry.Generation,
		Message:      "Dynamic plugin uninstalled on primary node.",
	}); err != nil {
		return err
	}
	return s.DispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginUninstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: manifest.ID,
			Name:     manifest.Name,
			Version:  manifest.Version,
		}),
	)
}

func (s *Service) rollbackDynamicInstallOrUpgrade(
	ctx context.Context,
	registry *entity.SysPlugin,
	restoreManifest *pluginManifest,
	failedManifest *pluginManifest,
	failedReleaseID int,
	reconcileErr error,
) error {
	if failedManifest != nil {
		if rollbackErr := s.executeManifestSQLFiles(ctx, failedManifest, pluginMigrationDirectionRollback); rollbackErr != nil {
			logger.Warningf(ctx, "rollback dynamic plugin SQL failed plugin=%s err=%v", failedManifest.ID, rollbackErr)
		}
		_ = s.deletePluginMenusByManifest(ctx, failedManifest)
	}
	if restoreManifest != nil {
		if restoreErr := s.syncPluginMenus(ctx, restoreManifest); restoreErr != nil {
			logger.Warningf(ctx, "restore previous plugin menus failed plugin=%s err=%v", restoreManifest.ID, restoreErr)
		}
	}
	return s.rollbackDynamicReleaseFailure(ctx, registry, failedReleaseID, reconcileErr)
}

func (s *Service) rollbackDynamicReleaseFailure(
	ctx context.Context,
	registry *entity.SysPlugin,
	releaseID int,
	reconcileErr error,
) error {
	if releaseID > 0 {
		_ = s.updatePluginReleaseState(ctx, releaseID, pluginReleaseStatusFailed, "")
	}
	_, _ = s.restorePluginRegistryStableState(ctx, registry)
	if registry != nil {
		_ = s.syncPluginNodeProjection(ctx, pluginNodeProjectionInput{
			PluginID:     registry.PluginId,
			ReleaseID:    registry.ReleaseId,
			DesiredState: registry.DesiredState,
			CurrentState: pluginNodeStateFailed.String(),
			Generation:   registry.Generation,
			Message:      reconcileErr.Error(),
		})
	}
	return reconcileErr
}
