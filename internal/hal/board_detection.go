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

	default:
		info.Name = "Unknown Board"
		info.NumGPIO = 26
		info.NumPWM = 2
		info.NumI2C = 1
		info.NumSPI = 1
		info.CPUCores = 1
		info.RAMSize = 512
	}

	return info, nil
}

func extractModel(cpuinfo string) BoardModel {
	lines := strings.Split(cpuinfo, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Model") {
			model := strings.ToLower(line)

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
		}
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
