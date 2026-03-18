package sysinfo

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Service provides system information operations.
type Service struct {
	startTime time.Time
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		startTime: time.Now(),
	}
}

// SystemInfo holds the system runtime information.
type SystemInfo struct {
	GoVersion   string
	GfVersion   string
	Os          string
	Arch        string
	DbVersion   string
	StartTime   string
	RunDuration string
}

// GetInfo returns system runtime information.
func (s *Service) GetInfo(ctx context.Context) (*SystemInfo, error) {
	info := &SystemInfo{
		GoVersion: runtime.Version(),
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

	// Get GoFrame version
	info.GfVersion = "v2.10.0"

	// Get database version
	dbVersion, err := s.getDbVersion(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "Failed to get database version: %v", err)
		info.DbVersion = "unknown"
	} else {
		info.DbVersion = dbVersion
	}

	return info, nil
}

// getDbVersion retrieves the database version.
func (s *Service) getDbVersion(ctx context.Context) (string, error) {
	result, err := g.DB().GetValue(ctx, "SELECT VERSION()")
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
