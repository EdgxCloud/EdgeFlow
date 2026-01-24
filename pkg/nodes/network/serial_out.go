package network

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
	"go.bug.st/serial"
)

type SerialOutNode struct {
	port       string
	baudRate   int
	dataBits   int
	stopBits   int
	parity     string
	addNewline bool

	serialPort serial.Port
}

func NewSerialOutNode() *SerialOutNode {
	return &SerialOutNode{
		baudRate:   9600,
		dataBits:   8,
		stopBits:   1,
		parity:     "none",
		addNewline: false,
	}
}

func (n *SerialOutNode) Init(config map[string]interface{}) error {
	if port, ok := config["port"].(string); ok {
		n.port = port
	}
	if baud, ok := config["baudRate"].(float64); ok {
		n.baudRate = int(baud)
	} else if baud, ok := config["baudRate"].(int); ok {
		n.baudRate = baud
	}
	if dataBits, ok := config["dataBits"].(float64); ok {
		n.dataBits = int(dataBits)
	} else if dataBits, ok := config["dataBits"].(int); ok {
		n.dataBits = dataBits
	}
	if stopBits, ok := config["stopBits"].(float64); ok {
		n.stopBits = int(stopBits)
	} else if stopBits, ok := config["stopBits"].(int); ok {
		n.stopBits = stopBits
	}
	if parity, ok := config["parity"].(string); ok {
		n.parity = parity
	}
	if addNewline, ok := config["addNewline"].(bool); ok {
		n.addNewline = addNewline
	}

	mode := &serial.Mode{
		BaudRate: n.baudRate,
		DataBits: n.dataBits,
		StopBits: serial.StopBits(n.stopBits),
	}

	switch n.parity {
	case "none":
		mode.Parity = serial.NoParity
	case "even":
		mode.Parity = serial.EvenParity
	case "odd":
		mode.Parity = serial.OddParity
	default:
		mode.Parity = serial.NoParity
	}

	port, err := serial.Open(n.port, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	n.serialPort = port
	return nil
}

func (n *SerialOutNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var data []byte

	// Extract data from payload
	if dataValue, ok := msg.Payload["data"]; ok {
		switch v := dataValue.(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		default:
			data = []byte(fmt.Sprintf("%v", v))
		}
	} else {
		// Fallback: use entire payload as string
		data = []byte(fmt.Sprintf("%v", msg.Payload))
	}

	if n.addNewline {
		data = append(data, '\n')
	}

	if _, err := n.serialPort.Write(data); err != nil {
		return msg, fmt.Errorf("failed to write to serial port: %w", err)
	}

	return msg, nil
}

func (n *SerialOutNode) Cleanup() error {
	if n.serialPort != nil {
		return n.serialPort.Close()
	}
	return nil
}

func NewSerialOutExecutor() node.Executor {
	return NewSerialOutNode()
}
