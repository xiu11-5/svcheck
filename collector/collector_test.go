package collector

import (
	"encoding/json"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1024 * 1024, "1.0M"},
		{1024*1024*1024, "1.0G"},
		{uint64(1024*1024) * 1024 * 1024, "1.0T"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.input)
		if got != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestHostInfoToJSON(t *testing.T) {
	h := &HostInfo{
		Hostname:   "testhost",
		OS:         "linux",
		Arch:       "amd64",
		CPUPercent: 50.5,
		MemTotal:   1024 * 1024 * 1024,
		MemUsed:    512 * 1024 * 1024,
		MemPercent: 50.0,
	}

	data, err := h.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// 验证能解析为有效 JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// 验证关键字段
	if parsed["hostname"] != "testhost" {
		t.Errorf("hostname = %v, want testhost", parsed["hostname"])
	}
	if parsed["os"] != "linux" {
		t.Errorf("os = %v, want linux", parsed["os"])
	}
	if parsed["cpu_percent"] != 50.5 {
		t.Errorf("cpu_percent = %v, want 50.5", parsed["cpu_percent"])
	}
}

func TestCollect(t *testing.T) {
	h, err := Collect()
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	// 验证基本字段
	if h.OS == "" {
		t.Error("OS should not be empty")
	}
	if h.Arch == "" {
		t.Error("Arch should not be empty")
	}
	if h.CPUPercent < 0 || h.CPUPercent > 100 {
		t.Errorf("CPUPercent = %v, should be between 0 and 100", h.CPUPercent)
	}
	if h.MemPercent < 0 || h.MemPercent > 100 {
		t.Errorf("MemPercent = %v, should be between 0 and 100", h.MemPercent)
	}
	if h.MemTotal == 0 {
		t.Error("MemTotal should not be zero")
	}
}

func TestDiskInfoStruct(t *testing.T) {
	d := DiskInfo{
		Device:     "/dev/sda1",
		Mountpoint: "/",
		Fstype:     "ext4",
		Total:      100 * 1024 * 1024 * 1024,
		Used:       50 * 1024 * 1024 * 1024,
		Percent:    50.0,
	}

	if d.Device != "/dev/sda1" {
		t.Errorf("Device = %s, want /dev/sda1", d.Device)
	}
	if d.Percent != 50.0 {
		t.Errorf("Percent = %v, want 50.0", d.Percent)
	}
}

func TestNetIOInfoStruct(t *testing.T) {
	n := NetIOInfo{
		BytesSent:   1000,
		BytesRecv:   2000,
		PacketsSent: 10,
		PacketsRecv: 20,
	}

	if n.BytesSent != 1000 {
		t.Errorf("BytesSent = %d, want 1000", n.BytesSent)
	}
	if n.BytesRecv != 2000 {
		t.Errorf("BytesRecv = %d, want 2000", n.BytesRecv)
	}
}

func TestProcessInfoStruct(t *testing.T) {
	p := ProcessInfo{
		Pid:        1234,
		Name:       "testproc",
		CPUPercent: 10.5,
		MemPercent: 5.5,
		MemRSS:     1024 * 1024,
	}

	if p.Pid != 1234 {
		t.Errorf("Pid = %d, want 1234", p.Pid)
	}
	if p.Name != "testproc" {
		t.Errorf("Name = %s, want testproc", p.Name)
	}
	if p.CPUPercent != 10.5 {
		t.Errorf("CPUPercent = %v, want 10.5", p.CPUPercent)
	}
}