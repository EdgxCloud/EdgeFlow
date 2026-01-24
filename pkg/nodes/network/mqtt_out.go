package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTOutConfig نود MQTT Out
type MQTTOutConfig struct {
	Broker        string `json:"broker"`        // MQTT broker URL
	Topic         string `json:"topic"`         // Topic to publish
	QoS           byte   `json:"qos"`           // Quality of Service (0, 1, 2)
	Retain        bool   `json:"retain"`        // Retain flag
	ClientID      string `json:"clientId"`      // Client ID (optional)
	Username      string `json:"username"`      // Username (optional)
	Password      string `json:"password"`      // Password (optional)
	CleanSession  bool   `json:"cleanSession"`  // Clean session flag
	AutoReconnect bool   `json:"autoReconnect"` // Auto reconnect

	// Last Will and Testament (LWT) configuration
	WillTopic   string `json:"willTopic"`   // LWT topic
	WillPayload string `json:"willPayload"` // LWT message payload
	WillQoS     byte   `json:"willQos"`     // LWT QoS (0, 1, 2)
	WillRetain  bool   `json:"willRetain"`  // LWT retain flag

	// Connection settings
	KeepAlive      int `json:"keepAlive"`      // Keep alive interval in seconds
	ConnectTimeout int `json:"connectTimeout"` // Connection timeout in seconds

	// Message settings
	MessageExpiry int `json:"messageExpiry"` // Message expiry in seconds (MQTT 5.0)
}

// MQTTOutExecutor اجراکننده نود MQTT Out
type MQTTOutExecutor struct {
	config    MQTTOutConfig
	client    mqtt.Client
	connected bool
	mu        sync.RWMutex
}

// NewMQTTOutExecutor ایجاد MQTTOutExecutor
func NewMQTTOutExecutor() node.Executor {
	return &MQTTOutExecutor{}
}

// Init initializes the executor with configuration
func (e *MQTTOutExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var mqttConfig MQTTOutConfig
	if err := json.Unmarshal(configJSON, &mqttConfig); err != nil {
		return fmt.Errorf("invalid mqtt config: %w", err)
	}

	// Validate required fields
	if mqttConfig.Broker == "" {
		return fmt.Errorf("broker is required")
	}
	if mqttConfig.Topic == "" {
		return fmt.Errorf("topic is required")
	}

	// Default values
	if mqttConfig.ClientID == "" {
		mqttConfig.ClientID = fmt.Sprintf("edgeflow_out_%d", time.Now().Unix())
	}
	if mqttConfig.QoS > 2 {
		mqttConfig.QoS = 2
	}

	e.config = mqttConfig
	return nil
}

// Execute اجرای نود
func (e *MQTTOutExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Connect to MQTT broker if not connected
	if !e.isConnected() {
		if err := e.connect(); err != nil {
			return node.Message{}, fmt.Errorf("failed to connect to MQTT broker: %w", err)
		}
	}

	// Get topic from message or config
	topic := e.config.Topic
	if topicFromMsg, ok := msg.Payload["topic"].(string); ok && topicFromMsg != "" {
		topic = topicFromMsg
	}

	// Get payload to publish
	var payload interface{}
	if p, ok := msg.Payload["payload"]; ok {
		payload = p
	} else {
		payload = msg.Payload
	}

	// Convert payload to bytes
	var payloadBytes []byte
	switch v := payload.(type) {
	case string:
		payloadBytes = []byte(v)
	case []byte:
		payloadBytes = v
	default:
		// Marshal to JSON
		var err error
		payloadBytes, err = json.Marshal(v)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	// Get QoS and Retain from message or config
	qos := e.config.QoS
	retain := e.config.Retain

	if qosFromMsg, ok := msg.Payload["qos"].(float64); ok {
		qos = byte(qosFromMsg)
	}
	if retainFromMsg, ok := msg.Payload["retain"].(bool); ok {
		retain = retainFromMsg
	}

	// Publish message
	token := e.client.Publish(topic, qos, retain, payloadBytes)
	token.Wait()

	if token.Error() != nil {
		return node.Message{}, fmt.Errorf("publish failed: %w", token.Error())
	}

	// Return original message with publish info
	outputMsg := node.Message{
		Payload: map[string]interface{}{
			"topic":     topic,
			"qos":       qos,
			"retain":    retain,
			"published": true,
			"size":      len(payloadBytes),
		},
	}

	return outputMsg, nil
}

// connect اتصال به MQTT broker
func (e *MQTTOutExecutor) connect() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.connected {
		return nil
	}

	// Create MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(e.config.Broker)
	opts.SetClientID(e.config.ClientID)
	opts.SetCleanSession(e.config.CleanSession)
	opts.SetAutoReconnect(e.config.AutoReconnect)

	// Set keep alive
	if e.config.KeepAlive > 0 {
		opts.SetKeepAlive(time.Duration(e.config.KeepAlive) * time.Second)
	} else {
		opts.SetKeepAlive(60 * time.Second)
	}

	// Set connection timeout
	if e.config.ConnectTimeout > 0 {
		opts.SetConnectTimeout(time.Duration(e.config.ConnectTimeout) * time.Second)
	} else {
		opts.SetConnectTimeout(30 * time.Second)
	}

	if e.config.Username != "" {
		opts.SetUsername(e.config.Username)
		opts.SetPassword(e.config.Password)
	}

	// Configure Last Will and Testament (LWT)
	if e.config.WillTopic != "" {
		opts.SetWill(e.config.WillTopic, e.config.WillPayload, e.config.WillQoS, e.config.WillRetain)
	}

	// Set connection handlers
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		e.mu.Lock()
		e.connected = true
		e.mu.Unlock()
	})

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		e.mu.Lock()
		e.connected = false
		e.mu.Unlock()
	})

	// Create and connect client
	e.client = mqtt.NewClient(opts)
	token := e.client.Connect()
	token.Wait()

	if token.Error() != nil {
		return fmt.Errorf("connection failed: %w", token.Error())
	}

	return nil
}

// isConnected بررسی وضعیت اتصال
func (e *MQTTOutExecutor) isConnected() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.connected && e.client != nil && e.client.IsConnected()
}

// Cleanup پاکسازی منابع
func (e *MQTTOutExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.client != nil && e.client.IsConnected() {
		e.client.Disconnect(250)
		e.connected = false
	}

	return nil
}
