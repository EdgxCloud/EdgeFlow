//go:build linux
// +build linux

package gpio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"go.bug.st/serial"
)

// NEOM8NConfig holds configuration for NEO-M8N GPS module
type NEOM8NConfig struct {
	SerialPort   string `json:"serial_port"`
	BaudRate     int    `json:"baud_rate"`
	PollInterval int    `json:"poll_interval_ms"`
}

// NEOM8NExecutor implements NEO-M8N GPS receiver
type NEOM8NExecutor struct {
	config      NEOM8NConfig
	port        serial.Port
	mu          sync.Mutex
	initialized bool
	lastData    GPSData
	reader      *bufio.Reader
}

func (e *NEOM8NExecutor) Init(config map[string]interface{}) error {
	e.config = NEOM8NConfig{
		SerialPort:   "/dev/ttyAMA0",
		BaudRate:     9600,
		PollInterval: 1000,
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse NEO-M8N config: %w", err)
		}
	}

	return nil
}

func (e *NEOM8NExecutor) initHardware() error {
	if e.initialized {
		return nil
	}

	mode := &serial.Mode{
		BaudRate: e.config.BaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(e.config.SerialPort, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", e.config.SerialPort, err)
	}
	e.port = port
	e.reader = bufio.NewReader(port)

	e.initialized = true
	return nil
}

func (e *NEOM8NExecutor) readNMEASentence() (string, error) {
	line, err := e.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (e *NEOM8NExecutor) parseNMEA(sentence string) {
	if len(sentence) < 6 || sentence[0] != '$' {
		return
	}

	// Verify checksum
	if !e.verifyChecksum(sentence) {
		return
	}

	parts := strings.Split(sentence[1:], ",")
	if len(parts) < 1 {
		return
	}

	msgType := parts[0]
	switch {
	case strings.HasSuffix(msgType, "GGA"):
		e.parseGGA(parts)
	case strings.HasSuffix(msgType, "RMC"):
		e.parseRMC(parts)
	case strings.HasSuffix(msgType, "GSA"):
		e.parseGSA(parts)
	case strings.HasSuffix(msgType, "VTG"):
		e.parseVTG(parts)
	}
}

func (e *NEOM8NExecutor) verifyChecksum(sentence string) bool {
	asterisk := strings.LastIndex(sentence, "*")
	if asterisk < 0 || asterisk+3 > len(sentence) {
		return false
	}

	data := sentence[1:asterisk]
	checksum := sentence[asterisk+1:]

	var calculated byte
	for i := 0; i < len(data); i++ {
		calculated ^= data[i]
	}

	expected, err := strconv.ParseUint(checksum, 16, 8)
	if err != nil {
		return false
	}

	return calculated == byte(expected)
}

func (e *NEOM8NExecutor) parseGGA(parts []string) {
	if len(parts) < 15 {
		return
	}

	// Parse time
	if len(parts[1]) >= 6 {
		e.parseTime(parts[1])
	}

	// Parse latitude
	if lat, ok := e.parseLatLon(parts[2], parts[3]); ok {
		e.lastData.Latitude = lat
	}

	// Parse longitude
	if lon, ok := e.parseLatLon(parts[4], parts[5]); ok {
		e.lastData.Longitude = lon
	}

	// Fix quality
	if fix, err := strconv.Atoi(parts[6]); err == nil {
		e.lastData.FixQuality = fix
		e.lastData.Valid = fix > 0
		switch fix {
		case 0:
			e.lastData.FixType = "invalid"
		case 1:
			e.lastData.FixType = "gps"
		case 2:
			e.lastData.FixType = "dgps"
		case 4:
			e.lastData.FixType = "rtk_fixed"
		case 5:
			e.lastData.FixType = "rtk_float"
		default:
			e.lastData.FixType = "unknown"
		}
	}

	// Number of satellites
	if sats, err := strconv.Atoi(parts[7]); err == nil {
		e.lastData.Satellites = sats
	}

	// HDOP
	if hdop, err := strconv.ParseFloat(parts[8], 64); err == nil {
		e.lastData.HDOP = hdop
	}

	// Altitude
	if alt, err := strconv.ParseFloat(parts[9], 64); err == nil {
		e.lastData.Altitude = alt
	}
}

func (e *NEOM8NExecutor) parseRMC(parts []string) {
	if len(parts) < 12 {
		return
	}

	// Status
	e.lastData.Valid = parts[2] == "A"

	// Parse latitude
	if lat, ok := e.parseLatLon(parts[3], parts[4]); ok {
		e.lastData.Latitude = lat
	}

	// Parse longitude
	if lon, ok := e.parseLatLon(parts[5], parts[6]); ok {
		e.lastData.Longitude = lon
	}

	// Speed over ground (knots to km/h)
	if speed, err := strconv.ParseFloat(parts[7], 64); err == nil {
		e.lastData.SpeedKmh = speed * 1.852 // knots to km/h
	}

	// Course over ground
	if course, err := strconv.ParseFloat(parts[8], 64); err == nil {
		e.lastData.Course = course
	}

	// Parse date and time
	if len(parts[1]) >= 6 && len(parts[9]) >= 6 {
		e.parseDateTime(parts[1], parts[9])
	}
}

func (e *NEOM8NExecutor) parseGSA(parts []string) {
	if len(parts) < 18 {
		return
	}

	// Fix type from GSA
	if fix, err := strconv.Atoi(parts[2]); err == nil {
		switch fix {
		case 1:
			e.lastData.FixType = "no_fix"
		case 2:
			e.lastData.FixType = "2d"
		case 3:
			e.lastData.FixType = "3d"
		}
	}

	// HDOP
	if hdop, err := strconv.ParseFloat(parts[16], 64); err == nil {
		e.lastData.HDOP = hdop
	}
}

func (e *NEOM8NExecutor) parseVTG(parts []string) {
	if len(parts) < 9 {
		return
	}

	// Course
	if course, err := strconv.ParseFloat(parts[1], 64); err == nil {
		e.lastData.Course = course
	}

	// Speed in km/h
	if speed, err := strconv.ParseFloat(parts[7], 64); err == nil {
		e.lastData.SpeedKmh = speed
	}
}

func (e *NEOM8NExecutor) parseLatLon(value, direction string) (float64, bool) {
	if len(value) < 4 {
		return 0, false
	}

	var degrees float64
	var minutes float64
	var err error

	if direction == "N" || direction == "S" {
		// Latitude: DDMM.MMMM
		degrees, err = strconv.ParseFloat(value[:2], 64)
		if err != nil {
			return 0, false
		}
		minutes, err = strconv.ParseFloat(value[2:], 64)
		if err != nil {
			return 0, false
		}
	} else {
		// Longitude: DDDMM.MMMM
		degrees, err = strconv.ParseFloat(value[:3], 64)
		if err != nil {
			return 0, false
		}
		minutes, err = strconv.ParseFloat(value[3:], 64)
		if err != nil {
			return 0, false
		}
	}

	result := degrees + minutes/60.0

	if direction == "S" || direction == "W" {
		result = -result
	}

	return result, true
}

func (e *NEOM8NExecutor) parseTime(timeStr string) {
	if len(timeStr) < 6 {
		return
	}

	hours, _ := strconv.Atoi(timeStr[0:2])
	minutes, _ := strconv.Atoi(timeStr[2:4])
	seconds, _ := strconv.Atoi(timeStr[4:6])

	now := time.Now().UTC()
	e.lastData.Timestamp = time.Date(
		now.Year(), now.Month(), now.Day(),
		hours, minutes, seconds, 0, time.UTC,
	).Unix()
}

func (e *NEOM8NExecutor) parseDateTime(timeStr, dateStr string) {
	if len(timeStr) < 6 || len(dateStr) < 6 {
		return
	}

	hours, _ := strconv.Atoi(timeStr[0:2])
	minutes, _ := strconv.Atoi(timeStr[2:4])
	seconds, _ := strconv.Atoi(timeStr[4:6])

	day, _ := strconv.Atoi(dateStr[0:2])
	month, _ := strconv.Atoi(dateStr[2:4])
	year, _ := strconv.Atoi(dateStr[4:6])
	year += 2000

	e.lastData.Timestamp = time.Date(
		year, time.Month(month), day,
		hours, minutes, seconds, 0, time.UTC,
	).Unix()
}

func (e *NEOM8NExecutor) sendUBXCommand(class, id byte, payload []byte) error {
	// UBX protocol header
	msg := []byte{0xB5, 0x62, class, id}

	// Length (little endian)
	length := uint16(len(payload))
	msg = append(msg, byte(length&0xFF), byte(length>>8))

	// Payload
	msg = append(msg, payload...)

	// Calculate checksum
	var ckA, ckB byte
	for i := 2; i < len(msg); i++ {
		ckA += msg[i]
		ckB += ckA
	}
	msg = append(msg, ckA, ckB)

	_, err := e.port.Write(msg)
	return err
}

func (e *NEOM8NExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.initHardware(); err != nil {
		return node.Message{}, err
	}

	action := "read"
	payload := msg.Payload
	if payload != nil {
		if a, ok := payload["action"].(string); ok {
			action = a
		}
	}

	switch action {
	case "read":
		return e.readGPS()
	case "configure":
		return e.handleConfigure(msg)
	case "set_rate":
		return e.handleSetRate(msg)
	case "cold_start":
		return e.handleColdStart()
	case "warm_start":
		return e.handleWarmStart()
	case "hot_start":
		return e.handleHotStart()
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *NEOM8NExecutor) readGPS() (node.Message, error) {
	// Read multiple NMEA sentences to get complete data
	timeout := time.After(2 * time.Second)
	sentences := 0

	for sentences < 5 {
		select {
		case <-timeout:
			break
		default:
			sentence, err := e.readNMEASentence()
			if err != nil {
				continue
			}
			e.parseNMEA(sentence)
			sentences++
		}
	}

	return node.Message{
		Payload: map[string]interface{}{
			"latitude":    e.lastData.Latitude,
			"longitude":   e.lastData.Longitude,
			"altitude":    e.lastData.Altitude,
			"speed_kmh":   e.lastData.SpeedKmh,
			"course":      e.lastData.Course,
			"satellites":  e.lastData.Satellites,
			"hdop":        e.lastData.HDOP,
			"fix_quality": e.lastData.FixQuality,
			"fix_type":    e.lastData.FixType,
			"valid":       e.lastData.Valid,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (e *NEOM8NExecutor) handleConfigure(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("payload is nil")
	}

	// Configure dynamic model
	if model, ok := payload["dynamic_model"].(string); ok {
		var dynModel byte
		switch model {
		case "portable":
			dynModel = 0
		case "stationary":
			dynModel = 2
		case "pedestrian":
			dynModel = 3
		case "automotive":
			dynModel = 4
		case "sea":
			dynModel = 5
		case "airborne_1g":
			dynModel = 6
		case "airborne_2g":
			dynModel = 7
		case "airborne_4g":
			dynModel = 8
		default:
			dynModel = 0
		}

		// UBX-CFG-NAV5 command
		cfgPayload := make([]byte, 36)
		cfgPayload[0] = 0x01 // Mask: apply dynamic model
		cfgPayload[1] = 0x00
		cfgPayload[2] = dynModel

		if err := e.sendUBXCommand(0x06, 0x24, cfgPayload); err != nil {
			return node.Message{}, fmt.Errorf("failed to set dynamic model: %w", err)
		}
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":  "configured",
			"message": "GPS configuration updated",
		},
	}, nil
}

func (e *NEOM8NExecutor) handleSetRate(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("payload is nil")
	}

	rateMs := 1000.0
	if r, ok := payload["rate_ms"].(float64); ok {
		rateMs = r
	}

	// UBX-CFG-RATE command
	cfgPayload := []byte{
		byte(int(rateMs) & 0xFF), byte(int(rateMs) >> 8), // measRate
		0x01, 0x00, // navRate
		0x01, 0x00, // timeRef (UTC)
	}

	if err := e.sendUBXCommand(0x06, 0x08, cfgPayload); err != nil {
		return node.Message{}, fmt.Errorf("failed to set rate: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":  "rate_set",
			"rate_ms": rateMs,
		},
	}, nil
}

func (e *NEOM8NExecutor) handleColdStart() (node.Message, error) {
	// UBX-CFG-RST cold start
	cfgPayload := []byte{0xFF, 0xFF, 0x00, 0x00}
	if err := e.sendUBXCommand(0x06, 0x04, cfgPayload); err != nil {
		return node.Message{}, fmt.Errorf("failed to cold start: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":  "cold_start",
			"message": "GPS cold start initiated",
		},
	}, nil
}

func (e *NEOM8NExecutor) handleWarmStart() (node.Message, error) {
	// UBX-CFG-RST warm start
	cfgPayload := []byte{0x01, 0x00, 0x00, 0x00}
	if err := e.sendUBXCommand(0x06, 0x04, cfgPayload); err != nil {
		return node.Message{}, fmt.Errorf("failed to warm start: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":  "warm_start",
			"message": "GPS warm start initiated",
		},
	}, nil
}

func (e *NEOM8NExecutor) handleHotStart() (node.Message, error) {
	// UBX-CFG-RST hot start
	cfgPayload := []byte{0x00, 0x00, 0x00, 0x00}
	if err := e.sendUBXCommand(0x06, 0x04, cfgPayload); err != nil {
		return node.Message{}, fmt.Errorf("failed to hot start: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":  "hot_start",
			"message": "GPS hot start initiated",
		},
	}, nil
}

func (e *NEOM8NExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized && e.port != nil {
		e.port.Close()
		e.initialized = false
	}
	return nil
}
