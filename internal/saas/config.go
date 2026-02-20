package saas

import (
	"time"
)

// Config holds SaaS connection configuration
type Config struct {
	// Enabled determines if SaaS connection is active
	Enabled bool `json:"enabled"`

	// ServerURL is the SaaS platform base URL
	ServerURL string `json:"server_url"`

	// DeviceID is the unique identifier for this edge device
	DeviceID string `json:"device_id"`

	// APIKey is the device authentication key (device_xxxxx format)
	APIKey string `json:"api_key"`

	// ProvisioningCode is the one-time code for initial provisioning
	// Only used if DeviceID/APIKey are not set
	ProvisioningCode string `json:"provisioning_code,omitempty"`

	// HeartbeatInterval defines how often to send ping messages
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`

	// ReconnectTimeout is max time to keep retrying connection
	ReconnectTimeout time.Duration `json:"reconnect_timeout"`

	// MaxReconnectAttempts limits retry attempts
	MaxReconnectAttempts int `json:"max_reconnect_attempts"`

	// EnableTLS uses wss:// instead of ws:// for tunnel
	EnableTLS bool `json:"enable_tls"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Enabled:              false,
		ServerURL:            "http://localhost:3000",
		HeartbeatInterval:    30 * time.Second,
		ReconnectTimeout:     5 * time.Minute,
		MaxReconnectAttempts: 5,
		EnableTLS:            true,
	}
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.ServerURL == "" {
		return ErrInvalidConfig("server_url is required")
	}

	// Need either provisioning code OR (device_id + api_key)
	if c.ProvisioningCode == "" {
		if c.DeviceID == "" {
			return ErrInvalidConfig("device_id is required when provisioning_code is not set")
		}
		if c.APIKey == "" {
			return ErrInvalidConfig("api_key is required when provisioning_code is not set")
		}
	}

	return nil
}

// IsProvisioned returns true if device has been provisioned
func (c *Config) IsProvisioned() bool {
	return c.DeviceID != "" && c.APIKey != ""
}

// TunnelURL returns the WebSocket tunnel URL
func (c *Config) TunnelURL() string {
	scheme := "ws"
	if c.EnableTLS {
		scheme = "wss"
	}
	return scheme + "://" + c.ServerURL + "/tunnel"
}

// APIURL returns the REST API base URL
func (c *Config) APIURL() string {
	scheme := "http"
	if c.EnableTLS {
		scheme = "https"
	}
	return scheme + "://" + c.ServerURL + "/api/v1"
}
