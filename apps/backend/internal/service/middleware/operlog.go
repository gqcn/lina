package middleware

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/grpool"

	"backend/internal/service/operlog"
)

const maxParamLen = 2000

// OperLog records operation logs for write operations and specially tagged GET operations.
func (s *Service) OperLog(r *ghttp.Request) {
	startTime := time.Now()
	r.Middleware.Next()

	// Collect all data synchronously (r.Response buffer is only available now)
	method := r.Method
	handler := r.GetServeHandler()
	operLogTag := ""
	if handler != nil {
		operLogTag = handler.GetMetaTag("operLog")
	}

	// Only log write operations (POST/PUT/DELETE) or GET with operLog tag
	shouldLog := false
	switch method {
	case "POST", "PUT", "DELETE":
		shouldLog = true
	case "GET":
		shouldLog = operLogTag != ""
	}
	if !shouldLog {
		return
	}

	title := ""
	if handler != nil {
		title = handler.GetMetaTag("tags")
	}

	operSummary := ""
	if handler != nil {
		operSummary = handler.GetMetaTag("summary")
	}

	operType := inferOperType(method, r.URL.Path, operLogTag)

	operName := ""
	if bizCtx := s.bizCtxSvc.Get(r.Context()); bizCtx != nil {
		operName = bizCtx.Username
	}

	// Get request parameters (skip binary content like file uploads)
	operParam := ""
	reqContentType := r.GetHeader("Content-Type")
	if isBinaryContentType(reqContentType) {
		operParam = "[二进制内容]"
	} else {
		operParam = truncate(maskPassword(getRequestParam(r)), maxParamLen)
	}

	// Get response result (skip binary content like xlsx exports)
	jsonResult := ""
	resContentType := r.Response.Header().Get("Content-Type")
	if isBinaryContentType(resContentType) {
		jsonResult = "[二进制内容]"
	} else {
		jsonResult = truncate(r.Response.BufferString(), maxParamLen)
	}

	status := 0
	errorMsg := ""
	if r.Response.Status >= 400 || r.GetError() != nil {
		status = 1
		if r.GetError() != nil {
			errorMsg = r.GetError().Error()
		}
	}

	costTime := int(time.Since(startTime).Milliseconds())
	urlPath := r.URL.Path
	urlString := r.URL.String()
	clientIp := r.GetClientIp()

	// Async write using grpool (goroutine pool) with NeverDoneCtx
	ctx := r.GetNeverDoneCtx()
	_ = grpool.AddWithRecover(ctx, func(ctx context.Context) {
		_ = s.operLogSvc.Create(ctx, operlog.CreateInput{
			Title:         title,
			OperSummary:   operSummary,
			OperType:      operType,
			Method:        urlPath,
			RequestMethod: method,
			OperName:      operName,
			OperUrl:       urlString,
			OperIp:        clientIp,
			OperParam:     operParam,
			JsonResult:    jsonResult,
			Status:        status,
			ErrorMsg:      errorMsg,
			CostTime:      costTime,
		})
	}, func(ctx context.Context, err error) {
		g.Log().Errorf(ctx, "operlog middleware panic: %v", err)
	})
}

// inferOperType determines operation type from HTTP method and path.
func inferOperType(method, path, operLogTag string) int {
	if operLogTag != "" {
		switch operLogTag {
		case "1":
			return 1
		case "2":
			return 2
		case "3":
			return 3
		case "4":
			return 4
		case "5":
			return 5
		default:
			return 6
		}
	}

	switch method {
	case "POST":
		if strings.Contains(strings.ToLower(path), "import") {
			return 5 // Import
		}
		return 1 // Create
	case "PUT":
		return 2 // Update
	case "DELETE":
		return 3 // Delete
	default:
		return 6 // Other
	}
}

// getRequestParam extracts request parameters as JSON string.
func getRequestParam(r *ghttp.Request) string {
	body := r.GetBodyString()
	if body != "" {
		return body
	}
	params := r.GetQueryMap()
	if len(params) > 0 {
		b, _ := json.Marshal(params)
		return string(b)
	}
	return ""
}

// maskPassword replaces password field values with ***.
func maskPassword(param string) string {
	if param == "" {
		return param
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(param), &data); err != nil {
		return param
	}
	masked := false
	for k := range data {
		lower := strings.ToLower(k)
		if lower == "password" || lower == "newpassword" || lower == "oldpassword" {
			data[k] = "***"
			masked = true
		}
	}
	if !masked {
		return param
	}
	b, _ := json.Marshal(data)
	return string(b)
}

// isBinaryContentType checks if the content type represents binary data.
func isBinaryContentType(contentType string) bool {
	if contentType == "" {
		return false
	}
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "multipart/form-data") ||
		strings.Contains(ct, "application/octet-stream") ||
		strings.Contains(ct, "spreadsheetml") ||
		strings.Contains(ct, "image/") ||
		strings.Contains(ct, "audio/") ||
		strings.Contains(ct, "video/")
}

// truncate truncates a string to maxLen and appends suffix if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}
