package hal

import (
	"fmt"
	"os"
	"strings"
)

type BoardModel int

const (
	BoardUnknown BoardModel = iota
	BoardRPiZero
	BoardRPiZeroW
	BoardRPiZero2W
	BoardRPi1
	BoardRPi2
	BoardRPi3
	BoardRPi3Plus
	BoardRPi4
	BoardRPi5
	BoardRPiCM3
	BoardRPiCM4
)

type BoardInfo struct {
	Model      BoardModel
	Name       string
	HasWiFi    bool
	HasBT      bool
	NumGPIO    int
	NumPWM     int
	NumI2C     int
	NumSPI     int
	CPUCores   int
	RAMSize    int
	GPIOChip   string
}

// GPIOChipName returns the GPIO character device name for this board model.
// Auto-detects by scanning /dev/gpiochip* for the RP1 or BCM2835 controller.
// Falls back to gpiochip0 if auto-detection fails.
func (b BoardModel) GPIOChipName() string {
	// Try to auto-detect the correct GPIO chip by reading chip labels
	// Pi 5 RP1 chip can be on gpiochip0 or gpiochip4 depending on OS version
	for _, chip := range []string{"gpiochip0", "gpiochip4"} {
		labelPath := fmt.Sprintf("/sys/bus/gpio/devices/%s/label", chip)
		data, err := os.ReadFile(labelPath)
		if err != nil {
			continue
		}
		label := strings.TrimSpace(string(data))
		// Pi 5 uses pinctrl-rp1, Pi 4 and earlier use pinctrl-bcm2835
		if strings.Contains(label, "pinctrl-rp1") || strings.Contains(label, "pinctrl-bcm2") {
			return chip
		}
	}
	// Fallback: gpiochip0 works for most boards
	return "gpiochip0"
}

func DetectBoard() (*BoardInfo, error) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read cpuinfo: %w", err)
	}

	cpuinfo := string(data)
	model := extractModel(cpuinfo)

	info := &BoardInfo{
		Model: model,
	}

	switch model {
	case BoardRPiZero:
		info.Name = "Raspberry Pi Zero"
		info.HasWiFi = false
		info.HasBT = false
		info.NumGPIO = 26
		info.NumPWM = 2
		info.NumI2C = 1
		info.NumSPI = 2
		info.CPUCores = 1
		info.RAMSize = 512
		info.GPIOChip = model.GPIOChipName()

	case BoardRPiZeroW:
		info.Name = "Raspberry Pi Zero W"
		info.HasWiFi = true
		info.HasBT = true
		info.NumGPIO = 26
		info.NumPWM = 2
		info.NumI2C = 1
		info.NumSPI = 2
		info.CPUCores = 1
		info.RAMSize = 512
		info.GPIOChip = model.GPIOChipName()

	case BoardRPiZero2W:
		info.Name = "Raspberry Pi Zero 2 W"
		info.HasWiFi = true
		info.HasBT = true
		info.NumGPIO = 26
		info.NumPWM = 2
		info.NumI2C = 1
		info.NumSPI = 2
		info.CPUCores = 4
		info.RAMSize = 512
		info.GPIOChip = model.GPIOChipName()

	case BoardRPi3:
		info.Name = "Raspberry Pi 3"
		info.HasWiFi = true
		info.HasBT = true
		info.NumGPIO = 26
		info.NumPWM = 4
		info.NumI2C = 1
		info.NumSPI = 2
		info.CPUCores = 4
		info.RAMSize = 1024
		info.GPIOChip = model.GPIOChipName()

	case BoardRPi3Plus:
		info.Name = "Raspberry Pi 3 Model B+"
		info.HasWiFi = true
		info.HasBT = true
		info.NumGPIO = 26
		info.NumPWM = 4
		info.NumI2C = 1
		info.NumSPI = 2
		info.CPUCores = 4
		info.RAMSize = 1024
		info.GPIOChip = model.GPIOChipName()

	case BoardRPi4:
		info.Name = "Raspberry Pi 4"
		info.HasWiFi = true
		info.HasBT = true
		info.NumGPIO = 26
		info.NumPWM = 4
		info.NumI2C = 6
		info.NumSPI = 5
		info.CPUCores = 4
		info.RAMSize = detectRAMSize()
		info.GPIOChip = model.GPIOChipName()

	case BoardRPi5:
		info.Name = "Raspberry Pi 5"
		info.HasWiFi = true
		info.HasBT = true
		info.NumGPIO = 26
		info.NumPWM = 4
		info.NumI2C = 8
		info.NumSPI = 5
		info.CPUCores = 4
		info.RAMSize = detectRAMSize()
		info.GPIOChip = model.GPIOChipName()

	case BoardRPiCM4:
		info.Name = "Raspberry Pi Compute Module 4"
		info.HasWiFi = false
		info.HasBT = false
		info.NumGPIO = 28
		info.NumPWM = 4
		info.NumI2C = 6
		info.NumSPI = 5
		info.CPUCores = 4
		info.RAMSize = detectRAMSize()
		info.GPIOChip = model.GPIOChipName()

	default:
		info.Name = "Unknown Board"
		info.NumGPIO = 26
		info.NumPWM = 2
		info.NumI2C = 1
		info.NumSPI = 1
		info.CPUCores = 1
		info.RAMSize = 512
		info.GPIOChip = "gpiochip0"
	}

	return info, nil
}

func extractModel(cpuinfo string) BoardModel {
	// First try /proc/cpuinfo Model line
	lines := strings.Split(cpuinfo, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Model") {
			if m := matchBoardModel(line); m != BoardUnknown {
				return m
			}
		}
	}

	// Fallback: Pi 5 doesn't have Model in cpuinfo, check device-tree
	if dtModel, err := os.ReadFile("/proc/device-tree/model"); err == nil {
		if m := matchBoardModel(string(dtModel)); m != BoardUnknown {
			return m
		}
	}

	return BoardUnknown
}

func matchBoardModel(text string) BoardModel {
	model := strings.ToLower(text)

	if strings.Contains(model, "pi 5") {
		return BoardRPi5
	} else if strings.Contains(model, "pi 4") {
		return BoardRPi4
	} else if strings.Contains(model, "pi 3 model b+") {
		return BoardRPi3Plus
	} else if strings.Contains(model, "pi 3") {
		return BoardRPi3
	} else if strings.Contains(model, "pi 2") {
		return BoardRPi2
	} else if strings.Contains(model, "pi 1") || strings.Contains(model, "model b") {
		return BoardRPi1
	} else if strings.Contains(model, "zero 2 w") {
		return BoardRPiZero2W
	} else if strings.Contains(model, "zero w") {
		return BoardRPiZeroW
	} else if strings.Contains(model, "zero") {
		return BoardRPiZero
	} else if strings.Contains(model, "compute module 4") {
		return BoardRPiCM4
	} else if strings.Contains(model, "compute module 3") {
		return BoardRPiCM3
	}
	return BoardUnknown
}

func detectRAMSize() int {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				var kb int
				fmt.Sscanf(parts[1], "%d", &kb)
				return kb / 1024
			}
		}
	}

	return 0
}

func (b BoardModel) String() string {
	switch b {
	case BoardRPiZero:
		return "Raspberry Pi Zero"
	case BoardRPiZeroW:
		return "Raspberry Pi Zero W"
	case BoardRPiZero2W:
		return "Raspberry Pi Zero 2 W"
	case BoardRPi1:
		return "Raspberry Pi 1"
	case BoardRPi2:
		return "Raspberry Pi 2"
	case BoardRPi3:
		return "Raspberry Pi 3"
	case BoardRPi3Plus:
		return "Raspberry Pi 3 B+"
	case BoardRPi4:
		return "Raspberry Pi 4"
	case BoardRPi5:
		return "Raspberry Pi 5"
	case BoardRPiCM3:
		return "Raspberry Pi Compute Module 3"
	case BoardRPiCM4:
		return "Raspberry Pi Compute Module 4"
	default:
		return "Unknown"
	}
}
