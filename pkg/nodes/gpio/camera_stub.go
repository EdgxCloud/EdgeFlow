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

// CameraConfig configuration for Raspberry Pi Camera node
type CameraConfig struct {
	Mode       string  `json:"mode"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Rotation   int     `json:"rotation"`
	Quality    int     `json:"quality"`
	OutputDir  string  `json:"outputDir"`
	Duration   int     `json:"duration"`
	Format     string  `json:"format"`
	HFlip      bool    `json:"hflip"`
	VFlip      bool    `json:"vflip"`
	Brightness float64 `json:"brightness"`
	Contrast   float64 `json:"contrast"`
	Exposure   string  `json:"exposure"`
	AWB        string  `json:"awb"`
}

// CameraExecutor stub for non-Linux platforms
type CameraExecutor struct {
	config CameraConfig
}

// NewCameraExecutor creates a new camera executor (stub)
func NewCameraExecutor() *CameraExecutor {
	return &CameraExecutor{}
}

// Init initializes the camera executor
func (e *CameraExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	json.Unmarshal(configJSON, &e.config)
	if e.config.Width == 0 {
		e.config.Width = 1920
	}
	if e.config.Height == 0 {
		e.config.Height = 1080
	}
	return nil
}

// Execute returns simulated data on non-Linux platforms
func (e *CameraExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	return node.Message{
		Payload: map[string]interface{}{
			"path":      "/tmp/simulated_capture.jpg",
			"filename":  "simulated_capture.jpg",
			"format":    "jpeg",
			"width":     e.config.Width,
			"height":    e.config.Height,
			"size":      0,
			"simulated": true,
			"platform":  "non-linux",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// Cleanup releases resources
func (e *CameraExecutor) Cleanup() error {
	return nil
}
