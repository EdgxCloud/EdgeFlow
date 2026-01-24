package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"go.bug.st/serial"
)

type SerialInNode struct {
	port     string
	baudRate int
	dataBits int
	stopBits int
	parity   string

	serialPort serial.Port
	mu         sync.Mutex
	running    bool
	msgChan    chan node.Message
}

func NewSerialInNode() *SerialInNode {
	return &SerialInNode{
		baudRate: 9600,
		dataBits: 8,
		stopBits: 1,
		parity:   "none",
		msgChan:  make(chan node.Message, 100),
	}
}

func (n *SerialInNode) Init(config map[string]interface{}) error {
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
	go n.readLoop()

	return nil
}

func (n *SerialInNode) readLoop() {
	n.mu.Lock()
	n.running = true
	n.mu.Unlock()

	buf := make([]byte, 1024)

	for n.running {
		numBytes, err := n.serialPort.Read(buf)
		if err != nil {
			n.msgChan <- node.Message{
				Payload: map[string]interface{}{
					"error": err.Error(),
				},
				Topic: "serial-error",
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if numBytes > 0 {
			data := make([]byte, numBytes)
			copy(data, buf[:numBytes])

			n.msgChan <- node.Message{
				Payload: map[string]interface{}{
					"data": string(data),
				},
				Topic: "serial-data",
			}
		}
	}
}

func (n *SerialInNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	select {
	case outMsg := <-n.msgChan:
		return outMsg, nil
	case <-time.After(100 * time.Millisecond):
		return node.Message{}, nil
	}
}

func (n *SerialInNode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.running = false
	if n.serialPort != nil {
		return n.serialPort.Close()
	}
	return nil
}

func NewSerialInExecutor() node.Executor {
	return NewSerialInNode()
}
