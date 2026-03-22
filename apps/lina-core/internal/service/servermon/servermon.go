package servermon

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
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

	"lina-core/internal/service/config"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// MonitorData represents all collected server metrics.
type MonitorData struct {
	Server  *ServerInfo  `json:"server"`
	CPU     *CPUInfo     `json:"cpu"`
	Memory  *MemoryInfo  `json:"memory"`
	Disks   []*DiskInfo  `json:"disks"`
	Network *NetworkInfo `json:"network"`
	GoInfo  *GoRuntimeInfo `json:"goInfo"`
}

type ServerInfo struct {
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	BootTime  string `json:"bootTime"`
	Uptime    uint64 `json:"uptime"`
	StartTime string `json:"startTime"`
}

type CPUInfo struct {
	Cores     int     `json:"cores"`
	ModelName string  `json:"modelName"`
	UsagePercent float64 `json:"usagePercent"`
}

type MemoryInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Available    uint64  `json:"available"`
	UsagePercent float64 `json:"usagePercent"`
}

type DiskInfo struct {
	Path         string  `json:"path"`
	FsType       string  `json:"fsType"`
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usagePercent"`
}

type NetworkInfo struct {
	BytesSent     uint64  `json:"bytesSent"`
	BytesRecv     uint64  `json:"bytesRecv"`
	SendRate      float64 `json:"sendRate"`
	RecvRate      float64 `json:"recvRate"`
}

type GoRuntimeInfo struct {
	Version    string `json:"version"`
	Goroutines int    `json:"goroutines"`
	HeapAlloc  uint64 `json:"heapAlloc"`
	HeapSys    uint64 `json:"heapSys"`
	GCPauseNs  uint64 `json:"gcPauseNs"`
	GfVersion  string `json:"gfVersion"`
}

// Service provides server monitoring operations.
type Service struct {
	configSvc     *config.Service
	startTime     time.Time
	lastNetBytes  *netutil.IOCountersStat
	lastCollectAt time.Time
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

func (s *Service) collectDisks() []*DiskInfo {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}
	var disks []*DiskInfo
	for _, p := range partitions {
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
	return &GoRuntimeInfo{
		Version:    runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		HeapAlloc:  m.HeapAlloc,
		HeapSys:    m.HeapSys,
		GCPauseNs:  m.PauseNs[(m.NumGC+255)%256],
		GfVersion:  "v2.10.0",
	}
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
	NodeName  string       `json:"nodeName"`
	NodeIp    string       `json:"nodeIp"`
	Data      *MonitorData `json:"data"`
	CollectAt string       `json:"collectAt"`
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
