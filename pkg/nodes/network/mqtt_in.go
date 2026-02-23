package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTInConfig configuration for the MQTT In node
type MQTTInConfig struct {
	Broker        string `json:"broker"`        // MQTT broker URL (e.g., tcp://localhost:1883)
	Topic         string `json:"topic"`         // Topic to subscribe (supports wildcards)
	QoS           byte   `json:"qos"`           // Quality of Service (0, 1, 2)
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
	KeepAlive       int  `json:"keepAlive"`       // Keep alive interval in seconds
	ConnectTimeout  int  `json:"connectTimeout"`  // Connection timeout in seconds
	RetainedHandler bool `json:"retainedHandler"` // Handle retained messages on subscribe
}

// MQTTInExecutor executor for the MQTT In node
type MQTTInExecutor struct {
	config     MQTTInConfig
	client     mqtt.Client
	outputChan chan node.Message
	connected  bool
	mu         sync.RWMutex
}

// NewMQTTInExecutor creates a new MQTTInExecutor
func NewMQTTInExecutor() node.Executor {
	return &MQTTInExecutor{
		outputChan: make(chan node.Message, 100),
	}
}

// Init initializes the executor with configuration
func (e *MQTTInExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var mqttConfig MQTTInConfig
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
		mqttConfig.ClientID = fmt.Sprintf("edgeflow_%d", time.Now().Unix())
	}
	if mqttConfig.QoS > 2 {
		mqttConfig.QoS = 2
	}

	e.config = mqttConfig
	return nil
}

// Execute executes the node
func (e *MQTTInExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Connect to MQTT broker if not connected
	if !e.isConnected() {
		if err := e.connect(); err != nil {
			return node.Message{}, fmt.Errorf("failed to connect to MQTT broker: %w", err)
		}
	}

	// Wait for incoming message or context cancellation
	select {
	case <-ctx.Done():
		return node.Message{}, ctx.Err()
	case mqttMsg := <-e.outputChan:
		return mqttMsg, nil
	}
}

// connect connects to the MQTT broker
func (e *MQTTInExecutor) connect() error {
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

		// Subscribe to topic
		token := c.Subscribe(e.config.Topic, e.config.QoS, e.messageHandler)
		token.Wait()
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

// messageHandler handler for incoming MQTT messages
func (e *MQTTInExecutor) messageHandler(client mqtt.Client, mqttMsg mqtt.Message) {
	// Parse payload as JSON if possible
	var payload interface{}
	if err := json.Unmarshal(mqttMsg.Payload(), &payload); err != nil {
		// If JSON parsing fails, use raw string
		payload = string(mqttMsg.Payload())
	}

	// Create message
	msg := node.Message{
		Payload: map[string]interface{}{
			"topic":    mqttMsg.Topic(),
			"payload":  payload,
			"qos":      mqttMsg.Qos(),
			"retained": mqttMsg.Retained(),
		},
	}

	// Send to output channel (non-blocking)
	select {
	case e.outputChan <- msg:
	default:
		// Channel full, log warning
	}
}

// isConnected checks the connection status
func (e *MQTTInExecutor) isConnected() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.connected && e.client != nil && e.client.IsConnected()
}

// Cleanup releases resources
func (e *MQTTInExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.client != nil && e.client.IsConnected() {
		// Unsubscribe
		e.client.Unsubscribe(e.config.Topic)
		// Disconnect
		e.client.Disconnect(250)
		e.connected = false
	}

	close(e.outputChan)
	return nil
}
