// This file defines internal metadata constants together with lightweight
// snapshot and descriptor models used by plugin governance persistence.

package plugin

// pluginMigrationDirection defines the install or uninstall phase persisted in migration records.
type pluginMigrationDirection string

// pluginReleaseStatus defines the normalized release status persisted in sys_plugin_release.
type pluginReleaseStatus string

// pluginMigrationExecutionStatus defines the migration execution result persisted in sys_plugin_migration.
type pluginMigrationExecutionStatus string

// pluginResourceKind defines the abstract resource category persisted in sys_plugin_resource_ref.
type pluginResourceKind string

// pluginResourceOwnerType defines the abstract owner category persisted in sys_plugin_resource_ref.
type pluginResourceOwnerType string

// pluginNodeStateValue defines the current node-state projection enum.
type pluginNodeStateValue string

// pluginHostStateValue defines the desired/current host lifecycle state enum.
type pluginHostStateValue string

// pluginLifecycleStateValue defines the lifecycle summary enum exposed by plugin governance.
type pluginLifecycleStateValue string

// pluginMigrationStateValue defines the review-friendly migration state enum.
type pluginMigrationStateValue string

// pluginResourceSpecType defines the supported plugin backend resource declaration type.
type pluginResourceSpecType string

// pluginResourceFilterOperator defines supported resource filter operators.
type pluginResourceFilterOperator string

// pluginResourceOrderDirection defines supported ordering directions in resource specs.
type pluginResourceOrderDirection string

const (
	pluginMigrationDirectionInstall   pluginMigrationDirection = "install"
	pluginMigrationDirectionUninstall pluginMigrationDirection = "uninstall"
	pluginMigrationDirectionUpgrade   pluginMigrationDirection = "upgrade"
	pluginMigrationDirectionRollback  pluginMigrationDirection = "rollback"

	pluginMigrationStatusFailed    = 0
	pluginMigrationStatusSucceeded = 1

	pluginReleaseStatusPrepared   pluginReleaseStatus = "prepared"
	pluginReleaseStatusUninstalled pluginReleaseStatus = "uninstalled"
	pluginReleaseStatusInstalled   pluginReleaseStatus = "installed"
	pluginReleaseStatusActive      pluginReleaseStatus = "active"
	pluginReleaseStatusFailed      pluginReleaseStatus = "failed"

	pluginMigrationExecutionStatusSucceeded pluginMigrationExecutionStatus = "succeeded"
	pluginMigrationExecutionStatusFailed    pluginMigrationExecutionStatus = "failed"

	pluginResourceKindManifest        pluginResourceKind = "manifest"
	pluginResourceKindBackendEntry    pluginResourceKind = "backend_entry"
	pluginResourceKindRuntimeWasm     pluginResourceKind = "runtime_wasm"
	pluginResourceKindRuntimeFrontend pluginResourceKind = "runtime_frontend"
	pluginResourceKindFrontendPage    pluginResourceKind = "frontend_page"
	pluginResourceKindFrontendSlot    pluginResourceKind = "frontend_slot"
	pluginResourceKindMenu            pluginResourceKind = "menu"
	pluginResourceKindInstallSQL      pluginResourceKind = "install_sql"
	pluginResourceKindUninstallSQL    pluginResourceKind = "uninstall_sql"

	pluginResourceOwnerTypeFile                pluginResourceOwnerType = "file"
	pluginResourceOwnerTypeBackendRegistration pluginResourceOwnerType = "backend-registration"
	pluginResourceOwnerTypeRuntimeArtifact     pluginResourceOwnerType = "runtime-artifact"
	pluginResourceOwnerTypeRuntimeFrontend     pluginResourceOwnerType = "runtime-frontend"
	pluginResourceOwnerTypeInstallSQL          pluginResourceOwnerType = "install-sql"
	pluginResourceOwnerTypeUninstallSQL        pluginResourceOwnerType = "uninstall-sql"
	pluginResourceOwnerTypeFrontendPageEntry   pluginResourceOwnerType = "frontend-page-entry"
	pluginResourceOwnerTypeFrontendSlotEntry   pluginResourceOwnerType = "frontend-slot-entry"
	pluginResourceOwnerTypeMenuEntry           pluginResourceOwnerType = "menu-entry"

	pluginNodeStateReconciling pluginNodeStateValue = "reconciling"
	pluginNodeStateFailed      pluginNodeStateValue = "failed"
	pluginNodeStateEnabled     pluginNodeStateValue = "enabled"
	pluginNodeStateInstalled   pluginNodeStateValue = "installed"
	pluginNodeStateUninstalled pluginNodeStateValue = "uninstalled"

	pluginHostStateReconciling pluginHostStateValue = "reconciling"
	pluginHostStateFailed      pluginHostStateValue = "failed"
	pluginHostStateEnabled     pluginHostStateValue = "enabled"
	pluginHostStateInstalled   pluginHostStateValue = "installed"
	pluginHostStateUninstalled pluginHostStateValue = "uninstalled"

	pluginLifecycleStateSourceEnabled      pluginLifecycleStateValue = "source_enabled"
	pluginLifecycleStateSourceDisabled     pluginLifecycleStateValue = "source_disabled"
	pluginLifecycleStateRuntimeUninstalled pluginLifecycleStateValue = "runtime_uninstalled"
	pluginLifecycleStateRuntimeInstalled   pluginLifecycleStateValue = "runtime_installed"
	pluginLifecycleStateRuntimeEnabled     pluginLifecycleStateValue = "runtime_enabled"

	pluginMigrationStateNone      pluginMigrationStateValue = "none"
	pluginMigrationStateSucceeded pluginMigrationStateValue = "succeeded"
	pluginMigrationStateFailed    pluginMigrationStateValue = "failed"

	pluginResourceSpecTypeTableList pluginResourceSpecType = "table-list"

	pluginResourceFilterOperatorEQ      pluginResourceFilterOperator = "eq"
	pluginResourceFilterOperatorLike    pluginResourceFilterOperator = "like"
	pluginResourceFilterOperatorGTEDate pluginResourceFilterOperator = "gte-date"
	pluginResourceFilterOperatorLTEDate pluginResourceFilterOperator = "lte-date"

	pluginResourceOrderDirectionASC  pluginResourceOrderDirection = "asc"
	pluginResourceOrderDirectionDESC pluginResourceOrderDirection = "desc"
)

// String returns the canonical migration direction value.
func (value pluginMigrationDirection) String() string { return string(value) }

// String returns the canonical release status value.
func (value pluginReleaseStatus) String() string { return string(value) }

// String returns the canonical migration execution status value.
func (value pluginMigrationExecutionStatus) String() string { return string(value) }

// String returns the canonical resource kind value.
func (value pluginResourceKind) String() string { return string(value) }

// String returns the canonical resource owner-type value.
func (value pluginResourceOwnerType) String() string { return string(value) }

// String returns the canonical node-state value.
func (value pluginNodeStateValue) String() string { return string(value) }

// String returns the canonical host-state value.
func (value pluginHostStateValue) String() string { return string(value) }

// String returns the canonical lifecycle-state value.
func (value pluginLifecycleStateValue) String() string { return string(value) }

// String returns the canonical migration-state value.
func (value pluginMigrationStateValue) String() string { return string(value) }

// String returns the canonical resource spec type value.
func (value pluginResourceSpecType) String() string { return string(value) }

// String returns the canonical resource filter-operator value.
func (value pluginResourceFilterOperator) String() string { return string(value) }

// String returns the canonical resource order-direction value.
func (value pluginResourceOrderDirection) String() string { return string(value) }

// pluginManifestSnapshot stores the review-friendly manifest snapshot persisted in sys_plugin_release.
type pluginManifestSnapshot struct {
	ID                        string `yaml:"id"`
	Name                      string `yaml:"name"`
	Version                   string `yaml:"version"`
	Type                      string `yaml:"type"`
	Description               string `yaml:"description,omitempty"`
	Author                    string `yaml:"author,omitempty"`
	Homepage                  string `yaml:"homepage,omitempty"`
	License                   string `yaml:"license,omitempty"`
	RuntimeKind               string `yaml:"runtimeKind,omitempty"`
	RuntimeABIVersion         string `yaml:"runtimeAbiVersion,omitempty"`
	ManifestDeclared          bool   `yaml:"manifestDeclared"`
	InstallSQLCount           int    `yaml:"installSqlCount,omitempty"`
	UninstallSQLCount         int    `yaml:"uninstallSqlCount,omitempty"`
	FrontendPageCount         int    `yaml:"frontendPageCount,omitempty"`
	FrontendSlotCount         int    `yaml:"frontendSlotCount,omitempty"`
	MenuCount                 int    `yaml:"menuCount,omitempty"`
	BackendHookCount          int    `yaml:"backendHookCount,omitempty"`
	ResourceSpecCount         int    `yaml:"resourceSpecCount,omitempty"`
	RuntimeFrontendAssetCount int    `yaml:"runtimeFrontendAssetCount,omitempty"`
	RuntimeSQLAssetCount      int    `yaml:"runtimeSqlAssetCount,omitempty"`
}

// pluginResourceRefDescriptor represents one discovered plugin asset recorded for later review.
type pluginResourceRefDescriptor struct {
	Kind      pluginResourceKind
	Key       string
	OwnerType pluginResourceOwnerType
	OwnerKey  string
	Remark    string
}

// derivePluginNodeState converts installation and enablement flags into one
// stable node-state key for the governance projection.
func derivePluginNodeState(installed int, enabled int) string {
	if installed != pluginInstalledYes {
		return pluginNodeStateUninstalled.String()
	}
	if enabled == pluginStatusEnabled {
		return pluginNodeStateEnabled.String()
	}
	return pluginNodeStateInstalled.String()
}

// derivePluginHostState converts install and enablement flags into the stable
// host lifecycle state stored in sys_plugin desired_state/current_state.
func derivePluginHostState(installed int, enabled int) string {
	if installed != pluginInstalledYes {
		return pluginHostStateUninstalled.String()
	}
	if enabled == pluginStatusEnabled {
		return pluginHostStateEnabled.String()
	}
	return pluginHostStateInstalled.String()
}

// derivePluginLifecycleState converts the plugin type and runtime flags into the
// lifecycle state exposed by the management API.
func derivePluginLifecycleState(pluginType string, installed int, enabled int) string {
	if normalizePluginType(pluginType) == pluginTypeSource {
		if enabled == pluginStatusEnabled {
			return pluginLifecycleStateSourceEnabled.String()
		}
		return pluginLifecycleStateSourceDisabled.String()
	}
	if installed != pluginInstalledYes {
		return pluginLifecycleStateRuntimeUninstalled.String()
	}
	if enabled == pluginStatusEnabled {
		return pluginLifecycleStateRuntimeEnabled.String()
	}
	return pluginLifecycleStateRuntimeInstalled.String()
}
