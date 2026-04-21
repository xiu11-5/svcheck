# svcheck

轻量级服务巡检 CLI 工具，一键采集本机系统状态，输出带颜色标记的巡检报告。

## 功能

- **CPU** — 型号 / 核数 / 使用率 / 负载 (1m/5m/15m)
- **内存** — 物理内存 / Swap 使用率
- **磁盘** — 各分区使用率 (自动过滤 <1GB 的分区)
- **网络** — 累计收发流量 / 包数
- **状态判定** — OK(<70%) / WARN(70-90%) / CRIT(>=90%)

## 安装

### 方式一：go install (需要 Go 1.21+)

```bash
go install github.com/xiu11-5/svcheck@latest
```

### 方式二：下载预编译二进制

从 [Releases](https://github.com/xiu11-5/svcheck/releases) 下载对应平台的二进制：

```bash
# Linux amd64 示例
curl -sL https://github.com/xiu11-5/svcheck/releases/latest/download/svcheck-linux-amd64 -o svcheck
chmod +x svcheck
sudo mv svcheck /usr/local/bin/
```

### 方式三：从源码编译

```bash
git clone https://github.com/xiu11-5/svcheck.git
cd svcheck
go build -o svcheck .
sudo mv svcheck /usr/local/bin/
```

## 使用

```bash
# 单次巡检
svcheck

# 持续监控 (默认 5 秒刷新)
svcheck -w

# 自定义刷新间隔
svcheck -w -i 3
```

## 输出示例

```
╔════════════════════════════════════════════════════════════════╗
║  主机巡检报告                                                  ║
║  主机: jdyunvm | 系统: linux/amd64                             ║
║  时间: 2026-04-21 17:03:05                                     ║
╚════════════════════════════════════════════════════════════════╝

── CPU
  型号: Intel(R) Xeon(R) Gold 6148 CPU @ 2.40GHz | 核数: 2
  使用率: [#-----------------------------] 5.0%  OK
  负载:   0.07 / 0.15 / 0.16 (1m/5m/15m)

── 内存
  物理内存: [###############---------------] 50.8%  997.5M / 1.9G  OK
  Swap:     [##----------------------------] 9.0%  369.3M / 4.0G  OK

── 磁盘
  /dev/vda3 ext4   [#########-----------] 49.6%  18.3G / 39.0G  /  OK

── 网络
 累计发送: 5.8G (9062449 pkts) | 累计接收: 5.7G (8449776 pkts)

── 磁盘IO
 读取: 313.7G (9629203 次) | 写入: 83.0G (5127760 次)

── Top进程 (按CPU)
  1. svcheck                        PID:1592883 CPU:1.8% MEM:0.5%
  2. sshd                           PID:1592830 CPU:1.6% MEM:0.5%
  3. jdog-kunlunmirror              PID:148870 CPU:0.9% MEM:2.7%
  4. ifrit-agent                    PID:1380871 CPU:0.2% MEM:0.5%
  5. MonitorPlugin                  PID:1825 CPU:0.2% MEM:0.6%

── Top进程 (按内存)
  1. python                         PID:1422627 MEM:5.6% RSS:109.7M
  2. jdog-kunlunmirror              PID:148870 MEM:2.7% RSS:53.4M
  3. xray                           PID:3699 MEM:1.7% RSS:34.4M
  4. multipathd                     PID:1137306 MEM:1.3% RSS:26.5M
  5. systemd-journald               PID:1125610 MEM:0.8% RSS:15.9M

── 巡检结论
  [OK] 所有指标正常
```

## 交叉编译

```bash
GOOS=linux   GOARCH=amd64 go build -o svcheck-linux-amd64 .
GOOS=linux   GOARCH=arm64 go build -o svcheck-linux-arm64 .
GOOS=darwin  GOARCH=amd64 go build -o svcheck-darwin-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o svcheck-darwin-arm64 .
```

## License

MIT
