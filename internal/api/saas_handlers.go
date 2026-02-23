package api

import (
	"github.com/EdgxCloud/EdgeFlow/internal/saas"
	"github.com/gofiber/fiber/v2"
)

// SaaSHandler handles SaaS-related API endpoints
type SaaSHandler struct {
	saasClient *saas.Client
}

// NewSaaSHandler creates a SaaS handler
func NewSaaSHandler(client *saas.Client) *SaaSHandler {
	return &SaaSHandler{
		saasClient: client,
	}
}

// GetConfig returns current SaaS configuration
func (h *SaaSHandler) GetConfig(c *fiber.Ctx) error {
	if h.saasClient == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "SaaS client not initialized",
		})
	}

	config := h.saasClient.GetConfig()

	// Hide API key in response (only show partial)
	maskedKey := config.APIKey
	if len(maskedKey) > 20 {
		maskedKey = maskedKey[:12] + "..." + maskedKey[len(maskedKey)-4:]
	}

	return c.JSON(fiber.Map{
		"enabled":          config.Enabled,
		"server_url":       config.ServerURL,
		"device_id":        config.DeviceID,
		"api_key":          maskedKey,
		"provisioning_code": config.ProvisioningCode,
		"enable_tls":       config.EnableTLS,
		"is_provisioned":   config.IsProvisioned(),
		"is_connected":     h.saasClient.IsConnected(),
	})
}

// UpdateConfig updates SaaS configuration
func (h *SaaSHandler) UpdateConfig(c *fiber.Ctx) error {
	if h.saasClient == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "SaaS client not initialized",
		})
	}

	var req struct {
		Enabled      bool   `json:"enabled"`
		ServerURL    string `json:"server_url"`
		DeviceID     string `json:"device_id"`
		APIKey       string `json:"api_key"`
		EnableTLS    bool   `json:"enable_tls"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update configuration
	config := h.saasClient.GetConfig()
	config.Enabled = req.Enabled
	config.ServerURL = req.ServerURL
	config.DeviceID = req.DeviceID
	if req.APIKey != "" && !contains(req.APIKey, "...") {
		// Only update API key if it's not the masked version
		config.APIKey = req.APIKey
	}
	config.EnableTLS = req.EnableTLS

	// TODO: Save config to disk

	return c.JSON(fiber.Map{
		"message": "Configuration updated",
	})
}

// GetStatus returns current connection status
func (h *SaaSHandler) GetStatus(c *fiber.Ctx) error {
	if h.saasClient == nil {
		return c.JSON(fiber.Map{
			"connected": false,
		})
	}

	config := h.saasClient.GetConfig()

	return c.JSON(fiber.Map{
		"connected":      h.saasClient.IsConnected(),
		"device_id":      config.DeviceID,
		"last_heartbeat": "", // TODO: Add heartbeat tracking
		"uptime":         "", // TODO: Add uptime tracking
	})
}

// Provision registers device with SaaS
func (h *SaaSHandler) Provision(c *fiber.Ctx) error {
	if h.saasClient == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "SaaS client not initialized",
		})
	}

	var req struct {
		ProvisioningCode string `json:"provisioning_code"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.ProvisioningCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Provisioning code is required",
		})
	}

	// Update config with provisioning code
	config := h.saasClient.GetConfig()
	config.ProvisioningCode = req.ProvisioningCode

	// Provision device
	if err := h.saasClient.Provision(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get updated config after provisioning
	config = h.saasClient.GetConfig()

	return c.JSON(fiber.Map{
		"device_id": config.DeviceID,
		"api_key":   config.APIKey, // Return full key only once during provisioning
	})
}

// Connect initiates SaaS connection
func (h *SaaSHandler) Connect(c *fiber.Ctx) error {
	if h.saasClient == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "SaaS client not initialized",
		})
	}

	if h.saasClient.IsConnected() {
		return c.JSON(fiber.Map{
			"message": "Already connected",
		})
	}

	// Start connection in background
	go func() {
		if err := h.saasClient.Start(); err != nil {
			// Log error but don't fail the request
			// Client will retry automatically
		}
	}()

	return c.JSON(fiber.Map{
		"message": "Connecting...",
	})
}

// Disconnect closes SaaS connection
func (h *SaaSHandler) Disconnect(c *fiber.Ctx) error {
	if h.saasClient == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "SaaS client not initialized",
		})
	}

	if err := h.saasClient.Stop(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Disconnected",
	})
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
