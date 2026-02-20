package saas

import "time"

// TunnelMessage represents a message over the WebSocket tunnel
type TunnelMessage struct {
	Type      string                 `json:"type"`       // connect, connected, ping, pong, command, response
	ID        string                 `json:"id,omitempty"`
	DeviceID  string                 `json:"device_id,omitempty"`
	APIKey    string                 `json:"api_key,omitempty"` // Only in connect message
	Version   string                 `json:"version,omitempty"`
	Action    string                 `json:"action,omitempty"`    // For command messages
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Status    string                 `json:"status,omitempty"`    // success, error
	Data      interface{}            `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
}

// ProvisionRequest is sent to provision a new device
type ProvisionRequest struct {
	ProvisioningCode string                 `json:"provisioning_code"`
	HardwareInfo     map[string]interface{} `json:"hardware_info"`
	NetworkInfo      map[string]interface{} `json:"network_info"`
}

// ProvisionResponse contains the provisioned device credentials
type ProvisionResponse struct {
	DeviceID  string    `json:"device_id"`
	APIKey    string    `json:"api_key"` // Only returned once
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// Shadow represents device state synchronization
type Shadow struct {
	DeviceID  string                 `json:"device_id"`
	Version   int                    `json:"version"`
	Desired   map[string]interface{} `json:"desired"`   // Cloud → Device
	Reported  map[string]interface{} `json:"reported"`  // Device → Cloud
	Delta     map[string]interface{} `json:"delta"`     // Needs sync
	Metadata  ShadowMetadata         `json:"metadata"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ShadowMetadata tracks timestamps for each field
type ShadowMetadata struct {
	Desired  map[string]time.Time `json:"desired,omitempty"`
	Reported map[string]time.Time `json:"reported,omitempty"`
}

// DeviceStatus represents the device online/offline state
type DeviceStatus struct {
	DeviceID     string    `json:"device_id"`
	Status       string    `json:"status"` // online, offline, error
	LastSeenAt   time.Time `json:"last_seen_at"`
	ConnectedAt  time.Time `json:"connected_at,omitempty"`
	FirmwareVer  string    `json:"firmware_version,omitempty"`
	AgentVersion string    `json:"agent_version,omitempty"`
}

// HardwareInfo contains device hardware details
type HardwareInfo struct {
	BoardModel string `json:"board_model"`
	CPU        string `json:"cpu"`
	RAMMB      int    `json:"ram_mb"`
	StorageGB  int    `json:"storage_gb"`
	Kernel     string `json:"kernel"`
	OS         string `json:"os"`
}

// NetworkInfo contains device network configuration
type NetworkInfo struct {
	Hostname       string   `json:"hostname"`
	IPAddress      string   `json:"ip_address"`
	MACAddress     string   `json:"mac_address"`
	Gateway        string   `json:"gateway"`
	DNS            []string `json:"dns"`
	ConnectionType string   `json:"connection_type"` // ethernet, wifi, cellular
}
