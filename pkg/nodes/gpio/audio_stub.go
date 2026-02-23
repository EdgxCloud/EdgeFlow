//go:build !linux
// +build !linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// AudioConfig configuration for Raspberry Pi Audio node
type AudioConfig struct {
	Operation  string `json:"operation"`
	Device     string `json:"device"`
	Format     string `json:"format"`
	SampleRate int    `json:"sampleRate"`
	Channels   int    `json:"channels"`
	BitDepth   int    `json:"bitDepth"`
	Duration   int    `json:"duration"`
	OutputDir  string `json:"outputDir"`
	Volume     int    `json:"volume"`
}

// AudioExecutor stub for non-Linux platforms
type AudioExecutor struct {
	config AudioConfig
}

// NewAudioExecutor creates a new audio executor (stub)
func NewAudioExecutor() *AudioExecutor {
	return &AudioExecutor{}
}

// Init initializes the audio executor
func (e *AudioExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	json.Unmarshal(configJSON, &e.config)
	return nil
}

// Execute returns simulated data on non-Linux platforms
func (e *AudioExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	return node.Message{
		Payload: map[string]interface{}{
			"simulated": true,
			"platform":  "non-linux",
			"playback":  []interface{}{},
			"capture":   []interface{}{},
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// Cleanup releases resources
func (e *AudioExecutor) Cleanup() error {
	return nil
}
