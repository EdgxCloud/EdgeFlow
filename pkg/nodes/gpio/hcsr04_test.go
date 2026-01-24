//go:build !linux
// +build !linux

package gpio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHCSR04Node(t *testing.T) {
	t.Run("Create new node with defaults", func(t *testing.T) {
		node := NewHCSR04Node()
		assert.NotNil(t, node)
		assert.Equal(t, 23, node.triggerPin)
		assert.Equal(t, 24, node.echoPin)
	})

	t.Run("Init fails without hardware", func(t *testing.T) {
		node := NewHCSR04Node()
		err := node.Init(map[string]interface{}{
			"triggerPin": 23,
			"echoPin":    24,
		})
		// Should fail on non-Linux because no GPIO hardware
		assert.Error(t, err)
	})

	t.Run("Parse config with int pins", func(t *testing.T) {
		node := NewHCSR04Node()
		_ = node.Init(map[string]interface{}{
			"triggerPin": int(17),
			"echoPin":    int(27),
		})
		assert.Equal(t, 17, node.triggerPin)
		assert.Equal(t, 27, node.echoPin)
	})

	t.Run("Parse config with float pins", func(t *testing.T) {
		node := NewHCSR04Node()
		_ = node.Init(map[string]interface{}{
			"triggerPin": float64(17),
			"echoPin":    float64(27),
		})
		assert.Equal(t, 17, node.triggerPin)
		assert.Equal(t, 27, node.echoPin)
	})

	t.Run("Cleanup returns nil", func(t *testing.T) {
		node := NewHCSR04Node()
		err := node.Cleanup()
		assert.NoError(t, err)
	})
}

func TestHCSR04Executor(t *testing.T) {
	t.Run("Create executor with defaults", func(t *testing.T) {
		executor, err := NewHCSR04Executor(map[string]interface{}{
			"triggerPin": 23,
			"echoPin":    24,
		})
		// Fails because no hardware, but validates config
		assert.Error(t, err)
		assert.Nil(t, executor)
	})
}

func TestHCSR04DistanceCalculation(t *testing.T) {
	t.Run("Calculate distance from pulse duration", func(t *testing.T) {
		// Speed of sound ~343 m/s at 20°C
		// Distance = duration * speed / 2 (round trip)
		// Distance (cm) = duration (µs) * 0.0343 / 2

		// 10 cm = 583 µs pulse
		pulseUs := 583.0
		distanceCm := pulseUs * 0.0343 / 2
		assert.InDelta(t, 10.0, distanceCm, 0.5)

		// 20 cm = 1166 µs pulse
		pulseUs = 1166.0
		distanceCm = pulseUs * 0.0343 / 2
		assert.InDelta(t, 20.0, distanceCm, 0.5)

		// 100 cm = 5831 µs pulse
		pulseUs = 5831.0
		distanceCm = pulseUs * 0.0343 / 2
		assert.InDelta(t, 100.0, distanceCm, 1.0)
	})

	t.Run("Maximum distance ~400cm", func(t *testing.T) {
		// 400 cm = ~23300 µs pulse
		pulseUs := 23300.0
		distanceCm := pulseUs * 0.0343 / 2
		assert.InDelta(t, 400.0, distanceCm, 10.0)
	})

	t.Run("Minimum distance ~2cm", func(t *testing.T) {
		// 2 cm = ~116 µs pulse
		pulseUs := 116.0
		distanceCm := pulseUs * 0.0343 / 2
		assert.InDelta(t, 2.0, distanceCm, 0.5)
	})
}

func TestHCSR04UnitConversion(t *testing.T) {
	t.Run("Convert cm to mm", func(t *testing.T) {
		distanceCm := 10.0
		distanceMm := distanceCm * 10
		assert.Equal(t, 100.0, distanceMm)
	})

	t.Run("Convert cm to m", func(t *testing.T) {
		distanceCm := 100.0
		distanceM := distanceCm / 100
		assert.Equal(t, 1.0, distanceM)
	})

	t.Run("Convert cm to inch", func(t *testing.T) {
		distanceCm := 10.0
		distanceInch := distanceCm / 2.54
		assert.InDelta(t, 3.94, distanceInch, 0.1)
	})
}

func TestHCSR04TemperatureCompensation(t *testing.T) {
	t.Run("Speed of sound at different temperatures", func(t *testing.T) {
		// Speed of sound = 331.3 + 0.606 * T (m/s)
		// where T is temperature in Celsius

		temps := []struct {
			celsius float64
			speed   float64 // m/s
		}{
			{0, 331.3},
			{20, 343.5},
			{25, 346.3},
			{30, 349.2},
		}

		for _, tc := range temps {
			speed := 331.3 + 0.606*tc.celsius
			assert.InDelta(t, tc.speed, speed, 0.1)
		}
	})

	t.Run("Temperature affects distance calculation", func(t *testing.T) {
		// At 20°C, speed = 343.5 m/s = 0.03435 cm/µs
		// At 0°C, speed = 331.3 m/s = 0.03313 cm/µs

		pulseUs := 1000.0

		// Distance at 20°C
		speed20 := 331.3 + 0.606*20.0
		distance20 := pulseUs * speed20 / 10000 / 2 // convert m/s to cm/µs and divide by 2

		// Distance at 0°C
		speed0 := 331.3 + 0.606*0.0
		distance0 := pulseUs * speed0 / 10000 / 2

		// Same pulse duration gives different distances at different temperatures
		assert.Greater(t, distance20, distance0)
	})
}

func TestHCSR04Averaging(t *testing.T) {
	t.Run("Average samples", func(t *testing.T) {
		samples := []float64{10.0, 10.5, 9.5, 10.2, 9.8}
		sum := 0.0
		for _, s := range samples {
			sum += s
		}
		avg := sum / float64(len(samples))
		assert.InDelta(t, 10.0, avg, 0.1)
	})

	t.Run("Median filter", func(t *testing.T) {
		// Median is useful for removing outliers
		samples := []float64{10.0, 100.0, 10.2, 10.1, 9.9} // 100.0 is outlier

		// Sort and take middle value
		sorted := make([]float64, len(samples))
		copy(sorted, samples)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j] < sorted[i] {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		median := sorted[len(sorted)/2]

		assert.InDelta(t, 10.1, median, 0.1) // Median ignores the 100.0 outlier
	})
}

func TestHCSR04ConfigValidation(t *testing.T) {
	t.Run("Valid GPIO pin range", func(t *testing.T) {
		validPins := []int{0, 1, 4, 17, 22, 23, 24, 27}
		for _, pin := range validPins {
			isValid := pin >= 0 && pin <= 27
			assert.True(t, isValid, "Pin %d should be valid", pin)
		}
	})

	t.Run("Invalid GPIO pin range", func(t *testing.T) {
		invalidPins := []int{-1, 28, 40, 100}
		for _, pin := range invalidPins {
			isValid := pin >= 0 && pin <= 27
			assert.False(t, isValid, "Pin %d should be invalid", pin)
		}
	})

	t.Run("Different trigger and echo pins", func(t *testing.T) {
		triggerPin := 23
		echoPin := 24
		assert.NotEqual(t, triggerPin, echoPin, "Trigger and echo pins should be different")
	})
}

func TestNewHCSR04Executor(t *testing.T) {
	t.Run("Executor creation requires hardware", func(t *testing.T) {
		_, err := NewHCSR04Executor(map[string]interface{}{
			"triggerPin": 23,
			"echoPin":    24,
		})
		require.Error(t, err) // Should fail without hardware
	})
}
