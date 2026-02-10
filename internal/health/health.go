package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Check represents a check structure
type Check struct {
	Name        string                           `json:"name"`
	Status      Status                           `json:"status"`
	Message     string                           `json:"message"`
	LastCheck   time.Time                        `json:"last_check"`
	CheckFunc   func(context.Context) (Status, string) `json:"-"`
	Interval    time.Duration                    `json:"-"`
}

// HealthChecker is the health checker system
type HealthChecker struct {
	checks map[string]*Check
	mu     sync.RWMutex
}

// NewHealthChecker creates a new HealthChecker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]*Check),
	}
}

// RegisterCheck registers a health check
func (h *HealthChecker) RegisterCheck(name string, checkFunc func(context.Context) (Status, string), interval time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checks[name] = &Check{
		Name:      name,
		Status:    StatusHealthy,
		Message:   "Not checked yet",
		LastCheck: time.Time{},
		CheckFunc: checkFunc,
		Interval:  interval,
	}
}

// RunChecks runs all health checks
func (h *HealthChecker) RunChecks(ctx context.Context) map[string]*Check {
	h.mu.Lock()
	defer h.mu.Unlock()

	results := make(map[string]*Check)

	for name, check := range h.checks {
		// Run check
		status, message := check.CheckFunc(ctx)

		// Update check
		check.Status = status
		check.Message = message
		check.LastCheck = time.Now()

		// Add to results
		results[name] = &Check{
			Name:      check.Name,
			Status:    check.Status,
			Message:   check.Message,
			LastCheck: check.LastCheck,
		}
	}

	return results
}

// GetOverallStatus returns the overall status
func (h *HealthChecker) GetOverallStatus() Status {
	h.mu.RLock()
	defer h.mu.RUnlock()

	hasUnhealthy := false
	hasDegraded := false

	for _, check := range h.checks {
		switch check.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

// GetCheckResults returns check results
func (h *HealthChecker) GetCheckResults() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make(map[string]interface{})
	checks := make([]map[string]interface{}, 0, len(h.checks))

	for _, check := range h.checks {
		checks = append(checks, map[string]interface{}{
			"name":       check.Name,
			"status":     check.Status,
			"message":    check.Message,
			"last_check": check.LastCheck,
		})
	}

	results["status"] = h.GetOverallStatus()
	results["checks"] = checks
	results["timestamp"] = time.Now()

	return results
}

// StartPeriodicChecks starts periodic health checks
func (h *HealthChecker) StartPeriodicChecks(ctx context.Context) {
	h.mu.RLock()
	checks := make([]*Check, 0, len(h.checks))
	for _, check := range h.checks {
		checks = append(checks, check)
	}
	h.mu.RUnlock()

	// Start a goroutine for each check
	for _, check := range checks {
		check := check // Capture variable
		go func() {
			ticker := time.NewTicker(check.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					status, message := check.CheckFunc(ctx)

					h.mu.Lock()
					check.Status = status
					check.Message = message
					check.LastCheck = time.Now()
					h.mu.Unlock()
				}
			}
		}()
	}
}

// Common health checks

// DatabaseHealthCheck performs a database health check
func DatabaseHealthCheck(pingFunc func(context.Context) error) func(context.Context) (Status, string) {
	return func(ctx context.Context) (Status, string) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := pingFunc(ctx); err != nil {
			return StatusUnhealthy, "Database connection failed: " + err.Error()
		}
		return StatusHealthy, "Database is healthy"
	}
}

// DiskSpaceHealthCheck performs a disk space health check
func DiskSpaceHealthCheck(getUsageFunc func() (used, total uint64)) func(context.Context) (Status, string) {
	return func(ctx context.Context) (Status, string) {
		used, total := getUsageFunc()
		if total == 0 {
			return StatusUnhealthy, "Could not determine disk usage"
		}

		usagePercent := float64(used) / float64(total) * 100

		if usagePercent >= 95 {
			return StatusUnhealthy, fmt.Sprintf("Disk usage critical: %.1f%%", usagePercent)
		}
		if usagePercent >= 85 {
			return StatusDegraded, fmt.Sprintf("Disk usage high: %.1f%%", usagePercent)
		}
		return StatusHealthy, fmt.Sprintf("Disk usage normal: %.1f%%", usagePercent)
	}
}

// MemoryHealthCheck performs a memory health check
func MemoryHealthCheck(getMemoryFunc func() (used, total uint64)) func(context.Context) (Status, string) {
	return func(ctx context.Context) (Status, string) {
		used, total := getMemoryFunc()
		if total == 0 {
			return StatusDegraded, "Could not determine memory usage"
		}

		usagePercent := float64(used) / float64(total) * 100

		if usagePercent >= 90 {
			return StatusDegraded, fmt.Sprintf("Memory usage high: %.1f%%", usagePercent)
		}
		return StatusHealthy, fmt.Sprintf("Memory usage normal: %.1f%%", usagePercent)
	}
}

// GoroutineHealthCheck performs a goroutine count health check
func GoroutineHealthCheck(getCountFunc func() int, maxGoroutines int) func(context.Context) (Status, string) {
	return func(ctx context.Context) (Status, string) {
		count := getCountFunc()

		if count >= maxGoroutines {
			return StatusDegraded, fmt.Sprintf("High number of goroutines: %d", count)
		}
		return StatusHealthy, fmt.Sprintf("Goroutine count normal: %d", count)
	}
}
