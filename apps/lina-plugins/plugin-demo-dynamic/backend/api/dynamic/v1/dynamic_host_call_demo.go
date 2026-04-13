package v1

import "github.com/gogf/gf/v2/frame/g"

// HostCallDemoReq is the request for invoking the host call demo endpoint.
type HostCallDemoReq struct {
	g.Meta `path:"/host-call-demo" method:"get" tags:"动态插件示例" summary:"宿主调用能力演示" dc:"演示动态插件通过 Host Functions 调用宿主能力，包括结构化日志输出和插件状态存储的读写操作，每次调用自增访问计数并返回当前值" access:"login" permission:"plugin-demo-dynamic:backend:view" operLog:"other"`
}

// HostCallDemoRes is the response for the host call demo endpoint.
type HostCallDemoRes struct {
	VisitCount int    `json:"visitCount" dc:"当前累计访问次数，通过宿主状态存储 host:state 实现持久化计数" eg:"1"`
	PluginID   string `json:"pluginId" dc:"当前插件唯一标识" eg:"plugin-demo-dynamic"`
	Message    string `json:"message" dc:"宿主调用演示说明信息" eg:"Host call demo: log written, state incremented."`
}
