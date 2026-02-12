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

// HTTPInConfig configuration for the HTTP-In node
type HTTPInConfig struct {
	Path     string `json:"path"`
	Method   string `json:"method"`
	AuthType string `json:"authType"`
	AuthVal  string `json:"authValue"`
	RawBody  bool   `json:"rawBody"`
	CORS     bool   `json:"cors"`
}

// HTTPInExecutor creates an HTTP endpoint that listens for requests
type HTTPInExecutor struct {
	config     HTTPInConfig
	outputChan chan node.Message
	registered bool
	mu         sync.RWMutex
}

// httpInRegistry global registry for http-in endpoints
var (
	httpInRegistry = make(map[string]*HTTPInExecutor)
	httpInMu       sync.RWMutex
)

// NewHTTPInExecutor creates a new HTTPInExecutor
func NewHTTPInExecutor() node.Executor {
	return &HTTPInExecutor{
		outputChan: make(chan node.Message, 100),
	}
}

// Init initializes the executor with configuration
func (e *HTTPInExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var cfg HTTPInConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return fmt.Errorf("invalid http-in config: %w", err)
	}

	if cfg.Path == "" {
		return fmt.Errorf("http-in path is required")
	}
	if !strings.HasPrefix(cfg.Path, "/") {
		cfg.Path = "/" + cfg.Path
	}
	if cfg.Method == "" {
		cfg.Method = "ALL"
	}
	cfg.Method = strings.ToUpper(cfg.Method)
	if cfg.AuthType == "" {
		cfg.AuthType = "none"
	}

	e.config = cfg
	return nil
}

// Execute waits for incoming HTTP requests on the configured path
func (e *HTTPInExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if !e.registered {
		if err := e.register(); err != nil {
			return node.Message{}, fmt.Errorf("failed to register http-in endpoint: %w", err)
		}
		e.registered = true
	}

	select {
	case <-ctx.Done():
		return node.Message{}, ctx.Err()
	case incoming := <-e.outputChan:
		return incoming, nil
	}
}

func (e *HTTPInExecutor) register() error {
	httpInMu.Lock()
	defer httpInMu.Unlock()

	key := e.config.Method + ":" + e.config.Path
	if _, exists := httpInRegistry[key]; exists {
		return fmt.Errorf("http-in endpoint %s %s already registered", e.config.Method, e.config.Path)
	}
	httpInRegistry[key] = e
	return nil
}

// HandleHTTPIn handles incoming HTTP requests for this endpoint
func (e *HTTPInExecutor) HandleHTTPIn(c *fiber.Ctx) error {
	// Check method
	if e.config.Method != "ALL" && c.Method() != e.config.Method {
		return c.Status(http.StatusMethodNotAllowed).JSON(fiber.Map{"error": "method not allowed"})
	}

	// Check auth
	if !e.checkAuth(c) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Parse body
	var body interface{}
	if e.config.RawBody {
		body = string(c.Body())
	} else {
		ct := c.Get("Content-Type")
		if strings.Contains(ct, "application/json") {
			if err := json.Unmarshal(c.Body(), &body); err != nil {
				body = string(c.Body())
			}
		} else if strings.Contains(ct, "application/x-www-form-urlencoded") {
			form := make(map[string]string)
			c.Request().PostArgs().VisitAll(func(key, value []byte) {
				form[string(key)] = string(value)
			})
			body = form
		} else {
			body = string(c.Body())
		}
	}

	// Extract path parameters
	params := make(map[string]string)
	for _, p := range c.Route().Params {
		params[p] = c.Params(p)
	}

	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"method":  c.Method(),
			"path":    c.Path(),
			"params":  params,
			"query":   c.Queries(),
			"headers": c.GetReqHeaders(),
			"body":    body,
			"ip":      c.IP(),
		},
	}

	select {
	case e.outputChan <- msg:
	default:
	}

	// HTTP-In does not auto-respond; the response is sent by an http-response node.
	// For now return 200 accepted.
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "request received",
	})
}

func (e *HTTPInExecutor) checkAuth(c *fiber.Ctx) bool {
	switch e.config.AuthType {
	case "none":
		return true
	case "basic":
		auth := c.Get("Authorization")
		return strings.HasPrefix(auth, "Basic ") && auth == "Basic "+e.config.AuthVal
	case "bearer":
		auth := c.Get("Authorization")
		return strings.HasPrefix(auth, "Bearer ") && auth == "Bearer "+e.config.AuthVal
	case "apikey":
		return c.Get("X-API-Key") == e.config.AuthVal
	default:
		return false
	}
}

// Cleanup releases resources
func (e *HTTPInExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.registered {
		httpInMu.Lock()
		key := e.config.Method + ":" + e.config.Path
		delete(httpInRegistry, key)
		httpInMu.Unlock()
		e.registered = false
	}
	close(e.outputChan)
	return nil
}

// GetHTTPInHandler retrieves the handler for a given method and path
func GetHTTPInHandler(method, path string) (*HTTPInExecutor, bool) {
	httpInMu.RLock()
	defer httpInMu.RUnlock()

	// Try exact match first
	key := method + ":" + path
	if ex, ok := httpInRegistry[key]; ok {
		return ex, true
	}
	// Try ALL method
	key = "ALL:" + path
	ex, ok := httpInRegistry[key]
	return ex, ok
}
