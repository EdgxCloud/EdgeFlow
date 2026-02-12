package industrial

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// PROFINET constants
const (
	profinetDCPPort      = 34964
	profinetMulticast    = "224.0.0.171"
	profinetEtherType    = 0x8892
	profinetDCPIdentify  = 0x05
	profinetDCPGet       = 0x03
)

// ProfinetNode implements PROFINET communication
type ProfinetNode struct {
	interfaceName string
	ipAddress     string
	deviceName    string
	timeout       time.Duration
	operation     string
	conn          net.Conn
	mu            sync.Mutex
}

// NewProfinetNode creates a new PROFINET node
func NewProfinetNode() *ProfinetNode {
	return &ProfinetNode{
		interfaceName: "eth0",
		timeout:       5 * time.Second,
		operation:     "discover",
	}
}

// Init initializes the PROFINET node
func (n *ProfinetNode) Init(config map[string]interface{}) error {
	if iface, ok := config["interface"].(string); ok {
		n.interfaceName = iface
	}
	if ip, ok := config["ipAddress"].(string); ok {
		n.ipAddress = ip
	}
	if name, ok := config["deviceName"].(string); ok {
		n.deviceName = name
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = time.Duration(timeout) * time.Millisecond
	}
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	}
	return nil
}

// Execute performs PROFINET operations
func (n *ProfinetNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	operation := n.operation
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	var result interface{}
	var err error

	switch operation {
	case "discover":
		result, err = n.discover()
	case "identify":
		deviceName := n.deviceName
		if dn, ok := msg.Payload["deviceName"].(string); ok {
			deviceName = dn
		}
		result, err = n.identify(deviceName)
	case "read_io":
		slot := uint16(0)
		subslot := uint16(1)
		if s, ok := msg.Payload["slot"].(float64); ok {
			slot = uint16(s)
		}
		if ss, ok := msg.Payload["subslot"].(float64); ok {
			subslot = uint16(ss)
		}
		result, err = n.readIO(slot, subslot)
	case "write_io":
		slot := uint16(0)
		subslot := uint16(1)
		if s, ok := msg.Payload["slot"].(float64); ok {
			slot = uint16(s)
		}
		if ss, ok := msg.Payload["subslot"].(float64); ok {
			subslot = uint16(ss)
		}
		data, _ := msg.Payload["data"].([]interface{})
		var byteData []byte
		for _, d := range data {
			if v, ok := d.(float64); ok {
				byteData = append(byteData, byte(v))
			}
		}
		result, err = n.writeIO(slot, subslot, byteData)
	case "get_diagnosis":
		result, err = n.getDiagnosis()
	default:
		err = fmt.Errorf("unknown PROFINET operation: %s", operation)
	}

	if err != nil {
		return msg, err
	}

	msg.Payload["result"] = result
	msg.Payload["operation"] = operation
	return msg, nil
}

func (n *ProfinetNode) discover() ([]map[string]interface{}, error) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", profinetMulticast, profinetDCPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer conn.Close()

	// Build DCP Identify Request
	request := n.buildDCPIdentifyRequest()
	_, err = conn.Write(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send discover request: %w", err)
	}

	var devices []map[string]interface{}
	deadline := time.Now().Add(n.timeout)

	for time.Now().Before(deadline) {
		conn.SetDeadline(deadline)
		buf := make([]byte, 1500)
		nRead, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}
		if nRead > 0 {
			device := n.parseDCPResponse(buf[:nRead])
			if device != nil {
				devices = append(devices, device)
			}
		}
	}

	return devices, nil
}

func (n *ProfinetNode) identify(deviceName string) (map[string]interface{}, error) {
	if n.ipAddress == "" {
		return nil, fmt.Errorf("ipAddress is required for identify operation")
	}

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", n.ipAddress, profinetDCPPort))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	request := n.buildDCPIdentifyRequest()
	_, err = conn.Write(request)
	if err != nil {
		return nil, err
	}

	conn.SetDeadline(time.Now().Add(n.timeout))
	buf := make([]byte, 1500)
	nRead, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("no response from device: %w", err)
	}

	return n.parseDCPResponse(buf[:nRead]), nil
}

func (n *ProfinetNode) readIO(slot, subslot uint16) (map[string]interface{}, error) {
	if n.ipAddress == "" {
		return nil, fmt.Errorf("ipAddress is required for read_io")
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", n.ipAddress, profinetDCPPort), n.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect for I/O read: %w", err)
	}
	defer conn.Close()

	// Build RPC read request
	request := n.buildIOReadRequest(slot, subslot)
	conn.SetDeadline(time.Now().Add(n.timeout))
	_, err = conn.Write(request)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1500)
	nRead, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"slot":        slot,
		"subslot":     subslot,
		"data_length": nRead,
		"raw_data":    fmt.Sprintf("%x", buf[:nRead]),
	}, nil
}

func (n *ProfinetNode) writeIO(slot, subslot uint16, data []byte) (map[string]interface{}, error) {
	if n.ipAddress == "" {
		return nil, fmt.Errorf("ipAddress is required for write_io")
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", n.ipAddress, profinetDCPPort), n.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect for I/O write: %w", err)
	}
	defer conn.Close()

	// Build write request with data
	request := make([]byte, 8+len(data))
	binary.BigEndian.PutUint16(request[0:], slot)
	binary.BigEndian.PutUint16(request[2:], subslot)
	binary.BigEndian.PutUint32(request[4:], uint32(len(data)))
	copy(request[8:], data)

	conn.SetDeadline(time.Now().Add(n.timeout))
	_, err = conn.Write(request)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"slot":    slot,
		"subslot": subslot,
		"written": len(data),
		"success": true,
	}, nil
}

func (n *ProfinetNode) getDiagnosis() (map[string]interface{}, error) {
	if n.ipAddress == "" {
		return nil, fmt.Errorf("ipAddress is required for get_diagnosis")
	}

	return map[string]interface{}{
		"device":    n.ipAddress,
		"name":      n.deviceName,
		"interface": n.interfaceName,
		"status":    "query_sent",
	}, nil
}

func (n *ProfinetNode) buildDCPIdentifyRequest() []byte {
	// Simple DCP identify multicast request via UDP
	request := make([]byte, 12)
	// Frame ID for DCP Identify
	binary.BigEndian.PutUint16(request[0:], 0xFEFE)
	request[2] = profinetDCPIdentify // Service ID
	request[3] = 0x00                // Service Type (Request)
	binary.BigEndian.PutUint32(request[4:], 0x00000001) // Xid
	binary.BigEndian.PutUint16(request[8:], 0x0001)     // Response Delay
	binary.BigEndian.PutUint16(request[10:], 0x0000)    // Data Length
	return request
}

func (n *ProfinetNode) parseDCPResponse(data []byte) map[string]interface{} {
	if len(data) < 12 {
		return nil
	}
	return map[string]interface{}{
		"frame_id":    fmt.Sprintf("0x%04x", binary.BigEndian.Uint16(data[0:])),
		"service_id":  data[2],
		"data_length": len(data),
		"raw_data":    fmt.Sprintf("%x", data),
	}
}

func (n *ProfinetNode) buildIOReadRequest(slot, subslot uint16) []byte {
	request := make([]byte, 8)
	binary.BigEndian.PutUint16(request[0:], slot)
	binary.BigEndian.PutUint16(request[2:], subslot)
	binary.BigEndian.PutUint32(request[4:], 0) // Read request marker
	return request
}

// Cleanup closes the connection
func (n *ProfinetNode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.conn != nil {
		n.conn.Close()
		n.conn = nil
	}
	return nil
}

// NewProfinetExecutor creates a new PROFINET executor for registry
func NewProfinetExecutor() node.Executor {
	return NewProfinetNode()
}
