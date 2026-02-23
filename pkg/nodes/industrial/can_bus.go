package industrial

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// CAN frame flags
const (
	canEFFFlag = 0x80000000 // Extended frame format
	canRTRFlag = 0x40000000 // Remote transmission request
	canERRFlag = 0x20000000 // Error frame
	canIDMask  = 0x1FFFFFFF // 29-bit ID mask
)

// CANBusNode implements CAN bus communication via SocketCAN
type CANBusNode struct {
	interfaceName string
	bitrate       int
	operation     string
	timeout       time.Duration
	filterIds     []uint32
	conn          net.Conn
	mu            sync.Mutex
	listening     bool
	stopChan      chan struct{}
	outputChan    chan node.Message
}

// NewCANBusNode creates a new CAN bus node
func NewCANBusNode() *CANBusNode {
	return &CANBusNode{
		interfaceName: "can0",
		bitrate:       500000,
		operation:     "receive",
		timeout:       5 * time.Second,
		stopChan:      make(chan struct{}),
		outputChan:    make(chan node.Message, 100),
	}
}

// Init initializes the CAN bus node
func (n *CANBusNode) Init(config map[string]interface{}) error {
	if iface, ok := config["interface"].(string); ok {
		n.interfaceName = iface
	}
	if bitrate, ok := config["bitrate"].(float64); ok {
		n.bitrate = int(bitrate)
	}
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = time.Duration(timeout) * time.Millisecond
	}
	if filters, ok := config["filterIds"].([]interface{}); ok {
		for _, f := range filters {
			if fv, ok := f.(float64); ok {
				n.filterIds = append(n.filterIds, uint32(fv))
			}
		}
	}
	return nil
}

// Execute performs CAN bus operations
func (n *CANBusNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	operation := n.operation
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	// Connect if needed
	if n.conn == nil && (operation == "send" || operation == "receive" || operation == "listen") {
		conn, err := net.DialTimeout("unix", "/var/run/can/"+n.interfaceName, n.timeout)
		if err != nil {
			// SocketCAN not available, return informational error
			return node.Message{
				Type: node.MessageTypeData,
				Payload: map[string]interface{}{
					"error":     "SocketCAN not available on this platform",
					"interface": n.interfaceName,
					"info":      "CAN bus requires Linux with SocketCAN support. Use 'ip link set can0 up type can bitrate 500000' to configure.",
				},
			}, nil
		}
		n.conn = conn
	}

	switch operation {
	case "send":
		return n.sendFrame(msg)
	case "receive":
		return n.receiveFrame()
	case "listen":
		return n.startListening(ctx)
	case "status":
		return n.getStatus()
	default:
		return node.Message{}, fmt.Errorf("unknown CAN operation: %s", operation)
	}
}

func (n *CANBusNode) sendFrame(msg node.Message) (node.Message, error) {
	canId := uint32(0)
	if id, ok := msg.Payload["id"].(float64); ok {
		canId = uint32(id)
	}

	extended := false
	if ext, ok := msg.Payload["extended"].(bool); ok {
		extended = ext
	}

	rtr := false
	if r, ok := msg.Payload["rtr"].(bool); ok {
		rtr = r
	}

	var data []byte
	if dataHex, ok := msg.Payload["data"].(string); ok {
		d, err := hex.DecodeString(dataHex)
		if err != nil {
			return node.Message{}, fmt.Errorf("invalid hex data: %w", err)
		}
		data = d
	} else if dataArr, ok := msg.Payload["data"].([]interface{}); ok {
		for _, d := range dataArr {
			if v, ok := d.(float64); ok {
				data = append(data, byte(v))
			}
		}
	}

	if len(data) > 8 {
		data = data[:8]
	}

	frame := buildCANFrame(canId, data, extended, rtr)

	if n.conn != nil {
		_, err := n.conn.Write(frame)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to send CAN frame: %w", err)
		}
	}

	return node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"sent":     true,
			"id":       canId,
			"data":     hex.EncodeToString(data),
			"dlc":      len(data),
			"extended": extended,
			"rtr":      rtr,
		},
	}, nil
}

func (n *CANBusNode) receiveFrame() (node.Message, error) {
	if n.conn == nil {
		return node.Message{}, fmt.Errorf("not connected to CAN interface")
	}

	n.conn.SetDeadline(time.Now().Add(n.timeout))
	buf := make([]byte, 16)
	nRead, err := n.conn.Read(buf)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read CAN frame: %w", err)
	}

	return parseCANFrame(buf[:nRead])
}

func (n *CANBusNode) startListening(ctx context.Context) (node.Message, error) {
	if n.listening {
		return node.Message{
			Payload: map[string]interface{}{"listening": true, "status": "already_listening"},
		}, nil
	}

	n.listening = true
	go func() {
		for {
			select {
			case <-n.stopChan:
				return
			case <-ctx.Done():
				return
			default:
				if n.conn != nil {
					n.conn.SetDeadline(time.Now().Add(100 * time.Millisecond))
					buf := make([]byte, 16)
					nRead, err := n.conn.Read(buf)
					if err != nil {
						continue
					}
					frameMsg, err := parseCANFrame(buf[:nRead])
					if err != nil {
						continue
					}
					select {
					case n.outputChan <- frameMsg:
					default:
					}
				}
			}
		}
	}()

	return node.Message{
		Payload: map[string]interface{}{"listening": true, "interface": n.interfaceName},
	}, nil
}

func (n *CANBusNode) getStatus() (node.Message, error) {
	return node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"interface": n.interfaceName,
			"bitrate":   n.bitrate,
			"connected": n.conn != nil,
			"listening": n.listening,
			"filters":   n.filterIds,
		},
	}, nil
}

func buildCANFrame(id uint32, data []byte, extended, rtr bool) []byte {
	frame := make([]byte, 16)
	canId := id
	if extended {
		canId |= canEFFFlag
	}
	if rtr {
		canId |= canRTRFlag
	}
	binary.LittleEndian.PutUint32(frame[0:], canId)
	frame[4] = byte(len(data)) // DLC
	copy(frame[8:], data)
	return frame
}

func parseCANFrame(raw []byte) (node.Message, error) {
	if len(raw) < 8 {
		return node.Message{}, fmt.Errorf("CAN frame too short")
	}

	canId := binary.LittleEndian.Uint32(raw[0:4])
	dlc := int(raw[4])
	extended := (canId & canEFFFlag) != 0
	rtr := (canId & canRTRFlag) != 0
	id := canId & canIDMask
	if !extended {
		id = canId & 0x7FF // 11-bit standard ID
	}

	var data []byte
	if len(raw) >= 8+dlc {
		data = raw[8 : 8+dlc]
	}

	return node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"id":        id,
			"dlc":       dlc,
			"data":      hex.EncodeToString(data),
			"extended":  extended,
			"rtr":       rtr,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		},
	}, nil
}

// Cleanup closes the connection
func (n *CANBusNode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.listening {
		close(n.stopChan)
		n.listening = false
	}
	if n.conn != nil {
		n.conn.Close()
		n.conn = nil
	}
	return nil
}

// NewCANBusExecutor creates a new CAN bus executor for registry
func NewCANBusExecutor() node.Executor {
	return NewCANBusNode()
}
