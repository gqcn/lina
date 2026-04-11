// This file provides small helpers for the host-level generation model used by
// dynamic plugin installs, upgrades, rollbacks, and node-state convergence.

package plugin

import (
	"strings"

	"lina-core/internal/model/entity"
)

// compareSemanticVersions compares two validated semantic-version strings.
// It returns -1 when left < right, 0 when equal, and 1 when left > right.
func compareSemanticVersions(left string, right string) (int, error) {
	leftVersion, err := parseSemanticVersion(left)
	if err != nil {
		return 0, err
	}
	rightVersion, err := parseSemanticVersion(right)
	if err != nil {
		return 0, err
	}

	switch {
	case leftVersion.Major < rightVersion.Major:
		return -1, nil
	case leftVersion.Major > rightVersion.Major:
		return 1, nil
	case leftVersion.Minor < rightVersion.Minor:
		return -1, nil
	case leftVersion.Minor > rightVersion.Minor:
		return 1, nil
	case leftVersion.Patch < rightVersion.Patch:
		return -1, nil
	case leftVersion.Patch > rightVersion.Patch:
		return 1, nil
	default:
		return 0, nil
	}
}

// nextPluginGeneration returns the next stable generation number for one plugin.
func nextPluginGeneration(registry *entity.SysPlugin) int64 {
	if registry != nil && registry.Generation > 0 {
		return registry.Generation + 1
	}
	return 1
}

// buildStablePluginHostState rebuilds the stable host-state enum from current
// install and enablement flags, ignoring transient reconciling/failed markers.
func buildStablePluginHostState(registry *entity.SysPlugin) string {
	if registry == nil {
		return pluginHostStateUninstalled.String()
	}
	return derivePluginHostState(registry.Installed, registry.Status)
}

// shouldTrackStagedDynamicRelease reports whether discovery found a newer
// dynamic artifact that should stay staged instead of immediately replacing the
// currently active registry version.
func shouldTrackStagedDynamicRelease(registry *entity.SysPlugin, manifest *pluginManifest) bool {
	if registry == nil || manifest == nil {
		return false
	}
	if normalizePluginType(registry.Type) != pluginTypeDynamic ||
		normalizePluginType(manifest.Type) != pluginTypeDynamic {
		return false
	}
	if registry.Installed != pluginInstalledYes {
		return false
	}
	if strings.TrimSpace(registry.Version) == "" || strings.TrimSpace(manifest.Version) == "" {
		return false
	}
	return strings.TrimSpace(registry.Version) != strings.TrimSpace(manifest.Version)
}
