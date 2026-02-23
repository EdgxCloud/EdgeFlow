package saas

import (
	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/resources"
)

// SystemServiceAdapter adapts api.Service to SystemService interface
type SystemServiceAdapter struct {
	apiService interface{} // Will hold *api.Service
}

// NewSystemServiceAdapter creates a system service adapter
func NewSystemServiceAdapter(apiService interface{}) SystemService {
	return &SystemServiceAdapter{
		apiService: apiService,
	}
}

// GetSystemInfo returns system metrics (CPU, memory, disk, etc.)
func (a *SystemServiceAdapter) GetSystemInfo() (map[string]interface{}, error) {
	info := resources.GetSystemInfo()

	return map[string]interface{}{
		"hostname":    info.Hostname,
		"os":          info.OS,
		"arch":        info.Arch,
		"board_model": info.BoardModel,
		"uptime":      info.Uptime,
		"temperature": info.Temperature,
		"cpu": map[string]interface{}{
			"usage_percent": info.CPUUsagePercent,
		},
		"memory": map[string]interface{}{
			"total_bytes":     info.OSMemTotal,
			"used_bytes":      info.OSMemUsed,
			"free_bytes":      info.OSMemFree,
			"available_bytes": info.OSMemAvailable,
			"percent":         info.OSMemPercent,
		},
		"swap": map[string]interface{}{
			"total_bytes": info.OSSwapTotal,
			"used_bytes":  info.OSSwapUsed,
		},
		"load_avg": map[string]interface{}{
			"1min":  info.LoadAvg1,
			"5min":  info.LoadAvg5,
			"15min": info.LoadAvg15,
		},
		"network": map[string]interface{}{
			"rx_bytes": info.NetRxBytes,
			"tx_bytes": info.NetTxBytes,
		},
	}, nil
}

// GetExecutions returns execution history
func (a *SystemServiceAdapter) GetExecutions() ([]interface{}, error) {
	if a.apiService == nil {
		return []interface{}{}, nil
	}

	// Type assert to access ListExecutions method
	type executionLister interface {
		ListExecutions() interface{}
	}

	if lister, ok := a.apiService.(executionLister); ok {
		execList := lister.ListExecutions()
		// Convert to []interface{}
		if execSlice, ok := execList.([]interface{}); ok {
			return execSlice, nil
		}
		// Try reflection if direct cast fails
		return []interface{}{execList}, nil
	}

	return []interface{}{}, nil
}

// GetGPIOState returns current GPIO pin states
func (a *SystemServiceAdapter) GetGPIOState() (map[string]interface{}, error) {
	h, err := hal.GetGlobalHAL()
	if err != nil || h == nil {
		return map[string]interface{}{
			"available": false,
			"message":   "HAL not initialized",
		}, nil
	}

	info := h.Info()
	gpio := h.GPIO()

	// Get active pins
	activePins := gpio.ActivePins()

	// Build pin states array
	pinStates := []map[string]interface{}{}
	for pin, mode := range activePins {
		modeStr := "unknown"
		switch mode {
		case hal.Input:
			modeStr = "input"
		case hal.Output:
			modeStr = "output"
		case hal.PWM:
			modeStr = "pwm"
		}

		pinStates = append(pinStates, map[string]interface{}{
			"pin":  pin,
			"mode": modeStr,
		})
	}

	// Build state map
	state := map[string]interface{}{
		"available": true,
		"board":     info.Name,
		"chip":      info.GPIOChip,
		"capabilities": map[string]interface{}{
			"num_gpio": info.NumGPIO,
			"num_pwm":  info.NumPWM,
			"num_i2c":  info.NumI2C,
			"num_spi":  info.NumSPI,
		},
		"active_pins": pinStates,
	}

	return state, nil
}

