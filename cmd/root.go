package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/spf13/cobra"
	"github.com/xiu11-5/svcheck/collector"
)

var (
	jsonOutput   bool
	watchMode    bool
	watchSeconds int
)

// 状态阈值
const (
	warnThreshold  = 70.0
	critThreshold  = 90.0
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)

// statusLabel 根据百分比返回带颜色的状态标签
func statusLabel(pct float64) string {
	switch {
	case pct >= critThreshold:
		return red("CRIT")
	case pct >= warnThreshold:
		return yellow("WARN")
	default:
		return green("OK  ")
	}
}

// bar 返回进度条字符串
func bar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "#"
		} else {
			bar += "-"
		}
	}
	return bar
}

var rootCmd = &cobra.Command{
	Use:   "svcheck",
	Short: "服务巡检工具 - 一键采集系统状态",
	Long:  `svcheck 采集本机 CPU/内存/磁盘/网络/负载指标，输出巡检报告。`,
	Run: func(cmd *cobra.Command, args []string) {
		if watchMode {
			runWatch()
		} else {
			runOnce()
		}
	},
}

func runOnce() {
	info, err := collector.Collect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "采集失败: %v\n", err)
		os.Exit(1)
	}
	printReport(info)
}

func runWatch() {
	for {
		info, err := collector.Collect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "采集失败: %v\n", err)
			os.Exit(1)
		}
		// 清屏
		fmt.Print("\033[2J\033[H")
		printReport(info)
		fmt.Printf("\n下次刷新: %ds 后 (Ctrl+C 退出)\n", watchSeconds)
		time.Sleep(time.Duration(watchSeconds) * time.Second)
	}
}

func printReport(h *collector.HostInfo) {
	// 主机头
	hostname, _ := host.HostID()
	hn, _ := host.Info()
	if hn != nil {
		hostname = hn.Hostname
	}

	fmt.Printf("\n%s %s\n", bold("╔══════════════════════════════════════════╗"), bold("══════════════════════════════════════════╗"))
	fmt.Printf("%s %-40s %s %s\n", bold("║"), cyan("主机巡检报告"), "", bold("║"))
	fmt.Printf("%s %-40s %s %s\n", bold("║"), fmt.Sprintf("主机: %s | 系统: %s/%s", hostname, h.OS, h.Arch), "", bold("║"))
	fmt.Printf("%s %-40s %s %s\n", bold("║"), fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")), "", bold("║"))
	fmt.Printf("%s\n", bold("╚══════════════════════════════════════════╝"))

	// CPU
	fmt.Printf("\n%s CPU\n", bold("──"))
	if h.CPUModel != "" {
		fmt.Printf("  型号: %s | 核数: %d\n", h.CPUModel, h.CPUCores)
	}
	fmt.Printf("  使用率: [%s] %.1f%%  %s\n", bar(h.CPUPercent, 30), h.CPUPercent, statusLabel(h.CPUPercent))
	if h.LoadAvg != nil {
		fmt.Printf("  负载:   %.2f / %.2f / %.2f (1m/5m/15m)\n", h.LoadAvg.Load1, h.LoadAvg.Load5, h.LoadAvg.Load15)
	}

	// 内存
	fmt.Printf("\n%s 内存\n", bold("──"))
	fmt.Printf("  物理内存: [%s] %.1f%%  %s / %s  %s\n",
		bar(h.MemPercent, 30), h.MemPercent,
		collector.FormatBytes(h.MemUsed), collector.FormatBytes(h.MemTotal),
		statusLabel(h.MemPercent))
	if h.SwapTotal > 0 {
		fmt.Printf("  Swap:     [%s] %.1f%%  %s / %s  %s\n",
			bar(h.SwapPercent, 30), h.SwapPercent,
			collector.FormatBytes(h.SwapUsed), collector.FormatBytes(h.SwapTotal),
			statusLabel(h.SwapPercent))
	} else {
		fmt.Printf("  Swap:     未配置\n")
	}

	// 磁盘
	fmt.Printf("\n%s 磁盘\n", bold("──"))
	if len(h.DiskParts) == 0 {
		fmt.Printf("  无可用分区信息\n")
	}
	for _, d := range h.DiskParts {
		fmt.Printf("  %-8s %-6s [%s] %.1f%%  %s / %s  %s  %s\n",
			d.Device, d.Fstype,
			bar(d.Percent, 20), d.Percent,
			collector.FormatBytes(d.Used), collector.FormatBytes(d.Total),
			d.Mountpoint,
			statusLabel(d.Percent))
	}

	// 网络
	fmt.Printf("\n%s 网络\n", bold("──"))
	if h.NetIO != nil {
		fmt.Printf("  累计发送: %s (%d pkts)  |  累计接收: %s (%d pkts)\n",
			collector.FormatBytes(h.NetIO.BytesSent), h.NetIO.PacketsSent,
			collector.FormatBytes(h.NetIO.BytesRecv), h.NetIO.PacketsRecv)
	}

	// 汇总
	fmt.Printf("\n%s 巡检结论\n", bold("──"))
	issues := 0
	if h.CPUPercent >= critThreshold {
		fmt.Printf("  %s CPU 使用率 %.1f%% 超过临界值 %.0f%%\n", red("[!!]"), h.CPUPercent, critThreshold)
		issues++
	} else if h.CPUPercent >= warnThreshold {
		fmt.Printf("  %s CPU 使用率 %.1f%% 超过警告值 %.0f%%\n", yellow("[!]"), h.CPUPercent, warnThreshold)
		issues++
	}
	if h.MemPercent >= critThreshold {
		fmt.Printf("  %s 内存使用率 %.1f%% 超过临界值 %.0f%%\n", red("[!!]"), h.MemPercent, critThreshold)
		issues++
	} else if h.MemPercent >= warnThreshold {
		fmt.Printf("  %s 内存使用率 %.1f%% 超过警告值 %.0f%%\n", yellow("[!]"), h.MemPercent, warnThreshold)
		issues++
	}
	for _, d := range h.DiskParts {
		if d.Percent >= critThreshold {
			fmt.Printf("  %s 磁盘 %s (%s) 使用率 %.1f%% 超过临界值 %.0f%%\n", red("[!!]"), d.Mountpoint, d.Device, d.Percent, critThreshold)
			issues++
		} else if d.Percent >= warnThreshold {
			fmt.Printf("  %s 磁盘 %s (%s) 使用率 %.1f%% 超过警告值 %.0f%%\n", yellow("[!]"), d.Mountpoint, d.Device, d.Percent, warnThreshold)
			issues++
		}
	}
	if issues == 0 {
		fmt.Printf("  %s 所有指标正常\n", green("[OK]"))
	}
	fmt.Println()
}

func init() {
	rootCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "输出 JSON 格式 (TODO)")
	rootCmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "持续监控模式")
	rootCmd.Flags().IntVarP(&watchSeconds, "interval", "i", 5, "监控刷新间隔(秒)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
