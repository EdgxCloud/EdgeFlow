package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// HTTP client for marketplace search with custom transport
var marketplaceHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
	},
}

// Handler holds the service dependencies for HTTP handlers
type Handler struct {
	service   *Service
	moduleAPI *ModuleAPI
}

// NewHandler creates a new HTTP handler
func NewHandler(service *Service) *Handler {
	moduleAPI, _ := NewModuleAPI("./modules")
	return &Handler{
		service:   service,
		moduleAPI: moduleAPI,
	}
}

// SetupRoutes configures all API routes with the handler
func (h *Handler) SetupRoutes(app *fiber.App) {
	// API v1 group
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", h.healthCheck)

	// Flow routes
	flowRoutes := api.Group("/flows")
	flowRoutes.Get("/", h.listFlows)
	flowRoutes.Post("/", h.createFlow)
	flowRoutes.Get("/:id", h.getFlow)
	flowRoutes.Put("/:id", h.updateFlow)
	flowRoutes.Delete("/:id", h.deleteFlow)
	flowRoutes.Post("/:id/start", h.startFlow)
	flowRoutes.Post("/:id/stop", h.stopFlow)

	// Node routes
	nodeRoutes := api.Group("/flows/:flowId/nodes")
	nodeRoutes.Get("/", h.listNodes)
	nodeRoutes.Post("/", h.addNode)
	nodeRoutes.Get("/:nodeId", h.getNode)
	nodeRoutes.Put("/:nodeId", h.updateNode)
	nodeRoutes.Delete("/:nodeId", h.deleteNode)

	// Connection routes
	connRoutes := api.Group("/flows/:flowId/connections")
	connRoutes.Get("/", h.listConnections)
	connRoutes.Post("/", h.createConnection)
	connRoutes.Delete("/:connId", h.deleteConnection)

	// Node types catalog
	api.Get("/node-types", h.listNodeTypes)
	api.Get("/node-types/:type", h.getNodeType)

	// Module search routes - use Handler methods
	api.Get("/modules/search/npm", h.searchNPM)
	api.Get("/modules/search/nodered", h.searchNodeRED)
	api.Get("/modules/search/github", h.searchGitHub)

	// Module install/upload routes
	if h.moduleAPI != nil {
		api.Post("/modules/install", h.moduleAPI.InstallModule)
		api.Post("/modules/upload", h.moduleAPI.UploadModule)
	}

	// Module routes
	moduleRoutes := api.Group("/modules")
	moduleRoutes.Get("/", h.listModules)
	moduleRoutes.Get("/stats", h.getModuleStats)
	moduleRoutes.Get("/:name", h.getModule)
	moduleRoutes.Post("/:name/load", h.loadModule)
	moduleRoutes.Post("/:name/unload", h.unloadModule)
	moduleRoutes.Post("/:name/enable", h.enableModule)
	moduleRoutes.Post("/:name/disable", h.disableModule)
	moduleRoutes.Post("/:name/reload", h.reloadModule)
	if h.moduleAPI != nil {
		moduleRoutes.Delete("/:name", h.moduleAPI.UninstallModule)
	}

	// Resource routes
	api.Get("/resources/stats", h.getResourceStats)

	// WebSocket for real-time updates
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		h.service.wsHub.HandleWebSocket(c)
	}))

	// Subflow routes
	h.SetupSubflowRoutes(app)
}

// Health check handlers
func (h *Handler) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":           "healthy",
		"service":          "edgeflow",
		"version":          "0.1.0",
		"websocket_clients": h.service.wsHub.GetClientCount(),
	})
}

// Flow handlers
func (h *Handler) listFlows(c *fiber.Ctx) error {
	// Get flows from storage to preserve node data
	storageFlows, err := h.service.ListStorageFlows()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert storage flows to frontend format
	flowsList := make([]fiber.Map, 0)
	for _, sf := range storageFlows {
		// Convert nodes array to map
		nodesMap := make(map[string]interface{})
		for _, nodeMap := range sf.Nodes {
			if id, ok := nodeMap["id"].(string); ok {
				nodesMap[id] = nodeMap
			}
		}

		// Get connections directly from storage
		connections := sf.Connections
		if connections == nil {
			connections = make([]map[string]interface{}, 0)
		}

		flowsList = append(flowsList, fiber.Map{
			"id":          sf.ID,
			"name":        sf.Name,
			"description": sf.Description,
			"status":      sf.Status,
			"nodes":       nodesMap,
			"connections": connections,
			"config":      make(map[string]interface{}),
		})
	}

	return c.JSON(fiber.Map{
		"flows": flowsList,
		"count": len(flowsList),
	})
}

func (h *Handler) createFlow(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Flow name is required",
		})
	}

	flow, err := h.service.CreateFlow(req.Name, req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(flow)
}

func (h *Handler) getFlow(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get flow from storage to preserve raw node data (config, position, etc.)
	storageFlow, err := h.service.GetStorageFlow(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Flow not found",
		})
	}

	// Convert storage nodes (array) to map format for frontend
	nodesMap := make(map[string]interface{})
	for _, nodeMap := range storageFlow.Nodes {
		if id, ok := nodeMap["id"].(string); ok {
			nodesMap[id] = nodeMap
		}
	}

	// Get connections directly from storage
	connections := storageFlow.Connections
	if connections == nil {
		connections = make([]map[string]interface{}, 0)
	}

	// Return flow data in frontend-compatible format
	return c.JSON(fiber.Map{
		"id":          storageFlow.ID,
		"name":        storageFlow.Name,
		"description": storageFlow.Description,
		"status":      storageFlow.Status,
		"nodes":       nodesMap,
		"connections": connections,
		"config":      make(map[string]interface{}),
	})
}

func (h *Handler) updateFlow(c *fiber.Ctx) error {
	id := c.Params("id")

	// Parse request body as raw map to preserve all node data
	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get existing flow
	flow, err := h.service.GetFlow(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Flow not found",
		})
	}

	// Update basic fields
	if name, ok := updateData["name"].(string); ok {
		flow.Name = name
	}
	if desc, ok := updateData["description"].(string); ok {
		flow.Description = desc
	}
	if config, ok := updateData["config"].(map[string]interface{}); ok {
		flow.Config = config
	}

	// Update flow
	if err := h.service.UpdateFlow(flow); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Also update the storage directly with raw node data to preserve positions/config
	if nodes, ok := updateData["nodes"].(map[string]interface{}); ok {
		storageFlow, err := h.service.GetStorageFlow(id)
		if err == nil {
			// Convert nodes map to slice for storage
			nodeSlice := make([]map[string]interface{}, 0)
			for nodeId, nodeData := range nodes {
				if nodeMap, ok := nodeData.(map[string]interface{}); ok {
					nodeMap["id"] = nodeId // Ensure ID is set
					nodeSlice = append(nodeSlice, nodeMap)
				}
			}
			storageFlow.Nodes = nodeSlice

			// Update connections
			if connections, ok := updateData["connections"].([]interface{}); ok {
				connSlice := make([]map[string]interface{}, 0)
				for _, conn := range connections {
					if connMap, ok := conn.(map[string]interface{}); ok {
						connSlice = append(connSlice, connMap)
					}
				}
				storageFlow.Connections = connSlice
			}

			// Save to storage
			h.service.UpdateStorageFlow(storageFlow)
		}
	}

	return c.JSON(flow)
}

func (h *Handler) deleteFlow(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.DeleteFlow(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Flow deleted successfully",
		"id":      id,
	})
}

func (h *Handler) startFlow(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.StartFlow(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Flow started successfully",
		"id":      id,
	})
}

func (h *Handler) stopFlow(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.StopFlow(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Flow stopped successfully",
		"id":      id,
	})
}

// Node handlers
func (h *Handler) listNodes(c *fiber.Ctx) error {
	flowID := c.Params("flowId")

	flow, err := h.service.GetFlow(flowID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Flow not found",
		})
	}

	return c.JSON(fiber.Map{
		"flow_id": flowID,
		"nodes":   flow.Nodes,
		"count":   len(flow.Nodes),
	})
}

func (h *Handler) addNode(c *fiber.Ctx) error {
	flowID := c.Params("flowId")

	var req struct {
		Type   string                 `json:"type"`
		Name   string                 `json:"name"`
		Config map[string]interface{} `json:"config"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	node, err := h.service.AddNodeToFlow(flowID, req.Type, req.Name, req.Config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(node)
}

func (h *Handler) getNode(c *fiber.Ctx) error {
	flowID := c.Params("flowId")
	nodeID := c.Params("nodeId")

	flow, err := h.service.GetFlow(flowID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Flow not found",
		})
	}

	node, err := flow.GetNode(nodeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	return c.JSON(node)
}

func (h *Handler) updateNode(c *fiber.Ctx) error {
	flowID := c.Params("flowId")
	nodeID := c.Params("nodeId")

	flow, err := h.service.GetFlow(flowID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Flow not found",
		})
	}

	node, err := flow.GetNode(nodeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	var req struct {
		Config map[string]interface{} `json:"config"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := node.UpdateConfig(req.Config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(node)
}

func (h *Handler) deleteNode(c *fiber.Ctx) error {
	flowID := c.Params("flowId")
	nodeID := c.Params("nodeId")

	if err := h.service.RemoveNodeFromFlow(flowID, nodeID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Node deleted successfully",
		"id":      nodeID,
	})
}

// Connection handlers
func (h *Handler) listConnections(c *fiber.Ctx) error {
	flowID := c.Params("flowId")

	flow, err := h.service.GetFlow(flowID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Flow not found",
		})
	}

	return c.JSON(fiber.Map{
		"flow_id":     flowID,
		"connections": flow.Connections,
		"count":       len(flow.Connections),
	})
}

func (h *Handler) createConnection(c *fiber.Ctx) error {
	flowID := c.Params("flowId")

	var req struct {
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.service.ConnectNodes(flowID, req.SourceID, req.TargetID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Connection created successfully",
		"source_id": req.SourceID,
		"target_id": req.TargetID,
	})
}

func (h *Handler) deleteConnection(c *fiber.Ctx) error {
	flowID := c.Params("flowId")
	connID := c.Params("connId")

	flow, err := h.service.GetFlow(flowID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Flow not found",
		})
	}

	if err := flow.Disconnect(connID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Connection deleted successfully",
		"id":      connID,
	})
}

// CategoryInfo contains metadata about a node category
type CategoryInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Order       int    `json:"order"`
}

// categoryDefinitions contains all category metadata
var categoryDefinitions = map[string]CategoryInfo{
	"input": {
		ID:          "input",
		Name:        "Input",
		Description: "Nodes that receive data from external sources",
		Icon:        "download",
		Color:       "#10b981",
		Order:       1,
	},
	"output": {
		ID:          "output",
		Name:        "Output",
		Description: "Nodes that send data to external destinations",
		Icon:        "upload",
		Color:       "#ef4444",
		Order:       2,
	},
	"function": {
		ID:          "function",
		Name:        "Function",
		Description: "Logic, transformation, and data processing",
		Icon:        "code",
		Color:       "#f59e0b",
		Order:       3,
	},
	"processing": {
		ID:          "processing",
		Name:        "Processing",
		Description: "Data processing and transformation nodes",
		Icon:        "cpu",
		Color:       "#8b5cf6",
		Order:       4,
	},
	"gpio": {
		ID:          "gpio",
		Name:        "GPIO",
		Description: "Raspberry Pi GPIO pins and basic I/O",
		Icon:        "circuit-board",
		Color:       "#16a34a",
		Order:       5,
	},
	"sensors": {
		ID:          "sensors",
		Name:        "Sensors",
		Description: "Temperature, humidity, light, and other sensors",
		Icon:        "thermometer",
		Color:       "#22c55e",
		Order:       6,
	},
	"actuators": {
		ID:          "actuators",
		Name:        "Actuators",
		Description: "Motors, relays, LEDs, and output devices",
		Icon:        "gauge",
		Color:       "#ec4899",
		Order:       7,
	},
	"communication": {
		ID:          "communication",
		Name:        "Communication",
		Description: "LoRa, NRF24, RF433, and wireless protocols",
		Icon:        "radio",
		Color:       "#0ea5e9",
		Order:       8,
	},
	"network": {
		ID:          "network",
		Name:        "Network",
		Description: "HTTP, MQTT, WebSocket, TCP/UDP protocols",
		Icon:        "network",
		Color:       "#06b6d4",
		Order:       9,
	},
	"database": {
		ID:          "database",
		Name:        "Database",
		Description: "MySQL, PostgreSQL, MongoDB, Redis, InfluxDB",
		Icon:        "database",
		Color:       "#3b82f6",
		Order:       10,
	},
	"storage": {
		ID:          "storage",
		Name:        "Storage",
		Description: "File storage, S3, Google Drive, FTP",
		Icon:        "hard-drive",
		Color:       "#6366f1",
		Order:       11,
	},
	"messaging": {
		ID:          "messaging",
		Name:        "Messaging",
		Description: "Telegram, Email, Slack, Discord notifications",
		Icon:        "message-square",
		Color:       "#14b8a6",
		Order:       12,
	},
	"ai": {
		ID:          "ai",
		Name:        "AI & ML",
		Description: "OpenAI, Anthropic, Ollama LLM integration",
		Icon:        "brain",
		Color:       "#a855f7",
		Order:       13,
	},
	"industrial": {
		ID:          "industrial",
		Name:        "Industrial",
		Description: "Modbus RTU/TCP, OPC-UA protocols",
		Icon:        "factory",
		Color:       "#f97316",
		Order:       14,
	},
	"dashboard": {
		ID:          "dashboard",
		Name:        "Dashboard",
		Description: "UI widgets, charts, gauges, buttons",
		Icon:        "layout-dashboard",
		Color:       "#0891b2",
		Order:       15,
	},
	"advanced": {
		ID:          "advanced",
		Name:        "Advanced",
		Description: "System commands, file operations, utilities",
		Icon:        "settings",
		Color:       "#64748b",
		Order:       99,
	},
}

// Node type handlers
func (h *Handler) listNodeTypes(c *fiber.Ctx) error {
	nodeTypes := h.service.GetNodeTypes()

	// Sort node types: first by category order, then by name within each category
	sort.Slice(nodeTypes, func(i, j int) bool {
		catI := string(nodeTypes[i].Category)
		catJ := string(nodeTypes[j].Category)

		// Get order for categories
		orderI := 99
		orderJ := 99
		if def, ok := categoryDefinitions[catI]; ok {
			orderI = def.Order
		}
		if def, ok := categoryDefinitions[catJ]; ok {
			orderJ = def.Order
		}

		// Sort by category first
		if orderI != orderJ {
			return orderI < orderJ
		}

		// Then sort by name within same category
		return nodeTypes[i].Name < nodeTypes[j].Name
	})

	// Count nodes per category
	categoryCount := make(map[string]int)
	for _, nt := range nodeTypes {
		catStr := string(nt.Category)
		categoryCount[catStr]++
	}

	// Build categories list with full metadata
	categories := make([]fiber.Map, 0)
	addedCategories := make(map[string]bool)

	for _, nt := range nodeTypes {
		catStr := string(nt.Category)
		if catStr == "" || addedCategories[catStr] {
			continue
		}
		addedCategories[catStr] = true

		// Get category definition or create default
		def, ok := categoryDefinitions[catStr]
		if !ok {
			def = CategoryInfo{
				ID:          catStr,
				Name:        catStr,
				Description: catStr + " nodes",
				Icon:        "box",
				Color:       "#64748b",
				Order:       99,
			}
		}

		categories = append(categories, fiber.Map{
			"id":          def.ID,
			"name":        def.Name,
			"description": def.Description,
			"icon":        def.Icon,
			"color":       def.Color,
			"order":       def.Order,
			"count":       categoryCount[catStr],
		})
	}

	// Sort categories by order
	sort.Slice(categories, func(i, j int) bool {
		orderI := categories[i]["order"].(int)
		orderJ := categories[j]["order"].(int)
		return orderI < orderJ
	})

	return c.JSON(fiber.Map{
		"categories": categories,
		"node_types": nodeTypes,
		"count":      len(nodeTypes),
	})
}

func (h *Handler) getNodeType(c *fiber.Ctx) error {
	nodeType := c.Params("type")

	info, err := h.service.GetNodeTypeInfo(nodeType)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node type not found",
		})
	}

	return c.JSON(info)
}

// Module handlers
func (h *Handler) listModules(c *fiber.Ctx) error {
	modules := h.service.ListModules()

	return c.JSON(fiber.Map{
		"modules": modules,
		"count":   len(modules),
	})
}

func (h *Handler) getModule(c *fiber.Ctx) error {
	name := c.Params("name")

	module, err := h.service.GetModuleInfo(name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Module not found",
		})
	}

	return c.JSON(module)
}

func (h *Handler) loadModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if err := h.service.LoadModule(name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Module loaded successfully",
		"module":  name,
	})
}

func (h *Handler) unloadModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if err := h.service.UnloadModule(name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Module unloaded successfully",
		"module":  name,
	})
}

func (h *Handler) enableModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if err := h.service.EnableModule(name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Module enabled successfully",
		"module":  name,
	})
}

func (h *Handler) disableModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if err := h.service.DisableModule(name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Module disabled successfully",
		"module":  name,
	})
}

func (h *Handler) reloadModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if err := h.service.ReloadModule(name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Module reloaded successfully",
		"module":  name,
	})
}

func (h *Handler) getModuleStats(c *fiber.Ctx) error {
	stats := h.service.GetModuleStats()

	return c.JSON(stats)
}

// Resource handlers
func (h *Handler) getResourceStats(c *fiber.Ctx) error {
	stats := h.service.GetResourceStats()

	// Transform to frontend format
	response := fiber.Map{
		"timestamp": stats.Timestamp,
		"cpu": fiber.Map{
			"usage_percent": 0.0, // Placeholder - need CPU monitoring
			"cores":         stats.CPUCores,
		},
		"memory": fiber.Map{
			"total_bytes":   stats.MemoryTotal,
			"used_bytes":    stats.MemoryUsed,
			"free_bytes":    stats.MemoryAvailable,
			"usage_percent": stats.MemoryPercent,
		},
		"disk": fiber.Map{
			"total_bytes":   stats.DiskTotal,
			"used_bytes":    stats.DiskUsed,
			"free_bytes":    stats.DiskAvailable,
			"usage_percent": stats.DiskPercent,
		},
		"goroutines": stats.GoroutineCount,
	}

	return c.JSON(response)
}

// ============================================
// Module Marketplace Search Handlers
// ============================================

// NPM Search types
type npmSearchResult struct {
	Objects []npmSearchObject `json:"objects"`
	Total   int               `json:"total"`
}

type npmSearchObject struct {
	Package npmSearchPackage `json:"package"`
	Score   npmSearchScore   `json:"score"`
}

type npmSearchPackage struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Publisher   struct {
		Username string `json:"username"`
	} `json:"publisher"`
	Links struct {
		NPM        string `json:"npm"`
		Homepage   string `json:"homepage"`
		Repository string `json:"repository"`
	} `json:"links"`
}

type npmSearchScore struct {
	Final float64 `json:"final"`
}

// searchNPM searches npm registry for Node-RED and n8n modules
func (h *Handler) searchNPM(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	// Search npm with "node-red" prefix to get Node-RED packages
	searchURL := fmt.Sprintf(
		"https://registry.npmjs.org/-/v1/search?text=node-red+%s&size=100",
		url.QueryEscape(query),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create search request",
		})
	}

	resp, err := marketplaceHTTPClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to reach npm registry: %v", err),
		})
	}
	defer resp.Body.Close()

	var result npmSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse npm response",
		})
	}

	// Return all results - they're already filtered by "node-red" in search query
	results := make([]fiber.Map, 0, len(result.Objects))
	for _, obj := range result.Objects {
		results = append(results, fiber.Map{
			"name":        obj.Package.Name,
			"version":     obj.Package.Version,
			"description": obj.Package.Description,
			"keywords":    obj.Package.Keywords,
			"author":      obj.Package.Publisher.Username,
			"url":         obj.Package.Links.NPM,
			"repository":  obj.Package.Links.Repository,
			"score":       obj.Score.Final,
			"source":      "npm",
		})
	}

	return c.JSON(fiber.Map{
		"results": results,
		"total":   len(results),
		"query":   query,
	})
}

// Node-RED catalog cache
var (
	nodeRedCatalog     []nodeRedCatalogModule
	nodeRedCatalogTime time.Time
	nodeRedCatalogLock sync.RWMutex
)

type nodeRedCatalogModule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Keywords    []string `json:"keywords"`
	Updated     string   `json:"updated_at"`
	Types       []string `json:"types"`
}

// searchNodeRED searches the Node-RED catalog
func (h *Handler) searchNodeRED(c *fiber.Ctx) error {
	query := strings.ToLower(c.Query("q"))
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	catalog, err := getNodeRedCatalog()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch Node-RED catalog: %v", err),
		})
	}

	// Search through catalog
	results := make([]fiber.Map, 0)
	for _, mod := range catalog {
		if matchesQuery(mod, query) {
			results = append(results, fiber.Map{
				"name":        mod.Name,
				"version":     mod.Version,
				"description": mod.Description,
				"keywords":    mod.Keywords,
				"types":       mod.Types,
				"updated":     mod.Updated,
				"source":      "node-red",
				"url":         fmt.Sprintf("https://flows.nodered.org/node/%s", mod.Name),
			})
		}
		if len(results) >= 50 {
			break
		}
	}

	return c.JSON(fiber.Map{
		"results": results,
		"total":   len(results),
		"query":   query,
	})
}

func getNodeRedCatalog() ([]nodeRedCatalogModule, error) {
	nodeRedCatalogLock.RLock()
	if time.Since(nodeRedCatalogTime) < time.Hour && nodeRedCatalog != nil {
		defer nodeRedCatalogLock.RUnlock()
		return nodeRedCatalog, nil
	}
	nodeRedCatalogLock.RUnlock()

	nodeRedCatalogLock.Lock()
	defer nodeRedCatalogLock.Unlock()

	// Double-check after acquiring write lock
	if time.Since(nodeRedCatalogTime) < time.Hour && nodeRedCatalog != nil {
		return nodeRedCatalog, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://catalogue.nodered.org/catalogue.json", nil)
	if err != nil {
		return nil, err
	}

	resp, err := marketplaceHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var catalog struct {
		Modules []nodeRedCatalogModule `json:"modules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return nil, err
	}

	nodeRedCatalog = catalog.Modules
	nodeRedCatalogTime = time.Now()

	return nodeRedCatalog, nil
}

func matchesQuery(mod nodeRedCatalogModule, query string) bool {
	if strings.Contains(strings.ToLower(mod.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(mod.Description), query) {
		return true
	}
	for _, kw := range mod.Keywords {
		if strings.Contains(strings.ToLower(kw), query) {
			return true
		}
	}
	return false
}

// GitHub search types
type githubSearchResult struct {
	TotalCount int                `json:"total_count"`
	Items      []githubRepository `json:"items"`
}

type githubRepository struct {
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	HTMLURL     string   `json:"html_url"`
	Stars       int      `json:"stargazers_count"`
	Forks       int      `json:"forks_count"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
	UpdatedAt   string   `json:"updated_at"`
	Owner       struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
}

// searchGitHub searches GitHub for module repositories
func (h *Handler) searchGitHub(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	// Build GitHub search query focusing on node-red and n8n
	searchQuery := fmt.Sprintf("%s node-red OR n8n", query)
	searchURL := fmt.Sprintf(
		"https://api.github.com/search/repositories?q=%s&sort=stars&order=desc&per_page=30",
		url.QueryEscape(searchQuery),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create search request",
		})
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "EdgeFlow/1.0")

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := marketplaceHTTPClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Failed to reach GitHub API",
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "GitHub rate limit exceeded",
		})
	}

	var result githubSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse GitHub response",
		})
	}

	// Transform results
	results := make([]fiber.Map, 0, len(result.Items))
	for _, repo := range result.Items {
		results = append(results, fiber.Map{
			"name":        repo.FullName,
			"description": repo.Description,
			"url":         repo.HTMLURL,
			"stars":       repo.Stars,
			"forks":       repo.Forks,
			"language":    repo.Language,
			"topics":      repo.Topics,
			"updated":     repo.UpdatedAt,
			"owner":       repo.Owner.Login,
			"avatar":      repo.Owner.AvatarURL,
			"source":      "github",
		})
	}

	return c.JSON(fiber.Map{
		"results": results,
		"total":   result.TotalCount,
		"query":   query,
	})
}
