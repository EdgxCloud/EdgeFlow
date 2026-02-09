//go:build linux
// +build linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

// NRF24L01 registers
const (
	nrf24RegConfig      = 0x00
	nrf24RegEnAA        = 0x01
	nrf24RegEnRxAddr    = 0x02
	nrf24RegSetupAW     = 0x03
	nrf24RegSetupRetr   = 0x04
	nrf24RegRFCh        = 0x05
	nrf24RegRFSetup     = 0x06
	nrf24RegStatus      = 0x07
	nrf24RegObserveTx   = 0x08
	nrf24RegRPD         = 0x09
	nrf24RegRxAddrP0    = 0x0A
	nrf24RegRxAddrP1    = 0x0B
	nrf24RegRxAddrP2    = 0x0C
	nrf24RegRxAddrP3    = 0x0D
	nrf24RegRxAddrP4    = 0x0E
	nrf24RegRxAddrP5    = 0x0F
	nrf24RegTxAddr      = 0x10
	nrf24RegRxPwP0      = 0x11
	nrf24RegRxPwP1      = 0x12
	nrf24RegRxPwP2      = 0x13
	nrf24RegRxPwP3      = 0x14
	nrf24RegRxPwP4      = 0x15
	nrf24RegRxPwP5      = 0x16
	nrf24RegFIFOStatus  = 0x17
	nrf24RegDynPD       = 0x1C
	nrf24RegFeature     = 0x1D
)

// NRF24L01 commands
const (
	nrf24CmdReadReg    = 0x00
	nrf24CmdWriteReg   = 0x20
	nrf24CmdRxPayload  = 0x61
	nrf24CmdTxPayload  = 0xA0
	nrf24CmdFlushTx    = 0xE1
	nrf24CmdFlushRx    = 0xE2
	nrf24CmdReuseTx    = 0xE3
	nrf24CmdNOP        = 0xFF
)

// NRF24L01 config bits
const (
	nrf24ConfigMaskRxDR  = 0x40
	nrf24ConfigMaskTxDS  = 0x20
	nrf24ConfigMaskMaxRT = 0x10
	nrf24ConfigEnCRC     = 0x08
	nrf24ConfigCRCO      = 0x04
	nrf24ConfigPwrUp     = 0x02
	nrf24ConfigPrimRX    = 0x01
)

// NRF24L01 status bits
const (
	nrf24StatusRxDR   = 0x40
	nrf24StatusTxDS   = 0x20
	nrf24StatusMaxRT  = 0x10
	nrf24StatusTxFull = 0x01
)

// NRF24L01Config holds configuration for NRF24L01 module
type NRF24L01Config struct {
	SPIBus       string   `json:"spi_bus"`
	SPIDevice    int      `json:"spi_device"`
	SPISpeed     int64    `json:"spi_speed"`
	CEPin        string   `json:"ce_pin"`
	Channel      int      `json:"channel"` // 0-125
	DataRate     string   `json:"data_rate"` // "250K", "1M", "2M"
	PALevel      string   `json:"pa_level"` // "MIN", "LOW", "HIGH", "MAX"
	AddressWidth int      `json:"address_width"` // 3, 4, or 5
	PayloadSize  int      `json:"payload_size"` // 1-32
	AutoAck      bool     `json:"auto_ack"`
	CRC          int      `json:"crc"` // 0, 1, or 2 bytes
	TxAddress    []byte   `json:"tx_address"`
	RxAddresses  [][]byte `json:"rx_addresses"`
}

// NRF24L01Executor implements 2.4GHz wireless transceiver
type NRF24L01Executor struct {
	config      NRF24L01Config
	spiPort     spi.PortCloser
	spiConn     spi.Conn
	cePin       gpio.PinIO
	mu          sync.Mutex
	hostInited  bool
	initialized bool
}

func (e *NRF24L01Executor) Init(config map[string]interface{}) error {
	e.config = NRF24L01Config{
		SPIBus:       "/dev/spidev0.0",
		SPIDevice:    0,
		SPISpeed:     10000000, // 10MHz
		CEPin:        "GPIO22",
		Channel:      76,
		DataRate:     "1M",
		PALevel:      "MAX",
		AddressWidth: 5,
		PayloadSize:  32,
		AutoAck:      true,
		CRC:          2,
		TxAddress:    []byte{0xE7, 0xE7, 0xE7, 0xE7, 0xE7},
		RxAddresses:  [][]byte{{0xE7, 0xE7, 0xE7, 0xE7, 0xE7}},
	}

	if config != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(configJSON, &e.config); err != nil {
			return fmt.Errorf("failed to parse NRF24L01 config: %w", err)
		}
	}

	return nil
}

func (e *NRF24L01Executor) initHardware() error {
	if e.initialized {
		return nil
	}

	if !e.hostInited {
		if _, err := host.Init(); err != nil {
			return fmt.Errorf("failed to initialize periph host: %w", err)
		}
		e.hostInited = true
	}

	// Initialize SPI
	port, err := spireg.Open(e.config.SPIBus)
	if err != nil {
		return fmt.Errorf("failed to open SPI: %w", err)
	}
	e.spiPort = port

	conn, err := port.Connect(spi.Mode0, 8, e.config.SPISpeed)
	if err != nil {
		port.Close()
		return fmt.Errorf("failed to configure SPI: %w", err)
	}
	e.spiConn = conn

	// Initialize CE pin
	cePin := gpioreg.ByName(e.config.CEPin)
	if cePin == nil {
		port.Close()
		return fmt.Errorf("failed to find CE pin %s", e.config.CEPin)
	}
	if err := cePin.Out(gpio.Low); err != nil {
		port.Close()
		return fmt.Errorf("failed to configure CE pin: %w", err)
	}
	e.cePin = cePin

	// Configure the radio
	if err := e.configureRadio(); err != nil {
		port.Close()
		return fmt.Errorf("failed to configure radio: %w", err)
	}

	e.initialized = true
	return nil
}

func (e *NRF24L01Executor) readRegister(reg byte) (byte, error) {
	write := []byte{nrf24CmdReadReg | reg, 0xFF}
	read := make([]byte, 2)
	if err := e.spiConn.Tx(write, read); err != nil {
		return 0, err
	}
	return read[1], nil
}

func (e *NRF24L01Executor) readRegisters(reg byte, length int) ([]byte, error) {
	write := make([]byte, length+1)
	write[0] = nrf24CmdReadReg | reg
	for i := 1; i < len(write); i++ {
		write[i] = 0xFF
	}
	read := make([]byte, length+1)
	if err := e.spiConn.Tx(write, read); err != nil {
		return nil, err
	}
	return read[1:], nil
}

func (e *NRF24L01Executor) writeRegister(reg, value byte) error {
	write := []byte{nrf24CmdWriteReg | reg, value}
	return e.spiConn.Tx(write, nil)
}

func (e *NRF24L01Executor) writeRegisters(reg byte, values []byte) error {
	write := append([]byte{nrf24CmdWriteReg | reg}, values...)
	return e.spiConn.Tx(write, nil)
}

func (e *NRF24L01Executor) sendCommand(cmd byte) (byte, error) {
	write := []byte{cmd}
	read := make([]byte, 1)
	if err := e.spiConn.Tx(write, read); err != nil {
		return 0, err
	}
	return read[0], nil
}

func (e *NRF24L01Executor) configureRadio() error {
	// Power down first
	e.cePin.Out(gpio.Low)
	time.Sleep(5 * time.Millisecond)

	// Set address width (3, 4, or 5 bytes)
	if err := e.writeRegister(nrf24RegSetupAW, byte(e.config.AddressWidth-2)); err != nil {
		return err
	}

	// Set channel
	if err := e.writeRegister(nrf24RegRFCh, byte(e.config.Channel)); err != nil {
		return err
	}

	// Set RF setup (data rate and PA level)
	var rfSetup byte = 0
	switch e.config.DataRate {
	case "250K":
		rfSetup |= 0x20 // RF_DR_LOW = 1
	case "2M":
		rfSetup |= 0x08 // RF_DR_HIGH = 1
	default: // "1M"
		// Both bits 0
	}
	switch e.config.PALevel {
	case "MIN":
		// RF_PWR = 00
	case "LOW":
		rfSetup |= 0x02
	case "HIGH":
		rfSetup |= 0x04
	default: // "MAX"
		rfSetup |= 0x06
	}
	if err := e.writeRegister(nrf24RegRFSetup, rfSetup); err != nil {
		return err
	}

	// Set retransmit (15 retries, 500us delay)
	if err := e.writeRegister(nrf24RegSetupRetr, 0x1F); err != nil {
		return err
	}

	// Enable auto-ack on pipe 0
	if e.config.AutoAck {
		if err := e.writeRegister(nrf24RegEnAA, 0x01); err != nil {
			return err
		}
	} else {
		if err := e.writeRegister(nrf24RegEnAA, 0x00); err != nil {
			return err
		}
	}

	// Enable RX pipe 0
	if err := e.writeRegister(nrf24RegEnRxAddr, 0x01); err != nil {
		return err
	}

	// Set payload size for pipe 0
	if err := e.writeRegister(nrf24RegRxPwP0, byte(e.config.PayloadSize)); err != nil {
		return err
	}

	// Set TX address
	if len(e.config.TxAddress) > 0 {
		if err := e.writeRegisters(nrf24RegTxAddr, e.config.TxAddress); err != nil {
			return err
		}
		// Also set RX address pipe 0 for auto-ack
		if err := e.writeRegisters(nrf24RegRxAddrP0, e.config.TxAddress); err != nil {
			return err
		}
	}

	// Configure: enable CRC, power up, PTX mode
	var configReg byte = nrf24ConfigPwrUp
	if e.config.CRC > 0 {
		configReg |= nrf24ConfigEnCRC
		if e.config.CRC == 2 {
			configReg |= nrf24ConfigCRCO
		}
	}
	if err := e.writeRegister(nrf24RegConfig, configReg); err != nil {
		return err
	}

	// Power up delay
	time.Sleep(5 * time.Millisecond)

	// Flush FIFOs
	e.sendCommand(nrf24CmdFlushTx)
	e.sendCommand(nrf24CmdFlushRx)

	// Clear status flags
	e.writeRegister(nrf24RegStatus, nrf24StatusRxDR|nrf24StatusTxDS|nrf24StatusMaxRT)

	return nil
}

func (e *NRF24L01Executor) powerUp(rxMode bool) error {
	config, err := e.readRegister(nrf24RegConfig)
	if err != nil {
		return err
	}

	config |= nrf24ConfigPwrUp
	if rxMode {
		config |= nrf24ConfigPrimRX
	} else {
		config &= ^byte(nrf24ConfigPrimRX)
	}

	return e.writeRegister(nrf24RegConfig, config)
}

func (e *NRF24L01Executor) transmit(data []byte) (bool, error) {
	if len(data) > 32 {
		return false, fmt.Errorf("payload too large: %d bytes (max 32)", len(data))
	}

	// Pad data to payload size
	payload := make([]byte, e.config.PayloadSize)
	copy(payload, data)

	// Ensure TX mode
	if err := e.powerUp(false); err != nil {
		return false, err
	}

	// Clear status flags
	e.writeRegister(nrf24RegStatus, nrf24StatusRxDR|nrf24StatusTxDS|nrf24StatusMaxRT)

	// Flush TX FIFO
	e.sendCommand(nrf24CmdFlushTx)

	// Write payload
	write := append([]byte{nrf24CmdTxPayload}, payload...)
	if err := e.spiConn.Tx(write, nil); err != nil {
		return false, err
	}

	// Pulse CE to transmit
	e.cePin.Out(gpio.High)
	time.Sleep(15 * time.Microsecond)
	e.cePin.Out(gpio.Low)

	// Wait for transmission complete
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return false, fmt.Errorf("transmission timeout")
		default:
			status, err := e.readRegister(nrf24RegStatus)
			if err != nil {
				return false, err
			}

			if status&nrf24StatusTxDS != 0 {
				// Success
				e.writeRegister(nrf24RegStatus, nrf24StatusTxDS)
				return true, nil
			}

			if status&nrf24StatusMaxRT != 0 {
				// Max retries reached
				e.writeRegister(nrf24RegStatus, nrf24StatusMaxRT)
				e.sendCommand(nrf24CmdFlushTx)
				return false, nil
			}

			time.Sleep(100 * time.Microsecond)
		}
	}
}

func (e *NRF24L01Executor) receive(timeoutMs int) ([]byte, bool, error) {
	// Enable RX mode
	if err := e.powerUp(true); err != nil {
		return nil, false, err
	}

	// Start listening
	e.cePin.Out(gpio.High)
	defer e.cePin.Out(gpio.Low)

	// Wait for data
	timeout := time.After(time.Duration(timeoutMs) * time.Millisecond)
	for {
		select {
		case <-timeout:
			return nil, false, nil // No data received
		default:
			status, err := e.readRegister(nrf24RegStatus)
			if err != nil {
				return nil, false, err
			}

			if status&nrf24StatusRxDR != 0 {
				// Data ready
				e.writeRegister(nrf24RegStatus, nrf24StatusRxDR)

				// Read payload
				write := make([]byte, e.config.PayloadSize+1)
				write[0] = nrf24CmdRxPayload
				for i := 1; i < len(write); i++ {
					write[i] = 0xFF
				}
				read := make([]byte, e.config.PayloadSize+1)
				if err := e.spiConn.Tx(write, read); err != nil {
					return nil, false, err
				}

				return read[1:], true, nil
			}

			time.Sleep(100 * time.Microsecond)
		}
	}
}

func (e *NRF24L01Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.initHardware(); err != nil {
		return node.Message{}, err
	}

	action := "status"
	if payload, ok := msg.Payload.(map[string]interface{}); ok {
		if a, ok := payload["action"].(string); ok {
			action = a
		}
	}

	switch action {
	case "send":
		return e.handleSend(msg)
	case "receive":
		return e.handleReceive(msg)
	case "configure":
		return e.handleConfigure(msg)
	case "set_address":
		return e.handleSetAddress(msg)
	case "status":
		return e.handleStatus()
	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *NRF24L01Executor) handleSend(msg node.Message) (node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	var data []byte

	if d, ok := payload["data"].(string); ok {
		data = []byte(d)
	} else if d, ok := payload["data"].([]interface{}); ok {
		data = make([]byte, len(d))
		for i, v := range d {
			if b, ok := v.(float64); ok {
				data[i] = byte(b)
			}
		}
	} else {
		return node.Message{}, fmt.Errorf("data required (string or byte array)")
	}

	success, err := e.transmit(data)
	if err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":    "sent",
			"success":   success,
			"data_size": len(data),
			"channel":   e.config.Channel,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

func (e *NRF24L01Executor) handleReceive(msg node.Message) (node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		payload = make(map[string]interface{})
	}

	timeoutMs := 1000
	if t, ok := payload["timeout_ms"].(float64); ok {
		timeoutMs = int(t)
	}

	data, received, err := e.receive(timeoutMs)
	if err != nil {
		return node.Message{}, err
	}

	result := map[string]interface{}{
		"received":  received,
		"channel":   e.config.Channel,
		"timestamp": time.Now().Unix(),
	}

	if received {
		result["data"] = data
		result["data_string"] = string(data)
		result["data_size"] = len(data)
	}

	return node.Message{
		Payload: result,
	}, nil
}

func (e *NRF24L01Executor) handleConfigure(msg node.Message) (node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	if channel, ok := payload["channel"].(float64); ok {
		e.config.Channel = int(channel)
		e.writeRegister(nrf24RegRFCh, byte(channel))
	}

	if dataRate, ok := payload["data_rate"].(string); ok {
		e.config.DataRate = dataRate
	}

	if paLevel, ok := payload["pa_level"].(string); ok {
		e.config.PALevel = paLevel
	}

	if payloadSize, ok := payload["payload_size"].(float64); ok {
		e.config.PayloadSize = int(payloadSize)
		e.writeRegister(nrf24RegRxPwP0, byte(payloadSize))
	}

	// Reconfigure radio with new settings
	if err := e.configureRadio(); err != nil {
		return node.Message{}, err
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":       "configured",
			"channel":      e.config.Channel,
			"data_rate":    e.config.DataRate,
			"pa_level":     e.config.PALevel,
			"payload_size": e.config.PayloadSize,
		},
	}, nil
}

func (e *NRF24L01Executor) handleSetAddress(msg node.Message) (node.Message, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	if tx, ok := payload["tx_address"].([]interface{}); ok {
		addr := make([]byte, len(tx))
		for i, v := range tx {
			if b, ok := v.(float64); ok {
				addr[i] = byte(b)
			}
		}
		e.config.TxAddress = addr
		e.writeRegisters(nrf24RegTxAddr, addr)
		e.writeRegisters(nrf24RegRxAddrP0, addr)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"status":     "address_set",
			"tx_address": e.config.TxAddress,
		},
	}, nil
}

func (e *NRF24L01Executor) handleStatus() (node.Message, error) {
	status, _ := e.readRegister(nrf24RegStatus)
	config, _ := e.readRegister(nrf24RegConfig)
	fifoStatus, _ := e.readRegister(nrf24RegFIFOStatus)
	observeTx, _ := e.readRegister(nrf24RegObserveTx)

	return node.Message{
		Payload: map[string]interface{}{
			"channel":        e.config.Channel,
			"data_rate":      e.config.DataRate,
			"pa_level":       e.config.PALevel,
			"payload_size":   e.config.PayloadSize,
			"auto_ack":       e.config.AutoAck,
			"status_reg":     fmt.Sprintf("0x%02X", status),
			"config_reg":     fmt.Sprintf("0x%02X", config),
			"fifo_status":    fmt.Sprintf("0x%02X", fifoStatus),
			"tx_full":        status&nrf24StatusTxFull != 0,
			"rx_ready":       status&nrf24StatusRxDR != 0,
			"tx_success":     status&nrf24StatusTxDS != 0,
			"max_retries":    status&nrf24StatusMaxRT != 0,
			"plos_cnt":       (observeTx >> 4) & 0x0F,
			"arc_cnt":        observeTx & 0x0F,
			"power_up":       config&nrf24ConfigPwrUp != 0,
			"rx_mode":        config&nrf24ConfigPrimRX != 0,
			"initialized":    e.initialized,
		},
	}, nil
}

func (e *NRF24L01Executor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		// Power down
		e.cePin.Out(gpio.Low)
		config, _ := e.readRegister(nrf24RegConfig)
		e.writeRegister(nrf24RegConfig, config&^nrf24ConfigPwrUp)

		if e.spiPort != nil {
			e.spiPort.Close()
		}
		if e.cePin != nil {
			e.cePin.Halt()
		}
		e.initialized = false
	}
	return nil
}
