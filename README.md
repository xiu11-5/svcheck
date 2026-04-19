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
╔══════════════════════════════════════════╗
║ 主机巡检报告
║ 主机: jdyunvm | 系统: linux/amd64
║ 时间: 2026-04-19 16:58:42
╚══════════════════════════════════════════╝

── CPU
  型号: Intel(R) Xeon(R) Gold 6148 CPU @ 2.40GHz | 核数: 2
  使用率: [#-----------------------------] 5.0%  OK
  负载:   0.04 / 0.14 / 0.17 (1m/5m/15m)

── 内存
  物理内存: [##################------------] 62.5%  1.2G / 1.9G  OK
  Swap:     未配置

── 磁盘
  /dev/vda3 ext4 [###########---------] 59.4%  22.0G / 39.0G  /  OK

── 网络
  累计发送: 3.4G (5504710 pkts)  |  累计接收: 2.8G (4845392 pkts)

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
