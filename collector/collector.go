package collector

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// HostInfo 主机基本信息
type HostInfo struct {
	Hostname    string         `json:"hostname"`
	OS          string         `json:"os"`
	Arch        string         `json:"arch"`
	Uptime      string         `json:"uptime"`
	CPUModel    string         `json:"cpu_model"`
	CPUCores    int            `json:"cpu_cores"`
	CPUPercent  float64        `json:"cpu_percent"`
	LoadAvg     *load.AvgStat  `json:"load_avg,omitempty"`
	MemTotal    uint64         `json:"mem_total"`
	MemUsed     uint64         `json:"mem_used"`
	MemPercent  float64        `json:"mem_percent"`
	SwapTotal   uint64         `json:"swap_total"`
	SwapUsed    uint64         `json:"swap_used"`
	SwapPercent float64        `json:"swap_percent"`
	DiskParts   []DiskInfo     `json:"disk_parts"`
	NetIO       *NetIOInfo     `json:"net_io,omitempty"`
	DiskIO      *DiskIOInfo    `json:"disk_io,omitempty"`
	TopCPU      []ProcessInfo  `json:"top_cpu,omitempty"`
	TopMem      []ProcessInfo  `json:"top_mem,omitempty"`
}

// DiskInfo 磁盘分区信息
type DiskInfo struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	Fstype     string  `json:"fstype"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Percent    float64 `json:"percent"`
}

// NetIOInfo 网络流量
type NetIOInfo struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// DiskIOInfo 磁盘IO统计
type DiskIOInfo struct {
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadCount  uint64 `json:"read_count"`
	WriteCount uint64 `json:"write_count"`
	ReadTime   uint64 `json:"read_time"`
	WriteTime  uint64 `json:"write_time"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	Pid       int32   `json:"pid"`
	Name      string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float64 `json:"mem_percent"`
	MemRSS     uint64  `json:"mem_rss"`
}

// Collect 采集本机所有指标
func Collect() (*HostInfo, error) {
	h := &HostInfo{}

	// 基本信息
	h.OS = runtime.GOOS
	h.Arch = runtime.GOARCH

	// CPU 使用率 (1秒采样)
	percents, err := cpu.Percent(1*time.Second, false)
	if err == nil && len(percents) > 0 {
		h.CPUPercent = percents[0]
	}

	// CPU 信息
	cpuInfos, err := cpu.Info()
	if err == nil && len(cpuInfos) > 0 {
		h.CPUModel = cpuInfos[0].ModelName
	}
	cores, err := cpu.Counts(true)
	if err == nil {
		h.CPUCores = cores
	}

	// 负载
	avg, err := load.Avg()
	if err == nil {
		h.LoadAvg = avg
	}

	// 内存
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		h.MemTotal = vmStat.Total
		h.MemUsed = vmStat.Used
		h.MemPercent = vmStat.UsedPercent
	}
	swStat, err := mem.SwapMemory()
	if err == nil {
		h.SwapTotal = swStat.Total
		h.SwapUsed = swStat.Used
		h.SwapPercent = swStat.UsedPercent
	}

	// 磁盘
	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil || usage == nil {
				continue
			}
			// 只显示有意义的分区 (>1GB)
			if usage.Total < 1<<30 {
				continue
			}
			h.DiskParts = append(h.DiskParts, DiskInfo{
				Device:     p.Device,
				Mountpoint: p.Mountpoint,
				Fstype:     p.Fstype,
				Total:      usage.Total,
				Used:       usage.Used,
				Percent:    usage.UsedPercent,
			})
		}
	}

	// 网络 IO
	counters, err := net.IOCounters(false)
	if err == nil && len(counters) > 0 {
		c := counters[0]
		h.NetIO = &NetIOInfo{
			BytesSent:   c.BytesSent,
			BytesRecv:   c.BytesRecv,
			PacketsSent: c.PacketsSent,
			PacketsRecv: c.PacketsRecv,
		}
	}

	// 磁盘 IO
	ioCounters, err := disk.IOCounters()
	if err == nil && len(ioCounters) > 0 {
		var totalReadBytes, totalWriteBytes uint64
		var totalReadCount, totalWriteCount uint64
		var totalReadTime, totalWriteTime uint64

		for _, io := range ioCounters {
			totalReadBytes += io.ReadBytes
			totalWriteBytes += io.WriteBytes
			totalReadCount += io.ReadCount
			totalWriteCount += io.WriteCount
			totalReadTime += io.ReadTime
			totalWriteTime += io.WriteTime
		}
		h.DiskIO = &DiskIOInfo{
			ReadBytes:  totalReadBytes,
			WriteBytes: totalWriteBytes,
			ReadCount:  totalReadCount,
			WriteCount: totalWriteCount,
			ReadTime:   totalReadTime,
			WriteTime:  totalWriteTime,
		}
	}

	// Top 进程
	h.TopCPU, h.TopMem = getTopProcesses(5)

	return h, nil
}

// getTopProcesses 获取CPU和内存占用最高的进程
func getTopProcesses(n int) (topCPU, topMem []ProcessInfo) {
	processes, err := process.Processes()
	if err != nil {
		return nil, nil
	}

	var cpuPercent, memPercent float64
	var memRSS uint64
	var name string

	for _, p := range processes {
		name, _ = p.Name()
		cpuPercent, _ = p.CPUPercent()
		memPercent32, _ := p.MemoryPercent()
		memPercent = float64(memPercent32)
		memInfo, _ := p.MemoryInfo()
		if memInfo != nil {
			memRSS = memInfo.RSS
		}

		proc := ProcessInfo{
			Pid:        p.Pid,
			Name:       name,
			CPUPercent: cpuPercent,
			MemPercent: float64(memPercent),
			MemRSS:     memRSS,
		}

		topCPU = append(topCPU, proc)
		topMem = append(topMem, proc)
	}

	// 按CPU排序
	sort.Slice(topCPU, func(i, j int) bool {
		return topCPU[i].CPUPercent > topCPU[j].CPUPercent
	})
	if len(topCPU) > n {
		topCPU = topCPU[:n]
	}

	// 按内存排序
	sort.Slice(topMem, func(i, j int) bool {
		return topMem[i].MemPercent > topMem[j].MemPercent
	})
	if len(topMem) > n {
		topMem = topMem[:n]
	}

	return topCPU, topMem
}

// ToJSON 输出JSON格式
func (h *HostInfo) ToJSON() ([]byte, error) {
	return json.MarshalIndent(h, "", "  ")
}

// FormatBytes 格式化字节数
func FormatBytes(b uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case b >= TB:
		return fmt.Sprintf("%.1fT", float64(b)/float64(TB))
	case b >= GB:
		return fmt.Sprintf("%.1fG", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1fM", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1fK", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%dB", b)
	}
}