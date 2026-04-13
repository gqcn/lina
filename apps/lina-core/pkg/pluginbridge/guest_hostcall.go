//go:build wasip1

// This file provides high-level guest-side helpers for invoking host functions
// through the lina_env.host_call import. It is only compiled for wasip1 targets.

package pluginbridge

import (
	"fmt"
	"strconv"
	"unsafe"
)

// linaHostCall is the imported host function provided by the lina_env module.
// It dispatches a host call identified by opcode and returns a packed
// (pointer << 32 | length) pair pointing to the response in guest memory.
//
//go:wasmimport lina_env host_call
func linaHostCall(opcode uint32, reqPtr uint32, reqLen uint32) uint64

// invokeHostCall sends one host call and returns the decoded response payload.
// On non-success status it returns an error describing the failure.
func invokeHostCall(opcode uint32, reqBytes []byte) ([]byte, error) {
	var reqPtr uint32
	var reqLen uint32
	if len(reqBytes) > 0 {
		reqPtr = uint32(uintptr(unsafe.Pointer(&reqBytes[0])))
		reqLen = uint32(len(reqBytes))
	}

	packed := linaHostCall(opcode, reqPtr, reqLen)
	respPtr := uint32(packed >> 32)
	respLen := uint32(packed & 0xffffffff)

	if respLen == 0 {
		return nil, nil
	}

	// The host wrote the response into guestHostCallResponseBuffer via the
	// lina_host_call_alloc export. Read it from there.
	buf := guestHostCallResponseBuffer
	if uint32(len(buf)) < respLen {
		return nil, fmt.Errorf("host call response buffer underflow: have %d, need %d", len(buf), respLen)
	}
	_ = respPtr // pointer is the start of guestHostCallResponseBuffer

	// Decode the generic response envelope.
	envelope, err := UnmarshalHostCallResponse(buf[:respLen])
	if err != nil {
		return nil, fmt.Errorf("host call response decode failed: %w", err)
	}
	if envelope.Status != HostCallStatusSuccess {
		msg := string(envelope.Payload)
		if msg == "" {
			msg = fmt.Sprintf("host call failed with status %d", envelope.Status)
		}
		return nil, fmt.Errorf("host call error (status=%d): %s", envelope.Status, msg)
	}
	return envelope.Payload, nil
}

// HostLog sends a structured log entry through the host logger.
func HostLog(level int, message string, fields map[string]string) error {
	req := &HostCallLogRequest{
		Level:   int32(level),
		Message: message,
		Fields:  fields,
	}
	_, err := invokeHostCall(OpcodeLog, MarshalHostCallLogRequest(req))
	return err
}

// HostStateGet reads a plugin-scoped state value by key.
// Returns the value, whether it was found, and any error.
func HostStateGet(key string) (string, bool, error) {
	req := &HostCallStateGetRequest{Key: key}
	payload, err := invokeHostCall(OpcodeStateGet, MarshalHostCallStateGetRequest(req))
	if err != nil {
		return "", false, err
	}
	if len(payload) == 0 {
		return "", false, nil
	}
	resp, err := UnmarshalHostCallStateGetResponse(payload)
	if err != nil {
		return "", false, err
	}
	return resp.Value, resp.Found, nil
}

// HostStateSet writes a plugin-scoped state value.
func HostStateSet(key, value string) error {
	req := &HostCallStateSetRequest{Key: key, Value: value}
	_, err := invokeHostCall(OpcodeStateSet, MarshalHostCallStateSetRequest(req))
	return err
}

// HostStateDelete removes a plugin-scoped state value.
func HostStateDelete(key string) error {
	req := &HostCallStateDeleteRequest{Key: key}
	_, err := invokeHostCall(OpcodeStateDelete, MarshalHostCallStateDeleteRequest(req))
	return err
}

// HostDBQueryResult holds the result of a read-only SQL query.
type HostDBQueryResult struct {
	Columns  []string
	Rows     [][]string
	RowCount int
}

// HostDBQuery executes a read-only SQL query and returns the result.
// maxRows limits the number of rows returned (capped at 1000 by the host).
func HostDBQuery(sql string, args []string, maxRows int) (*HostDBQueryResult, error) {
	req := &HostCallDBQueryRequest{
		SQL:     sql,
		Args:    args,
		MaxRows: int32(maxRows),
	}
	payload, err := invokeHostCall(OpcodeDBQuery, MarshalHostCallDBQueryRequest(req))
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return &HostDBQueryResult{}, nil
	}
	resp, err := UnmarshalHostCallDBQueryResponse(payload)
	if err != nil {
		return nil, err
	}
	return &HostDBQueryResult{
		Columns:  resp.Columns,
		Rows:     resp.Rows,
		RowCount: int(resp.RowCount),
	}, nil
}

// HostDBExecute executes a write SQL statement and returns rows affected and last insert ID.
func HostDBExecute(sql string, args []string) (rowsAffected int64, lastInsertID int64, err error) {
	req := &HostCallDBExecuteRequest{SQL: sql, Args: args}
	payload, err := invokeHostCall(OpcodeDBExecute, MarshalHostCallDBExecuteRequest(req))
	if err != nil {
		return 0, 0, err
	}
	if len(payload) == 0 {
		return 0, 0, nil
	}
	resp, err := UnmarshalHostCallDBExecuteResponse(payload)
	if err != nil {
		return 0, 0, err
	}
	return resp.RowsAffected, resp.LastInsertID, nil
}

// HostStateGetInt reads a plugin-scoped integer state value.
func HostStateGetInt(key string) (int, bool, error) {
	value, found, err := HostStateGet(key)
	if err != nil || !found {
		return 0, found, err
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, true, fmt.Errorf("state value for %q is not an integer: %s", key, value)
	}
	return n, true, nil
}

// HostStateSetInt writes a plugin-scoped integer state value.
func HostStateSetInt(key string, value int) error {
	return HostStateSet(key, strconv.Itoa(value))
}
