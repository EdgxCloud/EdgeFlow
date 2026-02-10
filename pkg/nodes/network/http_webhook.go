package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/gofiber/fiber/v2"
)

// HTTPWebhookConfig configuration for the HTTP Webhook node
type HTTPWebhookConfig struct {
	Path       string            `json:"path"`       // Webhook path (e.g., /webhook/myflow)
	Method     string            `json:"method"`     // HTTP method (default: POST)
	AuthType   string            `json:"authType"`   // none, basic, bearer, apikey
	AuthValue  string            `json:"authValue"`  // Auth value
	Headers    map[string]string `json:"headers"`    // Expected headers
	RawBody    bool              `json:"rawBody"`    // Return raw body instead of parsing
}

// HTTPWebhookExecutor executor for the HTTP Webhook node
type HTTPWebhookExecutor struct {
	config       HTTPWebhookConfig
	outputChan   chan node.Message
	server       *fiber.App
	registered   bool
	mu           sync.RWMutex
}

// WebhookRegistry global registry for webhooks
var (
	webhookRegistry = make(map[string]*HTTPWebhookExecutor)
	webhookMu       sync.RWMutex
)

// NewHTTPWebhookExecutor creates a new HTTPWebhookExecutor
func NewHTTPWebhookExecutor() node.Executor {
	return &HTTPWebhookExecutor{
		outputChan: make(chan node.Message, 100),
	}
}

// Init initializes the executor with configuration
func (e *HTTPWebhookExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var webhookConfig HTTPWebhookConfig
	if err := json.Unmarshal(configJSON, &webhookConfig); err != nil {
		return fmt.Errorf("invalid webhook config: %w", err)
	}

	// Validate path
	if webhookConfig.Path == "" {
		return fmt.Errorf("webhook path is required")
	}

	if !strings.HasPrefix(webhookConfig.Path, "/") {
		webhookConfig.Path = "/" + webhookConfig.Path
	}

	// Default method
	if webhookConfig.Method == "" {
		webhookConfig.Method = "POST"
	}
	webhookConfig.Method = strings.ToUpper(webhookConfig.Method)

	// Default auth type
	if webhookConfig.AuthType == "" {
		webhookConfig.AuthType = "none"
	}

	e.config = webhookConfig
	return nil
}

// Execute executes the node (webhook operates passively)
func (e *HTTPWebhookExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Register webhook if not already registered
	if !e.registered {
		if err := e.registerWebhook(); err != nil {
			return node.Message{}, fmt.Errorf("failed to register webhook: %w", err)
		}
		e.registered = true
	}

	// Wait for incoming webhook or context cancellation
	select {
	case <-ctx.Done():
		return node.Message{}, ctx.Err()
	case webhookMsg := <-e.outputChan:
		return webhookMsg, nil
	}
}

// registerWebhook registers the webhook on the HTTP server
func (e *HTTPWebhookExecutor) registerWebhook() error {
	webhookMu.Lock()
	defer webhookMu.Unlock()

	// Check if path already registered
	if _, exists := webhookRegistry[e.config.Path]; exists {
		return fmt.Errorf("webhook path %s already registered", e.config.Path)
	}

	// Register in registry
	webhookRegistry[e.config.Path] = e

	return nil
}

// HandleWebhook handler for webhook requests
func (e *HTTPWebhookExecutor) HandleWebhook(c *fiber.Ctx) error {
	// Check method
	if c.Method() != e.config.Method {
		return c.Status(http.StatusMethodNotAllowed).JSON(fiber.Map{
			"error": "Method not allowed",
		})
	}

	// Check authentication
	if !e.checkAuth(c) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Check headers
	for key, expectedValue := range e.config.Headers {
		actualValue := c.Get(key)
		if actualValue != expectedValue {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid header: %s", key),
			})
		}
	}

	// Parse body
	var body interface{}
	if e.config.RawBody {
		body = string(c.Body())
	} else {
		contentType := c.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			if err := c.BodyParser(&body); err != nil {
				body = string(c.Body())
			}
		} else {
			body = string(c.Body())
		}
	}

	// Create message
	msg := node.Message{
		Payload: map[string]interface{}{
			"method":  c.Method(),
			"path":    c.Path(),
			"query":   c.Queries(),
			"headers": c.GetReqHeaders(),
			"body":    body,
			"ip":      c.IP(),
		},
	}

	// Send to output channel (non-blocking)
	select {
	case e.outputChan <- msg:
	default:
		// Channel full, log warning
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Webhook received",
	})
}

// checkAuth verifies authentication
func (e *HTTPWebhookExecutor) checkAuth(c *fiber.Ctx) bool {
	switch e.config.AuthType {
	case "none":
		return true

	case "basic":
		auth := c.Get("Authorization")
		return strings.HasPrefix(auth, "Basic ") && auth == "Basic "+e.config.AuthValue

	case "bearer":
		auth := c.Get("Authorization")
		return strings.HasPrefix(auth, "Bearer ") && auth == "Bearer "+e.config.AuthValue

	case "apikey":
		apiKey := c.Get("X-API-Key")
		return apiKey == e.config.AuthValue

	default:
		return false
	}
}

// Cleanup releases resources
func (e *HTTPWebhookExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.registered {
		webhookMu.Lock()
		delete(webhookRegistry, e.config.Path)
		webhookMu.Unlock()
		e.registered = false
	}

	close(e.outputChan)
	return nil
}

// GetWebhookHandler retrieves the handler for a given path
func GetWebhookHandler(path string) (*HTTPWebhookExecutor, bool) {
	webhookMu.RLock()
	defer webhookMu.RUnlock()
	executor, exists := webhookRegistry[path]
	return executor, exists
}
