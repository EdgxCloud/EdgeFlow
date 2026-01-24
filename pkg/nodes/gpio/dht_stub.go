// +build !linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// DHTConfig configuration for DHT11/DHT22 sensor node
type DHTConfig struct {
	Pin         int     `json:"pin"`
	Type        string  `json:"type"`
	Retries     int     `json:"retries"`
	RetryDelay  int     `json:"retry_delay"`
	TempOffset  float64 `json:"temp_offset"`
	HumidOffset float64 `json:"humid_offset"`
}

// DHTExecutor stub for non-Linux platforms
type DHTExecutor struct {
	config DHTConfig
}

// NewDHTExecutor creates a new DHT executor (stub)
func NewDHTExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var dhtConfig DHTConfig
	if err := json.Unmarshal(configJSON, &dhtConfig); err != nil {
		return nil, fmt.Errorf("invalid dht config: %w", err)
	}

	if dhtConfig.Type == "" {
		dhtConfig.Type = "dht22"
	}
	dhtConfig.Type = strings.ToLower(dhtConfig.Type)

	return &DHTExecutor{config: dhtConfig}, nil
}

// Init initializes the DHT executor
func (e *DHTExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute returns stub data on non-Linux platforms
func (e *DHTExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Return simulated data for testing on non-Linux platforms
	return node.Message{
		Payload: map[string]interface{}{
			"temperature": 22.5 + e.config.TempOffset,
			"humidity":    50.0 + e.config.HumidOffset,
			"unit":        "C",
			"type":        e.config.Type,
			"pin":         e.config.Pin,
			"simulated":   true,
			"platform":    "non-linux",
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

// Cleanup releases resources
func (e *DHTExecutor) Cleanup() error {
	return nil
}
