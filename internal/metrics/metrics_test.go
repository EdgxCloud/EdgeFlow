package metrics

import (
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}
	if m.startTime.IsZero() {
		t.Error("Start time not set")
	}
}

func TestIncrementFlows(t *testing.T) {
	m := NewMetrics()

	initial := m.TotalFlows
	m.IncrementFlows()

	if m.TotalFlows != initial+1 {
		t.Errorf("Expected TotalFlows to be %d, got %d", initial+1, m.TotalFlows)
	}
}

func TestIncrementRunningFlows(t *testing.T) {
	m := NewMetrics()

	m.IncrementRunningFlows()
	if m.RunningFlows != 1 {
		t.Errorf("Expected RunningFlows to be 1, got %d", m.RunningFlows)
	}

	m.IncrementRunningFlows()
	if m.RunningFlows != 2 {
		t.Errorf("Expected RunningFlows to be 2, got %d", m.RunningFlows)
	}
}

func TestDecrementRunningFlows(t *testing.T) {
	m := NewMetrics()

	m.IncrementRunningFlows()
	m.IncrementRunningFlows()
	m.DecrementRunningFlows()

	if m.RunningFlows != 1 {
		t.Errorf("Expected RunningFlows to be 1, got %d", m.RunningFlows)
	}
	if m.StoppedFlows != 1 {
		t.Errorf("Expected StoppedFlows to be 1, got %d", m.StoppedFlows)
	}
}

func TestIncrementExecutions(t *testing.T) {
	m := NewMetrics()

	m.IncrementExecutions()
	m.IncrementExecutions()

	if m.TotalExecutions != 2 {
		t.Errorf("Expected TotalExecutions to be 2, got %d", m.TotalExecutions)
	}
}

func TestIncrementFailedExecutions(t *testing.T) {
	m := NewMetrics()

	m.IncrementExecutions()
	m.IncrementExecutions()
	m.IncrementFailedExecutions()

	if m.FailedExecutions != 1 {
		t.Errorf("Expected FailedExecutions to be 1, got %d", m.FailedExecutions)
	}
}

func TestRecordResponseTime(t *testing.T) {
	m := NewMetrics()

	// Record first response
	m.RecordResponseTime(100 * time.Millisecond)
	if m.AvgResponseTime == 0 {
		t.Error("Expected AvgResponseTime to be set")
	}

	// Record second response
	first := m.AvgResponseTime
	m.RecordResponseTime(200 * time.Millisecond)
	if m.AvgResponseTime == first {
		t.Error("Expected AvgResponseTime to change")
	}
}

func TestUpdateSystemMetrics(t *testing.T) {
	m := NewMetrics()
	time.Sleep(10 * time.Millisecond)

	m.UpdateSystemMetrics()

	if m.Uptime == 0 {
		t.Error("Expected Uptime to be greater than 0")
	}
	if m.MemoryUsed == 0 {
		t.Error("Expected MemoryUsed to be greater than 0")
	}
	if m.GoroutineCount == 0 {
		t.Error("Expected GoroutineCount to be greater than 0")
	}
}

func TestGetMetrics(t *testing.T) {
	m := NewMetrics()
	m.IncrementFlows()
	m.IncrementRunningFlows()
	m.IncrementExecutions()

	metrics := m.GetMetrics()

	if metrics == nil {
		t.Fatal("GetMetrics returned nil")
	}

	flows, ok := metrics["flows"].(map[string]interface{})
	if !ok {
		t.Fatal("flows not found in metrics")
	}

	if flows["total"] != int64(1) {
		t.Errorf("Expected flows.total to be 1, got %v", flows["total"])
	}
}

func TestPrometheusFormat(t *testing.T) {
	m := NewMetrics()
	m.IncrementFlows()
	m.IncrementExecutions()

	prometheus := m.PrometheusFormat()

	if prometheus == "" {
		t.Error("PrometheusFormat returned empty string")
	}

	if !contains(prometheus, "edgeflow_flows_total") {
		t.Error("Expected edgeflow_flows_total in Prometheus output")
	}
	if !contains(prometheus, "edgeflow_executions_total") {
		t.Error("Expected edgeflow_executions_total in Prometheus output")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr &&
		(len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkIncrementFlows(b *testing.B) {
	m := NewMetrics()
	for i := 0; i < b.N; i++ {
		m.IncrementFlows()
	}
}

func BenchmarkRecordResponseTime(b *testing.B) {
	m := NewMetrics()
	for i := 0; i < b.N; i++ {
		m.RecordResponseTime(100 * time.Millisecond)
	}
}

func BenchmarkGetMetrics(b *testing.B) {
	m := NewMetrics()
	m.IncrementFlows()
	m.IncrementExecutions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetMetrics()
	}
}
