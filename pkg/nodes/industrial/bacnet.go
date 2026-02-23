package industrial

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// BACnet constants
const (
	bacnetBVLCType         = 0x81
	bacnetBVLCUnicast      = 0x0A
	bacnetBVLCBroadcast    = 0x0B
	bacnetNPDUVersion      = 0x01
	bacnetAPDUConfirmedReq = 0x00
	bacnetAPDUUnconfirmed  = 0x10

	bacnetServiceReadProperty  = 0x0C
	bacnetServiceWriteProperty = 0x0F
	bacnetServiceWhoIs         = 0x08
	bacnetServiceIAm           = 0x00

	bacnetDefaultPort = 47808

	// Object types
	bacnetAnalogInput       = 0
	bacnetAnalogOutput      = 1
	bacnetAnalogValue       = 2
	bacnetBinaryInput       = 3
	bacnetBinaryOutput      = 4
	bacnetBinaryValue       = 5
	bacnetMultiStateInput   = 13
	bacnetMultiStateOutput  = 14
	bacnetMultiStateValue   = 19

	// Property IDs
	bacnetPropPresentValue = 85
	bacnetPropObjectName   = 77
	bacnetPropDescription  = 28
	bacnetPropStatusFlags  = 111
)

// BACnetNode implements BACnet/IP communication
type BACnetNode struct {
	host           string
	port           int
	deviceInstance uint32
	timeout        time.Duration
	operation      string
	objectType     uint16
	objectInstance uint32
	propertyId     uint32
	conn           *net.UDPConn
	mu             sync.Mutex
	invokeId       byte
}

// NewBACnetNode creates a new BACnet node
func NewBACnetNode() *BACnetNode {
	return &BACnetNode{
		host:       "127.0.0.1",
		port:       bacnetDefaultPort,
		timeout:    5 * time.Second,
		operation:  "read_property",
		propertyId: bacnetPropPresentValue,
	}
}

// Init initializes the BACnet node
func (n *BACnetNode) Init(config map[string]interface{}) error {
	if host, ok := config["host"].(string); ok {
		n.host = host
	}
	if port, ok := config["port"].(float64); ok {
		n.port = int(port)
	}
	if devInst, ok := config["deviceInstance"].(float64); ok {
		n.deviceInstance = uint32(devInst)
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = time.Duration(timeout) * time.Millisecond
	}
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	}
	if ot, ok := config["objectType"].(string); ok {
		n.objectType = bacnetParseObjectType(ot)
	}
	if oi, ok := config["objectInstance"].(float64); ok {
		n.objectInstance = uint32(oi)
	}
	if pid, ok := config["propertyId"].(float64); ok {
		n.propertyId = uint32(pid)
	}
	return nil
}

// Execute performs BACnet operations
func (n *BACnetNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	operation := n.operation
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	objectType := n.objectType
	objectInstance := n.objectInstance
	propertyId := n.propertyId

	if ot, ok := msg.Payload["objectType"].(string); ok {
		objectType = bacnetParseObjectType(ot)
	}
	if oi, ok := msg.Payload["objectInstance"].(float64); ok {
		objectInstance = uint32(oi)
	}
	if pid, ok := msg.Payload["propertyId"].(float64); ok {
		propertyId = uint32(pid)
	}

	// Connect if needed
	if n.conn == nil {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", n.host, n.port))
		if err != nil {
			return msg, fmt.Errorf("failed to resolve address: %w", err)
		}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			return msg, fmt.Errorf("failed to connect: %w", err)
		}
		n.conn = conn
	}

	n.conn.SetDeadline(time.Now().Add(n.timeout))

	var result interface{}
	var err error

	switch operation {
	case "read_property":
		result, err = n.readProperty(objectType, objectInstance, propertyId)
	case "write_property":
		value := msg.Payload["value"]
		result, err = n.writeProperty(objectType, objectInstance, propertyId, value)
	case "who_is":
		result, err = n.whoIs()
	default:
		err = fmt.Errorf("unknown BACnet operation: %s", operation)
	}

	if err != nil {
		n.conn.Close()
		n.conn = nil
		return msg, err
	}

	msg.Payload["result"] = result
	msg.Payload["operation"] = operation
	return msg, nil
}

func (n *BACnetNode) readProperty(objectType uint16, objectInstance uint32, propertyId uint32) (map[string]interface{}, error) {
	n.invokeId++
	packet := n.buildReadPropertyRequest(objectType, objectInstance, propertyId)

	_, err := n.conn.Write(packet)
	if err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}

	buf := make([]byte, 1500)
	nRead, err := n.conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return n.parseResponse(buf[:nRead])
}

func (n *BACnetNode) writeProperty(objectType uint16, objectInstance uint32, propertyId uint32, value interface{}) (map[string]interface{}, error) {
	n.invokeId++
	packet := n.buildWritePropertyRequest(objectType, objectInstance, propertyId, value)

	_, err := n.conn.Write(packet)
	if err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}

	buf := make([]byte, 1500)
	nRead, err := n.conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return n.parseResponse(buf[:nRead])
}

func (n *BACnetNode) whoIs() ([]map[string]interface{}, error) {
	// Build WhoIs broadcast
	packet := make([]byte, 12)
	packet[0] = bacnetBVLCType
	packet[1] = bacnetBVLCBroadcast
	binary.BigEndian.PutUint16(packet[2:], 12) // Length
	packet[4] = bacnetNPDUVersion
	packet[5] = 0x20 // Network layer message, expecting reply
	packet[6] = 0xFF // DNET broadcast
	packet[7] = 0xFF
	packet[8] = 0x00 // DLEN
	packet[9] = 0xFF // Hop count
	packet[10] = bacnetAPDUUnconfirmed
	packet[11] = bacnetServiceWhoIs

	_, err := n.conn.Write(packet)
	if err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}

	var devices []map[string]interface{}
	deadline := time.Now().Add(n.timeout)

	for time.Now().Before(deadline) {
		n.conn.SetDeadline(deadline)
		buf := make([]byte, 1500)
		nRead, err := n.conn.Read(buf)
		if err != nil {
			break
		}
		if nRead > 0 {
			device, _ := n.parseResponse(buf[:nRead])
			if device != nil {
				devices = append(devices, device)
			}
		}
	}

	return devices, nil
}

func (n *BACnetNode) buildReadPropertyRequest(objectType uint16, objectInstance uint32, propertyId uint32) []byte {
	apdu := []byte{
		0x00,       // Confirmed request
		0x04,       // Max segments unacknowledged, max APDU size
		n.invokeId, // Invoke ID
		bacnetServiceReadProperty,
	}
	// Object identifier (context tag 0)
	objId := bacnetEncodeObjectId(objectType, objectInstance)
	apdu = append(apdu, 0x0C) // Context tag 0, length 4
	apdu = append(apdu, objId...)
	// Property identifier (context tag 1)
	apdu = append(apdu, 0x19, byte(propertyId))

	// NPDU
	npdu := []byte{bacnetNPDUVersion, 0x04} // Expecting reply

	// BVLC
	totalLen := 4 + len(npdu) + len(apdu)
	bvlc := make([]byte, 4)
	bvlc[0] = bacnetBVLCType
	bvlc[1] = bacnetBVLCUnicast
	binary.BigEndian.PutUint16(bvlc[2:], uint16(totalLen))

	packet := append(bvlc, npdu...)
	packet = append(packet, apdu...)
	return packet
}

func (n *BACnetNode) buildWritePropertyRequest(objectType uint16, objectInstance uint32, propertyId uint32, value interface{}) []byte {
	apdu := []byte{
		0x00,
		0x04,
		n.invokeId,
		bacnetServiceWriteProperty,
	}
	objId := bacnetEncodeObjectId(objectType, objectInstance)
	apdu = append(apdu, 0x0C)
	apdu = append(apdu, objId...)
	apdu = append(apdu, 0x19, byte(propertyId))

	// Encode value (context tag 3, opening)
	apdu = append(apdu, 0x3E) // Opening tag 3
	// Encode as real value
	if fv, ok := value.(float64); ok {
		apdu = append(apdu, 0x44) // Application tag 4 (real), length 4
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(int32(fv*1)))
		apdu = append(apdu, buf...)
	}
	apdu = append(apdu, 0x3F) // Closing tag 3

	npdu := []byte{bacnetNPDUVersion, 0x04}
	totalLen := 4 + len(npdu) + len(apdu)
	bvlc := make([]byte, 4)
	bvlc[0] = bacnetBVLCType
	bvlc[1] = bacnetBVLCUnicast
	binary.BigEndian.PutUint16(bvlc[2:], uint16(totalLen))

	packet := append(bvlc, npdu...)
	packet = append(packet, apdu...)
	return packet
}

func (n *BACnetNode) parseResponse(data []byte) (map[string]interface{}, error) {
	if len(data) < 7 {
		return nil, fmt.Errorf("response too short")
	}
	return map[string]interface{}{
		"raw_length": len(data),
		"bvlc_type":  data[0],
		"bvlc_func":  data[1],
		"raw_data":   fmt.Sprintf("%x", data),
	}, nil
}

func bacnetEncodeObjectId(objectType uint16, instance uint32) []byte {
	val := (uint32(objectType) << 22) | (instance & 0x3FFFFF)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, val)
	return buf
}

func bacnetParseObjectType(s string) uint16 {
	switch s {
	case "analog_input":
		return bacnetAnalogInput
	case "analog_output":
		return bacnetAnalogOutput
	case "analog_value":
		return bacnetAnalogValue
	case "binary_input":
		return bacnetBinaryInput
	case "binary_output":
		return bacnetBinaryOutput
	case "binary_value":
		return bacnetBinaryValue
	case "multi_state_input":
		return bacnetMultiStateInput
	case "multi_state_output":
		return bacnetMultiStateOutput
	case "multi_state_value":
		return bacnetMultiStateValue
	default:
		return 0
	}
}

// Cleanup closes the connection
func (n *BACnetNode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.conn != nil {
		n.conn.Close()
		n.conn = nil
	}
	return nil
}

// NewBACnetExecutor creates a new BACnet executor for registry
func NewBACnetExecutor() node.Executor {
	return NewBACnetNode()
}
