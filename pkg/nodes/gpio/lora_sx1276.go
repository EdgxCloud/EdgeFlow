// +build linux

package gpio

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/node"
)

// SX1276 Register addresses
const (
	sx1276RegFifo          = 0x00
	sx1276RegOpMode        = 0x01
	sx1276RegFrfMsb        = 0x06
	sx1276RegFrfMid        = 0x07
	sx1276RegFrfLsb        = 0x08
	sx1276RegPaConfig      = 0x09
	sx1276RegOcp           = 0x0B
	sx1276RegLna           = 0x0C
	sx1276RegFifoAddrPtr   = 0x0D
	sx1276RegFifoTxBase    = 0x0E
	sx1276RegFifoRxBase    = 0x0F
	sx1276RegFifoRxCurrent = 0x10
	sx1276RegIrqFlags      = 0x12
	sx1276RegRxNbBytes     = 0x13
	sx1276RegPktSnrValue   = 0x19
	sx1276RegPktRssiValue  = 0x1A
	sx1276RegModemConfig1  = 0x1D
	sx1276RegModemConfig2  = 0x1E
	sx1276RegSymbTimeout   = 0x1F
	sx1276RegPreambleMsb   = 0x20
	sx1276RegPreambleLsb   = 0x21
	sx1276RegPayloadLength = 0x22
	sx1276RegModemConfig3  = 0x26
	sx1276RegDioMapping1   = 0x40
	sx1276RegVersion       = 0x42
	sx1276RegPaDac         = 0x4D
)

// SX1276 Modes
const (
	sx1276ModeSleep      = 0x00
	sx1276ModeStandby    = 0x01
	sx1276ModeFSTX       = 0x02
	sx1276ModeTX         = 0x03
	sx1276ModeFSRX       = 0x04
	sx1276ModeRXCont     = 0x05
	sx1276ModeRXSingle   = 0x06
	sx1276ModeCAD        = 0x07
	sx1276ModeLongRange  = 0x80
)

// SX1276 IRQ Flags
const (
	sx1276IrqRxTimeout    = 0x80
	sx1276IrqRxDone       = 0x40
	sx1276IrqPayloadCRC   = 0x20
	sx1276IrqValidHeader  = 0x10
	sx1276IrqTxDone       = 0x08
	sx1276IrqCADDone      = 0x04
	sx1276IrqFHSSChange   = 0x02
	sx1276IrqCADDetected  = 0x01
)

// SX1276Config configuration for SX1276 LoRa module
type SX1276Config struct {
	SPIBus      int     `json:"spi_bus"`
	SPIDevice   int     `json:"spi_device"`
	Speed       int     `json:"speed"`
	ResetPin    int     `json:"reset_pin"`
	DIO0Pin     int     `json:"dio0_pin"`
	Frequency   float64 `json:"frequency"`    // MHz (433, 868, 915)
	SpreadFactor int    `json:"spread_factor"` // 6-12 (default: 7)
	Bandwidth   int     `json:"bandwidth"`     // kHz: 7.8, 10.4, 15.6, 20.8, 31.25, 41.7, 62.5, 125, 250, 500
	CodingRate  int     `json:"coding_rate"`   // 5-8 (4/5 to 4/8)
	TxPower     int     `json:"tx_power"`      // dBm (2-20, default: 17)
	SyncWord    int     `json:"sync_word"`     // 0x12 private, 0x34 public LoRaWAN
}

// SX1276Executor executes SX1276 LoRa operations
type SX1276Executor struct {
	config      SX1276Config
	hal         hal.HAL
	mu          sync.Mutex
	initialized bool
	receiving   bool
	stopChan    chan struct{}
}

// NewSX1276Executor creates a new SX1276 executor
func NewSX1276Executor(config map[string]interface{}) (node.Executor, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var loraConfig SX1276Config
	if err := json.Unmarshal(configJSON, &loraConfig); err != nil {
		return nil, fmt.Errorf("invalid SX1276 config: %w", err)
	}

	// Defaults
	if loraConfig.Speed == 0 {
		loraConfig.Speed = 10000000
	}
	if loraConfig.Frequency == 0 {
		loraConfig.Frequency = 915.0
	}
	if loraConfig.SpreadFactor == 0 {
		loraConfig.SpreadFactor = 7
	}
	if loraConfig.Bandwidth == 0 {
		loraConfig.Bandwidth = 125
	}
	if loraConfig.CodingRate == 0 {
		loraConfig.CodingRate = 5
	}
	if loraConfig.TxPower == 0 {
		loraConfig.TxPower = 17
	}
	if loraConfig.SyncWord == 0 {
		loraConfig.SyncWord = 0x12
	}

	return &SX1276Executor{
		config:   loraConfig,
		stopChan: make(chan struct{}),
	}, nil
}

// Init initializes the SX1276 executor
func (e *SX1276Executor) Init(config map[string]interface{}) error {
	return nil
}

// Execute handles LoRa operations
func (e *SX1276Executor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.hal == nil {
		h, err := hal.GetGlobalHAL()
		if err != nil {
			return node.Message{}, fmt.Errorf("HAL not initialized: %w", err)
		}
		e.hal = h
	}

	if !e.initialized {
		if err := e.initSX1276(); err != nil {
			return node.Message{}, fmt.Errorf("failed to init SX1276: %w", err)
		}
		e.initialized = true
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return node.Message{}, fmt.Errorf("invalid payload type")
	}

	action, _ := payload["action"].(string)

	switch action {
	case "send", "transmit":
		var data []byte
		if dataIface, ok := payload["data"].([]interface{}); ok {
			for _, v := range dataIface {
				if b, ok := v.(float64); ok {
					data = append(data, byte(b))
				}
			}
		} else if text, ok := payload["text"].(string); ok {
			data = []byte(text)
		}
		return e.send(data)

	case "receive":
		timeout := int(getFloat(payload, "timeout_ms", 5000))
		return e.receive(timeout)

	case "receive_start":
		go e.continuousReceive(ctx)
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "receive_start",
				"sensor":    "SX1276",
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "receive_stop":
		e.stopReceive()
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "receive_stop",
				"sensor":    "SX1276",
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "set_frequency":
		freq := getFloat(payload, "frequency", 915.0)
		e.setFrequency(freq)
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "set_frequency",
				"frequency": freq,
				"sensor":    "SX1276",
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "set_power":
		power := int(getFloat(payload, "power", 17))
		e.setTxPower(power)
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "set_power",
				"power":     power,
				"sensor":    "SX1276",
				"timestamp": time.Now().Unix(),
			},
		}, nil

	case "cad":
		return e.channelActivityDetection()

	case "status":
		return e.getStatus()

	case "sleep":
		e.setMode(sx1276ModeSleep)
		return node.Message{
			Payload: map[string]interface{}{
				"action":    "sleep",
				"sensor":    "SX1276",
				"timestamp": time.Now().Unix(),
			},
		}, nil

	default:
		return node.Message{}, fmt.Errorf("unknown action: %s", action)
	}
}

// initSX1276 initializes the SX1276
func (e *SX1276Executor) initSX1276() error {
	gpio := e.hal.GPIO()

	// Reset if pin specified
	if e.config.ResetPin > 0 {
		gpio.SetMode(e.config.ResetPin, hal.Output)
		gpio.DigitalWrite(e.config.ResetPin, false)
		time.Sleep(10 * time.Millisecond)
		gpio.DigitalWrite(e.config.ResetPin, true)
		time.Sleep(10 * time.Millisecond)
	}

	// Check version
	version, err := e.readRegister(sx1276RegVersion)
	if err != nil {
		return err
	}
	if version != 0x12 {
		return fmt.Errorf("unexpected SX1276 version: 0x%02X", version)
	}

	// Set sleep mode
	e.setMode(sx1276ModeSleep)
	time.Sleep(10 * time.Millisecond)

	// Set LoRa mode
	e.writeRegister(sx1276RegOpMode, sx1276ModeLongRange|sx1276ModeSleep)
	time.Sleep(10 * time.Millisecond)

	// Set frequency
	e.setFrequency(e.config.Frequency)

	// Set FIFO base addresses
	e.writeRegister(sx1276RegFifoTxBase, 0x00)
	e.writeRegister(sx1276RegFifoRxBase, 0x00)

	// Set LNA boost
	e.writeRegister(sx1276RegLna, 0x23)

	// Set modem config
	e.setSpreadingFactor(e.config.SpreadFactor)
	e.setBandwidth(e.config.Bandwidth)
	e.setCodingRate(e.config.CodingRate)

	// Enable AGC
	e.writeRegister(sx1276RegModemConfig3, 0x04)

	// Set TX power
	e.setTxPower(e.config.TxPower)

	// Set sync word
	e.writeRegister(0x39, byte(e.config.SyncWord))

	// Set to standby
	e.setMode(sx1276ModeStandby)

	return nil
}

// send transmits data
func (e *SX1276Executor) send(data []byte) (node.Message, error) {
	if len(data) > 255 {
		return node.Message{}, fmt.Errorf("data too long (max 255 bytes)")
	}

	// Set standby mode
	e.setMode(sx1276ModeStandby)

	// Set FIFO pointer to TX base
	e.writeRegister(sx1276RegFifoAddrPtr, 0x00)

	// Write data to FIFO
	for _, b := range data {
		e.writeRegister(sx1276RegFifo, b)
	}

	// Set payload length
	e.writeRegister(sx1276RegPayloadLength, byte(len(data)))

	// Clear IRQ flags
	e.writeRegister(sx1276RegIrqFlags, 0xFF)

	// Start TX
	e.setMode(sx1276ModeTX)

	// Wait for TX done
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			return node.Message{}, fmt.Errorf("TX timeout")
		default:
			irq, _ := e.readRegister(sx1276RegIrqFlags)
			if irq&sx1276IrqTxDone != 0 {
				e.writeRegister(sx1276RegIrqFlags, sx1276IrqTxDone)
				e.setMode(sx1276ModeStandby)
				return node.Message{
					Payload: map[string]interface{}{
						"action":    "send",
						"bytes":     len(data),
						"sensor":    "SX1276",
						"timestamp": time.Now().Unix(),
					},
				}, nil
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// receive waits for and receives data
func (e *SX1276Executor) receive(timeoutMs int) (node.Message, error) {
	// Set standby mode
	e.setMode(sx1276ModeStandby)

	// Set FIFO pointer to RX base
	e.writeRegister(sx1276RegFifoAddrPtr, 0x00)

	// Clear IRQ flags
	e.writeRegister(sx1276RegIrqFlags, 0xFF)

	// Start RX single
	e.setMode(sx1276ModeRXSingle)

	// Wait for RX done
	timeout := time.After(time.Duration(timeoutMs) * time.Millisecond)
	for {
		select {
		case <-timeout:
			e.setMode(sx1276ModeStandby)
			return node.Message{
				Payload: map[string]interface{}{
					"received":  false,
					"sensor":    "SX1276",
					"timestamp": time.Now().Unix(),
				},
			}, nil
		default:
			irq, _ := e.readRegister(sx1276RegIrqFlags)
			if irq&sx1276IrqRxDone != 0 {
				return e.processRxPacket(irq)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// processRxPacket processes a received packet
func (e *SX1276Executor) processRxPacket(irq byte) (node.Message, error) {
	// Check CRC
	crcError := irq&sx1276IrqPayloadCRC != 0

	// Get packet info
	currentAddr, _ := e.readRegister(sx1276RegFifoRxCurrent)
	rxLen, _ := e.readRegister(sx1276RegRxNbBytes)

	// Get RSSI and SNR
	pktRssi, _ := e.readRegister(sx1276RegPktRssiValue)
	pktSnr, _ := e.readRegister(sx1276RegPktSnrValue)

	rssi := -157 + int(pktRssi)
	snr := float64(int8(pktSnr)) / 4.0

	// Read data
	e.writeRegister(sx1276RegFifoAddrPtr, currentAddr)
	data := make([]byte, rxLen)
	for i := 0; i < int(rxLen); i++ {
		data[i], _ = e.readRegister(sx1276RegFifo)
	}

	// Clear IRQ
	e.writeRegister(sx1276RegIrqFlags, 0xFF)
	e.setMode(sx1276ModeStandby)

	return node.Message{
		Payload: map[string]interface{}{
			"received":  true,
			"data":      data,
			"text":      string(data),
			"length":    rxLen,
			"rssi":      rssi,
			"snr":       snr,
			"crc_error": crcError,
			"sensor":    "SX1276",
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// continuousReceive runs continuous receive mode
func (e *SX1276Executor) continuousReceive(ctx context.Context) {
	e.receiving = true
	e.setMode(sx1276ModeRXCont)

	for {
		select {
		case <-ctx.Done():
			e.receiving = false
			e.setMode(sx1276ModeStandby)
			return
		case <-e.stopChan:
			e.receiving = false
			e.setMode(sx1276ModeStandby)
			return
		default:
			e.mu.Lock()
			irq, _ := e.readRegister(sx1276RegIrqFlags)
			if irq&sx1276IrqRxDone != 0 {
				e.processRxPacket(irq)
			}
			e.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// stopReceive stops continuous receive
func (e *SX1276Executor) stopReceive() {
	if e.receiving {
		select {
		case e.stopChan <- struct{}{}:
		default:
		}
	}
}

// channelActivityDetection performs CAD
func (e *SX1276Executor) channelActivityDetection() (node.Message, error) {
	e.setMode(sx1276ModeStandby)
	e.writeRegister(sx1276RegIrqFlags, 0xFF)
	e.setMode(sx1276ModeCAD)

	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-timeout:
			return node.Message{}, fmt.Errorf("CAD timeout")
		default:
			irq, _ := e.readRegister(sx1276RegIrqFlags)
			if irq&sx1276IrqCADDone != 0 {
				detected := irq&sx1276IrqCADDetected != 0
				e.writeRegister(sx1276RegIrqFlags, 0xFF)
				e.setMode(sx1276ModeStandby)
				return node.Message{
					Payload: map[string]interface{}{
						"channel_active": detected,
						"sensor":         "SX1276",
						"timestamp":      time.Now().Unix(),
					},
				}, nil
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// getStatus returns radio status
func (e *SX1276Executor) getStatus() (node.Message, error) {
	opMode, _ := e.readRegister(sx1276RegOpMode)
	version, _ := e.readRegister(sx1276RegVersion)

	mode := "unknown"
	switch opMode & 0x07 {
	case sx1276ModeSleep:
		mode = "sleep"
	case sx1276ModeStandby:
		mode = "standby"
	case sx1276ModeTX:
		mode = "tx"
	case sx1276ModeRXCont:
		mode = "rx_continuous"
	case sx1276ModeRXSingle:
		mode = "rx_single"
	case sx1276ModeCAD:
		mode = "cad"
	}

	return node.Message{
		Payload: map[string]interface{}{
			"mode":          mode,
			"lora_mode":     opMode&sx1276ModeLongRange != 0,
			"frequency":     e.config.Frequency,
			"spread_factor": e.config.SpreadFactor,
			"bandwidth":     e.config.Bandwidth,
			"coding_rate":   e.config.CodingRate,
			"tx_power":      e.config.TxPower,
			"version":       version,
			"sensor":        "SX1276",
			"timestamp":     time.Now().Unix(),
		},
	}, nil
}

// setMode sets the operating mode
func (e *SX1276Executor) setMode(mode byte) {
	current, _ := e.readRegister(sx1276RegOpMode)
	e.writeRegister(sx1276RegOpMode, (current&0xF8)|mode)
}

// setFrequency sets the frequency in MHz
func (e *SX1276Executor) setFrequency(freqMHz float64) {
	frf := uint32((freqMHz * 1000000.0) / 61.035)
	e.writeRegister(sx1276RegFrfMsb, byte(frf>>16))
	e.writeRegister(sx1276RegFrfMid, byte(frf>>8))
	e.writeRegister(sx1276RegFrfLsb, byte(frf))
	e.config.Frequency = freqMHz
}

// setTxPower sets transmit power
func (e *SX1276Executor) setTxPower(power int) {
	if power < 2 {
		power = 2
	}
	if power > 20 {
		power = 20
	}

	if power > 17 {
		e.writeRegister(sx1276RegPaDac, 0x87) // High power mode
		e.writeRegister(sx1276RegPaConfig, 0x80|(byte(power-5)))
	} else {
		e.writeRegister(sx1276RegPaDac, 0x84) // Normal mode
		e.writeRegister(sx1276RegPaConfig, 0x80|(byte(power-2)))
	}
	e.config.TxPower = power
}

// setSpreadingFactor sets the spreading factor
func (e *SX1276Executor) setSpreadingFactor(sf int) {
	if sf < 6 {
		sf = 6
	}
	if sf > 12 {
		sf = 12
	}

	config2, _ := e.readRegister(sx1276RegModemConfig2)
	e.writeRegister(sx1276RegModemConfig2, (config2&0x0F)|(byte(sf)<<4))
	e.config.SpreadFactor = sf
}

// setBandwidth sets the bandwidth
func (e *SX1276Executor) setBandwidth(bwKHz int) {
	var bw byte
	switch bwKHz {
	case 7:
		bw = 0
	case 10:
		bw = 1
	case 15:
		bw = 2
	case 20:
		bw = 3
	case 31:
		bw = 4
	case 41:
		bw = 5
	case 62:
		bw = 6
	case 125:
		bw = 7
	case 250:
		bw = 8
	case 500:
		bw = 9
	default:
		bw = 7
	}

	config1, _ := e.readRegister(sx1276RegModemConfig1)
	e.writeRegister(sx1276RegModemConfig1, (config1&0x0F)|(bw<<4))
	e.config.Bandwidth = bwKHz
}

// setCodingRate sets the coding rate
func (e *SX1276Executor) setCodingRate(cr int) {
	if cr < 5 {
		cr = 5
	}
	if cr > 8 {
		cr = 8
	}

	config1, _ := e.readRegister(sx1276RegModemConfig1)
	e.writeRegister(sx1276RegModemConfig1, (config1&0xF1)|(byte(cr-4)<<1))
	e.config.CodingRate = cr
}

// readRegister reads a register
func (e *SX1276Executor) readRegister(reg byte) (byte, error) {
	spi := e.hal.SPI()
	data := []byte{reg & 0x7F, 0x00}
	if err := spi.Transfer(e.config.SPIBus, e.config.SPIDevice, data, 2); err != nil {
		return 0, err
	}
	return data[1], nil
}

// writeRegister writes a register
func (e *SX1276Executor) writeRegister(reg, value byte) error {
	spi := e.hal.SPI()
	return spi.Transfer(e.config.SPIBus, e.config.SPIDevice, []byte{reg | 0x80, value}, 2)
}

// Cleanup releases resources
func (e *SX1276Executor) Cleanup() error {
	e.stopReceive()
	e.setMode(sx1276ModeSleep)
	return nil
}
