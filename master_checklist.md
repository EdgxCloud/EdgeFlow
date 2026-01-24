# Raspberry Pi Hardware Support - Master Checklist

## Overview
This checklist tracks all Raspberry Pi hardware functionality implementation status in EdgeFlow.

**Last Updated:** 2026-01-23
**Target Platform:** Raspberry Pi 5 (ARM64) + Edge Computing

---

## Core Hardware Interfaces

### GPIO (General Purpose I/O)
| Feature | Status | File | Notes |
|---------|--------|------|-------|
| Digital Input | ✅ Done | `pkg/nodes/gpio/gpio_in.go` | With debounce, edge detection |
| Digital Output | ✅ Done | `pkg/nodes/gpio/gpio_out.go` | With invert, persistence |
| PWM Output | ✅ Done | `pkg/nodes/gpio/pwm.go` | General, servo, LED modes |
| Edge Detection | ✅ Done | `pkg/nodes/gpio/gpio_in.go` | Rising, falling, both |
| Pull-up/Pull-down | ✅ Done | `internal/hal/hal.go` | Configurable resistors |

### Communication Protocols
| Feature | Status | File | Notes |
|---------|--------|------|-------|
| I2C Read/Write | ✅ Done | `pkg/nodes/gpio/i2c.go` | Register-level access |
| SPI Transfer | ✅ Done | `pkg/nodes/gpio/spi.go` | Full duplex, modes 0-3 |
| Serial/UART | ✅ Done | `pkg/nodes/gpio/serial.go` | 9600-115200 baud |
| 1-Wire | ✅ Done | `internal/hal/onewire.go` | Linux sysfs interface |

---

## Temperature Sensors

| Sensor | Status | File | Interface | Notes |
|--------|--------|------|-----------|-------|
| DS18B20 | ✅ Done | `pkg/nodes/gpio/ds18b20.go` | 1-Wire | Digital temp sensor |
| DHT11/DHT22 | ✅ Done | `pkg/nodes/gpio/dht.go` | GPIO | Kernel driver + bit-bang |
| BMP280 | ✅ Done | `pkg/nodes/gpio/bmp280.go` | I2C | Pressure + Temp |
| BME280 | ✅ Done | `pkg/nodes/gpio/bme280.go` | I2C | Pressure + Temp + Humidity + Altitude |
| BME680 | ✅ Done | `pkg/nodes/gpio/bme680.go` | I2C | Air Quality (VOC) + IAQ |
| AHT20 | ✅ Done | `pkg/nodes/gpio/aht20.go` | I2C | Humidity + Temp |
| SHT3x | ✅ Done | `pkg/nodes/gpio/sht3x.go` | I2C | Humidity + Temp |
| MAX31855 | ✅ Done | `pkg/nodes/gpio/max31855.go` | SPI | Thermocouple + fault detection |
| MAX31865 | ✅ Done | `pkg/nodes/gpio/max31865.go` | SPI | RTD (PT100/PT1000) |

---

## Environmental Sensors

| Sensor | Status | File | Interface | Notes |
|--------|--------|------|-----------|-------|
| BH1750 | ✅ Done | `pkg/nodes/gpio/bh1750.go` | I2C | Light intensity (lux) |
| TSL2561 | ✅ Done | `pkg/nodes/gpio/tsl2561.go` | I2C | Advanced lux sensor |
| VEML7700 | ✅ Done | `pkg/nodes/gpio/veml7700.go` | I2C | High accuracy lux, auto-gain |
| CCS811 | ✅ Done | `pkg/nodes/gpio/ccs811.go` | I2C | CO2 + TVOC, baseline |
| SGP30 | ✅ Done | `pkg/nodes/gpio/sgp30.go` | I2C | Air quality (eCO2, TVOC), self-test |
| MQ-x Series | ❌ TODO | - | Analog/ADC | Gas sensors (via ADC) |

---

## Distance & Motion Sensors

| Sensor | Status | File | Interface | Notes |
|--------|--------|------|-----------|-------|
| HC-SR04 | ✅ Done | `pkg/nodes/gpio/hcsr04.go` | GPIO | Ultrasonic (Linux only) |
| VL53L0X | ✅ Done | `pkg/nodes/gpio/vl53l0x.go` | I2C | Laser ToF distance |
| VL53L1X | ✅ Done | `pkg/nodes/gpio/vl53l1x.go` | I2C | Long range ToF (4m) |
| PIR (HC-SR501) | ✅ Done | `pkg/nodes/gpio/pir.go` | GPIO | Motion (Linux only) |
| RCWL-0516 | ✅ Done | `pkg/nodes/gpio/rcwl0516.go` | GPIO | Microwave motion, debounce |

---

## ADC (Analog to Digital)

| Device | Status | File | Interface | Notes |
|--------|--------|------|-----------|-------|
| ADS1115 | ✅ Done | `pkg/nodes/gpio/voltage_monitor.go` | I2C | 16-bit, 4 channels |
| ADS1015 | ✅ Done | `pkg/nodes/gpio/ads1015.go` | I2C | 12-bit, 4 channels, differential |
| MCP3008 | ✅ Done | `pkg/nodes/gpio/mcp3008.go` | SPI | 10-bit, 8 channels |
| MCP3208 | ✅ Done | `pkg/nodes/gpio/mcp3008.go` | SPI | 12-bit, 8 channels |
| PCF8591 | ✅ Done | `pkg/nodes/gpio/pcf8591.go` | I2C | 8-bit ADC + DAC |

---

## Output Devices

| Device | Status | File | Interface | Notes |
|--------|--------|------|-----------|-------|
| Relay | ✅ Done | `pkg/nodes/gpio/relay.go` | GPIO | On/Off/Toggle/Pulse |
| LED Strip (WS2812) | ✅ Done | `pkg/nodes/gpio/ws2812.go` | GPIO | Addressable RGB + Effects |
| Servo Motor | ✅ Done | `pkg/nodes/gpio/pwm.go` | PWM | Via PWM node |
| DC Motor (L298N) | ✅ Done | `pkg/nodes/gpio/motor_l298n.go` | GPIO+PWM | H-Bridge driver |
| Stepper Motor | ✅ Done | `pkg/nodes/gpio/motor_l298n.go` | GPIO | A4988/DRV8825 |
| Buzzer | ✅ Done | `pkg/nodes/gpio/buzzer.go` | GPIO/PWM | Tones/Melodies/Patterns |
| LCD I2C (16x2) | ✅ Done | `pkg/nodes/gpio/lcd_i2c.go` | I2C | HD44780 + PCF8574 |
| OLED (SSD1306) | ✅ Done | `pkg/nodes/gpio/oled_ssd1306.go` | I2C | 128x64 display |

---

## Industrial Protocols

| Protocol | Status | File | Interface | Notes |
|----------|--------|------|-----------|-------|
| Modbus RTU | ✅ Done | `pkg/nodes/gpio/modbus.go` | Serial | RS485 support |
| Modbus TCP | ✅ Done | `pkg/nodes/gpio/modbus.go` | Ethernet | TCP/IP |
| CAN Bus | ✅ Done | `pkg/nodes/gpio/can_mcp2515.go` | SPI (MCP2515) | Send/Receive/Filters |
| RS485 | ✅ Done | `pkg/nodes/gpio/modbus.go` | Serial | Via Modbus RTU |

---

## Wireless Communication

| Feature | Status | File | Interface | Notes |
|---------|--------|------|-----------|-------|
| RF 433MHz | ✅ Done | `pkg/nodes/gpio/rf433.go` | GPIO | TX/RX, multi-protocol |
| NRF24L01 | ✅ Done | `pkg/nodes/gpio/nrf24l01.go` | SPI | 2.4GHz, auto-ack, configurable |
| LoRa (SX1276) | ✅ Done | `pkg/nodes/gpio/lora_sx1276.go` | SPI | Long range, CAD, configurable SF/BW |
| RFID (RC522) | ✅ Done | `pkg/nodes/gpio/rfid_rc522.go` | SPI | Read/Write MIFARE |
| NFC (PN532) | ✅ Done | `pkg/nodes/gpio/nfc_pn532.go` | I2C | Read/Write tags, NDEF |

---

## GPS & Location

| Device | Status | File | Interface | Notes |
|--------|--------|------|-----------|-------|
| NEO-6M GPS | ✅ Done | `pkg/nodes/gpio/gps.go` | Serial | NMEA parsing + Haversine |
| NEO-M8N GPS | ✅ Done | `pkg/nodes/gpio/gps_neom8n.go` | Serial | UBX protocol, dynamic model |
| BN-880 | ✅ Done | `pkg/nodes/gpio/compass_bn880.go` | Serial+I2C | GPS + QMC5883L/HMC5883L compass |

---

## Real-Time Clock

| Device | Status | File | Interface | Notes |
|--------|--------|------|-----------|-------|
| DS1307 | ✅ Done | `pkg/nodes/gpio/rtc_ds1307.go` | I2C | Basic RTC + RAM + SQW |
| DS3231 | ✅ Done | `pkg/nodes/gpio/rtc_ds3231.go` | I2C | High precision RTC + Alarms |
| PCF8523 | ✅ Done | `pkg/nodes/gpio/rtc_pcf8523.go` | I2C | Low power RTC, battery status, alarms |

---

## Board Detection & HAL

| Feature | Status | File | Notes |
|---------|--------|------|-------|
| RPi Model Detection | ✅ Done | `internal/hal/board_detection.go` | All models supported |
| Pin Mapping (BCM) | ✅ Done | `internal/hal/pin_mapping.go` | 40-pin header |
| HAL Abstraction | ✅ Done | `internal/hal/hal.go` | GPIO/I2C/SPI/Serial |
| Mock HAL | ✅ Done | `internal/hal/mock.go` | For testing |
| RPi HAL Implementation | ✅ Done | `internal/hal/rpi.go` | Uses periph.io |

---

## Implementation Progress

### Summary
| Category | Implemented | Total | Percentage |
|----------|-------------|-------|------------|
| Core Interfaces | 5 | 5 | 100% |
| Temperature Sensors | 9 | 9 | 100% |
| Environmental Sensors | 5 | 6 | 83% |
| Distance & Motion | 5 | 5 | 100% |
| ADC | 5 | 5 | 100% |
| Output Devices | 8 | 8 | 100% |
| Industrial Protocols | 4 | 4 | 100% |
| Wireless | 5 | 5 | 100% |
| GPS | 3 | 3 | 100% |
| RTC | 3 | 3 | 100% |
| **TOTAL** | **52** | **53** | **98%** |

---

## Recently Implemented (2026-01-23)

### Batch 1 - Core Sensors
1. [x] DHT11/DHT22 - Real driver with kernel + bit-bang support
2. [x] BME280 - Temperature + Humidity + Pressure + Altitude
3. [x] BH1750 - Light sensor with multiple modes
4. [x] MCP3008/MCP3208 - SPI ADC (10-bit/12-bit)
5. [x] Modbus RTU/TCP - Industrial protocol

### Batch 2 - Displays & Motors
6. [x] LCD I2C (HD44780) - 16x2/20x4 character display
7. [x] OLED SSD1306 - Graphics display with text/shapes
8. [x] DC Motor (L298N) - H-Bridge driver
9. [x] Stepper Motor - A4988/DRV8825 support

### Batch 3 - Advanced Sensors & Peripherals
10. [x] BME680 - Air quality sensor with IAQ calculation
11. [x] WS2812 LED Strip - Addressable RGB with rainbow/gradient/chase effects
12. [x] Buzzer - Tones, melodies, patterns, siren effects
13. [x] VL53L0X - ToF laser distance sensor (0-2000mm)
14. [x] GPS NEO-6M - NMEA parsing, distance calculation, UBX config
15. [x] RTC DS3231 - High precision clock with alarms & temperature

### Batch 4 - Industrial & Specialized
16. [x] CAN Bus (MCP2515) - Full CAN controller with TX/RX, filters, modes
17. [x] MAX31855 - Thermocouple interface with fault detection
18. [x] MAX31865 - RTD (PT100/PT1000) with Callendar-Van Dusen
19. [x] VL53L1X - Long range ToF sensor (up to 4m)
20. [x] DS1307 - Basic RTC with RAM and square wave output
21. [x] RFID RC522 - Full MIFARE read/write with authentication

### Batch 5 - Wireless & Final Sensors
22. [x] NFC (PN532) - I2C NFC reader with NDEF support
23. [x] LoRa (SX1276) - Long range wireless with CAD
24. [x] TSL2561 - Advanced lux sensor with visible/IR channels
25. [x] VEML7700 - High accuracy lux with auto-gain
26. [x] CCS811 - CO2 + TVOC with baseline management
27. [x] SGP30 - Air quality (eCO2, TVOC) with humidity compensation
28. [x] RCWL-0516 - Microwave motion sensor with debounce
29. [x] NEO-M8N GPS - UBX protocol with dynamic model config
30. [x] BN-880 - GPS + QMC5883L/HMC5883L compass
31. [x] PCF8523 - Low power RTC with battery status
32. [x] ADS1015 - 12-bit ADC with differential mode
33. [x] PCF8591 - 8-bit ADC + DAC
34. [x] RF 433MHz - Multi-protocol TX with tri-state support
35. [x] NRF24L01 - 2.4GHz wireless with auto-ack

---

## Implementation Queue (Remaining)

### Only Remaining Item
1. [ ] MQ-x Series - Gas sensors (requires external ADC)

---

## Dependencies

```go
periph.io/x/host/v3     // Hardware initialization
periph.io/x/conn/v3     // I2C/SPI protocols
periph.io/x/devices/v3  // Sensor drivers (BME280, SSD1306)
go-rpio/v4              // GPIO control
go.bug.st/serial        // UART communication
```

---

## New Files Added

```
pkg/nodes/gpio/
├── dht.go              # DHT11/DHT22 (Linux)
├── dht_stub.go         # DHT11/DHT22 (non-Linux stub)
├── bme280.go           # BME280 sensor
├── bme680.go           # BME680 air quality sensor
├── bh1750.go           # BH1750 light sensor
├── mcp3008.go          # MCP3008/MCP3208 ADC
├── modbus.go           # Modbus RTU/TCP
├── lcd_i2c.go          # LCD I2C display
├── oled_ssd1306.go     # OLED SSD1306 display
├── motor_l298n.go      # L298N + Stepper motor
├── ws2812.go           # WS2812/NeoPixel LED strip
├── buzzer.go           # Buzzer (active/passive)
├── vl53l0x.go          # VL53L0X ToF distance sensor
├── vl53l1x.go          # VL53L1X long range ToF
├── gps.go              # GPS NEO-6M (Linux)
├── gps_neom8n.go       # GPS NEO-M8N with UBX protocol
├── rtc_ds3231.go       # DS3231 RTC
├── rtc_ds1307.go       # DS1307 basic RTC
├── rtc_pcf8523.go      # PCF8523 low power RTC
├── can_mcp2515.go      # MCP2515 CAN controller
├── max31855.go         # MAX31855 thermocouple
├── max31865.go         # MAX31865 RTD (PT100/PT1000)
├── rfid_rc522.go       # RC522 RFID reader
├── nfc_pn532.go        # PN532 NFC reader
├── lora_sx1276.go      # SX1276 LoRa transceiver
├── tsl2561.go          # TSL2561 lux sensor
├── veml7700.go         # VEML7700 high accuracy lux
├── ccs811.go           # CCS811 CO2/TVOC sensor
├── sgp30.go            # SGP30 air quality sensor
├── rcwl0516.go         # RCWL-0516 microwave motion
├── compass_bn880.go    # BN-880 GPS + Compass
├── ads1015.go          # ADS1015 12-bit ADC
├── pcf8591.go          # PCF8591 ADC/DAC
├── rf433.go            # RF 433MHz TX/RX
└── nrf24l01.go         # NRF24L01 2.4GHz wireless
```

---

## Notes

- All GPIO nodes work on Linux (Raspberry Pi) only
- Windows/Mac builds use stub implementations
- Mock HAL available for unit testing
- periph.io provides cross-platform sensor drivers
- DHT sensors use kernel driver when available, fall back to bit-banging
- Modbus supports both RTU (serial/RS485) and TCP modes
- BME680 provides IAQ (Indoor Air Quality) calculation based on humidity and gas resistance
- WS2812 uses software timing (for production use rpi_ws281x library)
- Buzzer supports musical notes C0-C8, melodies, patterns, and siren effects
- VL53L0X ToF sensor range: 0-2000mm with calibration support
- VL53L1X long range ToF up to 4m with ROI (Region of Interest) support
- GPS NEO-6M parses NMEA sentences (GGA, RMC, GSA, VTG) and supports Haversine distance calculation
- GPS NEO-M8N adds UBX protocol support with dynamic model configuration
- DS3231 RTC includes alarms, temperature sensor, and aging offset calibration
- DS1307 RTC includes 56-byte battery-backed RAM and square wave output
- PCF8523 RTC includes battery switchover detection and low-power alarms
- CAN Bus (MCP2515) supports standard/extended frames, filters, and multiple bitrates
- MAX31855 thermocouple with cold junction compensation and fault detection
- MAX31865 RTD using Callendar-Van Dusen equation for PT100/PT1000
- RFID RC522 supports MIFARE Classic read/write with key authentication
- NFC PN532 supports NDEF text records and tag scanning
- LoRa SX1276 supports CAD (Channel Activity Detection) and configurable SF/BW/CR
- TSL2561 provides visible and IR channel readings with lux calculation
- VEML7700 features auto-gain adjustment for 0.0036 - 120,000 lux range
- CCS811 measures eCO2 (400-8192 ppm) and TVOC (0-1187 ppb) with baseline management
- SGP30 provides eCO2/TVOC with humidity compensation and self-test
- RCWL-0516 microwave motion sensor with configurable debounce and hold time
- BN-880 combines GPS with QMC5883L/HMC5883L magnetometer compass
- ADS1015 12-bit ADC with programmable gain and differential input modes
- PCF8591 8-bit ADC with DAC output capability
- RF 433MHz supports multiple protocols (1-6) with tri-state encoding
- NRF24L01 2.4GHz transceiver with auto-acknowledgment and configurable data rates
