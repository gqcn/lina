// This file implements the host:db:query and host:db:execute capability
// handlers that provide SQL access for WASM guest plugins.

package wasm

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/pluginbridge"
)

// dbQueryMaxRowsCeiling is the absolute maximum number of rows a plugin can
// request in a single query to prevent excessive memory usage.
const dbQueryMaxRowsCeiling = 1000

// dbForbiddenKeywords lists DDL and administrative keywords that are never
// allowed in plugin SQL statements.
var dbForbiddenKeywords = []string{
	"DROP", "ALTER", "CREATE", "TRUNCATE",
	"GRANT", "REVOKE",
}

// handleHostDBQuery processes OpcodeDBQuery requests. Only SELECT statements
// are allowed, and a row limit is enforced.
func handleHostDBQuery(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *pluginbridge.HostCallResponseEnvelope {
	req, err := pluginbridge.UnmarshalHostCallDBQueryRequest(reqBytes)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	sql := strings.TrimSpace(req.SQL)
	if sql == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, "SQL statement must not be empty")
	}

	// Validate statement type: only SELECT allowed for query.
	if !strings.HasPrefix(strings.ToUpper(sql), "SELECT") {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest,
			"host:db:query only allows SELECT statements")
	}

	// Reject forbidden keywords.
	if keyword := containsForbiddenKeyword(sql); keyword != "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest,
			fmt.Sprintf("SQL statement contains forbidden keyword: %s", keyword))
	}

	// Enforce row limit.
	maxRows := int(req.MaxRows)
	if maxRows <= 0 || maxRows > dbQueryMaxRowsCeiling {
		maxRows = dbQueryMaxRowsCeiling
	}

	// Build args slice for parameterized query.
	args := buildDBArgs(req.Args)
	args = append(args, maxRows)

	// Execute with LIMIT appended.
	records, err := g.DB().Ctx(ctx).GetAll(ctx, sql+" LIMIT ?", args...)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}

	// Build response.
	resp := &pluginbridge.HostCallDBQueryResponse{}
	if len(records) > 0 {
		// Extract column names from the first record.
		for key := range records[0] {
			resp.Columns = append(resp.Columns, key)
		}
		// Ensure deterministic column order.
		sortStrings(resp.Columns)

		// Build rows using the column order.
		for _, record := range records {
			row := make([]string, len(resp.Columns))
			for i, col := range resp.Columns {
				if val := record[col]; val != nil {
					row[i] = val.String()
				}
			}
			resp.Rows = append(resp.Rows, row)
		}
	}
	resp.RowCount = int32(len(resp.Rows))

	return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostCallDBQueryResponse(resp))
}

// handleHostDBExecute processes OpcodeDBExecute requests. Only INSERT, UPDATE,
// DELETE, and REPLACE statements are allowed.
func handleHostDBExecute(ctx context.Context, hcc *hostCallContext, reqBytes []byte) *pluginbridge.HostCallResponseEnvelope {
	req, err := pluginbridge.UnmarshalHostCallDBExecuteRequest(reqBytes)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	sql := strings.TrimSpace(req.SQL)
	if sql == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, "SQL statement must not be empty")
	}

	// Validate statement type: only DML allowed.
	upper := strings.ToUpper(sql)
	if !strings.HasPrefix(upper, "INSERT") &&
		!strings.HasPrefix(upper, "UPDATE") &&
		!strings.HasPrefix(upper, "DELETE") &&
		!strings.HasPrefix(upper, "REPLACE") {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest,
			"host:db:execute only allows INSERT, UPDATE, DELETE, or REPLACE statements")
	}

	// Reject forbidden keywords.
	if keyword := containsForbiddenKeyword(sql); keyword != "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest,
			fmt.Sprintf("SQL statement contains forbidden keyword: %s", keyword))
	}

	args := buildDBArgs(req.Args)
	result, err := g.DB().Ctx(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}

	resp := &pluginbridge.HostCallDBExecuteResponse{}
	if result != nil {
		resp.RowsAffected, _ = result.RowsAffected()
		resp.LastInsertID, _ = result.LastInsertId()
	}

	return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostCallDBExecuteResponse(resp))
}

// containsForbiddenKeyword checks if the SQL contains any forbidden DDL keyword
// as a standalone word boundary. Returns the matched keyword or empty string.
func containsForbiddenKeyword(sql string) string {
	upper := strings.ToUpper(sql)
	for _, keyword := range dbForbiddenKeywords {
		// Check for the keyword as a standalone token by looking for word boundaries.
		idx := strings.Index(upper, keyword)
		for idx >= 0 {
			// Check left boundary: start of string or non-alphanumeric.
			leftOK := idx == 0 || !isAlphaNumeric(upper[idx-1])
			// Check right boundary: end of string or non-alphanumeric.
			rightIdx := idx + len(keyword)
			rightOK := rightIdx >= len(upper) || !isAlphaNumeric(upper[rightIdx])
			if leftOK && rightOK {
				return keyword
			}
			// Search for next occurrence.
			nextIdx := strings.Index(upper[idx+1:], keyword)
			if nextIdx < 0 {
				break
			}
			idx = idx + 1 + nextIdx
		}
	}
	return ""
}

// isAlphaNumeric checks if a byte is a letter, digit, or underscore.
func isAlphaNumeric(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '_'
}

// buildDBArgs converts string args to interface slice for parameterized queries.
func buildDBArgs(args []string) []interface{} {
	result := make([]interface{}, 0, len(args))
	for _, arg := range args {
		result = append(result, arg)
	}
	return result
}

// sortStrings sorts a string slice in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
