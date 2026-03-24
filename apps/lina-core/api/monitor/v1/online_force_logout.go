package v1

import "github.com/gogf/gf/v2/frame/g"

// Online User Force Logout API

type OnlineForceLogoutReq struct {
	g.Meta  `path:"/monitor/online/{tokenId}" method:"delete" tags:"系统监控" summary:"强制下线" dc:"强制下线指定在线用户，被下线用户的后续请求将返回401"`
	TokenId string `json:"tokenId" v:"required#请指定会话ID" dc:"要强制下线的会话Token ID" eg:"abc123"`
}

// OnlineForceLogoutRes 强制下线响应
type OnlineForceLogoutRes struct{}
