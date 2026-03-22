package v1

import "github.com/gogf/gf/v2/frame/g"

// Server Monitor API

type ServerMonitorReq struct {
	g.Meta   `path:"/monitor/server" method:"get" tags:"系统监控" summary:"服务监控" dc:"查询服务器监控数据，返回各节点最新的CPU、内存、磁盘、网络、Go运行时等指标信息"`
	NodeName string `json:"nodeName" dc:"按节点名称过滤，不传则返回所有节点" eg:"my-server"`
}

type ServerMonitorRes struct {
	Nodes []*ServerNodeInfo `json:"nodes" dc:"各节点监控数据"`
}

type ServerNodeInfo struct {
	NodeName  string         `json:"nodeName" dc:"节点名称（hostname）" eg:"my-server"`
	NodeIp    string         `json:"nodeIp" dc:"节点IP地址" eg:"192.168.1.100"`
	CollectAt string         `json:"collectAt" dc:"采集时间" eg:"2025-01-01 12:00:00"`
	Server    *ServerBasic   `json:"server" dc:"服务器基本信息"`
	CPU       *CPUMetrics    `json:"cpu" dc:"CPU指标"`
	Memory    *MemoryMetrics `json:"memory" dc:"内存指标"`
	Disks     []*DiskMetrics `json:"disks" dc:"磁盘使用情况"`
	Network   *NetMetrics    `json:"network" dc:"网络流量指标"`
	GoInfo    *GoMetrics     `json:"goInfo" dc:"Go运行时指标"`
}

type ServerBasic struct {
	Hostname  string `json:"hostname" dc:"主机名" eg:"my-server"`
	OS        string `json:"os" dc:"操作系统" eg:"linux"`
	Arch      string `json:"arch" dc:"系统架构" eg:"amd64"`
	BootTime  string `json:"bootTime" dc:"系统启动时间" eg:"2025-01-01 00:00:00"`
	Uptime    uint64 `json:"uptime" dc:"系统运行时长（秒）" eg:"86400"`
	StartTime string `json:"startTime" dc:"服务启动时间" eg:"2025-01-01 08:00:00"`
}

type CPUMetrics struct {
	Cores        int     `json:"cores" dc:"CPU核心数" eg:"8"`
	ModelName    string  `json:"modelName" dc:"CPU型号" eg:"Intel Core i7-12700"`
	UsagePercent float64 `json:"usagePercent" dc:"CPU使用率（百分比）" eg:"45.5"`
}

type MemoryMetrics struct {
	Total        uint64  `json:"total" dc:"总内存（字节）" eg:"17179869184"`
	Used         uint64  `json:"used" dc:"已用内存（字节）" eg:"8589934592"`
	Available    uint64  `json:"available" dc:"可用内存（字节）" eg:"8589934592"`
	UsagePercent float64 `json:"usagePercent" dc:"内存使用率（百分比）" eg:"50.0"`
}

type DiskMetrics struct {
	Path         string  `json:"path" dc:"挂载点路径" eg:"/"`
	FsType       string  `json:"fsType" dc:"文件系统类型" eg:"ext4"`
	Total        uint64  `json:"total" dc:"总容量（字节）" eg:"107374182400"`
	Used         uint64  `json:"used" dc:"已用容量（字节）" eg:"53687091200"`
	Free         uint64  `json:"free" dc:"可用容量（字节）" eg:"53687091200"`
	UsagePercent float64 `json:"usagePercent" dc:"使用率（百分比）" eg:"50.0"`
}

type NetMetrics struct {
	BytesSent uint64  `json:"bytesSent" dc:"总发送字节数" eg:"1073741824"`
	BytesRecv uint64  `json:"bytesRecv" dc:"总接收字节数" eg:"2147483648"`
	SendRate  float64 `json:"sendRate" dc:"发送速率（字节/秒）" eg:"102400"`
	RecvRate  float64 `json:"recvRate" dc:"接收速率（字节/秒）" eg:"204800"`
}

type GoMetrics struct {
	Version    string `json:"version" dc:"Go版本" eg:"go1.22.0"`
	Goroutines int    `json:"goroutines" dc:"Goroutine数量" eg:"42"`
	HeapAlloc  uint64 `json:"heapAlloc" dc:"堆内存分配量（字节）" eg:"10485760"`
	HeapSys    uint64 `json:"heapSys" dc:"堆内存系统分配（字节）" eg:"20971520"`
	GCPauseNs  uint64 `json:"gcPauseNs" dc:"最近一次GC暂停时间（纳秒）" eg:"150000"`
	GfVersion  string `json:"gfVersion" dc:"GoFrame版本" eg:"v2.10.0"`
}
