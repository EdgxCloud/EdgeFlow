//go:build linux
// +build linux

package gpio

import (
	"context"
	"encoding/base64"
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

// CameraConfig configuration for Raspberry Pi Camera node
type CameraConfig struct {
	Mode       string `json:"mode"`       // "photo" or "video"
	Width      int    `json:"width"`      // Image width in pixels
	Height     int    `json:"height"`     // Image height in pixels
	Rotation   int    `json:"rotation"`   // Image rotation: 0, 90, 180, 270
	Quality    int    `json:"quality"`    // JPEG quality: 1-100
	OutputDir  string `json:"outputDir"`  // Output directory for files
	Duration   int    `json:"duration"`   // Video duration in seconds
	Format     string `json:"format"`     // "jpeg", "png", "bmp"
	HFlip      bool   `json:"hflip"`      // Horizontal flip
	VFlip      bool   `json:"vflip"`      // Vertical flip
	Brightness float64 `json:"brightness"` // Brightness adjustment (-1.0 to 1.0)
	Contrast   float64 `json:"contrast"`   // Contrast adjustment (0.0 to 2.0)
	Exposure   string `json:"exposure"`   // Exposure mode: auto, short, long
	AWB        string `json:"awb"`        // Auto white balance: auto, daylight, cloudy, tungsten
}

// CameraExecutor executes camera capture operations on Raspberry Pi
type CameraExecutor struct {
	config   CameraConfig
	cameraCmd string // detected camera command
}

// NewCameraExecutor creates a new camera executor
func NewCameraExecutor() *CameraExecutor {
	return &CameraExecutor{}
}

// Init initializes the camera executor
func (e *CameraExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := json.Unmarshal(configJSON, &e.config); err != nil {
		return fmt.Errorf("invalid camera config: %w", err)
	}

	// Defaults
	if e.config.Mode == "" {
		e.config.Mode = "photo"
	}
	if e.config.Width == 0 {
		e.config.Width = 1920
	}
	if e.config.Height == 0 {
		e.config.Height = 1080
	}
	if e.config.Quality == 0 {
		e.config.Quality = 85
	}
	if e.config.Format == "" {
		e.config.Format = "jpeg"
	}
	if e.config.OutputDir == "" {
		e.config.OutputDir = "/tmp/edgeflow-camera"
	}
	if e.config.Duration == 0 {
		e.config.Duration = 5
	}
	if e.config.Exposure == "" {
		e.config.Exposure = "auto"
	}
	if e.config.AWB == "" {
		e.config.AWB = "auto"
	}

	// Detect available camera command
	e.cameraCmd = detectCameraCommand()
	if e.cameraCmd == "" {
		return fmt.Errorf("no camera command found (tried rpicam-still, libcamera-still, raspistill)")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(e.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// detectCameraCommand finds the available camera CLI tool
func detectCameraCommand() string {
	// Pi 5 / Bookworm uses rpicam-still
	if _, err := exec.LookPath("rpicam-still"); err == nil {
		return "rpicam-still"
	}
	// Pi 4 / Bullseye uses libcamera-still
	if _, err := exec.LookPath("libcamera-still"); err == nil {
		return "libcamera-still"
	}
	// Legacy (Buster and older) uses raspistill
	if _, err := exec.LookPath("raspistill"); err == nil {
		return "raspistill"
	}
	return ""
}

// detectVideoCommand finds the available video CLI tool
func detectVideoCommand() string {
	if _, err := exec.LookPath("rpicam-vid"); err == nil {
		return "rpicam-vid"
	}
	if _, err := exec.LookPath("libcamera-vid"); err == nil {
		return "libcamera-vid"
	}
	if _, err := exec.LookPath("raspivid"); err == nil {
		return "raspivid"
	}
	return ""
}

// Execute captures a photo or video
func (e *CameraExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Allow message to override config
	mode := e.config.Mode
	if m, ok := msg.Payload["mode"].(string); ok {
		mode = m
	}

	switch mode {
	case "photo":
		return e.capturePhoto(ctx, msg)
	case "video":
		return e.captureVideo(ctx, msg)
	case "detect":
		return e.detectCamera(ctx, msg)
	default:
		return node.Message{}, fmt.Errorf("unknown camera mode: %s", mode)
	}
}

// capturePhoto takes a still photo
func (e *CameraExecutor) capturePhoto(ctx context.Context, msg node.Message) (node.Message, error) {
	timestamp := time.Now().Format("20060102-150405")
	ext := e.config.Format
	if ext == "jpeg" {
		ext = "jpg"
	}
	filename := fmt.Sprintf("capture_%s.%s", timestamp, ext)
	outputPath := filepath.Join(e.config.OutputDir, filename)

	args := []string{
		"-o", outputPath,
		"--width", strconv.Itoa(e.config.Width),
		"--height", strconv.Itoa(e.config.Height),
		"-q", strconv.Itoa(e.config.Quality),
		"-n", // no preview
		"-t", "1", // minimal timeout
	}

	if e.config.Rotation != 0 {
		args = append(args, "--rotation", strconv.Itoa(e.config.Rotation))
	}
	if e.config.HFlip {
		args = append(args, "--hflip")
	}
	if e.config.VFlip {
		args = append(args, "--vflip")
	}
	if e.config.Brightness != 0 {
		args = append(args, "--brightness", fmt.Sprintf("%.2f", e.config.Brightness))
	}
	if e.config.Contrast != 0 {
		args = append(args, "--contrast", fmt.Sprintf("%.2f", e.config.Contrast))
	}
	if e.config.Exposure != "auto" {
		args = append(args, "--exposure", e.config.Exposure)
	}
	if e.config.AWB != "auto" {
		args = append(args, "--awb", e.config.AWB)
	}

	// For legacy raspistill, adjust args
	cmd := e.cameraCmd
	if strings.Contains(cmd, "raspistill") {
		// raspistill uses -w/-h instead of --width/--height
		args = adjustForLegacy(args)
	}

	execCmd := exec.CommandContext(ctx, cmd, args...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return node.Message{}, fmt.Errorf("camera capture failed: %w (output: %s)", err, string(output))
	}

	// Read file info
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to stat captured file: %w", err)
	}

	result := map[string]interface{}{
		"path":      outputPath,
		"filename":  filename,
		"format":    e.config.Format,
		"width":     e.config.Width,
		"height":    e.config.Height,
		"quality":   e.config.Quality,
		"size":      fileInfo.Size(),
		"timestamp": time.Now().Unix(),
		"camera":    cmd,
	}

	// Optionally encode as base64 for small images
	if fileInfo.Size() < 2*1024*1024 { // < 2MB
		data, err := os.ReadFile(outputPath)
		if err == nil {
			result["base64"] = base64.StdEncoding.EncodeToString(data)
		}
	}

	return node.Message{
		Payload: result,
	}, nil
}

// captureVideo records a video
func (e *CameraExecutor) captureVideo(ctx context.Context, msg node.Message) (node.Message, error) {
	videoCmd := detectVideoCommand()
	if videoCmd == "" {
		return node.Message{}, fmt.Errorf("no video command found (tried rpicam-vid, libcamera-vid, raspivid)")
	}

	duration := e.config.Duration
	if d, ok := msg.Payload["duration"].(float64); ok {
		duration = int(d)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("video_%s.h264", timestamp)
	outputPath := filepath.Join(e.config.OutputDir, filename)

	args := []string{
		"-o", outputPath,
		"--width", strconv.Itoa(e.config.Width),
		"--height", strconv.Itoa(e.config.Height),
		"-t", strconv.Itoa(duration * 1000), // ms
		"-n", // no preview
	}

	execCmd := exec.CommandContext(ctx, videoCmd, args...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return node.Message{}, fmt.Errorf("video capture failed: %w (output: %s)", err, string(output))
	}

	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to stat captured video: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"path":      outputPath,
			"filename":  filename,
			"format":    "h264",
			"width":     e.config.Width,
			"height":    e.config.Height,
			"duration":  duration,
			"size":      fileInfo.Size(),
			"timestamp": time.Now().Unix(),
			"camera":    videoCmd,
		},
	}, nil
}

// detectCamera checks camera availability and returns info
func (e *CameraExecutor) detectCamera(ctx context.Context, msg node.Message) (node.Message, error) {
	result := map[string]interface{}{
		"available": false,
		"command":   "",
		"cameras":   []interface{}{},
	}

	// Try to list cameras
	listCmds := []string{"rpicam-still", "libcamera-still"}
	for _, cmd := range listCmds {
		if _, err := exec.LookPath(cmd); err != nil {
			continue
		}
		execCmd := exec.CommandContext(ctx, cmd, "--list-cameras")
		output, err := execCmd.CombinedOutput()
		if err == nil {
			result["available"] = true
			result["command"] = cmd
			result["info"] = string(output)
			break
		}
	}

	// Check V4L2 devices
	devices, _ := filepath.Glob("/dev/video*")
	result["v4l2Devices"] = devices

	// Check if camera module is loaded
	if data, err := os.ReadFile("/proc/modules"); err == nil {
		modules := string(data)
		result["bcm2835_v4l2"] = strings.Contains(modules, "bcm2835_v4l2")
		result["imx708"] = strings.Contains(modules, "imx708") // Pi Camera Module 3
		result["imx219"] = strings.Contains(modules, "imx219") // Pi Camera Module 2
		result["ov5647"] = strings.Contains(modules, "ov5647") // Pi Camera Module 1
	}

	return node.Message{Payload: result}, nil
}

// adjustForLegacy converts libcamera args to raspistill args
func adjustForLegacy(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--width":
			out = append(out, "-w")
		case "--height":
			out = append(out, "-h")
		case "--rotation":
			out = append(out, "-rot")
		default:
			out = append(out, args[i])
			continue
		}
		if i+1 < len(args) {
			i++
			out = append(out, args[i])
		}
	}
	return out
}

// Cleanup releases camera resources
func (e *CameraExecutor) Cleanup() error {
	return nil
}
