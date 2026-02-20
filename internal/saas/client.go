package saas

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Client is the main SaaS integration client
type Client struct {
	config     *Config
	logger     *zap.Logger
	tunnel     *TunnelAgent
	shadow     *ShadowManager
	cmdHandler *EdgeFlowCommandHandler
	provisioning *ProvisioningClient
	configPath string
}

// NewClient creates a new SaaS client
func NewClient(config *Config, logger *zap.Logger, configPath string) *Client {
	return &Client{
		config:     config,
		logger:     logger,
		configPath: configPath,
	}
}

// Initialize sets up the SaaS client with flow service
func (c *Client) Initialize(flowService FlowService, apiService ...interface{}) error {
	if !c.config.Enabled {
		c.logger.Info("SaaS integration disabled")
		return nil
	}

	c.logger.Info("Initializing SaaS client",
		zap.String("server", c.config.ServerURL),
		zap.Bool("provisioned", c.config.IsProvisioned()))

	// Create shadow manager
	c.shadow = NewShadowManager(c.config, c.logger)

	// Create system service adapter for metrics/executions/GPIO
	var systemService SystemService
	if len(apiService) > 0 && apiService[0] != nil {
		systemService = NewSystemServiceAdapter(apiService[0])
	} else {
		systemService = NewSystemServiceAdapter(nil)
	}

	// Create command handler
	c.cmdHandler = NewEdgeFlowCommandHandler(c.logger, flowService, c.shadow)
	c.cmdHandler.SetSystemService(systemService)

	// Create tunnel agent
	c.tunnel = NewTunnelAgent(c.config, c.logger)
	c.tunnel.SetCommandHandler(c.cmdHandler)
	c.tunnel.SetCallbacks(
		c.onConnected,
		c.onDisconnected,
	)

	// Create provisioning client
	c.provisioning = NewProvisioningClient(c.config, c.logger)

	return nil
}

// Start initiates SaaS connection
func (c *Client) Start() error {
	if !c.config.Enabled {
		return nil
	}

	// Check if provisioning is needed
	if !c.config.IsProvisioned() {
		if c.config.ProvisioningCode != "" {
			c.logger.Info("Device not provisioned, starting provisioning...")
			if err := c.Provision(); err != nil {
				return fmt.Errorf("provisioning failed: %w", err)
			}
		} else {
			return ErrNotProvisioned("device needs provisioning code or API key")
		}
	}

	// Start tunnel connection
	if err := c.tunnel.Start(); err != nil {
		return fmt.Errorf("tunnel connection failed: %w", err)
	}

	return nil
}

// Stop gracefully stops SaaS connection
func (c *Client) Stop() error {
	if c.tunnel != nil {
		return c.tunnel.Stop()
	}
	return nil
}

// Provision registers device with SaaS
func (c *Client) Provision() error {
	resp, err := c.provisioning.Provision()
	if err != nil {
		return err
	}

	// Update config with credentials
	c.config.DeviceID = resp.DeviceID
	c.config.APIKey = resp.APIKey
	c.config.ProvisioningCode = "" // Clear one-time code

	// Save updated config to disk
	if c.configPath != "" {
		if err := c.saveConfig(); err != nil {
			c.logger.Error("Failed to save config after provisioning", zap.Error(err))
			// Don't fail - credentials are still in memory
		}
	}

	c.logger.Info("Device provisioned",
		zap.String("device_id", resp.DeviceID))

	return nil
}

// IsConnected returns tunnel connection status
func (c *Client) IsConnected() bool {
	if c.tunnel == nil {
		return false
	}
	return c.tunnel.IsConnected()
}

// GetConfig returns the current configuration
func (c *Client) GetConfig() *Config {
	return c.config
}

// GetShadow retrieves device shadow
func (c *Client) GetShadow() (*Shadow, error) {
	if c.shadow == nil {
		return nil, fmt.Errorf("shadow manager not initialized")
	}
	return c.shadow.GetShadow()
}

// UpdateReportedState updates device state to cloud
func (c *Client) UpdateReportedState(state map[string]interface{}) error {
	if c.shadow == nil {
		return fmt.Errorf("shadow manager not initialized")
	}
	return c.shadow.UpdateReported(state)
}

// SendCommand sends a command to SaaS and waits for response
func (c *Client) SendCommand(action string, payload map[string]interface{}) (*TunnelMessage, error) {
	if c.tunnel == nil {
		return nil, fmt.Errorf("tunnel not initialized")
	}
	return c.tunnel.SendCommand(action, payload, 30*time.Second)
}

// onConnected is called when tunnel connects
func (c *Client) onConnected() {
	c.logger.Info("SaaS tunnel connected")

	// Start shadow sync
	if c.shadow != nil {
		go c.shadow.StartPeriodicSync(5 * time.Minute)

		// Do initial sync
		if _, err := c.shadow.GetShadow(); err != nil {
			c.logger.Error("Initial shadow sync failed", zap.Error(err))
		}
	}

	// Report initial device state
	c.reportDeviceState()
}

// onDisconnected is called when tunnel disconnects
func (c *Client) onDisconnected() {
	c.logger.Warn("SaaS tunnel disconnected")
}

// reportDeviceState reports current device status to cloud
func (c *Client) reportDeviceState() {
	if c.shadow == nil {
		return
	}

	// Basic device state
	state := map[string]interface{}{
		"online":        true,
		"agent_version": "1.0.0",
		"last_updated":  time.Now().Format(time.RFC3339),
	}

	if err := c.shadow.UpdateReported(state); err != nil {
		c.logger.Error("Failed to report device state", zap.Error(err))
	}
}

// saveConfig saves configuration to disk (implementation depends on config format)
func (c *Client) saveConfig() error {
	// TODO: Implement based on your config file format (JSON/YAML/TOML)
	c.logger.Info("Config save not implemented, credentials only in memory")
	return nil
}
