//go:build !linux

package resources

import (
	"os"
	"runtime"
)

// SystemInfo اطلاعات سخت‌افزاری سیستم (غیرلینوکس)
type SystemInfo struct {
	Hostname    string  `json:"hostname"`
	OS          string  `json:"os"`
	Arch        string  `json:"arch"`
	Uptime      uint64  `json:"uptime_seconds"`
	Temperature float64 `json:"temperature"`
	LoadAvg1    float64 `json:"load_avg_1"`
	LoadAvg5    float64 `json:"load_avg_5"`
	LoadAvg15   float64 `json:"load_avg_15"`
	BoardModel  string  `json:"board_model"`

	OSMemTotal     uint64  `json:"os_mem_total"`
	OSMemUsed      uint64  `json:"os_mem_used"`
	OSMemFree      uint64  `json:"os_mem_free"`
	OSMemAvailable uint64  `json:"os_mem_available"`
	OSMemPercent   float64 `json:"os_mem_percent"`
	OSSwapTotal    uint64  `json:"os_swap_total"`
	OSSwapUsed     uint64  `json:"os_swap_used"`

	CPUUsagePercent float64 `json:"cpu_usage_percent"`

	NetRxBytes uint64 `json:"net_rx_bytes"`
	NetTxBytes uint64 `json:"net_tx_bytes"`
}

// GetSystemInfo دریافت اطلاعات سیستم (غیرلینوکس - مقادیر محدود)
func GetSystemInfo() SystemInfo {
	hostname, _ := os.Hostname()

	// استفاده از Go runtime برای اطلاعات حافظه
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return SystemInfo{
		Hostname:       hostname,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		Uptime:         0,
		Temperature:    0,
		BoardModel:     runtime.GOOS + "/" + runtime.GOARCH,
		OSMemTotal:     memStats.Sys,
		OSMemUsed:      memStats.Alloc,
		OSMemFree:      memStats.Sys - memStats.Alloc,
		OSMemAvailable: memStats.Sys - memStats.Alloc,
		OSMemPercent:   float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		CPUUsagePercent: 0,
	}
}

// GetDiskUsage اطلاعات دیسک (غیرلینوکس - مقادیر placeholder)
func GetDiskUsage(path string) DiskStats {
	// در ویندوز/مک بدون syscall اختصاصی
	return DiskStats{
		Total:     0,
		Used:      0,
		Available: 0,
		Percent:   0,
	}
}
