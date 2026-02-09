//go:build linux
// +build linux

package wireless

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// scanDevicesLinux performs BLE scan using bluetoothctl on Linux
func (n *BLENode) scanDevices(ctx context.Context) ([]BLEDevice, error) {
	// Check if bluetoothctl is available
	if _, err := exec.LookPath("bluetoothctl"); err != nil {
		return nil, fmt.Errorf("bluetoothctl not found - install bluez package: sudo apt install bluez")
	}

	// Power on adapter
	exec.CommandContext(ctx, "bluetoothctl", "power", "on").Run()

	// Start scan with timeout
	scanCtx, cancel := context.WithTimeout(ctx, n.scanDuration)
	defer cancel()

	cmd := exec.CommandContext(scanCtx, "bluetoothctl", "--timeout", strconv.Itoa(int(n.scanDuration.Seconds())), "scan", "on")
	output, _ := cmd.CombinedOutput()

	// Parse discovered devices using bluetoothctl devices
	devCmd := exec.CommandContext(ctx, "bluetoothctl", "devices")
	devOutput, err := devCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w (scan output: %s)", err, string(output))
	}

	devices := []BLEDevice{}
	scanner := bufio.NewScanner(strings.NewReader(string(devOutput)))
	devicePattern := regexp.MustCompile(`Device\s+([0-9A-Fa-f:]{17})\s+(.*)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := devicePattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			addr := matches[1]
			name := matches[2]

			// Get RSSI via hcitool if available
			rssi := -100
			rssi = getRSSI(ctx, addr)

			devices = append(devices, BLEDevice{
				Address:     addr,
				Name:        name,
				RSSI:        rssi,
				Connectable: true,
			})
		}
	}

	return devices, nil
}

// getRSSI attempts to get RSSI for a device
func getRSSI(ctx context.Context, addr string) int {
	cmd := exec.CommandContext(ctx, "hcitool", "rssi", addr)
	output, err := cmd.Output()
	if err != nil {
		return -100
	}
	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) > 0 {
		if val, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
			return val
		}
	}
	return -100
}

// connect connects to a BLE device using bluetoothctl
func (n *BLENode) connect(ctx context.Context, address string) error {
	if address == "" {
		return fmt.Errorf("device address is required")
	}

	cmd := exec.CommandContext(ctx, "bluetoothctl", "connect", address)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w (output: %s)", address, err, string(output))
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Connected: yes") || strings.Contains(outputStr, "Connection successful") {
		n.connected = true
		return nil
	}

	return fmt.Errorf("connection to %s may have failed: %s", address, outputStr)
}

// disconnect disconnects from a BLE device
func (n *BLENode) disconnect(ctx context.Context, address string) error {
	if address == "" {
		return fmt.Errorf("device address is required")
	}

	cmd := exec.CommandContext(ctx, "bluetoothctl", "disconnect", address)
	cmd.Run()
	n.connected = false
	return nil
}

// readCharacteristic reads a BLE GATT characteristic using gatttool
func (n *BLENode) readCharacteristic(ctx context.Context, address, serviceUUID, charUUID string) ([]byte, error) {
	if address == "" || charUUID == "" {
		return nil, fmt.Errorf("device address and characteristic UUID are required")
	}

	// Use gatttool for GATT read operations
	if _, err := exec.LookPath("gatttool"); err != nil {
		return nil, fmt.Errorf("gatttool not found - install bluez package: sudo apt install bluez")
	}

	handle := charUUID
	// If UUID provided, try to use it as handle
	cmd := exec.CommandContext(ctx, "gatttool", "-b", address, "--char-read", "--uuid", handle)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to read characteristic: %w (output: %s)", err, string(output))
	}

	// Parse hex output from gatttool
	outputStr := strings.TrimSpace(string(output))
	if idx := strings.Index(outputStr, "value:"); idx >= 0 {
		hexStr := strings.TrimSpace(outputStr[idx+6:])
		return parseHexBytes(hexStr), nil
	}

	return []byte(outputStr), nil
}

// writeCharacteristic writes to a BLE GATT characteristic
func (n *BLENode) writeCharacteristic(ctx context.Context, address, serviceUUID, charUUID string, value []byte) error {
	if address == "" || charUUID == "" {
		return fmt.Errorf("device address and characteristic UUID are required")
	}

	if _, err := exec.LookPath("gatttool"); err != nil {
		return fmt.Errorf("gatttool not found - install bluez package: sudo apt install bluez")
	}

	// Convert value to hex string
	hexParts := make([]string, len(value))
	for i, b := range value {
		hexParts[i] = fmt.Sprintf("%02x", b)
	}
	hexStr := strings.Join(hexParts, "")

	cmd := exec.CommandContext(ctx, "gatttool", "-b", address, "--char-write-req", "--uuid", charUUID, "--value", hexStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write characteristic: %w (output: %s)", err, string(output))
	}

	return nil
}

// subscribeCharacteristic subscribes to BLE notifications
func (n *BLENode) subscribeCharacteristic(ctx context.Context, address, serviceUUID, charUUID string) error {
	// gatttool --listen mode for notifications
	if _, err := exec.LookPath("gatttool"); err != nil {
		return fmt.Errorf("gatttool not found - install bluez package: sudo apt install bluez")
	}

	// Start listening in background - in a real flow, this would send events
	listenCtx, _ := context.WithTimeout(ctx, n.timeout)
	cmd := exec.CommandContext(listenCtx, "gatttool", "-b", address, "--char-write-req", "--uuid", charUUID, "--value", "0100", "--listen")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Don't wait for it to complete - it runs as long as the context lives
	go func() {
		cmd.Wait()
	}()

	return nil
}

// discoverServices discovers GATT services on a connected device
func (n *BLENode) discoverServices(ctx context.Context, address string) ([]string, error) {
	if _, err := exec.LookPath("gatttool"); err != nil {
		return nil, fmt.Errorf("gatttool not found - install bluez package: sudo apt install bluez")
	}

	cmd := exec.CommandContext(ctx, "gatttool", "-b", address, "--primary")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w (output: %s)", err, string(output))
	}

	services := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	uuidPattern := regexp.MustCompile(`uuid:\s*([0-9a-fA-F-]+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := uuidPattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			services = append(services, matches[1])
		}
	}

	return services, nil
}

// discoverCharacteristics discovers characteristics of a GATT service
func (n *BLENode) discoverCharacteristics(ctx context.Context, address, serviceUUID string) ([]BLECharacteristic, error) {
	if _, err := exec.LookPath("gatttool"); err != nil {
		return nil, fmt.Errorf("gatttool not found - install bluez package: sudo apt install bluez")
	}

	cmd := exec.CommandContext(ctx, "gatttool", "-b", address, "--characteristics")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to discover characteristics: %w (output: %s)", err, string(output))
	}

	chars := []BLECharacteristic{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	charPattern := regexp.MustCompile(`uuid\s*=\s*([0-9a-fA-F-]+).*properties\s*=\s*0x([0-9a-fA-F]+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := charPattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			uuid := matches[1]
			propByte, _ := strconv.ParseUint(matches[2], 16, 8)
			props := decodeCharProperties(byte(propByte))
			chars = append(chars, BLECharacteristic{
				UUID:       uuid,
				Properties: props,
			})
		}
	}

	return chars, nil
}

// decodeCharProperties decodes GATT characteristic properties byte
func decodeCharProperties(props byte) []string {
	var result []string
	if props&0x02 != 0 {
		result = append(result, "read")
	}
	if props&0x04 != 0 {
		result = append(result, "write-no-response")
	}
	if props&0x08 != 0 {
		result = append(result, "write")
	}
	if props&0x10 != 0 {
		result = append(result, "notify")
	}
	if props&0x20 != 0 {
		result = append(result, "indicate")
	}
	return result
}

// parseHexBytes parses space-separated hex bytes
func parseHexBytes(hex string) []byte {
	parts := strings.Fields(hex)
	result := make([]byte, 0, len(parts))
	for _, p := range parts {
		if val, err := strconv.ParseUint(strings.TrimSpace(p), 16, 8); err == nil {
			result = append(result, byte(val))
		}
	}
	return result
}
