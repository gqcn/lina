package servermon

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gtime"
	cpuutil "github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	netutil "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"lina-core/internal/service/config"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// MonitorData represents all collected server metrics.
type MonitorData struct {
	Server   *ServerInfo     `json:"server"`   // 服务器信息
	CPU      *CPUInfo        `json:"cpu"`       // CPU信息
	Memory   *MemoryInfo     `json:"memory"`    // 内存信息
	Disks    []*DiskInfo     `json:"disks"`     // 磁盘列表
	Network  *NetworkInfo    `json:"network"`   // 网络信息
	GoInfo   *GoRuntimeInfo  `json:"goInfo"`    // Go运行时信息
}

// ServerInfo represents server basic information.
type ServerInfo struct {
	Hostname  string `json:"hostname"`  // 主机名
	OS        string `json:"os"`        // 操作系统
	Arch      string `json:"arch"`      // 系统架构
	BootTime  string `json:"bootTime"`  // 系统启动时间
	Uptime    uint64 `json:"uptime"`    // 系统运行时长（秒）
	StartTime string `json:"startTime"` // 服务启动时间
}

// CPUInfo represents CPU metrics.
type CPUInfo struct {
	Cores        int     `json:"cores"`        // CPU核心数
	ModelName    string  `json:"modelName"`    // CPU型号
	UsagePercent float64 `json:"usagePercent"` // CPU使用率（百分比）
}

// MemoryInfo represents memory metrics.
type MemoryInfo struct {
	Total        uint64  `json:"total"`        // 总内存（字节）
	Used         uint64  `json:"used"`         // 已用内存（字节）
	Available    uint64  `json:"available"`    // 可用内存（字节）
	UsagePercent float64 `json:"usagePercent"` // 内存使用率（百分比）
}

// DiskInfo represents disk metrics.
type DiskInfo struct {
	Path         string  `json:"path"`         // 挂载路径
	FsType       string  `json:"fsType"`       // 文件系统类型
	Total        uint64  `json:"total"`        // 总空间（字节）
	Used         uint64  `json:"used"`         // 已用空间（字节）
	Free         uint64  `json:"free"`         // 可用空间（字节）
	UsagePercent float64 `json:"usagePercent"` // 使用率（百分比）
}

// NetworkInfo represents network metrics.
type NetworkInfo struct {
	BytesSent uint64  `json:"bytesSent"` // 发送字节数
	BytesRecv uint64  `json:"bytesRecv"` // 接收字节数
	SendRate  float64 `json:"sendRate"`  // 发送速率（字节/秒）
	RecvRate  float64 `json:"recvRate"`  // 接收速率（字节/秒）
}

// GoRuntimeInfo represents Go runtime metrics.
type GoRuntimeInfo struct {
	Version       string  `json:"version"`       // Go版本
	Goroutines    int     `json:"goroutines"`    // Goroutine数量
	ProcessCPU    float64 `json:"processCpu"`    // 进程CPU使用率
	ProcessMemory float64 `json:"processMemory"` // 进程内存使用率
	GCPauseNs     uint64  `json:"gcPauseNs"`     // GC暂停时间（纳秒）
	GfVersion     string  `json:"gfVersion"`     // GoFrame版本
	ServiceUptime string  `json:"serviceUptime"` // 服务运行时长
}

// DBInfo represents database metrics.
type DBInfo struct {
	Version      string `json:"version"`      // 数据库版本
	MaxOpenConns int    `json:"maxOpenConns"` // 最大连接数
	OpenConns    int    `json:"openConns"`    // 已打开连接数
	InUse        int    `json:"inUse"`        // 正在使用的连接数
	Idle         int    `json:"idle"`         // 空闲连接数
}

// Service provides server monitoring operations.
type Service struct {
	configSvc     *config.Service       // 配置服务
	startTime     time.Time             // 服务启动时间
	lastNetBytes  *netutil.IOCountersStat // 上次网络统计数据
	lastCollectAt time.Time             // 上次采集时间
}

// New creates a new Service.
func New() *Service {
	return &Service{
		configSvc: config.New(),
		startTime: time.Now(),
	}
}

// StartCollector starts the periodic metrics collector.
func (s *Service) StartCollector(ctx context.Context) {
	monCfg := s.configSvc.GetMonitor(ctx)

	// Collect immediately on startup
	s.collectAndStore(ctx)

	// Then collect periodically via gcron
	cronPattern := fmt.Sprintf("*/%d * * * * *", monCfg.IntervalSeconds)
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		s.collectAndStore(ctx)
	}, "server-monitor-collector")
	if err != nil {
		g.Log().Warningf(ctx, "failed to start server monitor cron: %v", err)
	}
}

// collectAndStore collects metrics and stores them in the database.
func (s *Service) collectAndStore(ctx context.Context) {
	data := s.Collect(ctx)
	jsonData, err := gjson.Encode(data)
	if err != nil {
		g.Log().Errorf(ctx, "Failed to encode monitor data: %v", err)
		return
	}

	nodeName, _ := os.Hostname()
	nodeIp := getLocalIP()

	_, err = dao.SysServerMonitor.Ctx(ctx).Data(do.SysServerMonitor{
		NodeName:  nodeName,
		NodeIp:    nodeIp,
		Data:      string(jsonData),
		CreatedAt: gtime.Now(),
	}).Save()
	if err != nil {
		g.Log().Errorf(ctx, "Failed to store monitor data: %v", err)
	}
}

// Collect gathers all server metrics.
func (s *Service) Collect(ctx context.Context) *MonitorData {
	data := &MonitorData{}
	data.Server = s.collectServer()
	data.CPU = s.collectCPU()
	data.Memory = s.collectMemory()
	data.Disks = s.collectDisks()
	data.Network = s.collectNetwork()
	data.GoInfo = s.collectGoRuntime()
	return data
}

func (s *Service) collectServer() *ServerInfo {
	hostname, _ := os.Hostname()
	info, _ := host.Info()
	bootTime := ""
	var uptime uint64
	if info != nil {
		bootTime = time.Unix(int64(info.BootTime), 0).Format("2006-01-02 15:04:05")
		uptime = info.Uptime
	}
	return &ServerInfo{
		Hostname:  hostname,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		BootTime:  bootTime,
		Uptime:    uptime,
		StartTime: s.startTime.Format("2006-01-02 15:04:05"),
	}
}

func (s *Service) collectCPU() *CPUInfo {
	info := &CPUInfo{}
	info.Cores = runtime.NumCPU()
	cpuInfos, err := cpuutil.Info()
	if err == nil && len(cpuInfos) > 0 {
		info.ModelName = cpuInfos[0].ModelName
	}
	percents, err := cpuutil.Percent(time.Second, false)
	if err == nil && len(percents) > 0 {
		info.UsagePercent = percents[0]
	}
	return info
}

func (s *Service) collectMemory() *MemoryInfo {
	v, err := mem.VirtualMemory()
	if err != nil {
		return &MemoryInfo{}
	}
	return &MemoryInfo{
		Total:        v.Total,
		Used:         v.Used,
		Available:    v.Available,
		UsagePercent: v.UsedPercent,
	}
}

// virtualFsTypes lists filesystem types to exclude (common in containers).
var virtualFsTypes = map[string]bool{
	"overlay":   true,
	"tmpfs":     true,
	"devtmpfs":  true,
	"devfs":     true,
	"proc":      true,
	"sysfs":     true,
	"cgroup":    true,
	"cgroup2":   true,
	"squashfs":  true,
	"aufs":      true,
	"shm":       true,
	"nsfs":      true,
	"fuse":      true,
}

func (s *Service) collectDisks() []*DiskInfo {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}
	var disks []*DiskInfo
	for _, p := range partitions {
		// Skip virtual/pseudo filesystems
		if virtualFsTypes[p.Fstype] {
			continue
		}
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}
		disks = append(disks, &DiskInfo{
			Path:         p.Mountpoint,
			FsType:       p.Fstype,
			Total:        usage.Total,
			Used:         usage.Used,
			Free:         usage.Free,
			UsagePercent: usage.UsedPercent,
		})
	}
	return disks
}

func (s *Service) collectNetwork() *NetworkInfo {
	counters, err := netutil.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return &NetworkInfo{}
	}
	current := &counters[0]
	info := &NetworkInfo{
		BytesSent: current.BytesSent,
		BytesRecv: current.BytesRecv,
	}

	// Calculate rate from previous sample
	if s.lastNetBytes != nil && !s.lastCollectAt.IsZero() {
		elapsed := time.Since(s.lastCollectAt).Seconds()
		if elapsed > 0 {
			info.SendRate = float64(current.BytesSent-s.lastNetBytes.BytesSent) / elapsed
			info.RecvRate = float64(current.BytesRecv-s.lastNetBytes.BytesRecv) / elapsed
		}
	}
	s.lastNetBytes = current
	s.lastCollectAt = time.Now()
	return info
}

func (s *Service) collectGoRuntime() *GoRuntimeInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info := &GoRuntimeInfo{
		Version:    runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		GCPauseNs:  m.PauseNs[(m.NumGC+255)%256],
		GfVersion:  "v2.10.0",
	}

	// Collect process CPU and memory usage
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err == nil {
		if cpuPercent, err := proc.CPUPercent(); err == nil {
			info.ProcessCPU = cpuPercent
		}
		if memPercent, err := proc.MemoryPercent(); err == nil {
			info.ProcessMemory = float64(memPercent)
		}
	}

	// Calculate service uptime
	duration := time.Since(s.startTime)
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	mins := int(duration.Minutes()) % 60
	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d天", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d小时", hours))
	}
	if mins > 0 {
		parts = append(parts, fmt.Sprintf("%d分钟", mins))
	}
	if len(parts) == 0 {
		info.ServiceUptime = "刚启动"
	} else {
		info.ServiceUptime = strings.Join(parts, " ")
	}

	return info
}

// GetDBInfo collects database metrics on-demand.
func (s *Service) GetDBInfo(ctx context.Context) *DBInfo {
	info := &DBInfo{}

	// Get database version
	result, err := g.DB().GetValue(ctx, "SELECT VERSION()")
	if err == nil {
		info.Version = result.String()
	}

	// Get connection pool stats
	statsItems := g.DB().GetCore().Stats(ctx)
	if len(statsItems) > 0 {
		stats := statsItems[0].Stats()
		info.MaxOpenConns = stats.MaxOpenConnections
		info.OpenConns = stats.OpenConnections
		info.InUse = stats.InUse
		info.Idle = stats.Idle
	}

	return info
}

// GetLatest returns the latest monitor records for each node.
func (s *Service) GetLatest(ctx context.Context, nodeName string) ([]*NodeMonitorData, error) {
	cols := dao.SysServerMonitor.Columns()
	m := dao.SysServerMonitor.Ctx(ctx)
	if nodeName != "" {
		m = m.Where(cols.NodeName, nodeName)
	}

	// Get distinct node names
	var allRecords []*entity.SysServerMonitor
	err := m.OrderDesc(cols.CreatedAt).Scan(&allRecords)
	if err != nil {
		return nil, err
	}

	// Group by node, keep latest for each
	seen := make(map[string]bool)
	var result []*NodeMonitorData
	for _, record := range allRecords {
		key := record.NodeName + "|" + record.NodeIp
		if seen[key] {
			continue
		}
		seen[key] = true

		var data MonitorData
		if err := gjson.DecodeTo([]byte(record.Data), &data); err != nil {
			continue
		}
		result = append(result, &NodeMonitorData{
			NodeName:  record.NodeName,
			NodeIp:    record.NodeIp,
			Data:      &data,
			CollectAt: record.CreatedAt.Format("Y-m-d H:i:s"),
		})
	}
	return result, nil
}

// NodeMonitorData wraps monitor data with node info.
type NodeMonitorData struct {
	NodeName  string       `json:"nodeName"`  // 节点名称
	NodeIp    string       `json:"nodeIp"`    // 节点IP
	Data      *MonitorData `json:"data"`      // 监控数据
	CollectAt string       `json:"collectAt"` // 采集时间
}

// getLocalIP returns the first non-loopback IPv4 address.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return "unknown"
}
