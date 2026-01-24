package health

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthChecker(t *testing.T) {
	checker := NewHealthChecker()
	assert.NotNil(t, checker)
	assert.NotNil(t, checker.checks)
	assert.Empty(t, checker.checks)
}

func TestHealthChecker_RegisterCheck(t *testing.T) {
	checker := NewHealthChecker()

	checkFunc := func(ctx context.Context) (Status, string) {
		return StatusHealthy, "OK"
	}

	checker.RegisterCheck("test-check", checkFunc, 30*time.Second)

	// Verify check was registered
	assert.Len(t, checker.checks, 1)
	assert.Contains(t, checker.checks, "test-check")

	check := checker.checks["test-check"]
	assert.Equal(t, "test-check", check.Name)
	assert.Equal(t, StatusHealthy, check.Status)
	assert.Equal(t, "Not checked yet", check.Message)
	assert.Equal(t, 30*time.Second, check.Interval)
}

func TestHealthChecker_RegisterMultipleChecks(t *testing.T) {
	checker := NewHealthChecker()

	checks := []struct {
		name     string
		interval time.Duration
	}{
		{"database", 30 * time.Second},
		{"disk", 60 * time.Second},
		{"memory", 10 * time.Second},
		{"goroutines", 5 * time.Second},
	}

	for _, c := range checks {
		checker.RegisterCheck(c.name, func(ctx context.Context) (Status, string) {
			return StatusHealthy, "OK"
		}, c.interval)
	}

	assert.Len(t, checker.checks, 4)
	for _, c := range checks {
		assert.Contains(t, checker.checks, c.name)
	}
}

func TestHealthChecker_RunChecks(t *testing.T) {
	checker := NewHealthChecker()

	// Register checks with different statuses
	checker.RegisterCheck("healthy-check", func(ctx context.Context) (Status, string) {
		return StatusHealthy, "All good"
	}, time.Minute)

	checker.RegisterCheck("degraded-check", func(ctx context.Context) (Status, string) {
		return StatusDegraded, "Some issues"
	}, time.Minute)

	checker.RegisterCheck("unhealthy-check", func(ctx context.Context) (Status, string) {
		return StatusUnhealthy, "Critical error"
	}, time.Minute)

	ctx := context.Background()
	results := checker.RunChecks(ctx)

	assert.Len(t, results, 3)

	// Verify each check result
	assert.Equal(t, StatusHealthy, results["healthy-check"].Status)
	assert.Equal(t, "All good", results["healthy-check"].Message)

	assert.Equal(t, StatusDegraded, results["degraded-check"].Status)
	assert.Equal(t, "Some issues", results["degraded-check"].Message)

	assert.Equal(t, StatusUnhealthy, results["unhealthy-check"].Status)
	assert.Equal(t, "Critical error", results["unhealthy-check"].Message)

	// Verify LastCheck was updated
	for _, result := range results {
		assert.False(t, result.LastCheck.IsZero())
		assert.WithinDuration(t, time.Now(), result.LastCheck, time.Second)
	}
}

func TestHealthChecker_GetOverallStatus_AllHealthy(t *testing.T) {
	checker := NewHealthChecker()

	checker.RegisterCheck("check1", func(ctx context.Context) (Status, string) {
		return StatusHealthy, "OK"
	}, time.Minute)

	checker.RegisterCheck("check2", func(ctx context.Context) (Status, string) {
		return StatusHealthy, "OK"
	}, time.Minute)

	// Run checks to update statuses
	checker.RunChecks(context.Background())

	assert.Equal(t, StatusHealthy, checker.GetOverallStatus())
}

func TestHealthChecker_GetOverallStatus_WithDegraded(t *testing.T) {
	checker := NewHealthChecker()

	checker.RegisterCheck("healthy-check", func(ctx context.Context) (Status, string) {
		return StatusHealthy, "OK"
	}, time.Minute)

	checker.RegisterCheck("degraded-check", func(ctx context.Context) (Status, string) {
		return StatusDegraded, "Warning"
	}, time.Minute)

	// Run checks to update statuses
	checker.RunChecks(context.Background())

	assert.Equal(t, StatusDegraded, checker.GetOverallStatus())
}

func TestHealthChecker_GetOverallStatus_WithUnhealthy(t *testing.T) {
	checker := NewHealthChecker()

	checker.RegisterCheck("healthy-check", func(ctx context.Context) (Status, string) {
		return StatusHealthy, "OK"
	}, time.Minute)

	checker.RegisterCheck("degraded-check", func(ctx context.Context) (Status, string) {
		return StatusDegraded, "Warning"
	}, time.Minute)

	checker.RegisterCheck("unhealthy-check", func(ctx context.Context) (Status, string) {
		return StatusUnhealthy, "Critical"
	}, time.Minute)

	// Run checks to update statuses
	checker.RunChecks(context.Background())

	// Unhealthy takes precedence
	assert.Equal(t, StatusUnhealthy, checker.GetOverallStatus())
}

func TestHealthChecker_GetCheckResults(t *testing.T) {
	checker := NewHealthChecker()

	checker.RegisterCheck("test-check", func(ctx context.Context) (Status, string) {
		return StatusHealthy, "All good"
	}, time.Minute)

	// Run checks
	checker.RunChecks(context.Background())

	results := checker.GetCheckResults()

	assert.Equal(t, StatusHealthy, results["status"])
	assert.NotNil(t, results["checks"])
	assert.NotNil(t, results["timestamp"])

	checks := results["checks"].([]map[string]interface{})
	assert.Len(t, checks, 1)
	assert.Equal(t, "test-check", checks[0]["name"])
	assert.Equal(t, StatusHealthy, checks[0]["status"])
	assert.Equal(t, "All good", checks[0]["message"])
}

func TestHealthChecker_ConcurrentAccess(t *testing.T) {
	checker := NewHealthChecker()

	// Register a check
	checker.RegisterCheck("concurrent-check", func(ctx context.Context) (Status, string) {
		return StatusHealthy, "OK"
	}, time.Minute)

	// Run concurrent operations
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			checker.RunChecks(context.Background())
		}()

		go func() {
			defer wg.Done()
			checker.GetOverallStatus()
		}()

		go func() {
			defer wg.Done()
			checker.GetCheckResults()
		}()
	}

	wg.Wait()
	// Test passes if no race conditions occurred
}

func TestDatabaseHealthCheck_Healthy(t *testing.T) {
	pingFunc := func(ctx context.Context) error {
		return nil
	}

	checkFunc := DatabaseHealthCheck(pingFunc)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusHealthy, status)
	assert.Equal(t, "Database is healthy", message)
}

func TestDatabaseHealthCheck_Unhealthy(t *testing.T) {
	pingFunc := func(ctx context.Context) error {
		return errors.New("connection refused")
	}

	checkFunc := DatabaseHealthCheck(pingFunc)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusUnhealthy, status)
	assert.Contains(t, message, "Database connection failed")
	assert.Contains(t, message, "connection refused")
}

func TestDatabaseHealthCheck_Timeout(t *testing.T) {
	pingFunc := func(ctx context.Context) error {
		// Simulate slow response
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			return nil
		}
	}

	checkFunc := DatabaseHealthCheck(pingFunc)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	status, message := checkFunc(ctx)

	// The internal timeout is 5 seconds, but we pass a shorter context
	// This should result in context deadline exceeded
	assert.Equal(t, StatusUnhealthy, status)
	assert.Contains(t, message, "Database connection failed")
}

func TestDiskSpaceHealthCheck_Healthy(t *testing.T) {
	getUsageFunc := func() (used, total uint64) {
		return 50 * 1024 * 1024 * 1024, 100 * 1024 * 1024 * 1024 // 50GB used of 100GB
	}

	checkFunc := DiskSpaceHealthCheck(getUsageFunc)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusHealthy, status)
	assert.Contains(t, message, "normal")
	assert.Contains(t, message, "50.0%")
}

func TestDiskSpaceHealthCheck_Degraded(t *testing.T) {
	getUsageFunc := func() (used, total uint64) {
		return 90 * 1024 * 1024 * 1024, 100 * 1024 * 1024 * 1024 // 90GB used of 100GB
	}

	checkFunc := DiskSpaceHealthCheck(getUsageFunc)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusDegraded, status)
	assert.Contains(t, message, "high")
	assert.Contains(t, message, "90.0%")
}

func TestDiskSpaceHealthCheck_Unhealthy(t *testing.T) {
	getUsageFunc := func() (used, total uint64) {
		return 98 * 1024 * 1024 * 1024, 100 * 1024 * 1024 * 1024 // 98GB used of 100GB
	}

	checkFunc := DiskSpaceHealthCheck(getUsageFunc)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusUnhealthy, status)
	assert.Contains(t, message, "critical")
	assert.Contains(t, message, "98.0%")
}

func TestDiskSpaceHealthCheck_ZeroTotal(t *testing.T) {
	getUsageFunc := func() (used, total uint64) {
		return 0, 0
	}

	checkFunc := DiskSpaceHealthCheck(getUsageFunc)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusUnhealthy, status)
	assert.Contains(t, message, "Could not determine")
}

func TestMemoryHealthCheck_Healthy(t *testing.T) {
	getMemoryFunc := func() (used, total uint64) {
		return 4 * 1024 * 1024 * 1024, 16 * 1024 * 1024 * 1024 // 4GB used of 16GB
	}

	checkFunc := MemoryHealthCheck(getMemoryFunc)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusHealthy, status)
	assert.Contains(t, message, "normal")
	assert.Contains(t, message, "25.0%")
}

func TestMemoryHealthCheck_Degraded(t *testing.T) {
	getMemoryFunc := func() (used, total uint64) {
		return 15 * 1024 * 1024 * 1024, 16 * 1024 * 1024 * 1024 // 15GB used of 16GB
	}

	checkFunc := MemoryHealthCheck(getMemoryFunc)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusDegraded, status)
	assert.Contains(t, message, "high")
	assert.Contains(t, message, "93.7%")
}

func TestMemoryHealthCheck_ZeroTotal(t *testing.T) {
	getMemoryFunc := func() (used, total uint64) {
		return 0, 0
	}

	checkFunc := MemoryHealthCheck(getMemoryFunc)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusDegraded, status)
	assert.Contains(t, message, "Could not determine")
}

func TestGoroutineHealthCheck_Healthy(t *testing.T) {
	getCountFunc := func() int {
		return 50
	}

	checkFunc := GoroutineHealthCheck(getCountFunc, 1000)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusHealthy, status)
	assert.Contains(t, message, "normal")
	assert.Contains(t, message, "50")
}

func TestGoroutineHealthCheck_Degraded(t *testing.T) {
	getCountFunc := func() int {
		return 1500
	}

	checkFunc := GoroutineHealthCheck(getCountFunc, 1000)
	status, message := checkFunc(context.Background())

	assert.Equal(t, StatusDegraded, status)
	assert.Contains(t, message, "High")
	assert.Contains(t, message, "1500")
}

func TestStatus_Values(t *testing.T) {
	assert.Equal(t, Status("healthy"), StatusHealthy)
	assert.Equal(t, Status("degraded"), StatusDegraded)
	assert.Equal(t, Status("unhealthy"), StatusUnhealthy)
}

func TestHealthChecker_StartPeriodicChecks(t *testing.T) {
	checker := NewHealthChecker()

	checkCount := 0
	var mu sync.Mutex

	checker.RegisterCheck("periodic-check", func(ctx context.Context) (Status, string) {
		mu.Lock()
		checkCount++
		mu.Unlock()
		return StatusHealthy, "OK"
	}, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	// Start periodic checks
	checker.StartPeriodicChecks(ctx)

	// Wait for a few check cycles
	time.Sleep(200 * time.Millisecond)

	// Cancel context to stop periodic checks
	cancel()

	// Give time for goroutines to exit
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	finalCount := checkCount
	mu.Unlock()

	// Should have run at least 2-3 times
	require.GreaterOrEqual(t, finalCount, 2, "Expected at least 2 check runs")
}

func TestHealthChecker_EmptyChecks(t *testing.T) {
	checker := NewHealthChecker()

	// GetOverallStatus with no checks should return healthy
	assert.Equal(t, StatusHealthy, checker.GetOverallStatus())

	// RunChecks with no checks should return empty map
	results := checker.RunChecks(context.Background())
	assert.Empty(t, results)

	// GetCheckResults with no checks should still work
	checkResults := checker.GetCheckResults()
	assert.Equal(t, StatusHealthy, checkResults["status"])
	assert.Empty(t, checkResults["checks"])
}

func BenchmarkRunChecks(b *testing.B) {
	checker := NewHealthChecker()

	for i := 0; i < 10; i++ {
		checker.RegisterCheck("check-"+string(rune('a'+i)), func(ctx context.Context) (Status, string) {
			return StatusHealthy, "OK"
		}, time.Minute)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.RunChecks(ctx)
	}
}

func BenchmarkGetOverallStatus(b *testing.B) {
	checker := NewHealthChecker()

	for i := 0; i < 10; i++ {
		checker.RegisterCheck("check-"+string(rune('a'+i)), func(ctx context.Context) (Status, string) {
			return StatusHealthy, "OK"
		}, time.Minute)
	}

	checker.RunChecks(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.GetOverallStatus()
	}
}
