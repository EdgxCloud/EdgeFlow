package resources

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

// ResourceStats نمایش وضعیت منابع سیستم
type ResourceStats struct {
	MemoryTotal     uint64  `json:"memory_total"`
	MemoryUsed      uint64  `json:"memory_used"`
	MemoryAvailable uint64  `json:"memory_available"`
	MemoryPercent   float64 `json:"memory_percent"`
	DiskTotal       uint64  `json:"disk_total"`
	DiskUsed        uint64  `json:"disk_used"`
	DiskAvailable   uint64  `json:"disk_available"`
	DiskPercent     float64 `json:"disk_percent"`
	CPUCores        int     `json:"cpu_cores"`
	GoroutineCount  int     `json:"goroutine_count"`
	Timestamp       time.Time `json:"timestamp"`

	// اطلاعات سخت‌افزاری سیستم
	SysInfo SystemInfo `json:"sys_info"`
}

// ResourceLimits محدودیت‌های منابع
type ResourceLimits struct {
	MemoryLimit           uint64 `json:"memory_limit"`
	MemoryHardLimit       uint64 `json:"memory_hard_limit"`
	DiskLimit             uint64 `json:"disk_limit"`
	LowMemoryThreshold    uint64 `json:"low_memory_threshold"`
	AutoDisableOnLowMemory bool  `json:"auto_disable_on_low_memory"`
}

// Monitor سیستم نظارت بر منابع
type Monitor struct {
	limits      ResourceLimits
	currentStats ResourceStats
	mu          sync.RWMutex

	// Callbacks برای اقدامات خودکار
	onLowMemory     func()
	onHighMemory    func()
	onDiskFull      func()

	// ماژول‌های فعال
	enabledModules  map[string]bool
	modulesMu       sync.RWMutex
}

// NewMonitor ایجاد نمونه جدید مانیتور
func NewMonitor(limits ResourceLimits) *Monitor {
	return &Monitor{
		limits:         limits,
		enabledModules: make(map[string]bool),
	}
}

// Start شروع نظارت دوره‌ای
func (m *Monitor) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.Update()
			m.checkLimits()
		}
	}
}

// Update به‌روزرسانی آمار فعلی
func (m *Monitor) Update() {
	stats := m.getSystemStats()

	m.mu.Lock()
	m.currentStats = stats
	m.mu.Unlock()
}

// GetStats دریافت آمار فعلی
func (m *Monitor) GetStats() ResourceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentStats
}

// getSystemStats دریافت آمار سیستم
func (m *Monitor) getSystemStats() ResourceStats {
	// دریافت اطلاعات سخت‌افزاری سیستم (پلتفرم-اختصاصی)
	sysInfo := GetSystemInfo()

	stats := ResourceStats{
		Timestamp:      time.Now(),
		CPUCores:       runtime.NumCPU(),
		GoroutineCount: runtime.NumGoroutine(),
		SysInfo:        sysInfo,
	}

	// حافظه سیستم‌عامل (واقعی)
	if sysInfo.OSMemTotal > 0 {
		stats.MemoryTotal = sysInfo.OSMemTotal
		stats.MemoryUsed = sysInfo.OSMemUsed
		stats.MemoryAvailable = sysInfo.OSMemAvailable
		stats.MemoryPercent = sysInfo.OSMemPercent
	} else {
		// فال‌بک به حافظه Go
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		stats.MemoryUsed = memStats.Alloc
		stats.MemoryTotal = memStats.Sys
		if stats.MemoryTotal > 0 {
			stats.MemoryPercent = float64(stats.MemoryUsed) / float64(stats.MemoryTotal) * 100
		}
	}

	// دیسک واقعی
	diskStats := GetDiskUsage("/")
	if diskStats.Total > 0 {
		stats.DiskTotal = diskStats.Total
		stats.DiskUsed = diskStats.Used
		stats.DiskAvailable = diskStats.Available
		stats.DiskPercent = diskStats.Percent
	}

	return stats
}

// DiskStats آمار دیسک
type DiskStats struct {
	Total     uint64
	Used      uint64
	Available uint64
	Percent   float64
}


// checkLimits بررسی محدودیت‌ها و اقدام خودکار
func (m *Monitor) checkLimits() {
	stats := m.GetStats()

	// بررسی حافظه کم
	if m.limits.AutoDisableOnLowMemory && stats.MemoryAvailable < m.limits.LowMemoryThreshold {
		log.Printf("[WARN] Low memory detected: %dMB available (threshold: %dMB)",
			stats.MemoryAvailable/1024/1024,
			m.limits.LowMemoryThreshold/1024/1024)

		if m.onLowMemory != nil {
			m.onLowMemory()
		} else {
			m.autoDisableNonEssentialModules()
		}
	}

	// بررسی حافظه بالا (برگشت به حالت عادی)
	if stats.MemoryAvailable > m.limits.LowMemoryThreshold*2 {
		if m.onHighMemory != nil {
			m.onHighMemory()
		}
	}

	// بررسی دیسک پر
	if stats.DiskPercent > 95 {
		log.Printf("[WARN] Disk nearly full: %.1f%% used", stats.DiskPercent)
		if m.onDiskFull != nil {
			m.onDiskFull()
		}
	}

	// بررسی محدودیت حافظه سخت
	if m.limits.MemoryHardLimit > 0 && stats.MemoryUsed > m.limits.MemoryHardLimit {
		log.Printf("[CRITICAL] Hard memory limit exceeded: %dMB used (limit: %dMB)",
			stats.MemoryUsed/1024/1024,
			m.limits.MemoryHardLimit/1024/1024)

		// Force garbage collection
		runtime.GC()
	}
}

// autoDisableNonEssentialModules غیرفعال‌سازی خودکار ماژول‌های غیرضروری
func (m *Monitor) autoDisableNonEssentialModules() {
	stats := m.GetStats()
	availableMB := stats.MemoryAvailable / 1024 / 1024

	log.Printf("[ACTION] Auto-disabling non-essential modules (available: %dMB)", availableMB)

	// اولویت غیرفعال‌سازی (از بزرگ‌ترین به کوچک‌ترین)
	priorityOrder := []string{
		"collaboration", // ~40MB
		"ui_advanced",   // ~30MB
		"ai",            // ~30MB
		"advanced",      // ~50MB
		"industrial",    // ~20MB
		"database",      // ~15MB
		"messaging",     // ~10MB
	}

	for _, module := range priorityOrder {
		if m.IsModuleEnabled(module) {
			m.DisableModule(module)
			log.Printf("[ACTION] Disabled module: %s", module)

			// Force GC after disabling
			runtime.GC()

			// بررسی مجدد
			newStats := m.getSystemStats()
			if newStats.MemoryAvailable >= m.limits.LowMemoryThreshold {
				log.Printf("[ACTION] Memory recovered: %dMB available", newStats.MemoryAvailable/1024/1024)
				return
			}
		}
	}
}

// CanLoadModule بررسی امکان بارگذاری ماژول
func (m *Monitor) CanLoadModule(moduleName string, requiredMemory uint64) (bool, string) {
	stats := m.GetStats()

	// بررسی حافظه کافی
	if stats.MemoryAvailable < requiredMemory {
		return false, fmt.Sprintf(
			"insufficient memory: need %dMB, have %dMB",
			requiredMemory/1024/1024,
			stats.MemoryAvailable/1024/1024,
		)
	}

	// بررسی محدودیت کلی
	if m.limits.MemoryLimit > 0 {
		projectedUsage := stats.MemoryUsed + requiredMemory
		if projectedUsage > m.limits.MemoryLimit {
			return false, fmt.Sprintf(
				"would exceed memory limit: projected %dMB, limit %dMB",
				projectedUsage/1024/1024,
				m.limits.MemoryLimit/1024/1024,
			)
		}
	}

	return true, ""
}

// EnableModule فعال‌سازی ماژول
func (m *Monitor) EnableModule(moduleName string) {
	m.modulesMu.Lock()
	defer m.modulesMu.Unlock()
	m.enabledModules[moduleName] = true
	log.Printf("[RESOURCE] Module enabled: %s", moduleName)
}

// DisableModule غیرفعال‌سازی ماژول
func (m *Monitor) DisableModule(moduleName string) {
	m.modulesMu.Lock()
	defer m.modulesMu.Unlock()
	m.enabledModules[moduleName] = false
	log.Printf("[RESOURCE] Module disabled: %s", moduleName)
}

// IsModuleEnabled بررسی فعال بودن ماژول
func (m *Monitor) IsModuleEnabled(moduleName string) bool {
	m.modulesMu.RLock()
	defer m.modulesMu.RUnlock()
	enabled, exists := m.enabledModules[moduleName]
	return exists && enabled
}

// GetEnabledModules دریافت لیست ماژول‌های فعال
func (m *Monitor) GetEnabledModules() []string {
	m.modulesMu.RLock()
	defer m.modulesMu.RUnlock()

	modules := make([]string, 0, len(m.enabledModules))
	for name, enabled := range m.enabledModules {
		if enabled {
			modules = append(modules, name)
		}
	}
	return modules
}

// SetOnLowMemory تنظیم کالبک حافظه کم
func (m *Monitor) SetOnLowMemory(callback func()) {
	m.onLowMemory = callback
}

// SetOnHighMemory تنظیم کالبک حافظه بالا
func (m *Monitor) SetOnHighMemory(callback func()) {
	m.onHighMemory = callback
}

// SetOnDiskFull تنظیم کالبک دیسک پر
func (m *Monitor) SetOnDiskFull(callback func()) {
	m.onDiskFull = callback
}

// ForceGC اجرای اجباری garbage collection
func (m *Monitor) ForceGC() {
	before := m.GetStats()

	runtime.GC()

	time.Sleep(100 * time.Millisecond)
	after := m.getSystemStats()

	freed := int64(before.MemoryUsed) - int64(after.MemoryUsed)
	log.Printf("[GC] Garbage collection: freed %dMB", freed/1024/1024)
}

// GetMemoryProfile دریافت پروفایل حافظه
func (m *Monitor) GetMemoryProfile() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"alloc_mb":         memStats.Alloc / 1024 / 1024,
		"total_alloc_mb":   memStats.TotalAlloc / 1024 / 1024,
		"sys_mb":           memStats.Sys / 1024 / 1024,
		"num_gc":           memStats.NumGC,
		"gc_cpu_fraction":  memStats.GCCPUFraction,
		"heap_alloc_mb":    memStats.HeapAlloc / 1024 / 1024,
		"heap_sys_mb":      memStats.HeapSys / 1024 / 1024,
		"heap_objects":     memStats.HeapObjects,
		"goroutines":       runtime.NumGoroutine(),
	}
}

// GetResourceReport دریافت گزارش کامل منابع
func (m *Monitor) GetResourceReport() map[string]interface{} {
	stats := m.GetStats()

	return map[string]interface{}{
		"timestamp": stats.Timestamp,
		"memory": map[string]interface{}{
			"total_mb":     stats.MemoryTotal / 1024 / 1024,
			"used_mb":      stats.MemoryUsed / 1024 / 1024,
			"available_mb": stats.MemoryAvailable / 1024 / 1024,
			"percent":      fmt.Sprintf("%.1f%%", stats.MemoryPercent),
		},
		"disk": map[string]interface{}{
			"total_mb":     stats.DiskTotal / 1024 / 1024,
			"used_mb":      stats.DiskUsed / 1024 / 1024,
			"available_mb": stats.DiskAvailable / 1024 / 1024,
			"percent":      fmt.Sprintf("%.1f%%", stats.DiskPercent),
		},
		"cpu": map[string]interface{}{
			"cores":      stats.CPUCores,
			"goroutines": stats.GoroutineCount,
		},
		"limits": map[string]interface{}{
			"memory_limit_mb":      m.limits.MemoryLimit / 1024 / 1024,
			"memory_hard_limit_mb": m.limits.MemoryHardLimit / 1024 / 1024,
			"low_memory_threshold_mb": m.limits.LowMemoryThreshold / 1024 / 1024,
		},
		"modules": map[string]interface{}{
			"enabled": m.GetEnabledModules(),
		},
	}
}
