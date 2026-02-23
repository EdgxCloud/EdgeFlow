package wireless

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// ZWaveNode implements Z-Wave communication
type ZWaveNode struct {
	controller    string // Controller type: zwave-js, openzwave, zwave2mqtt
	wsURL         string // WebSocket URL (for zwave-js-ui)
	mqttBroker    string // MQTT broker URL (for zwave2mqtt)
	mqttTopic     string // Base MQTT topic
	serialPort    string // Serial port for direct communication
	operation     string // get_nodes, get_node, set_value, add_node, remove_node, heal_network
	nodeID        int    // Target node ID
	commandClass  int    // Command class ID
	property      string // Property name
	timeout       time.Duration
	mu            sync.Mutex
}

// ZWaveNodeInfo represents a Z-Wave node
type ZWaveNodeInfo struct {
	NodeID           int                    `json:"node_id"`
	Name             string                 `json:"name"`
	Location         string                 `json:"location"`
	Manufacturer     string                 `json:"manufacturer"`
	Product          string                 `json:"product"`
	ProductType      string                 `json:"product_type"`
	Type             string                 `json:"type"` // Static Controller, Portable Controller, Slave
	IsListening      bool                   `json:"is_listening"`
	IsRouting        bool                   `json:"is_routing"`
	IsSecure         bool                   `json:"is_secure"`
	IsBeaming        bool                   `json:"is_beaming"`
	IsAwake          bool                   `json:"is_awake"`
	IsFailed         bool                   `json:"is_failed"`
	Status           string                 `json:"status"` // alive, dead, sleeping, unknown
	CommandClasses   []int                  `json:"command_classes"`
	Values           map[string]interface{} `json:"values,omitempty"`
}

// ZWaveValue represents a Z-Wave value
type ZWaveValue struct {
	CommandClass int         `json:"command_class"`
	Property     string      `json:"property"`
	PropertyKey  string      `json:"property_key,omitempty"`
	Endpoint     int         `json:"endpoint"`
	Value        interface{} `json:"value"`
	Label        string      `json:"label"`
	Type         string      `json:"type"` // number, boolean, string, etc.
	Readable     bool        `json:"readable"`
	Writeable    bool        `json:"writeable"`
	Min          interface{} `json:"min,omitempty"`
	Max          interface{} `json:"max,omitempty"`
	Unit         string      `json:"unit,omitempty"`
}

// NewZWaveNode creates a new Z-Wave node
func NewZWaveNode() *ZWaveNode {
	return &ZWaveNode{
		controller: "zwave2mqtt",
		mqttBroker: "mqtt://localhost:1883",
		mqttTopic:  "zwave",
		operation:  "get_nodes",
		timeout:    30 * time.Second,
	}
}

// Init initializes the Z-Wave node
func (n *ZWaveNode) Init(config map[string]interface{}) error {
	if ctrl, ok := config["controller"].(string); ok {
		n.controller = ctrl
	}
	if url, ok := config["wsUrl"].(string); ok {
		n.wsURL = url
	}
	if broker, ok := config["mqttBroker"].(string); ok {
		n.mqttBroker = broker
	}
	if topic, ok := config["mqttTopic"].(string); ok {
		n.mqttTopic = topic
	}
	if port, ok := config["serialPort"].(string); ok {
		n.serialPort = port
	}
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	}
	if nodeID, ok := config["nodeId"].(float64); ok {
		n.nodeID = int(nodeID)
	}
	if cc, ok := config["commandClass"].(float64); ok {
		n.commandClass = int(cc)
	}
	if prop, ok := config["property"].(string); ok {
		n.property = prop
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = time.Duration(timeout) * time.Millisecond
	}

	return nil
}

// Execute performs Z-Wave operation
func (n *ZWaveNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get parameters from message
	operation := n.operation
	nodeID := n.nodeID

	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}
	if nid, ok := msg.Payload["nodeId"].(float64); ok {
		nodeID = int(nid)
	}

	var result interface{}
	var err error

	switch n.controller {
	case "zwave2mqtt":
		result, err = n.executeZWave2MQTT(ctx, operation, nodeID, msg.Payload)
	case "zwave-js":
		result, err = n.executeZWaveJS(ctx, operation, nodeID, msg.Payload)
	default:
		err = fmt.Errorf("unsupported controller: %s", n.controller)
	}

	if err != nil {
		return msg, err
	}

	msg.Payload["result"] = result
	msg.Payload["operation"] = operation
	msg.Payload["controller"] = n.controller

	return msg, nil
}

// executeZWave2MQTT executes via zwave2mqtt MQTT interface
func (n *ZWaveNode) executeZWave2MQTT(ctx context.Context, operation string, nodeID int, payload map[string]interface{}) (interface{}, error) {
	// Note: In full implementation would use MQTT client
	// to communicate with zwave2mqtt
	//
	// Example topics:
	// zwave/_CLIENTS/ZWAVE_GATEWAY-xxx/api/getNodes/set - request all nodes
	// zwave/<nodeName>/<commandClass>/<property>/set - set value
	// zwave/<nodeName>/<commandClass>/<property> - subscribe to value changes
	//
	// API topics (JSON-RPC style):
	// zwave/_CLIENTS/ZWAVE_GATEWAY-<uuid>/api/<command>/set
	// Supported commands: getNodes, getInfo, addNode, removeNode, healNetwork, etc.

	switch operation {
	case "get_nodes":
		return map[string]interface{}{
			"topic":   fmt.Sprintf("%s/_CLIENTS/ZWAVE_GATEWAY-*/api/getNodes/set", n.mqttTopic),
			"payload": map[string]interface{}{"args": []interface{}{}},
			"hint":    "Send this via MQTT Out node, subscribe to response topic",
		}, nil
	case "get_node":
		return map[string]interface{}{
			"topic":   fmt.Sprintf("%s/_CLIENTS/ZWAVE_GATEWAY-*/api/getNodeInfo/set", n.mqttTopic),
			"payload": map[string]interface{}{"args": []interface{}{nodeID}},
			"hint":    "Send this via MQTT Out node",
		}, nil
	case "set_value":
		cc, _ := payload["commandClass"].(float64)
		prop, _ := payload["property"].(string)
		value := payload["value"]
		return map[string]interface{}{
			"topic":   fmt.Sprintf("%s/nodeID_%d/%d/%s/set", n.mqttTopic, nodeID, int(cc), prop),
			"payload": value,
			"hint":    "Send this via MQTT Out node",
		}, nil
	case "add_node":
		secure := true
		if s, ok := payload["secure"].(bool); ok {
			secure = s
		}
		return map[string]interface{}{
			"topic":   fmt.Sprintf("%s/_CLIENTS/ZWAVE_GATEWAY-*/api/addNode/set", n.mqttTopic),
			"payload": map[string]interface{}{"args": []interface{}{secure}},
			"hint":    "Send this via MQTT Out node to start inclusion",
		}, nil
	case "remove_node":
		return map[string]interface{}{
			"topic":   fmt.Sprintf("%s/_CLIENTS/ZWAVE_GATEWAY-*/api/removeNode/set", n.mqttTopic),
			"payload": map[string]interface{}{"args": []interface{}{}},
			"hint":    "Send this via MQTT Out node to start exclusion",
		}, nil
	case "heal_network":
		return map[string]interface{}{
			"topic":   fmt.Sprintf("%s/_CLIENTS/ZWAVE_GATEWAY-*/api/healNetwork/set", n.mqttTopic),
			"payload": map[string]interface{}{"args": []interface{}{}},
			"hint":    "Send this via MQTT Out node",
		}, nil
	case "heal_node":
		return map[string]interface{}{
			"topic":   fmt.Sprintf("%s/_CLIENTS/ZWAVE_GATEWAY-*/api/healNode/set", n.mqttTopic),
			"payload": map[string]interface{}{"args": []interface{}{nodeID}},
			"hint":    "Send this via MQTT Out node",
		}, nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// executeZWaveJS executes via zwave-js-ui WebSocket
func (n *ZWaveNode) executeZWaveJS(ctx context.Context, operation string, nodeID int, payload map[string]interface{}) (interface{}, error) {
	// Note: In full implementation would use WebSocket client
	// to communicate with zwave-js-ui
	//
	// zwave-js-ui provides a WebSocket API and REST API
	// WebSocket endpoint: ws://localhost:8091
	// REST endpoint: http://localhost:8091/api

	switch operation {
	case "get_nodes":
		return map[string]interface{}{
			"url":    fmt.Sprintf("%s/api/zwave/nodes", n.wsURL),
			"method": "GET",
			"hint":   "Use HTTP Request node to fetch nodes",
		}, nil
	case "get_node":
		return map[string]interface{}{
			"url":    fmt.Sprintf("%s/api/zwave/nodes/%d", n.wsURL, nodeID),
			"method": "GET",
			"hint":   "Use HTTP Request node to fetch node",
		}, nil
	case "set_value":
		cc, _ := payload["commandClass"].(float64)
		prop, _ := payload["property"].(string)
		value := payload["value"]
		body, _ := json.Marshal(map[string]interface{}{
			"nodeId":       nodeID,
			"commandClass": int(cc),
			"property":     prop,
			"value":        value,
		})
		return map[string]interface{}{
			"url":     fmt.Sprintf("%s/api/zwave/setValue", n.wsURL),
			"method":  "POST",
			"payload": string(body),
			"hint":    "Use HTTP Request node to set value",
		}, nil
	default:
		return nil, fmt.Errorf("unknown operation for zwave-js: %s", operation)
	}
}

// Cleanup releases Z-Wave resources
func (n *ZWaveNode) Cleanup() error {
	return nil
}

// NewZWaveExecutor creates a new Z-Wave executor
func NewZWaveExecutor() node.Executor {
	return NewZWaveNode()
}
