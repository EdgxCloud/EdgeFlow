//go:build !linux
// +build !linux

package wireless

import (
	"context"
	"fmt"
)

// scanDevices stub for non-Linux platforms
func (n *BLENode) scanDevices(ctx context.Context) ([]BLEDevice, error) {
	return nil, fmt.Errorf("BLE scanning requires Linux with bluez - not available on this platform")
}

// connect stub for non-Linux platforms
func (n *BLENode) connect(ctx context.Context, address string) error {
	return fmt.Errorf("BLE connection requires Linux with bluez - not available on this platform")
}

// disconnect stub for non-Linux platforms
func (n *BLENode) disconnect(ctx context.Context, address string) error {
	n.connected = false
	return nil
}

// readCharacteristic stub for non-Linux platforms
func (n *BLENode) readCharacteristic(ctx context.Context, address, serviceUUID, charUUID string) ([]byte, error) {
	return nil, fmt.Errorf("BLE read requires Linux with bluez - not available on this platform")
}

// writeCharacteristic stub for non-Linux platforms
func (n *BLENode) writeCharacteristic(ctx context.Context, address, serviceUUID, charUUID string, value []byte) error {
	return fmt.Errorf("BLE write requires Linux with bluez - not available on this platform")
}

// subscribeCharacteristic stub for non-Linux platforms
func (n *BLENode) subscribeCharacteristic(ctx context.Context, address, serviceUUID, charUUID string) error {
	return fmt.Errorf("BLE subscribe requires Linux with bluez - not available on this platform")
}

// discoverServices stub for non-Linux platforms
func (n *BLENode) discoverServices(ctx context.Context, address string) ([]string, error) {
	return nil, fmt.Errorf("BLE service discovery requires Linux with bluez - not available on this platform")
}

// discoverCharacteristics stub for non-Linux platforms
func (n *BLENode) discoverCharacteristics(ctx context.Context, address, serviceUUID string) ([]BLECharacteristic, error) {
	return nil, fmt.Errorf("BLE characteristic discovery requires Linux with bluez - not available on this platform")
}
