//go:build linux
// +build linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// AudioConfig configuration for Raspberry Pi Audio node
type AudioConfig struct {
	Operation  string `json:"operation"`  // "record", "play", "detect", "volume"
	Device     string `json:"device"`     // ALSA device (e.g., "hw:0,0", "plughw:0,0", "default")
	Format     string `json:"format"`     // "wav", "raw"
	SampleRate int    `json:"sampleRate"` // Sample rate in Hz (8000, 16000, 44100, 48000)
	Channels   int    `json:"channels"`   // 1 (mono) or 2 (stereo)
	BitDepth   int    `json:"bitDepth"`   // 8, 16, 24, 32
	Duration   int    `json:"duration"`   // Recording duration in seconds
	OutputDir  string `json:"outputDir"`  // Output directory for recordings
	Volume     int    `json:"volume"`     // Volume level 0-100
}

// AudioExecutor executes audio operations on Raspberry Pi
type AudioExecutor struct {
	config AudioConfig
}

// NewAudioExecutor creates a new audio executor
func NewAudioExecutor() *AudioExecutor {
	return &AudioExecutor{}
}

// Init initializes the audio executor
func (e *AudioExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := json.Unmarshal(configJSON, &e.config); err != nil {
		return fmt.Errorf("invalid audio config: %w", err)
	}

	// Defaults
	if e.config.Operation == "" {
		e.config.Operation = "detect"
	}
	if e.config.Device == "" {
		e.config.Device = "default"
	}
	if e.config.Format == "" {
		e.config.Format = "wav"
	}
	if e.config.SampleRate == 0 {
		e.config.SampleRate = 44100
	}
	if e.config.Channels == 0 {
		e.config.Channels = 1
	}
	if e.config.BitDepth == 0 {
		e.config.BitDepth = 16
	}
	if e.config.Duration == 0 {
		e.config.Duration = 5
	}
	if e.config.OutputDir == "" {
		e.config.OutputDir = "/tmp/edgeflow-audio"
	}
	if e.config.Volume == 0 {
		e.config.Volume = 75
	}

	if err := os.MkdirAll(e.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// Execute performs an audio operation
func (e *AudioExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	operation := e.config.Operation
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	switch operation {
	case "record":
		return e.record(ctx, msg)
	case "play":
		return e.play(ctx, msg)
	case "detect":
		return e.detectDevices(ctx)
	case "volume":
		return e.setVolume(ctx, msg)
	default:
		return node.Message{}, fmt.Errorf("unknown audio operation: %s", operation)
	}
}

// record captures audio using arecord
func (e *AudioExecutor) record(ctx context.Context, msg node.Message) (node.Message, error) {
	if _, err := exec.LookPath("arecord"); err != nil {
		return node.Message{}, fmt.Errorf("arecord not found - install ALSA utils: sudo apt install alsa-utils")
	}

	duration := e.config.Duration
	if d, ok := msg.Payload["duration"].(float64); ok {
		duration = int(d)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("recording_%s.%s", timestamp, e.config.Format)
	outputPath := filepath.Join(e.config.OutputDir, filename)

	formatStr := fmt.Sprintf("S%dLE", e.config.BitDepth) // e.g. S16LE
	if e.config.BitDepth == 8 {
		formatStr = "U8"
	}

	args := []string{
		"-D", e.config.Device,
		"-f", formatStr,
		"-r", strconv.Itoa(e.config.SampleRate),
		"-c", strconv.Itoa(e.config.Channels),
		"-d", strconv.Itoa(duration),
		"-t", e.config.Format,
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "arecord", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return node.Message{}, fmt.Errorf("recording failed: %w (output: %s)", err, string(output))
	}

	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to stat recording: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"path":       outputPath,
			"filename":   filename,
			"format":     e.config.Format,
			"sampleRate": e.config.SampleRate,
			"channels":   e.config.Channels,
			"bitDepth":   e.config.BitDepth,
			"duration":   duration,
			"size":       fileInfo.Size(),
			"device":     e.config.Device,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

// play plays audio using aplay
func (e *AudioExecutor) play(ctx context.Context, msg node.Message) (node.Message, error) {
	if _, err := exec.LookPath("aplay"); err != nil {
		return node.Message{}, fmt.Errorf("aplay not found - install ALSA utils: sudo apt install alsa-utils")
	}

	// Get file path from message
	filePath, ok := msg.Payload["path"].(string)
	if !ok || filePath == "" {
		return node.Message{}, fmt.Errorf("audio file path is required in payload.path")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return node.Message{}, fmt.Errorf("audio file not found: %s", filePath)
	}

	args := []string{"-D", e.config.Device, filePath}

	cmd := exec.CommandContext(ctx, "aplay", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return node.Message{}, fmt.Errorf("playback failed: %w (output: %s)", err, string(output))
	}

	return node.Message{
		Payload: map[string]interface{}{
			"played":    true,
			"path":      filePath,
			"device":    e.config.Device,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// detectDevices lists available audio devices
func (e *AudioExecutor) detectDevices(ctx context.Context) (node.Message, error) {
	result := map[string]interface{}{
		"playback":  []interface{}{},
		"capture":   []interface{}{},
		"timestamp": time.Now().Unix(),
	}

	// List playback devices
	if _, err := exec.LookPath("aplay"); err == nil {
		cmd := exec.CommandContext(ctx, "aplay", "-l")
		output, err := cmd.CombinedOutput()
		if err == nil {
			result["playback"] = parseALSADevices(string(output))
		}
	}

	// List capture devices
	if _, err := exec.LookPath("arecord"); err == nil {
		cmd := exec.CommandContext(ctx, "arecord", "-l")
		output, err := cmd.CombinedOutput()
		if err == nil {
			result["capture"] = parseALSADevices(string(output))
		}
	}

	// Get current volume
	if _, err := exec.LookPath("amixer"); err == nil {
		cmd := exec.CommandContext(ctx, "amixer", "get", "Master")
		output, _ := cmd.CombinedOutput()
		result["mixer"] = string(output)
	}

	return node.Message{Payload: result}, nil
}

// setVolume sets the audio volume using amixer
func (e *AudioExecutor) setVolume(ctx context.Context, msg node.Message) (node.Message, error) {
	if _, err := exec.LookPath("amixer"); err != nil {
		return node.Message{}, fmt.Errorf("amixer not found - install ALSA utils: sudo apt install alsa-utils")
	}

	volume := e.config.Volume
	if v, ok := msg.Payload["volume"].(float64); ok {
		volume = int(v)
	}
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}

	cmd := exec.CommandContext(ctx, "amixer", "set", "Master", fmt.Sprintf("%d%%", volume))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return node.Message{}, fmt.Errorf("volume set failed: %w (output: %s)", err, string(output))
	}

	return node.Message{
		Payload: map[string]interface{}{
			"volume":    volume,
			"success":   true,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// parseALSADevices parses aplay -l / arecord -l output
func parseALSADevices(output string) []map[string]interface{} {
	devices := []map[string]interface{}{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "card ") {
			// Parse "card 0: bcm2835 [bcm2835 Headphones], device 0: ..."
			parts := strings.SplitN(line, ":", 2)
			if len(parts) >= 2 {
				cardStr := strings.TrimPrefix(parts[0], "card ")
				cardNum, _ := strconv.Atoi(strings.TrimSpace(cardStr))

				devNum := 0
				if idx := strings.Index(parts[1], "device"); idx >= 0 {
					devPart := parts[1][idx+7:]
					colonIdx := strings.Index(devPart, ":")
					if colonIdx > 0 {
						devNum, _ = strconv.Atoi(strings.TrimSpace(devPart[:colonIdx]))
					}
				}

				name := strings.TrimSpace(parts[1])

				devices = append(devices, map[string]interface{}{
					"card":     cardNum,
					"device":   devNum,
					"name":     name,
					"alsaName": fmt.Sprintf("hw:%d,%d", cardNum, devNum),
				})
			}
		}
	}
	return devices
}

// Cleanup releases audio resources
func (e *AudioExecutor) Cleanup() error {
	return nil
}
