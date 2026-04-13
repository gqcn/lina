package plugin

import (
	"testing"

	"lina-core/pkg/pluginbridge"
)

func TestOpcodeToCapabilityMapsAllOpcodes(t *testing.T) {
	cases := []struct {
		opcode     uint32
		capability string
	}{
		{pluginbridge.OpcodeLog, pluginbridge.CapabilityLog},
		{pluginbridge.OpcodeStateGet, pluginbridge.CapabilityState},
		{pluginbridge.OpcodeStateSet, pluginbridge.CapabilityState},
		{pluginbridge.OpcodeStateDelete, pluginbridge.CapabilityState},
		{pluginbridge.OpcodeDBQuery, pluginbridge.CapabilityDBQuery},
		{pluginbridge.OpcodeDBExecute, pluginbridge.CapabilityDBExecute},
	}
	for _, tc := range cases {
		got := pluginbridge.OpcodeToCapability(tc.opcode)
		if got != tc.capability {
			t.Errorf("OpcodeToCapability(0x%04x): got %q, want %q", tc.opcode, got, tc.capability)
		}
	}
}

func TestOpcodeToCapabilityUnknownReturnsEmpty(t *testing.T) {
	got := pluginbridge.OpcodeToCapability(0xFFFF)
	if got != "" {
		t.Errorf("expected empty for unknown opcode, got %q", got)
	}
}

func TestValidateCapabilitiesAcceptsValid(t *testing.T) {
	err := pluginbridge.ValidateCapabilities([]string{"host:log", "host:state", "host:db:query"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateCapabilitiesRejectsUnknown(t *testing.T) {
	err := pluginbridge.ValidateCapabilities([]string{"host:log", "host:unknown"})
	if err == nil {
		t.Error("expected error for unknown capability")
	}
}

func TestValidateCapabilitiesRejectsEmpty(t *testing.T) {
	err := pluginbridge.ValidateCapabilities([]string{""})
	if err == nil {
		t.Error("expected error for empty capability")
	}
}

func TestHostCallContextHasCapability(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "test-plugin",
		capabilities: map[string]struct{}{
			"host:log":   {},
			"host:state": {},
		},
	}
	if !hcc.hasCapability("host:log") {
		t.Error("expected host:log to be granted")
	}
	if hcc.hasCapability("host:db:query") {
		t.Error("expected host:db:query to not be granted")
	}
}

func TestHostCallContextNilCapabilities(t *testing.T) {
	hcc := &hostCallContext{pluginID: "test-plugin"}
	if hcc.hasCapability("host:log") {
		t.Error("expected nil capabilities to deny all")
	}
}

func TestContainsForbiddenKeyword(t *testing.T) {
	cases := []struct {
		sql    string
		expect string
	}{
		{"SELECT * FROM sys_user", ""},
		{"DROP TABLE sys_user", "DROP"},
		{"SELECT * FROM sys_user; DROP TABLE sys_user", "DROP"},
		{"ALTER TABLE sys_user ADD COLUMN x INT", "ALTER"},
		{"CREATE TABLE test (id INT)", "CREATE"},
		{"TRUNCATE TABLE sys_user", "TRUNCATE"},
		{"SELECT dropped FROM sys_user", ""},
		// Note: keyword inside string literals is a known false positive; full SQL
		// parsing is intentionally out of scope for Phase 1.
		{"INSERT INTO create_log (msg) VALUES ('test')", ""},
		{"GRANT ALL ON *.* TO 'user'@'host'", "GRANT"},
	}
	for _, tc := range cases {
		got := containsForbiddenKeyword(tc.sql)
		if got != tc.expect {
			t.Errorf("containsForbiddenKeyword(%q): got %q, want %q", tc.sql, got, tc.expect)
		}
	}
}

func TestDBQueryRejectsNonSelect(t *testing.T) {
	hcc := &hostCallContext{
		pluginID:     "test-plugin",
		capabilities: map[string]struct{}{pluginbridge.CapabilityDBQuery: {}},
	}
	reqBytes := pluginbridge.MarshalHostCallDBQueryRequest(&pluginbridge.HostCallDBQueryRequest{
		SQL:     "INSERT INTO sys_user (name) VALUES ('test')",
		MaxRows: 10,
	})
	resp := handleHostDBQuery(nil, hcc, reqBytes)
	if resp.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Errorf("expected invalid_request for non-SELECT, got status %d", resp.Status)
	}
}

func TestDBExecuteRejectsDDL(t *testing.T) {
	hcc := &hostCallContext{
		pluginID:     "test-plugin",
		capabilities: map[string]struct{}{pluginbridge.CapabilityDBExecute: {}},
	}
	reqBytes := pluginbridge.MarshalHostCallDBExecuteRequest(&pluginbridge.HostCallDBExecuteRequest{
		SQL: "DROP TABLE sys_user",
	})
	resp := handleHostDBExecute(nil, hcc, reqBytes)
	if resp.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Errorf("expected invalid_request for DDL, got status %d", resp.Status)
	}
}

func TestDBExecuteRejectsSelect(t *testing.T) {
	hcc := &hostCallContext{
		pluginID:     "test-plugin",
		capabilities: map[string]struct{}{pluginbridge.CapabilityDBExecute: {}},
	}
	reqBytes := pluginbridge.MarshalHostCallDBExecuteRequest(&pluginbridge.HostCallDBExecuteRequest{
		SQL: "SELECT * FROM sys_user",
	})
	resp := handleHostDBExecute(nil, hcc, reqBytes)
	if resp.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Errorf("expected invalid_request for SELECT in execute, got status %d", resp.Status)
	}
}
