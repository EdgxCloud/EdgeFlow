package metrics

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Metrics ساختار metrics
type Metrics struct {
	// Flow metrics
	TotalFlows       int64 `json:"total_flows"`
	RunningFlows     int64 `json:"running_flows"`
	StoppedFlows     int64 `json:"stopped_flows"`
	TotalExecutions  int64 `json:"total_executions"`
	FailedExecutions int64 `json:"failed_executions"`

	// Node metrics
	TotalNodes        int64 `json:"total_nodes"`
	ActiveNodes       int64 `json:"active_nodes"`
	RegisteredNodeTypes int64 `json:"registered_node_types"`

	// System metrics
	Uptime           int64   `json:"uptime_seconds"`
	CPUUsage         float64 `json:"cpu_usage_percent"`
	MemoryUsed       uint64  `json:"memory_used_bytes"`
	MemoryTotal      uint64  `json:"memory_total_bytes"`
	GoroutineCount   int     `json:"goroutine_count"`

	// API metrics
	TotalRequests    int64 `json:"total_requests"`
	TotalErrors      int64 `json:"total_errors"`
	AvgResponseTime  float64 `json:"avg_response_time_ms"`

	mu sync.RWMutex
	startTime time.Time
}

// NewMetrics ایجاد Metrics
func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// IncrementFlows افزایش تعداد flows
func (m *Metrics) IncrementFlows() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalFlows++
}

// IncrementRunningFlows افزایش flows در حال اجرا
func (m *Metrics) IncrementRunningFlows() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RunningFlows++
}

// DecrementRunningFlows کاهش flows در حال اجرا
func (m *Metrics) DecrementRunningFlows() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.RunningFlows > 0 {
		m.RunningFlows--
	}
	m.StoppedFlows++
}

// IncrementExecutions افزایش تعداد اجراها
func (m *Metrics) IncrementExecutions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalExecutions++
}

// IncrementFailedExecutions افزایش اجراهای ناموفق
func (m *Metrics) IncrementFailedExecutions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailedExecutions++
}

// SetNodeMetrics تنظیم metrics نودها
func (m *Metrics) SetNodeMetrics(total, active, registered int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalNodes = total
	m.ActiveNodes = active
	m.RegisteredNodeTypes = registered
}

// IncrementRequests افزایش تعداد درخواست‌ها
func (m *Metrics) IncrementRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests++
}

// IncrementErrors افزایش تعداد خطاها
func (m *Metrics) IncrementErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalErrors++
}

// RecordResponseTime ثبت زمان پاسخ
func (m *Metrics) RecordResponseTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate moving average
	ms := float64(duration.Milliseconds())
	if m.AvgResponseTime == 0 {
		m.AvgResponseTime = ms
	} else {
		m.AvgResponseTime = (m.AvgResponseTime * 0.9) + (ms * 0.1)
	}
}

// UpdateSystemMetrics به‌روزرسانی metrics سیستم
func (m *Metrics) UpdateSystemMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Uptime
	m.Uptime = int64(time.Since(m.startTime).Seconds())

	// Memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.MemoryUsed = memStats.Alloc
	m.MemoryTotal = memStats.Sys

	// Goroutine count
	m.GoroutineCount = runtime.NumGoroutine()
}

// GetMetrics دریافت metrics
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"flows": map[string]interface{}{
			"total":   m.TotalFlows,
			"running": m.RunningFlows,
			"stopped": m.StoppedFlows,
		},
		"executions": map[string]interface{}{
			"total":  m.TotalExecutions,
			"failed": m.FailedExecutions,
			"success_rate": func() float64 {
				if m.TotalExecutions == 0 {
					return 100.0
				}
				return float64(m.TotalExecutions-m.FailedExecutions) / float64(m.TotalExecutions) * 100
			}(),
		},
		"nodes": map[string]interface{}{
			"total":            m.TotalNodes,
			"active":           m.ActiveNodes,
			"registered_types": m.RegisteredNodeTypes,
		},
		"system": map[string]interface{}{
			"uptime_seconds":     m.Uptime,
			"memory_used_bytes":  m.MemoryUsed,
			"memory_total_bytes": m.MemoryTotal,
			"memory_used_mb":     m.MemoryUsed / 1024 / 1024,
			"goroutines":         m.GoroutineCount,
		},
		"api": map[string]interface{}{
			"total_requests":      m.TotalRequests,
			"total_errors":        m.TotalErrors,
			"avg_response_time_ms": m.AvgResponseTime,
			"error_rate": func() float64 {
				if m.TotalRequests == 0 {
					return 0.0
				}
				return float64(m.TotalErrors) / float64(m.TotalRequests) * 100
			}(),
		},
	}
}

// PrometheusFormat تبدیل به فرمت Prometheus
func (m *Metrics) PrometheusFormat() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return `# HELP edgeflow_flows_total Total number of flows
# TYPE edgeflow_flows_total counter
edgeflow_flows_total ` + formatInt64(m.TotalFlows) + `

# HELP edgeflow_flows_running Number of running flows
# TYPE edgeflow_flows_running gauge
edgeflow_flows_running ` + formatInt64(m.RunningFlows) + `

# HELP edgeflow_executions_total Total number of executions
# TYPE edgeflow_executions_total counter
edgeflow_executions_total ` + formatInt64(m.TotalExecutions) + `

# HELP edgeflow_executions_failed Total number of failed executions
# TYPE edgeflow_executions_failed counter
edgeflow_executions_failed ` + formatInt64(m.FailedExecutions) + `

# HELP edgeflow_nodes_total Total number of nodes
# TYPE edgeflow_nodes_total gauge
edgeflow_nodes_total ` + formatInt64(m.TotalNodes) + `

# HELP edgeflow_nodes_active Number of active nodes
# TYPE edgeflow_nodes_active gauge
edgeflow_nodes_active ` + formatInt64(m.ActiveNodes) + `

# HELP edgeflow_uptime_seconds Uptime in seconds
# TYPE edgeflow_uptime_seconds gauge
edgeflow_uptime_seconds ` + formatInt64(m.Uptime) + `

# HELP edgeflow_memory_used_bytes Memory used in bytes
# TYPE edgeflow_memory_used_bytes gauge
edgeflow_memory_used_bytes ` + formatUint64(m.MemoryUsed) + `

# HELP edgeflow_goroutines Number of goroutines
# TYPE edgeflow_goroutines gauge
edgeflow_goroutines ` + formatInt(m.GoroutineCount) + `

# HELP edgeflow_api_requests_total Total number of API requests
# TYPE edgeflow_api_requests_total counter
edgeflow_api_requests_total ` + formatInt64(m.TotalRequests) + `

# HELP edgeflow_api_errors_total Total number of API errors
# TYPE edgeflow_api_errors_total counter
edgeflow_api_errors_total ` + formatInt64(m.TotalErrors) + `

# HELP edgeflow_api_response_time_ms Average API response time in milliseconds
# TYPE edgeflow_api_response_time_ms gauge
edgeflow_api_response_time_ms ` + formatFloat64(m.AvgResponseTime) + `
`
}

// MetricsMiddleware میدلور metrics
func MetricsMiddleware(m *Metrics) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Increment requests
		m.IncrementRequests()

		// Process request
		err := c.Next()

		// Record response time
		duration := time.Since(start)
		m.RecordResponseTime(duration)

		// Increment errors if status >= 400
		if c.Response().StatusCode() >= 400 {
			m.IncrementErrors()
		}

		return err
	}
}

// Helper functions
func formatInt64(n int64) string {
	return fmt.Sprintf("%d", n)
}

func formatUint64(n uint64) string {
	return fmt.Sprintf("%d", n)
}

func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}

func formatFloat64(n float64) string {
	return fmt.Sprintf("%.2f", n)
}
