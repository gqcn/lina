// This file tests host call request and response codec round trips.

package pluginbridge

import (
	"testing"
)

func TestHostCallResponseEnvelopeRoundTrip(t *testing.T) {
	original := &HostCallResponseEnvelope{
		Status:  HostCallStatusCapabilityDenied,
		Payload: []byte("missing host:log capability"),
	}
	data := MarshalHostCallResponse(original)
	decoded, err := UnmarshalHostCallResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != original.Status {
		t.Errorf("status: got %d, want %d", decoded.Status, original.Status)
	}
	if string(decoded.Payload) != string(original.Payload) {
		t.Errorf("payload: got %q, want %q", decoded.Payload, original.Payload)
	}
}

func TestHostCallSuccessResponseRoundTrip(t *testing.T) {
	original := NewHostCallEmptySuccessResponse()
	data := MarshalHostCallResponse(original)
	decoded, err := UnmarshalHostCallResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != HostCallStatusSuccess {
		t.Errorf("status: got %d, want %d", decoded.Status, HostCallStatusSuccess)
	}
}

func TestHostCallLogRequestRoundTrip(t *testing.T) {
	original := &HostCallLogRequest{
		Level:   LogLevelWarning,
		Message: "test warning message",
		Fields:  map[string]string{"key1": "val1", "key2": "val2"},
	}
	data := MarshalHostCallLogRequest(original)
	decoded, err := UnmarshalHostCallLogRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Level != original.Level {
		t.Errorf("level: got %d, want %d", decoded.Level, original.Level)
	}
	if decoded.Message != original.Message {
		t.Errorf("message: got %q, want %q", decoded.Message, original.Message)
	}
	if len(decoded.Fields) != 2 || decoded.Fields["key1"] != "val1" {
		t.Errorf("fields: got %v, want %v", decoded.Fields, original.Fields)
	}
}

func TestHostCallStateGetRequestRoundTrip(t *testing.T) {
	original := &HostCallStateGetRequest{Key: "counter"}
	data := MarshalHostCallStateGetRequest(original)
	decoded, err := UnmarshalHostCallStateGetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Errorf("key: got %q, want %q", decoded.Key, original.Key)
	}
}

func TestHostCallStateGetResponseRoundTrip(t *testing.T) {
	original := &HostCallStateGetResponse{Value: "42", Found: true}
	data := MarshalHostCallStateGetResponse(original)
	decoded, err := UnmarshalHostCallStateGetResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Value != original.Value {
		t.Errorf("value: got %q, want %q", decoded.Value, original.Value)
	}
	if decoded.Found != original.Found {
		t.Errorf("found: got %v, want %v", decoded.Found, original.Found)
	}
}

func TestHostCallStateSetRequestRoundTrip(t *testing.T) {
	original := &HostCallStateSetRequest{Key: "counter", Value: "43"}
	data := MarshalHostCallStateSetRequest(original)
	decoded, err := UnmarshalHostCallStateSetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Errorf("key: got %q, want %q", decoded.Key, original.Key)
	}
	if decoded.Value != original.Value {
		t.Errorf("value: got %q, want %q", decoded.Value, original.Value)
	}
}

func TestHostCallStateDeleteRequestRoundTrip(t *testing.T) {
	original := &HostCallStateDeleteRequest{Key: "counter"}
	data := MarshalHostCallStateDeleteRequest(original)
	decoded, err := UnmarshalHostCallStateDeleteRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Errorf("key: got %q, want %q", decoded.Key, original.Key)
	}
}

func TestHostCallDBQueryRequestRoundTrip(t *testing.T) {
	original := &HostCallDBQueryRequest{
		SQL:     "SELECT id, name FROM sys_user WHERE status = ?",
		Args:    []string{"1"},
		MaxRows: 100,
	}
	data := MarshalHostCallDBQueryRequest(original)
	decoded, err := UnmarshalHostCallDBQueryRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.SQL != original.SQL {
		t.Errorf("sql: got %q, want %q", decoded.SQL, original.SQL)
	}
	if len(decoded.Args) != 1 || decoded.Args[0] != "1" {
		t.Errorf("args: got %v, want %v", decoded.Args, original.Args)
	}
	if decoded.MaxRows != original.MaxRows {
		t.Errorf("maxRows: got %d, want %d", decoded.MaxRows, original.MaxRows)
	}
}

func TestHostCallDBQueryResponseRoundTrip(t *testing.T) {
	original := &HostCallDBQueryResponse{
		Columns:  []string{"id", "name"},
		Rows:     [][]string{{"1", "admin"}, {"2", "user1"}},
		RowCount: 2,
	}
	data := MarshalHostCallDBQueryResponse(original)
	decoded, err := UnmarshalHostCallDBQueryResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(decoded.Columns) != 2 || decoded.Columns[0] != "id" || decoded.Columns[1] != "name" {
		t.Errorf("columns: got %v, want %v", decoded.Columns, original.Columns)
	}
	if len(decoded.Rows) != 2 {
		t.Fatalf("rows: got %d, want 2", len(decoded.Rows))
	}
	if decoded.Rows[0][0] != "1" || decoded.Rows[0][1] != "admin" {
		t.Errorf("row[0]: got %v, want %v", decoded.Rows[0], original.Rows[0])
	}
	if decoded.RowCount != 2 {
		t.Errorf("rowCount: got %d, want 2", decoded.RowCount)
	}
}

func TestHostCallDBExecuteRequestRoundTrip(t *testing.T) {
	original := &HostCallDBExecuteRequest{
		SQL:  "INSERT INTO plg_demo_items (name) VALUES (?)",
		Args: []string{"item1"},
	}
	data := MarshalHostCallDBExecuteRequest(original)
	decoded, err := UnmarshalHostCallDBExecuteRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.SQL != original.SQL {
		t.Errorf("sql: got %q, want %q", decoded.SQL, original.SQL)
	}
	if len(decoded.Args) != 1 || decoded.Args[0] != "item1" {
		t.Errorf("args: got %v, want %v", decoded.Args, original.Args)
	}
}

func TestHostCallDBExecuteResponseRoundTrip(t *testing.T) {
	original := &HostCallDBExecuteResponse{
		RowsAffected: 3,
		LastInsertID: 42,
	}
	data := MarshalHostCallDBExecuteResponse(original)
	decoded, err := UnmarshalHostCallDBExecuteResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.RowsAffected != original.RowsAffected {
		t.Errorf("rowsAffected: got %d, want %d", decoded.RowsAffected, original.RowsAffected)
	}
	if decoded.LastInsertID != original.LastInsertID {
		t.Errorf("lastInsertId: got %d, want %d", decoded.LastInsertID, original.LastInsertID)
	}
}
