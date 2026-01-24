package hal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// OneWireProvider defines the interface for 1-Wire protocol communication
type OneWireProvider interface {
	// ScanDevices scans for 1-Wire devices on the bus
	ScanDevices() ([]OneWireDevice, error)

	// ReadDevice reads data from a specific 1-Wire device
	ReadDevice(deviceID string) ([]byte, error)

	// WriteDevice writes data to a specific 1-Wire device
	WriteDevice(deviceID string, data []byte) error

	// GetDeviceInfo retrieves information about a device
	GetDeviceInfo(deviceID string) (*OneWireDevice, error)

	// Close closes the 1-Wire bus
	Close() error
}

// OneWireDevice represents a 1-Wire device
type OneWireDevice struct {
	ID       string // Device unique ID (e.g., "28-00000123abcd")
	Family   string // Device family code
	Type     string // Device type (e.g., "DS18B20")
	Path     string // System path to device
}

// LinuxOneWire implements OneWireProvider using Linux's w1 kernel module
// Devices are accessed via /sys/bus/w1/devices/
type LinuxOneWire struct {
	basePath    string
	devices     map[string]*OneWireDevice
	mu          sync.RWMutex
	lastScan    time.Time
	scanInterval time.Duration
}

// NewLinuxOneWire creates a new Linux 1-Wire provider
func NewLinuxOneWire() (*LinuxOneWire, error) {
	basePath := "/sys/bus/w1/devices"

	// Check if w1 kernel module is loaded
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, errors.New("1-Wire kernel module not loaded. Run: modprobe w1-gpio")
	}

	return &LinuxOneWire{
		basePath:     basePath,
		devices:      make(map[string]*OneWireDevice),
		scanInterval: 5 * time.Second,
	}, nil
}

// ScanDevices scans for all available 1-Wire devices
func (w *LinuxOneWire) ScanDevices() ([]OneWireDevice, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Rate limit scans
	if time.Since(w.lastScan) < w.scanInterval {
		// Return cached results
		result := make([]OneWireDevice, 0, len(w.devices))
		for _, dev := range w.devices {
			result = append(result, *dev)
		}
		return result, nil
	}

	entries, err := os.ReadDir(w.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read 1-Wire devices: %w", err)
	}

	devices := []OneWireDevice{}
	w.devices = make(map[string]*OneWireDevice)

	for _, entry := range entries {
		name := entry.Name()

		// Skip master controller (w1_bus_master1, etc.)
		if strings.HasPrefix(name, "w1_bus_master") {
			continue
		}

		// Valid device IDs are in format: XX-XXXXXXXXXXXX
		// where XX is the family code
		if len(name) < 3 || name[2] != '-' {
			continue
		}

		family := name[:2]
		deviceType := w.getFamilyType(family)

		device := &OneWireDevice{
			ID:     name,
			Family: family,
			Type:   deviceType,
			Path:   filepath.Join(w.basePath, name),
		}

		devices = append(devices, *device)
		w.devices[name] = device
	}

	w.lastScan = time.Now()

	return devices, nil
}

// ReadDevice reads data from a 1-Wire device
func (w *LinuxOneWire) ReadDevice(deviceID string) ([]byte, error) {
	w.mu.RLock()
	device, exists := w.devices[deviceID]
	w.mu.RUnlock()

	if !exists {
		// Try to rescan
		if _, err := w.ScanDevices(); err != nil {
			return nil, err
		}

		w.mu.RLock()
		device, exists = w.devices[deviceID]
		w.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("device %s not found", deviceID)
		}
	}

	// Read from w1_slave file (standard for most 1-Wire sensors)
	dataPath := filepath.Join(device.Path, "w1_slave")
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read device %s: %w", deviceID, err)
	}

	return data, nil
}

// WriteDevice writes data to a 1-Wire device
func (w *LinuxOneWire) WriteDevice(deviceID string, data []byte) error {
	w.mu.RLock()
	device, exists := w.devices[deviceID]
	w.mu.RUnlock()

	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	// Most 1-Wire sensors are read-only from Linux's perspective
	// Write operations are device-specific
	writePath := filepath.Join(device.Path, "output")
	if err := os.WriteFile(writePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write to device %s: %w", deviceID, err)
	}

	return nil
}

// GetDeviceInfo retrieves information about a specific device
func (w *LinuxOneWire) GetDeviceInfo(deviceID string) (*OneWireDevice, error) {
	w.mu.RLock()
	device, exists := w.devices[deviceID]
	w.mu.RUnlock()

	if !exists {
		// Try to rescan
		if _, err := w.ScanDevices(); err != nil {
			return nil, err
		}

		w.mu.RLock()
		device, exists = w.devices[deviceID]
		w.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("device %s not found", deviceID)
		}
	}

	return device, nil
}

// Close closes the 1-Wire provider
func (w *LinuxOneWire) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.devices = make(map[string]*OneWireDevice)
	return nil
}

// getFamilyType returns the device type based on family code
func (w *LinuxOneWire) getFamilyType(family string) string {
	familyTypes := map[string]string{
		"10": "DS18S20",  // Temperature sensor
		"28": "DS18B20",  // Temperature sensor
		"3b": "DS1825",   // Temperature sensor
		"22": "DS1822",   // Temperature sensor
		"42": "DS28EA00", // Temperature sensor with alarm
		"01": "DS2401",   // Silicon serial number
		"1D": "DS2423",   // 4Kb RAM with counter
		"26": "DS2438",   // Smart battery monitor
		"12": "DS2406",   // Dual addressable switch
		"29": "DS2408",   // 8-Channel addressable switch
		"3A": "DS2413",   // Dual channel addressable switch
	}

	if deviceType, ok := familyTypes[family]; ok {
		return deviceType
	}

	return fmt.Sprintf("Unknown (0x%s)", family)
}

// MockOneWire implements a mock 1-Wire provider for testing
type MockOneWire struct {
	devices map[string]*OneWireDevice
	data    map[string][]byte
	mu      sync.RWMutex
}

// NewMockOneWire creates a new mock 1-Wire provider
func NewMockOneWire() *MockOneWire {
	return &MockOneWire{
		devices: make(map[string]*OneWireDevice),
		data:    make(map[string][]byte),
	}
}

// AddDevice adds a mock device
func (m *MockOneWire) AddDevice(id, family, deviceType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.devices[id] = &OneWireDevice{
		ID:     id,
		Family: family,
		Type:   deviceType,
		Path:   fmt.Sprintf("/sys/bus/w1/devices/%s", id),
	}
}

// SetDeviceData sets the data that will be returned when reading a device
func (m *MockOneWire) SetDeviceData(id string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[id] = data
}

// ScanDevices returns all mock devices
func (m *MockOneWire) ScanDevices() ([]OneWireDevice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]OneWireDevice, 0, len(m.devices))
	for _, dev := range m.devices {
		result = append(result, *dev)
	}

	return result, nil
}

// ReadDevice returns mock data for a device
func (m *MockOneWire) ReadDevice(deviceID string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.devices[deviceID]; !exists {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}

	if data, exists := m.data[deviceID]; exists {
		return data, nil
	}

	// Return default DS18B20-like data
	return []byte("73 01 4b 46 7f ff 0c 10 1c : crc=1c YES\n73 01 4b 46 7f ff 0c 10 1c t=23187"), nil
}

// WriteDevice stores data for a mock device
func (m *MockOneWire) WriteDevice(deviceID string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.devices[deviceID]; !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	m.data[deviceID] = data
	return nil
}

// GetDeviceInfo returns information about a mock device
func (m *MockOneWire) GetDeviceInfo(deviceID string) (*OneWireDevice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if device, exists := m.devices[deviceID]; exists {
		return device, nil
	}

	return nil, fmt.Errorf("device %s not found", deviceID)
}

// Close closes the mock provider
func (m *MockOneWire) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.devices = make(map[string]*OneWireDevice)
	m.data = make(map[string][]byte)
	return nil
}
