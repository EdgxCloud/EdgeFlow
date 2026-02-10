//go:build linux

package resources

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// SystemInfo holds system hardware information (Linux / Raspberry Pi)
type SystemInfo struct {
	Hostname    string  `json:"hostname"`
	OS          string  `json:"os"`
	Arch        string  `json:"arch"`
	Uptime      uint64  `json:"uptime_seconds"`
	Temperature float64 `json:"temperature"` // Celsius
	LoadAvg1    float64 `json:"load_avg_1"`
	LoadAvg5    float64 `json:"load_avg_5"`
	LoadAvg15   float64 `json:"load_avg_15"`
	BoardModel  string  `json:"board_model"`

	// OS memory (not just Go)
	OSMemTotal     uint64  `json:"os_mem_total"`
	OSMemUsed      uint64  `json:"os_mem_used"`
	OSMemFree      uint64  `json:"os_mem_free"`
	OSMemAvailable uint64  `json:"os_mem_available"`
	OSMemPercent   float64 `json:"os_mem_percent"`
	OSSwapTotal    uint64  `json:"os_swap_total"`
	OSSwapUsed     uint64  `json:"os_swap_used"`

	// CPU
	CPUUsagePercent float64 `json:"cpu_usage_percent"`

	// Network
	NetRxBytes uint64 `json:"net_rx_bytes"`
	NetTxBytes uint64 `json:"net_tx_bytes"`
}

// readProcFile reads a proc file and returns its trimmed content
func readProcFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// GetSystemInfo returns system information (Linux)
func GetSystemInfo() SystemInfo {
	info := SystemInfo{
		OS:   "linux",
		Arch: getArch(),
	}

	info.Hostname, _ = os.Hostname()

	// CPU temperature (Raspberry Pi)
	info.Temperature = getCPUTemperature()

	// Uptime
	info.Uptime = getUptime()

	// Load average
	info.LoadAvg1, info.LoadAvg5, info.LoadAvg15 = getLoadAvg()

	// Board model
	info.BoardModel = getBoardModel()

	// OS memory
	info.OSMemTotal, info.OSMemUsed, info.OSMemFree, info.OSMemAvailable, info.OSMemPercent,
		info.OSSwapTotal, info.OSSwapUsed = getOSMemory()

	// CPU usage
	info.CPUUsagePercent = getCPUUsage()

	// Network traffic
	info.NetRxBytes, info.NetTxBytes = getNetworkBytes()

	return info
}

// getCPUTemperature returns CPU temperature from thermal zone
func getCPUTemperature() float64 {
	content, err := readProcFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0
	}
	temp, err := strconv.ParseFloat(content, 64)
	if err != nil {
		return 0
	}
	return temp / 1000.0 // millidegree to degree
}

// getUptime returns system uptime
func getUptime() uint64 {
	content, err := readProcFile("/proc/uptime")
	if err != nil {
		return 0
	}
	parts := strings.Fields(content)
	if len(parts) < 1 {
		return 0
	}
	uptime, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	return uint64(uptime)
}

// getLoadAvg returns system load averages
func getLoadAvg() (float64, float64, float64) {
	content, err := readProcFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0
	}
	parts := strings.Fields(content)
	if len(parts) < 3 {
		return 0, 0, 0
	}
	l1, _ := strconv.ParseFloat(parts[0], 64)
	l5, _ := strconv.ParseFloat(parts[1], 64)
	l15, _ := strconv.ParseFloat(parts[2], 64)
	return l1, l5, l15
}

// getBoardModel returns the board model (Raspberry Pi)
func getBoardModel() string {
	content, err := readProcFile("/proc/device-tree/model")
	if err != nil {
		// fallback
		content, err = readProcFile("/sys/firmware/devicetree/base/model")
		if err != nil {
			return "Unknown"
		}
	}
	// Remove trailing null character
	return strings.TrimRight(content, "\x00")
}

// getOSMemory returns OS memory information from /proc/meminfo
func getOSMemory() (total, used, free, available uint64, percent float64, swapTotal, swapUsed uint64) {
	content, err := readProcFile("/proc/meminfo")
	if err != nil {
		return
	}

	memMap := make(map[string]uint64)
	for _, line := range strings.Split(content, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		val, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		// /proc/meminfo values are in kB
		memMap[key] = val * 1024
	}

	total = memMap["MemTotal"]
	free = memMap["MemFree"]
	available = memMap["MemAvailable"]
	if available == 0 {
		// If MemAvailable is missing (old kernel)
		available = free + memMap["Buffers"] + memMap["Cached"]
	}
	used = total - available
	if total > 0 {
		percent = float64(used) / float64(total) * 100
	}

	swapTotal = memMap["SwapTotal"]
	swapFree := memMap["SwapFree"]
	if swapTotal > swapFree {
		swapUsed = swapTotal - swapFree
	}

	return
}

// CPUTimes stores CPU state for usage calculation
var prevCPUIdle, prevCPUTotal uint64

// getCPUUsage calculates CPU usage percentage
func getCPUUsage() float64 {
	content, err := readProcFile("/proc/stat")
	if err != nil {
		return 0
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 1 {
		return 0
	}

	// First line: cpu  user nice system idle iowait irq softirq steal
	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0
	}

	var values []uint64
	for _, f := range fields[1:] {
		v, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			v = 0
		}
		values = append(values, v)
	}

	if len(values) < 4 {
		return 0
	}

	idle := values[3]
	if len(values) > 4 {
		idle += values[4] // iowait
	}

	var total uint64
	for _, v := range values {
		total += v
	}

	// Calculate difference from previous sample
	if prevCPUTotal == 0 {
		prevCPUIdle = idle
		prevCPUTotal = total
		return 0
	}

	diffIdle := idle - prevCPUIdle
	diffTotal := total - prevCPUTotal

	prevCPUIdle = idle
	prevCPUTotal = total

	if diffTotal == 0 {
		return 0
	}

	return (1.0 - float64(diffIdle)/float64(diffTotal)) * 100
}

// getNetworkBytes returns network traffic bytes
func getNetworkBytes() (rx, tx uint64) {
	content, err := readProcFile("/proc/net/dev")
	if err != nil {
		return 0, 0
	}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		// Skip header and loopback
		if strings.HasPrefix(line, "Inter") || strings.HasPrefix(line, "face") || strings.HasPrefix(line, "lo:") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 10 {
			continue
		}
		// Strip interface name
		rxVal, _ := strconv.ParseUint(parts[1], 10, 64)
		txVal, _ := strconv.ParseUint(parts[9], 10, 64)
		rx += rxVal
		tx += txVal
	}

	return
}

// GetDiskUsage returns actual disk usage information
func GetDiskUsage(path string) DiskStats {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskStats{}
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	percent := 0.0
	if total > 0 {
		percent = float64(used) / float64(total) * 100
	}

	return DiskStats{
		Total:     total,
		Used:      used,
		Available: free,
		Percent:   percent,
	}
}

// getArch returns the processor architecture
func getArch() string {
	content, err := readProcFile("/proc/cpuinfo")
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "model name") || strings.HasPrefix(line, "Hardware") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "arm"
}

func init() {
	// Initial CPU read to establish a baseline value
	getCPUUsage()
	time.Sleep(100 * time.Millisecond)
}
