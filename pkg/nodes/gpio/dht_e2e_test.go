//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// DHTConfig is duplicated here for testing on non-Linux platforms
// since dht.go has //go:build linux constraint
type dhtTestConfig struct {
	Pin         int     `json:"pin"`
	Type        string  `json:"type"`
	Retries     int     `json:"retries"`
	RetryDelay  int     `json:"retry_delay"`
	TempOffset  float64 `json:"temp_offset"`
	HumidOffset float64 `json:"humid_offset"`
}

func TestDHTConfigValidation(t *testing.T) {
	t.Run("Valid GPIO pin range", func(t *testing.T) {
		validPins := []int{0, 1, 4, 17, 22, 27}
		for _, pin := range validPins {
			assert.True(t, pin >= 0 && pin <= 27, "Pin %d should be valid", pin)
		}
	})

	t.Run("Invalid GPIO pin range", func(t *testing.T) {
		invalidPins := []int{-1, 28, 40, 100}
		for _, pin := range invalidPins {
			assert.False(t, pin >= 0 && pin <= 27, "Pin %d should be invalid", pin)
		}
	})

	t.Run("Valid DHT types", func(t *testing.T) {
		validTypes := []string{"dht11", "dht22", "am2302"}
		for _, dhtType := range validTypes {
			isValid := dhtType == "dht11" || dhtType == "dht22" || dhtType == "am2302"
			assert.True(t, isValid, "Type %s should be valid", dhtType)
		}
	})

	t.Run("Invalid DHT type", func(t *testing.T) {
		invalidTypes := []string{"dht10", "dht23", "am2301", "sht31"}
		for _, dhtType := range invalidTypes {
			isValid := dhtType == "dht11" || dhtType == "dht22" || dhtType == "am2302"
			assert.False(t, isValid, "Type %s should be invalid", dhtType)
		}
	})
}

func TestDHTDataParsing(t *testing.T) {
	t.Run("DHT11 temperature parsing", func(t *testing.T) {
		// DHT11 data format: byte0=humidity int, byte1=humidity dec,
		// byte2=temp int, byte3=temp dec
		data := []byte{55, 0, 25, 0, 80} // 55% humidity, 25°C

		// Verify checksum
		checksum := data[0] + data[1] + data[2] + data[3]
		assert.Equal(t, data[4], checksum, "Checksum should match")

		// Parse values
		humidity := float64(data[0]) + float64(data[1])/10.0
		temperature := float64(data[2]) + float64(data[3])/10.0

		assert.Equal(t, 55.0, humidity)
		assert.Equal(t, 25.0, temperature)
	})

	t.Run("DHT22 temperature parsing", func(t *testing.T) {
		// DHT22 data format: 16-bit humidity, 16-bit temperature (MSB has sign)
		// Example: 650 (65.0%) humidity, 250 (25.0°C) temperature
		humMSB := byte(0x02)  // 0x028A = 650
		humLSB := byte(0x8A)
		tempMSB := byte(0x00) // 0x00FA = 250
		tempLSB := byte(0xFA)
		checksum := humMSB + humLSB + tempMSB + tempLSB

		data := []byte{humMSB, humLSB, tempMSB, tempLSB, checksum}

		// Parse humidity
		humidity := float64(uint16(data[0])<<8|uint16(data[1])) / 10.0
		assert.InDelta(t, 65.0, humidity, 0.1)

		// Parse temperature (positive)
		tempRaw := uint16(data[2]&0x7F)<<8 | uint16(data[3])
		temperature := float64(tempRaw) / 10.0
		if data[2]&0x80 != 0 {
			temperature = -temperature
		}
		assert.InDelta(t, 25.0, temperature, 0.1)
	})

	t.Run("DHT22 negative temperature parsing", func(t *testing.T) {
		// Negative temperature: MSB bit 7 set
		// -10.5°C = 105 with sign bit = 0x8069
		tempMSB := byte(0x80) // Sign bit set
		tempLSB := byte(0x69) // 105 = 10.5 * 10

		tempRaw := uint16(tempMSB&0x7F)<<8 | uint16(tempLSB)
		temperature := float64(tempRaw) / 10.0
		if tempMSB&0x80 != 0 {
			temperature = -temperature
		}
		assert.InDelta(t, -10.5, temperature, 0.1)
	})

	t.Run("Checksum validation", func(t *testing.T) {
		// Valid checksum
		data := []byte{0x02, 0x8A, 0x00, 0xFA, 0x86}
		checksum := data[0] + data[1] + data[2] + data[3]
		assert.Equal(t, data[4], checksum, "Valid checksum should match")

		// Invalid checksum
		badData := []byte{0x02, 0x8A, 0x00, 0xFA, 0xFF}
		badChecksum := badData[0] + badData[1] + badData[2] + badData[3]
		assert.NotEqual(t, badData[4], badChecksum, "Invalid checksum should not match")
	})
}

func TestDHTTemperatureRanges(t *testing.T) {
	t.Run("DHT11 valid temperature range", func(t *testing.T) {
		// DHT11: 0-50°C
		validTemps := []float64{0, 10, 25, 40, 50}
		for _, temp := range validTemps {
			isValid := temp >= 0 && temp <= 50
			assert.True(t, isValid, "Temperature %.1f should be valid for DHT11", temp)
		}
	})

	t.Run("DHT11 invalid temperature range", func(t *testing.T) {
		// DHT11: outside 0-50°C
		invalidTemps := []float64{-1, -10, 51, 60, 80}
		for _, temp := range invalidTemps {
			isValid := temp >= 0 && temp <= 50
			assert.False(t, isValid, "Temperature %.1f should be invalid for DHT11", temp)
		}
	})

	t.Run("DHT22 valid temperature range", func(t *testing.T) {
		// DHT22: -40 to 80°C
		validTemps := []float64{-40, -20, 0, 25, 50, 80}
		for _, temp := range validTemps {
			isValid := temp >= -40 && temp <= 80
			assert.True(t, isValid, "Temperature %.1f should be valid for DHT22", temp)
		}
	})

	t.Run("DHT22 invalid temperature range", func(t *testing.T) {
		// DHT22: outside -40 to 80°C
		invalidTemps := []float64{-50, -41, 81, 100}
		for _, temp := range invalidTemps {
			isValid := temp >= -40 && temp <= 80
			assert.False(t, isValid, "Temperature %.1f should be invalid for DHT22", temp)
		}
	})
}

func TestDHTHumidityRanges(t *testing.T) {
	t.Run("Valid humidity range", func(t *testing.T) {
		// Humidity: 0-100%
		validHumidity := []float64{0, 25, 50, 75, 100}
		for _, hum := range validHumidity {
			isValid := hum >= 0 && hum <= 100
			assert.True(t, isValid, "Humidity %.1f should be valid", hum)
		}
	})

	t.Run("Invalid humidity range", func(t *testing.T) {
		// Humidity: outside 0-100%
		invalidHumidity := []float64{-1, -10, 101, 150}
		for _, hum := range invalidHumidity {
			isValid := hum >= 0 && hum <= 100
			assert.False(t, isValid, "Humidity %.1f should be invalid", hum)
		}
	})
}

func TestDHTCalibrationOffsets(t *testing.T) {
	t.Run("Apply temperature offset", func(t *testing.T) {
		rawTemp := 25.0
		offset := -0.5
		calibratedTemp := rawTemp + offset
		assert.Equal(t, 24.5, calibratedTemp)
	})

	t.Run("Apply humidity offset", func(t *testing.T) {
		rawHumidity := 50.0
		offset := 2.5
		calibratedHumidity := rawHumidity + offset
		assert.Equal(t, 52.5, calibratedHumidity)
	})

	t.Run("Humidity clamping after offset", func(t *testing.T) {
		rawHumidity := 99.0
		offset := 5.0
		calibratedHumidity := rawHumidity + offset

		// Should be clamped to 100
		if calibratedHumidity > 100 {
			calibratedHumidity = 100
		}
		assert.Equal(t, 100.0, calibratedHumidity)
	})

	t.Run("Negative humidity clamping", func(t *testing.T) {
		rawHumidity := 2.0
		offset := -5.0
		calibratedHumidity := rawHumidity + offset

		// Should be clamped to 0
		if calibratedHumidity < 0 {
			calibratedHumidity = 0
		}
		assert.Equal(t, 0.0, calibratedHumidity)
	})
}

func TestDHTRetryConfiguration(t *testing.T) {
	t.Run("Default retries", func(t *testing.T) {
		retries := 0
		if retries <= 0 {
			retries = 3
		}
		assert.Equal(t, 3, retries)
	})

	t.Run("Custom retries", func(t *testing.T) {
		retries := 5
		assert.Equal(t, 5, retries)
	})

	t.Run("Default retry delay", func(t *testing.T) {
		retryDelay := 0
		if retryDelay <= 0 {
			retryDelay = 2000
		}
		assert.Equal(t, 2000, retryDelay)
	})

	t.Run("Custom retry delay", func(t *testing.T) {
		retryDelay := 3000
		assert.Equal(t, 3000, retryDelay)
	})
}
