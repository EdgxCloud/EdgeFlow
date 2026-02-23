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

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"go.bug.st/serial"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// BN880 combines GPS (NEO-M8N compatible) and compass (QMC5883L/HMC5883L)
const (
	// QMC5883L registers
	qmc5883lAddr         = 0x0D
	qmc5883lRegDataXLSB  = 0x00
	qmc5883lRegStatus    = 0x06
	qmc5883lRegControl1  = 0x09
	qmc5883lRegControl2  = 0x0A
	qmc5883lRegSetReset  = 0x0B

	// HMC5883L registers (alternate)
	hmc5883lAddr        = 0x1E
	hmc5883lRegConfigA  = 0x00
	hmc5883lRegConfigB  = 0x01
	hmc5883lRegMode     = 0x02
	hmc5883lRegDataXMSB = 0x03
)

// BN880Config holds configuration for BN-880 module
type BN880Config struct {
	SerialPort      string  `json:"serial_port"`
	BaudRate        int     `json:"baud_rate"`
	I2CBus          string  `json:"i2c_bus"`
	CompassAddress  uint16  `json:"compass_address"`
	CompassType     string  `json:"compass_type"` // "qmc5883l" or "hmc5883l"
	DeclinationDeg  float64 `json:"declination_deg"`
	CalibrationX    float64 `json:"calibration_x"`
	CalibrationY    float64 `json:"calibration_y"`
	CalibrationZ    float64 `json:"calibration_z"`
	PollInterval    int     `json:"poll_interval_ms"`
}

// BN880Executor implements BN-880 GPS + Compass module
type BN880Executor struct {
	config      BN880Config
	port        serial.Port
	bus         i2c.BusCloser
	compassDev  i2c.Dev
	mu          sync.Mutex
	hostInited  bool
	initialized bool
	reader      *bufio.Reader

	// GPS data
	latitude   float64
	longitude  float64
	altitude   float64
	speed      float64
	course     float64
	satellites int
	hdop       float64
	gpsValid   bool
	gpsTime    time.Time

	// Compass data
	heading    float64
	magX       int16
	magY       int16
	magZ       int16
}

func (e *BN880Executor) Init(config map[string]interface{}) error {
	e.config = BN880Config{
		SerialPort:     "/dev/ttyAMA0",
		BaudRate:       9600,
		I2CBus:         "/dev/i2c-1",
		CompassAddress: qmc5883lAddr,
		CompassType:    "qmc5883l",
		DeclinationDeg: 0,
		CalibrationX:   1.0,
		CalibrationY:   1.0,
		CalibrationZ:   1.0,
		PollInterval:   1000,
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse BN880 config: %w", err)
		}
	}

	return nil
}

func (e *BN880Executor) initHardware() error {
	if e.initialized {
		return nil
	}

	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return fmt.Errorf("failed to initialize periph host: %w", err)
		}
		e.hostInited = true
	}

	// Initialize GPS serial using go.bug.st/serial
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

	// Initialize compass I2C
	bus, err := i2creg.Open(e.config.I2CBus)
	if err != nil {
		port.Close()
		return fmt.Errorf("failed to open I2C bus %s: %w", e.config.I2CBus, err)
	}
	e.bus = bus
	e.compassDev = i2c.Dev{Bus: bus, Addr: e.config.CompassAddress}

	// Initialize compass
	if err := e.initCompass(); err != nil {
		port.Close()
		bus.Close()
		return fmt.Errorf("failed to initialize compass: %w", err)
	}

	e.initialized = true
	return nil
}

func (e *BN880Executor) initCompass() error {
	switch e.config.CompassType {
	case "qmc5883l":
		return e.initQMC5883L()
	case "hmc5883l":
		return e.initHMC5883L()
	default:
		return e.initQMC5883L()
	}
}

func (e *BN880Executor) initQMC5883L() error {
	// Set/Reset period
	if err := e.writeCompassReg(qmc5883lRegSetReset, 0x01); err != nil {
		return err
	}

	// Control register 1: continuous mode, 200Hz, 8G range, 512 OSR
	if err := e.writeCompassReg(qmc5883lRegControl1, 0x1D); err != nil {
		return err
	}

	// Control register 2: enable interrupt, disable pointer rollover
	if err := e.writeCompassReg(qmc5883lRegControl2, 0x00); err != nil {
		return err
	}

	return nil
}

func (e *BN880Executor) initHMC5883L() error {
	// Config A: 8 samples averaged, 75Hz, normal measurement
	if err := e.writeCompassReg(hmc5883lRegConfigA, 0x78); err != nil {
		return err
	}

	// Config B: Gain = 5 (1.3 Ga)
	if err := e.writeCompassReg(hmc5883lRegConfigB, 0xA0); err != nil {
		return err
	}

	// Mode: Continuous measurement
	if err := e.writeCompassReg(hmc5883lRegMode, 0x00); err != nil {
		return err
	}

	return nil
}

func (e *BN880Executor) writeCompassReg(reg, value byte) error {
	return e.compassDev.Tx([]byte{reg, value}, nil)
}

func (e *BN880Executor) readCompassReg(reg byte, length int) ([]byte, error) {
	write := []byte{reg}
	read := make([]byte, length)
	if err := e.compassDev.Tx(write, read); err != nil {
		return nil, err
	}
	return read, nil
}

func (e *BN880Executor) readCompass() error {
	var data []byte
	var err error

	switch e.config.CompassType {
	case "qmc5883l":
		data, err = e.readCompassReg(qmc5883lRegDataXLSB, 6)
		if err != nil {
			return err
		}
		// QMC5883L: LSB first
		e.magX = int16(uint16(data[0]) | uint16(data[1])<<8)
		e.magY = int16(uint16(data[2]) | uint16(data[3])<<8)
		e.magZ = int16(uint16(data[4]) | uint16(data[5])<<8)

	case "hmc5883l":
		data, err = e.readCompassReg(hmc5883lRegDataXMSB, 6)
		if err != nil {
			return err
		}
		// HMC5883L: MSB first, order is X, Z, Y
		e.magX = int16(uint16(data[0])<<8 | uint16(data[1]))
		e.magZ = int16(uint16(data[2])<<8 | uint16(data[3]))
		e.magY = int16(uint16(data[4])<<8 | uint16(data[5]))
	}

	// Apply calibration
	calX := float64(e.magX) * e.config.CalibrationX
	calY := float64(e.magY) * e.config.CalibrationY

	// Calculate heading
	heading := math.Atan2(calY, calX) * 180 / math.Pi

	// Apply magnetic declination
	heading += e.config.DeclinationDeg

	// Normalize to 0-360
	if heading < 0 {
		heading += 360
	}
	if heading >= 360 {
		heading -= 360
	}

	e.heading = heading
	return nil
}

func (e *BN880Executor) readNMEASentence() (string, error) {
	line, err := e.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (e *BN880Executor) parseNMEA(sentence string) {
	if len(sentence) < 6 || sentence[0] != '$' {
		return
	}

	// Verify checksum
	asterisk := strings.LastIndex(sentence, "*")
	if asterisk < 0 {
		return
	}

	parts := strings.Split(sentence[1:asterisk], ",")
	if len(parts) < 1 {
		return
	}

	msgType := parts[0]
	switch {
	case strings.HasSuffix(msgType, "GGA"):
		e.parseGGA(parts)
	case strings.HasSuffix(msgType, "RMC"):
		e.parseRMC(parts)
	}
}

func (e *BN880Executor) parseGGA(parts []string) {
	if len(parts) < 15 {
		return
	}

	// Latitude
	if lat, ok := e.parseLatLon(parts[2], parts[3]); ok {
		e.latitude = lat
	}

	// Longitude
	if lon, ok := e.parseLatLon(parts[4], parts[5]); ok {
		e.longitude = lon
	}

	// Fix quality
	if fix, err := strconv.Atoi(parts[6]); err == nil {
		e.gpsValid = fix > 0
	}

	// Satellites
	if sats, err := strconv.Atoi(parts[7]); err == nil {
		e.satellites = sats
	}

	// HDOP
	if hdop, err := strconv.ParseFloat(parts[8], 64); err == nil {
		e.hdop = hdop
	}

	// Altitude
	if alt, err := strconv.ParseFloat(parts[9], 64); err == nil {
		e.altitude = alt
	}
}

func (e *BN880Executor) parseRMC(parts []string) {
	if len(parts) < 12 {
		return
	}

	e.gpsValid = parts[2] == "A"

	// Latitude
	if lat, ok := e.parseLatLon(parts[3], parts[4]); ok {
		e.latitude = lat
	}

	// Longitude
	if lon, ok := e.parseLatLon(parts[5], parts[6]); ok {
		e.longitude = lon
	}

	// Speed (knots to km/h)
	if speed, err := strconv.ParseFloat(parts[7], 64); err == nil {
		e.speed = speed * 1.852
	}

	// Course
	if course, err := strconv.ParseFloat(parts[8], 64); err == nil {
		e.course = course
	}
}

func (e *BN880Executor) parseLatLon(value, direction string) (float64, bool) {
	if len(value) < 4 {
		return 0, false
	}

	var degrees, minutes float64
	var err error

	if direction == "N" || direction == "S" {
		degrees, err = strconv.ParseFloat(value[:2], 64)
		if err != nil {
			return 0, false
		}
		minutes, err = strconv.ParseFloat(value[2:], 64)
	} else {
		degrees, err = strconv.ParseFloat(value[:3], 64)
		if err != nil {
			return 0, false
		}
		minutes, err = strconv.ParseFloat(value[3:], 64)
	}
	if err != nil {
		return 0, false
	}

	result := degrees + minutes/60.0
	if direction == "S" || direction == "W" {
		result = -result
	}

	return result, true
}

func (e *BN880Executor) getCardinalDirection(heading float64) string {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	index := int((heading+11.25)/22.5) % 16
	return directions[index]
}

func (e *BN880Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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
		return e.readAll()
	case "read_gps":
		return e.readGPS()
	case "read_compass":
		return e.readCompassData()
	case "calibrate":
		return e.handleCalibrate(msg)
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *BN880Executor) readAll() (node.Message, error) {
	// Read GPS data
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

	// Read compass data
	if err := e.readCompass(); err != nil {
		return node.Message{}, fmt.Errorf("failed to read compass: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			// GPS data
			"latitude":   e.latitude,
			"longitude":  e.longitude,
			"altitude":   e.altitude,
			"speed":      e.speed,
			"course":     e.course,
			"satellites": e.satellites,
			"hdop":       e.hdop,
			"gps_valid":  e.gpsValid,
			// Compass data
			"heading":           e.heading,
			"cardinal":          e.getCardinalDirection(e.heading),
			"mag_x":             e.magX,
			"mag_y":             e.magY,
			"mag_z":             e.magZ,
			"declination":       e.config.DeclinationDeg,
			"timestamp":         time.Now().Unix(),
		},
	}, nil
}

func (e *BN880Executor) readGPS() (node.Message, error) {
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
			"latitude":   e.latitude,
			"longitude":  e.longitude,
			"altitude":   e.altitude,
			"speed":      e.speed,
			"course":     e.course,
			"satellites": e.satellites,
			"hdop":       e.hdop,
			"valid":      e.gpsValid,
			"timestamp":  time.Now().Unix(),
		},
	}, nil
}

func (e *BN880Executor) readCompassData() (node.Message, error) {
	if err := e.readCompass(); err != nil {
		return node.Message{}, fmt.Errorf("failed to read compass: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"heading":     e.heading,
			"cardinal":    e.getCardinalDirection(e.heading),
			"mag_x":       e.magX,
			"mag_y":       e.magY,
			"mag_z":       e.magZ,
			"declination": e.config.DeclinationDeg,
			"timestamp":   time.Now().Unix(),
		},
	}, nil
}

func (e *BN880Executor) handleCalibrate(msg node.Message) (node.Message, error) {
	payload := msg.Payload
	if payload == nil {
		return node.Message{}, fmt.Errorf("payload is nil")
	}

	if declination, ok := payload["declination"].(float64); ok {
		e.config.DeclinationDeg = declination
	}

	if calX, ok := payload["calibration_x"].(float64); ok {
		e.config.CalibrationX = calX
	}
	if calY, ok := payload["calibration_y"].(float64); ok {
		e.config.CalibrationY = calY
	}
	if calZ, ok := payload["calibration_z"].(float64); ok {
		e.config.CalibrationZ = calZ
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":        "calibrated",
			"declination":   e.config.DeclinationDeg,
			"calibration_x": e.config.CalibrationX,
			"calibration_y": e.config.CalibrationY,
			"calibration_z": e.config.CalibrationZ,
		},
	}, nil
}

func (e *BN880Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		if e.port != nil {
			e.port.Close()
		}
		if e.bus != nil {
			e.bus.Close()
		}
		e.initialized = false
	}
	return nil
}
