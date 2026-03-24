package sysinfo

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gogf/gf/v2"
	"github.com/gogf/gf/v2/frame/g"
)

// Service provides system information operations.
type Service struct {
	startTime time.Time // 服务启动时间
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		startTime: time.Now(),
	}
}

// SystemInfo holds the system runtime information.
type SystemInfo struct {
	GoVersion          string          // Go版本
	GfVersion          string          // GoFrame版本
	Os                 string          // 操作系统
	Arch               string          // 系统架构
	DbVersion          string          // 数据库版本
	StartTime          string          // 服务启动时间
	RunDuration        string          // 运行时长
	BackendComponents  []ComponentInfo // 后端组件列表
	FrontendComponents []ComponentInfo // 前端组件列表
}

// ComponentInfo holds component display information.
type ComponentInfo struct {
	Name        string // 组件名称
	Version     string // 组件版本
	Url         string // 组件链接
	Description string // 组件描述
}

// GetInfo returns system runtime information.
func (s *Service) GetInfo(ctx context.Context) (*SystemInfo, error) {
	info := &SystemInfo{
		GoVersion: runtime.Version(),
		GfVersion: gf.VERSION,
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		StartTime: s.startTime.Format("2006-01-02 15:04:05"),
	}

	// Calculate run duration
	duration := time.Since(s.startTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	if hours > 0 {
		info.RunDuration = fmt.Sprintf("%d小时%d分钟%d秒", hours, minutes, seconds)
	} else if minutes > 0 {
		info.RunDuration = fmt.Sprintf("%d分钟%d秒", minutes, seconds)
	} else {
		info.RunDuration = fmt.Sprintf("%d秒", seconds)
	}

	// Get database version
	dbVersion, err := s.getDbVersion(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "Failed to get database version: %v", err)
		info.DbVersion = "unknown"
	} else {
		info.DbVersion = dbVersion
	}

	// Load component info from config
	info.BackendComponents = s.loadComponents(ctx, "components.backend", dbVersion)
	info.FrontendComponents = s.loadComponents(ctx, "components.frontend", "")

	return info, nil
}

// loadComponents reads component configuration from config file.
func (s *Service) loadComponents(ctx context.Context, configKey string, dbVersion string) []ComponentInfo {
	cfg := g.Cfg()
	val, err := cfg.Get(ctx, configKey)
	if err != nil || val.IsEmpty() {
		return nil
	}

	var components []ComponentInfo
	if err = val.Scan(&components); err != nil {
		g.Log().Warningf(ctx, "Failed to scan components config '%s': %v", configKey, err)
		return nil
	}

	// Replace "auto" versions with runtime values
	for i := range components {
		if components[i].Version == "auto" {
			switch components[i].Name {
			case "GoFrame":
				components[i].Version = gf.VERSION
			case "MySQL":
				if dbVersion != "" {
					components[i].Version = dbVersion
				}
			}
		}
	}

	return components
}

// getDbVersion retrieves the database version.
func (s *Service) getDbVersion(ctx context.Context) (string, error) {
	result, err := g.DB().GetValue(ctx, "SELECT VERSION()")
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
