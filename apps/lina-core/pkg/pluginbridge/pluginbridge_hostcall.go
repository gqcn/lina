// This file defines the host call protocol constants, opcode definitions,
// and capability validation used by both host and guest sides of the
// dynamic plugin bridge.

package pluginbridge

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// HostModuleName is the wazero host module namespace for Lina host functions.
	HostModuleName = "lina_env"
	// HostCallFunctionName is the single host call dispatch function name.
	HostCallFunctionName = "host_call"

	// DefaultGuestHostCallAllocExport is the guest export used by the host to
	// allocate response buffers during host call processing.
	DefaultGuestHostCallAllocExport = "lina_host_call_alloc"
)

// Host call opcodes identify specific host capabilities. Each opcode maps to
// exactly one capability string that must be declared in the plugin manifest.
const (
	// OpcodeLog writes a structured log entry through the host logger.
	OpcodeLog uint32 = 0x0001

	// OpcodeStateGet reads a plugin-scoped key-value state entry.
	OpcodeStateGet uint32 = 0x0101
	// OpcodeStateSet writes a plugin-scoped key-value state entry.
	OpcodeStateSet uint32 = 0x0102
	// OpcodeStateDelete removes a plugin-scoped key-value state entry.
	OpcodeStateDelete uint32 = 0x0103

	// OpcodeDBQuery executes a read-only SQL query (SELECT only).
	OpcodeDBQuery uint32 = 0x0201
	// OpcodeDBExecute executes a write SQL statement (INSERT/UPDATE/DELETE).
	OpcodeDBExecute uint32 = 0x0202
)

// Capability identifiers declared in plugin.yaml to request host functions.
const (
	// CapabilityLog grants access to host structured logging.
	CapabilityLog = "host:log"
	// CapabilityState grants access to plugin-scoped key-value state storage.
	CapabilityState = "host:state"
	// CapabilityDBQuery grants access to read-only SQL queries.
	CapabilityDBQuery = "host:db:query"
	// CapabilityDBExecute grants access to write SQL statements.
	CapabilityDBExecute = "host:db:execute"
)

// Host call response status codes.
const (
	// HostCallStatusSuccess indicates the host call completed successfully.
	HostCallStatusSuccess uint32 = 0
	// HostCallStatusCapabilityDenied indicates the plugin lacks the required capability.
	HostCallStatusCapabilityDenied uint32 = 1
	// HostCallStatusNotFound indicates an unknown opcode.
	HostCallStatusNotFound uint32 = 2
	// HostCallStatusInvalidRequest indicates a malformed request payload.
	HostCallStatusInvalidRequest uint32 = 3
	// HostCallStatusInternalError indicates a host-side processing failure.
	HostCallStatusInternalError uint32 = 4
)

// Log level constants used by the host:log capability.
const (
	// LogLevelDebug maps to logger.Debug.
	LogLevelDebug int32 = 1
	// LogLevelInfo maps to logger.Info.
	LogLevelInfo int32 = 2
	// LogLevelWarning maps to logger.Warning.
	LogLevelWarning int32 = 3
	// LogLevelError maps to logger.Error.
	LogLevelError int32 = 4
)

// opcodeCapabilityMap maps each opcode to its required capability string.
var opcodeCapabilityMap = map[uint32]string{
	OpcodeLog:         CapabilityLog,
	OpcodeStateGet:    CapabilityState,
	OpcodeStateSet:    CapabilityState,
	OpcodeStateDelete: CapabilityState,
	OpcodeDBQuery:     CapabilityDBQuery,
	OpcodeDBExecute:   CapabilityDBExecute,
}

// allCapabilities lists all known capability strings for validation.
var allCapabilities = map[string]struct{}{
	CapabilityLog:       {},
	CapabilityState:     {},
	CapabilityDBQuery:   {},
	CapabilityDBExecute: {},
}

// OpcodeToCapability returns the capability string required by the given opcode,
// or an empty string if the opcode is unknown.
func OpcodeToCapability(opcode uint32) string {
	return opcodeCapabilityMap[opcode]
}

// AllCapabilities returns a sorted list of all known capability identifiers.
func AllCapabilities() []string {
	result := make([]string, 0, len(allCapabilities))
	for cap := range allCapabilities {
		result = append(result, cap)
	}
	sort.Strings(result)
	return result
}

// ValidateCapabilities checks that every capability string is recognized.
func ValidateCapabilities(capabilities []string) error {
	for _, cap := range capabilities {
		normalized := strings.TrimSpace(cap)
		if normalized == "" {
			return fmt.Errorf("插件能力声明不能为空")
		}
		if _, ok := allCapabilities[normalized]; !ok {
			return fmt.Errorf("未知的插件能力声明: %s，支持的值: %v", normalized, AllCapabilities())
		}
	}
	return nil
}

// NormalizeCapabilities trims whitespace and removes duplicates from a capability list.
func NormalizeCapabilities(capabilities []string) []string {
	seen := make(map[string]struct{}, len(capabilities))
	result := make([]string, 0, len(capabilities))
	for _, cap := range capabilities {
		normalized := strings.TrimSpace(cap)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

// CapabilitySliceToMap converts a capability slice to a set for O(1) lookup.
func CapabilitySliceToMap(capabilities []string) map[string]struct{} {
	result := make(map[string]struct{}, len(capabilities))
	for _, cap := range capabilities {
		normalized := strings.TrimSpace(cap)
		if normalized != "" {
			result[normalized] = struct{}{}
		}
	}
	return result
}
