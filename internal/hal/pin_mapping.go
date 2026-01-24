package hal

type PinCapability int

const (
	CapGPIO PinCapability = 1 << iota
	CapPWM
	CapI2C
	CapSPI
	CapUART
	Cap1Wire
)

type PinInfo struct {
	Physical     int
	BCM          int
	Name         string
	Capabilities PinCapability
	AltFunctions map[string]int
}

var RaspberryPiPinMap = map[int]*PinInfo{
	3: {
		Physical:     3,
		BCM:          2,
		Name:         "GPIO2 (SDA1)",
		Capabilities: CapGPIO | CapI2C,
		AltFunctions: map[string]int{
			"I2C1_SDA": 0,
		},
	},
	5: {
		Physical:     5,
		BCM:          3,
		Name:         "GPIO3 (SCL1)",
		Capabilities: CapGPIO | CapI2C,
		AltFunctions: map[string]int{
			"I2C1_SCL": 0,
		},
	},
	7: {
		Physical:     7,
		BCM:          4,
		Name:         "GPIO4 (GPCLK0)",
		Capabilities: CapGPIO | Cap1Wire,
		AltFunctions: map[string]int{
			"GPCLK0": 0,
		},
	},
	8: {
		Physical:     8,
		BCM:          14,
		Name:         "GPIO14 (TXD0)",
		Capabilities: CapGPIO | CapUART,
		AltFunctions: map[string]int{
			"UART0_TXD": 0,
		},
	},
	10: {
		Physical:     10,
		BCM:          15,
		Name:         "GPIO15 (RXD0)",
		Capabilities: CapGPIO | CapUART,
		AltFunctions: map[string]int{
			"UART0_RXD": 0,
		},
	},
	11: {
		Physical:     11,
		BCM:          17,
		Name:         "GPIO17",
		Capabilities: CapGPIO,
	},
	12: {
		Physical:     12,
		BCM:          18,
		Name:         "GPIO18 (PWM0)",
		Capabilities: CapGPIO | CapPWM,
		AltFunctions: map[string]int{
			"PWM0": 5,
		},
	},
	13: {
		Physical:     13,
		BCM:          27,
		Name:         "GPIO27",
		Capabilities: CapGPIO,
	},
	15: {
		Physical:     15,
		BCM:          22,
		Name:         "GPIO22",
		Capabilities: CapGPIO,
	},
	16: {
		Physical:     16,
		BCM:          23,
		Name:         "GPIO23",
		Capabilities: CapGPIO,
	},
	18: {
		Physical:     18,
		BCM:          24,
		Name:         "GPIO24",
		Capabilities: CapGPIO,
	},
	19: {
		Physical:     19,
		BCM:          10,
		Name:         "GPIO10 (MOSI)",
		Capabilities: CapGPIO | CapSPI,
		AltFunctions: map[string]int{
			"SPI0_MOSI": 0,
		},
	},
	21: {
		Physical:     21,
		BCM:          9,
		Name:         "GPIO9 (MISO)",
		Capabilities: CapGPIO | CapSPI,
		AltFunctions: map[string]int{
			"SPI0_MISO": 0,
		},
	},
	22: {
		Physical:     22,
		BCM:          25,
		Name:         "GPIO25",
		Capabilities: CapGPIO,
	},
	23: {
		Physical:     23,
		BCM:          11,
		Name:         "GPIO11 (SCLK)",
		Capabilities: CapGPIO | CapSPI,
		AltFunctions: map[string]int{
			"SPI0_SCLK": 0,
		},
	},
	24: {
		Physical:     24,
		BCM:          8,
		Name:         "GPIO8 (CE0)",
		Capabilities: CapGPIO | CapSPI,
		AltFunctions: map[string]int{
			"SPI0_CE0_N": 0,
		},
	},
	26: {
		Physical:     26,
		BCM:          7,
		Name:         "GPIO7 (CE1)",
		Capabilities: CapGPIO | CapSPI,
		AltFunctions: map[string]int{
			"SPI0_CE1_N": 0,
		},
	},
	29: {
		Physical:     29,
		BCM:          5,
		Name:         "GPIO5",
		Capabilities: CapGPIO,
	},
	31: {
		Physical:     31,
		BCM:          6,
		Name:         "GPIO6",
		Capabilities: CapGPIO,
	},
	32: {
		Physical:     32,
		BCM:          12,
		Name:         "GPIO12 (PWM0)",
		Capabilities: CapGPIO | CapPWM,
		AltFunctions: map[string]int{
			"PWM0": 0,
		},
	},
	33: {
		Physical:     33,
		BCM:          13,
		Name:         "GPIO13 (PWM1)",
		Capabilities: CapGPIO | CapPWM,
		AltFunctions: map[string]int{
			"PWM1": 0,
		},
	},
	35: {
		Physical:     35,
		BCM:          19,
		Name:         "GPIO19 (PWM1)",
		Capabilities: CapGPIO | CapPWM,
		AltFunctions: map[string]int{
			"PWM1": 5,
		},
	},
	36: {
		Physical:     36,
		BCM:          16,
		Name:         "GPIO16",
		Capabilities: CapGPIO,
	},
	37: {
		Physical:     37,
		BCM:          26,
		Name:         "GPIO26",
		Capabilities: CapGPIO,
	},
	38: {
		Physical:     38,
		BCM:          20,
		Name:         "GPIO20",
		Capabilities: CapGPIO,
	},
	40: {
		Physical:     40,
		BCM:          21,
		Name:         "GPIO21",
		Capabilities: CapGPIO,
	},
}

func GetPinInfo(physical int) *PinInfo {
	return RaspberryPiPinMap[physical]
}

func GetPinByBCM(bcm int) *PinInfo {
	for _, pin := range RaspberryPiPinMap {
		if pin.BCM == bcm {
			return pin
		}
	}
	return nil
}

func HasCapability(physical int, cap PinCapability) bool {
	pin := GetPinInfo(physical)
	if pin == nil {
		return false
	}
	return pin.Capabilities&cap != 0
}

func GetPWMPins() []int {
	pins := make([]int, 0)
	for physical, pin := range RaspberryPiPinMap {
		if pin.Capabilities&CapPWM != 0 {
			pins = append(pins, physical)
		}
	}
	return pins
}

func GetI2CPins() []int {
	pins := make([]int, 0)
	for physical, pin := range RaspberryPiPinMap {
		if pin.Capabilities&CapI2C != 0 {
			pins = append(pins, physical)
		}
	}
	return pins
}

func GetSPIPins() []int {
	pins := make([]int, 0)
	for physical, pin := range RaspberryPiPinMap {
		if pin.Capabilities&CapSPI != 0 {
			pins = append(pins, physical)
		}
	}
	return pins
}
