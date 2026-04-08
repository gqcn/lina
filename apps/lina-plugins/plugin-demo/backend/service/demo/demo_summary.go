package demo

import (
	"context"
)

const summaryMessage = "这是一条来自 plugin-demo 接口的简要介绍，用于验证插件页面可读取插件后端数据。"

// SummaryOutput defines one concise plugin summary payload.
type SummaryOutput struct {
	// Message is the concise page introduction returned from the plugin API.
	Message string
}

// Summary returns one concise plugin summary payload.
func (s *Service) Summary(ctx context.Context) (out *SummaryOutput, err error) {
	return &SummaryOutput{
		Message: summaryMessage,
	}, nil
}
