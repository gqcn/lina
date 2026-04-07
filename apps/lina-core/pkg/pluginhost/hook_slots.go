package pluginhost

// HookSlot defines one published backend plugin hook slot.
type HookSlot string

const (
	// HookSlotAuthLoginSucceeded is fired after user login succeeds.
	HookSlotAuthLoginSucceeded HookSlot = "auth.login.succeeded"
	// HookSlotAuthLogoutSucceeded is fired after user logout succeeds.
	HookSlotAuthLogoutSucceeded HookSlot = "auth.logout.succeeded"
	// HookSlotPluginInstalled is fired after a runtime plugin is installed.
	HookSlotPluginInstalled HookSlot = "plugin.installed"
	// HookSlotPluginEnabled is fired after a plugin is enabled.
	HookSlotPluginEnabled HookSlot = "plugin.enabled"
	// HookSlotPluginDisabled is fired after a plugin is disabled.
	HookSlotPluginDisabled HookSlot = "plugin.disabled"
	// HookSlotPluginUninstalled is fired after a runtime plugin is uninstalled.
	HookSlotPluginUninstalled HookSlot = "plugin.uninstalled"
	// HookSlotSystemStarted is fired after host http server startup.
	HookSlotSystemStarted HookSlot = "system.started"
)

// HookAction defines one supported plugin hook action.
type HookAction string

const (
	// HookActionInsert inserts one row into plugin-owned table.
	HookActionInsert HookAction = "insert"
)

var publishedHookSlots = map[HookSlot]struct{}{
	HookSlotAuthLoginSucceeded:  {},
	HookSlotAuthLogoutSucceeded: {},
	HookSlotPluginInstalled:     {},
	HookSlotPluginEnabled:       {},
	HookSlotPluginDisabled:      {},
	HookSlotPluginUninstalled:   {},
	HookSlotSystemStarted:       {},
}

var publishedHookActions = map[HookAction]struct{}{
	HookActionInsert: {},
}

// String returns the canonical slot key.
func (slot HookSlot) String() string {
	return string(slot)
}

// String returns the canonical hook action key.
func (action HookAction) String() string {
	return string(action)
}

// IsPublishedHookSlot reports whether the hook slot is part of the published host contract.
func IsPublishedHookSlot(slot HookSlot) bool {
	_, ok := publishedHookSlots[slot]
	return ok
}

// IsSupportedHookAction reports whether the hook action is supported by current host runtime.
func IsSupportedHookAction(action HookAction) bool {
	_, ok := publishedHookActions[action]
	return ok
}
