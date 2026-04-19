package collector

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// HostInfo 主机基本信息
type HostInfo struct {
	Hostname    string
	OS          string
	Arch        string
	Uptime      string
	CPUModel    string
	CPUCores    int
	CPUPercent  float64
	LoadAvg     *load.AvgStat
	MemTotal    uint64
	MemUsed     uint64
	MemPercent  float64
	SwapTotal   uint64
	SwapUsed    uint64
	SwapPercent float64
	DiskParts   []DiskInfo
	NetIO       *NetIOInfo
}

// DiskInfo 磁盘分区信息
type DiskInfo struct {
	Device     string
	Mountpoint string
	Fstype     string
	Total      uint64
	Used       uint64
	Percent    float64
}

// NetIOInfo 网络流量
type NetIOInfo struct {
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64
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
			// 只显示有意义的分区 (>1GB) 且常见文件系统
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

	return h, nil
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
