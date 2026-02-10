//go:build linux
// +build linux

package gpio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"go.bug.st/serial"
)

// GPSConfig configuration for GPS NEO-6M module
type GPSConfig struct {
	Port     string `json:"port"`      // Serial port (e.g., "/dev/ttyAMA0", "/dev/serial0")
	BaudRate int    `json:"baud_rate"` // Baud rate (default: 9600)
}

// GPSData holds parsed GPS information
type GPSData struct {
	// Position
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Altitude   float64 `json:"altitude"`
	LatDir     string  `json:"lat_dir"`
	LonDir     string  `json:"lon_dir"`
	AltUnit    string  `json:"alt_unit"`

	// Time
	UTCTime    string `json:"utc_time"`
	UTCDate    string `json:"utc_date"`
	Timestamp  int64  `json:"timestamp"`

	// Quality
	FixQuality int     `json:"fix_quality"`
	FixType    string  `json:"fix_type"`
	Satellites int     `json:"satellites"`
	HDOP       float64 `json:"hdop"`
	PDOP       float64 `json:"pdop"`
	VDOP       float64 `json:"vdop"`

	// Speed and course
	SpeedKnots float64 `json:"speed_knots"`
	SpeedKmh   float64 `json:"speed_kmh"`
	SpeedMph   float64 `json:"speed_mph"`
	Course     float64 `json:"course"`

	// Status
	Valid      bool   `json:"valid"`
	Mode       string `json:"mode"`
	LastUpdate int64  `json:"last_update"`
}

// GPSExecutor executes GPS module readings
type GPSExecutor struct {
	config      GPSConfig
	port        serial.Port
	mu          sync.Mutex
	data        GPSData
	running     bool
	stopChan    chan struct{}
	initialized bool
}

// NewGPSExecutor creates a new GPS executor
func NewGPSExecutor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var gpsConfig GPSConfig
	if err := json.Unmarshal(configJSON, &gpsConfig); err != nil {
		return nil, fmt.Errorf("invalid GPS config: %w", err)
	}

	// Defaults
	if gpsConfig.Port == "" {
		gpsConfig.Port = "/dev/ttyAMA0" // Default Raspberry Pi serial port
	}
	if gpsConfig.BaudRate == 0 {
		gpsConfig.BaudRate = 9600 // NEO-6M default baud rate
	}

	return &GPSExecutor{
		config:   gpsConfig,
		stopChan: make(chan struct{}),
	}, nil
}

// Init initializes the GPS executor
func (e *GPSExecutor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles GPS operations
func (e *GPSExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Initialize serial port if needed
	if e.port == nil {
		mode := &serial.Mode{
			BaudRate: e.config.BaudRate,
			DataBits: 8,
			Parity:   serial.NoParity,
			StopBits: serial.OneStopBit,
		}
		port, err := serial.Open(e.config.Port, mode)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to open serial port: %w", err)
		}
		e.port = port
	}

	// Start background NMEA reading if not running
	if !e.running {
		e.running = true
		go e.readNMEA(ctx)
	}

	// Parse command
	payload := msg.Payload
	if payload == nil {
		// Default: return current GPS data
		return e.getCurrentData()
	}

	action, _ := payload["action"].(string)

	switch action {
	case "read", "get", "":
		return e.getCurrentData()

	case "raw":
		// Return raw NMEA sentence
		return e.readRawSentence()

	case "distance":
		// Calculate distance to target
		lat := getFloat(payload, "latitude", 0)
		lon := getFloat(payload, "longitude", 0)
		if lat == 0 || lon == 0 {
			return node.Message{}, fmt.Errorf("latitude and longitude required")
		}
		return e.calculateDistance(lat, lon)

	case "configure":
		// Configure GPS module
		rate := int(getFloat(payload, "rate_hz", 1))
		return e.configureGPS(rate)

	case "stop":
		e.stopReading()
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "stop",
				"timestamp": time.Now().Unix(),
			},
		}, nil

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// getCurrentData returns the current GPS data
func (e *GPSExecutor) getCurrentData() (node.Message, error) {
	return node.Message{
		Payload: map[string]interface{}{
			"latitude":    e.data.Latitude,
			"longitude":   e.data.Longitude,
			"altitude":    e.data.Altitude,
			"lat_dir":     e.data.LatDir,
			"lon_dir":     e.data.LonDir,
			"utc_time":    e.data.UTCTime,
			"utc_date":    e.data.UTCDate,
			"fix_quality": e.data.FixQuality,
			"fix_type":    e.data.FixType,
			"satellites":  e.data.Satellites,
			"hdop":        e.data.HDOP,
			"speed_knots": e.data.SpeedKnots,
			"speed_kmh":   e.data.SpeedKmh,
			"course":      e.data.Course,
			"valid":       e.data.Valid,
			"last_update": e.data.LastUpdate,
			"port":        e.config.Port,
			"sensor":      "GPS-NEO6M",
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

// readNMEA continuously reads NMEA sentences from GPS
func (e *GPSExecutor) readNMEA(ctx context.Context) {
	reader := bufio.NewReader(e.port)

	for {
		select {
		case <-ctx.Done():
			e.running = false
			return
		case <-e.stopChan:
			e.running = false
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			line = strings.TrimSpace(line)
			if len(line) == 0 || line[0] != '$' {
				continue
			}

			e.mu.Lock()
			e.parseNMEA(line)
			e.mu.Unlock()
		}
	}
}

// parseNMEA parses NMEA sentences
func (e *GPSExecutor) parseNMEA(sentence string) {
	// Verify checksum
	if !e.verifyChecksum(sentence) {
		return
	}

	// Remove checksum
	if idx := strings.Index(sentence, "*"); idx > 0 {
		sentence = sentence[:idx]
	}

	parts := strings.Split(sentence, ",")
	if len(parts) < 2 {
		return
	}

	msgType := parts[0]

	switch msgType {
	case "$GPGGA", "$GNGGA":
		e.parseGGA(parts)
	case "$GPRMC", "$GNRMC":
		e.parseRMC(parts)
	case "$GPGSA", "$GNGSA":
		e.parseGSA(parts)
	case "$GPVTG", "$GNVTG":
		e.parseVTG(parts)
	}
}

// parseGGA parses GGA sentence (fix data)
func (e *GPSExecutor) parseGGA(parts []string) {
	if len(parts) < 15 {
		return
	}

	// UTC Time
	if parts[1] != "" {
		e.data.UTCTime = parts[1]
	}

	// Latitude
	if parts[2] != "" && parts[3] != "" {
		e.data.Latitude = e.parseCoordinate(parts[2], parts[3])
		e.data.LatDir = parts[3]
	}

	// Longitude
	if parts[4] != "" && parts[5] != "" {
		e.data.Longitude = e.parseCoordinate(parts[4], parts[5])
		e.data.LonDir = parts[5]
	}

	// Fix quality
	if parts[6] != "" {
		e.data.FixQuality, _ = strconv.Atoi(parts[6])
		switch e.data.FixQuality {
		case 0:
			e.data.FixType = "invalid"
		case 1:
			e.data.FixType = "gps"
		case 2:
			e.data.FixType = "dgps"
		case 3:
			e.data.FixType = "pps"
		case 4:
			e.data.FixType = "rtk"
		case 5:
			e.data.FixType = "float_rtk"
		case 6:
			e.data.FixType = "estimated"
		case 7:
			e.data.FixType = "manual"
		case 8:
			e.data.FixType = "simulation"
		}
	}

	// Satellites
	if parts[7] != "" {
		e.data.Satellites, _ = strconv.Atoi(parts[7])
	}

	// HDOP
	if parts[8] != "" {
		e.data.HDOP, _ = strconv.ParseFloat(parts[8], 64)
	}

	// Altitude
	if parts[9] != "" {
		e.data.Altitude, _ = strconv.ParseFloat(parts[9], 64)
		e.data.AltUnit = parts[10]
	}

	e.data.Valid = e.data.FixQuality > 0
	e.data.LastUpdate = time.Now().Unix()
}

// parseRMC parses RMC sentence (recommended minimum)
func (e *GPSExecutor) parseRMC(parts []string) {
	if len(parts) < 12 {
		return
	}

	// UTC Time
	if parts[1] != "" {
		e.data.UTCTime = parts[1]
	}

	// Status
	e.data.Valid = parts[2] == "A"

	// Latitude
	if parts[3] != "" && parts[4] != "" {
		e.data.Latitude = e.parseCoordinate(parts[3], parts[4])
		e.data.LatDir = parts[4]
	}

	// Longitude
	if parts[5] != "" && parts[6] != "" {
		e.data.Longitude = e.parseCoordinate(parts[5], parts[6])
		e.data.LonDir = parts[6]
	}

	// Speed (knots)
	if parts[7] != "" {
		e.data.SpeedKnots, _ = strconv.ParseFloat(parts[7], 64)
		e.data.SpeedKmh = e.data.SpeedKnots * 1.852
		e.data.SpeedMph = e.data.SpeedKnots * 1.15078
	}

	// Course
	if parts[8] != "" {
		e.data.Course, _ = strconv.ParseFloat(parts[8], 64)
	}

	// Date (DDMMYY)
	if parts[9] != "" {
		e.data.UTCDate = parts[9]
	}

	e.data.LastUpdate = time.Now().Unix()
}

// parseGSA parses GSA sentence (DOP and active satellites)
func (e *GPSExecutor) parseGSA(parts []string) {
	if len(parts) < 18 {
		return
	}

	// Mode
	e.data.Mode = parts[1]

	// Fix type (1=no fix, 2=2D, 3=3D)
	if parts[2] != "" {
		fixType, _ := strconv.Atoi(parts[2])
		switch fixType {
		case 1:
			e.data.FixType = "no_fix"
		case 2:
			e.data.FixType = "2d"
		case 3:
			e.data.FixType = "3d"
		}
	}

	// PDOP
	if parts[15] != "" {
		e.data.PDOP, _ = strconv.ParseFloat(parts[15], 64)
	}

	// HDOP
	if parts[16] != "" {
		e.data.HDOP, _ = strconv.ParseFloat(parts[16], 64)
	}

	// VDOP
	if parts[17] != "" {
		e.data.VDOP, _ = strconv.ParseFloat(parts[17], 64)
	}
}

// parseVTG parses VTG sentence (velocity)
func (e *GPSExecutor) parseVTG(parts []string) {
	if len(parts) < 9 {
		return
	}

	// Course
	if parts[1] != "" {
		e.data.Course, _ = strconv.ParseFloat(parts[1], 64)
	}

	// Speed in knots
	if parts[5] != "" {
		e.data.SpeedKnots, _ = strconv.ParseFloat(parts[5], 64)
	}

	// Speed in km/h
	if parts[7] != "" {
		e.data.SpeedKmh, _ = strconv.ParseFloat(parts[7], 64)
		e.data.SpeedMph = e.data.SpeedKmh / 1.60934
	}
}

// parseCoordinate converts NMEA coordinate to decimal degrees
func (e *GPSExecutor) parseCoordinate(coord string, dir string) float64 {
	if len(coord) < 4 {
		return 0
	}

	var degrees float64
	var minutes float64

	// Latitude: DDMM.MMMM, Longitude: DDDMM.MMMM
	if dir == "N" || dir == "S" {
		degrees, _ = strconv.ParseFloat(coord[:2], 64)
		minutes, _ = strconv.ParseFloat(coord[2:], 64)
	} else {
		degrees, _ = strconv.ParseFloat(coord[:3], 64)
		minutes, _ = strconv.ParseFloat(coord[3:], 64)
	}

	decimal := degrees + minutes/60.0

	if dir == "S" || dir == "W" {
		decimal = -decimal
	}

	return decimal
}

// verifyChecksum verifies NMEA sentence checksum
func (e *GPSExecutor) verifyChecksum(sentence string) bool {
	idx := strings.Index(sentence, "*")
	if idx < 0 || idx+3 > len(sentence) {
		return false
	}

	// Calculate XOR of all characters between $ and *
	var checksum byte
	for i := 1; i < idx; i++ {
		checksum ^= sentence[i]
	}

	// Parse expected checksum
	expected, err := strconv.ParseUint(sentence[idx+1:idx+3], 16, 8)
	if err != nil {
		return false
	}

	return checksum == byte(expected)
}

// readRawSentence reads a single NMEA sentence
func (e *GPSExecutor) readRawSentence() (node.Message, error) {
	reader := bufio.NewReader(e.port)
	line, err := reader.ReadString('\n')
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"sentence":  strings.TrimSpace(line),
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// calculateDistance calculates distance to a target coordinate using Haversine formula
func (e *GPSExecutor) calculateDistance(targetLat, targetLon float64) (node.Message, error) {
	const earthRadius = 6371000 // meters

	lat1 := e.data.Latitude * math.Pi / 180
	lat2 := targetLat * math.Pi / 180
	deltaLat := (targetLat - e.data.Latitude) * math.Pi / 180
	deltaLon := (targetLon - e.data.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distanceM := earthRadius * c
	distanceKm := distanceM / 1000
	distanceMi := distanceM / 1609.34

	// Calculate bearing
	y := math.Sin(deltaLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(deltaLon)
	bearing := math.Atan2(y, x) * 180 / math.Pi
	if bearing < 0 {
		bearing += 360
	}

	return node.Message{
		Payload: map[string]interface{}{
			"distance_m":      distanceM,
			"distance_km":     distanceKm,
			"distance_miles":  distanceMi,
			"bearing":         bearing,
			"from_lat":        e.data.Latitude,
			"from_lon":        e.data.Longitude,
			"to_lat":          targetLat,
			"to_lon":          targetLon,
			"timestamp":       time.Now().Unix(),
		},
	}, nil
}

// configureGPS sends UBX commands to configure GPS module
func (e *GPSExecutor) configureGPS(rateHz int) (node.Message, error) {
	// UBX-CFG-RATE command to set navigation rate
	// Rate in ms: 1000/rateHz
	if rateHz < 1 || rateHz > 10 {
		rateHz = 1
	}
	rateMs := 1000 / rateHz

	// UBX message structure: sync(2) + class(1) + id(1) + length(2) + payload + checksum(2)
	// CFG-RATE: class=0x06, id=0x08, length=6
	payload := []byte{
		byte(rateMs & 0xFF), byte(rateMs >> 8), // measRate (ms)
		0x01, 0x00, // navRate (cycles)
		0x01, 0x00, // timeRef (0=UTC, 1=GPS)
	}

	msg := e.buildUBXMessage(0x06, 0x08, payload)
	_, err := e.port.Write(msg)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to configure GPS: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"action":    "configure",
			"rate_hz":   rateHz,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// buildUBXMessage builds a UBX protocol message
func (e *GPSExecutor) buildUBXMessage(class, id byte, payload []byte) []byte {
	length := len(payload)
	msg := make([]byte, 8+length)

	// Sync chars
	msg[0] = 0xB5
	msg[1] = 0x62

	// Class, ID
	msg[2] = class
	msg[3] = id

	// Length (little endian)
	msg[4] = byte(length & 0xFF)
	msg[5] = byte(length >> 8)

	// Payload
	copy(msg[6:], payload)

	// Calculate checksum (Fletcher-8)
	ckA, ckB := byte(0), byte(0)
	for i := 2; i < 6+length; i++ {
		ckA += msg[i]
		ckB += ckA
	}
	msg[6+length] = ckA
	msg[7+length] = ckB

	return msg
}

// stopReading stops the background NMEA reading
func (e *GPSExecutor) stopReading() {
	if e.running {
		select {
		case e.stopChan <- struct{}{}:
		default:
		}
	}
}

// Cleanup releases resources
func (e *GPSExecutor) Cleanup() error {
	e.stopReading()
	if e.port != nil {
		e.port.Close()
		e.port = nil
	}
	return nil
}
