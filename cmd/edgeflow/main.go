package main

import (
	"fmt"
	"log"
	"os"

	"github.com/edgeflow/edgeflow/internal/api"
	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/edgeflow/edgeflow/internal/storage"
	"github.com/edgeflow/edgeflow/internal/websocket"
	aiNodes "github.com/edgeflow/edgeflow/pkg/nodes/ai"
	coreNodes "github.com/edgeflow/edgeflow/pkg/nodes/core"
	dashboardNodes "github.com/edgeflow/edgeflow/pkg/nodes/dashboard"
	databaseNodes "github.com/edgeflow/edgeflow/pkg/nodes/database"
	gpioNodes "github.com/edgeflow/edgeflow/pkg/nodes/gpio"
	industrialNodes "github.com/edgeflow/edgeflow/pkg/nodes/industrial"
	messagingNodes "github.com/edgeflow/edgeflow/pkg/nodes/messaging"
	networkNodes "github.com/edgeflow/edgeflow/pkg/nodes/network"
	parserNodes "github.com/edgeflow/edgeflow/pkg/nodes/parser"
	storageNodes "github.com/edgeflow/edgeflow/pkg/nodes/storage"
	wirelessNodes "github.com/edgeflow/edgeflow/pkg/nodes/wireless"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var Version = "0.1.0"

func main() {
	fmt.Println("╔═══════════════════════════════════════╗")
	fmt.Printf("║       EdgeFlow v%-20s ║\n", Version)
	fmt.Println("║   پلتفرم اتوماسیون سبک Edge و IoT    ║")
	fmt.Println("╚═══════════════════════════════════════╝")

	// Initialize Hardware Abstraction Layer (GPIO, I2C, SPI)
	initHAL()

	// Initialize storage
	storageBackend, err := storage.NewFileStorage("./data")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storageBackend.Close()

	// Initialize node registry and register all modules
	// Use GetGlobalRegistry() so that nodes registered via init() are included
	registry := node.GetGlobalRegistry()
	registerModules(registry)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize API service
	service := api.NewService(storageBackend, registry, wsHub)
	handler := api.NewHandler(service)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "EdgeFlow v" + Version,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to EdgeFlow!",
			"version": Version,
			"status":  "running",
		})
	})

	// Legacy health check (for compatibility)
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"version": Version,
		})
	})

	// Setup API routes
	handler.SetupRoutes(app)

	port := getEnv("EDGEFLOW_SERVER_PORT", "8080")
	host := getEnv("EDGEFLOW_SERVER_HOST", "0.0.0.0")
	addr := fmt.Sprintf("%s:%s", host, port)

	log.Printf("Server starting on http://%s\n", addr)
	log.Printf("Health check: http://%s/api/health\n", addr)
	log.Printf("API v1: http://%s/api/v1\n", addr)
	log.Printf("WebSocket: ws://%s/ws\n", addr)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func registerModules(registry *node.Registry) {
	log.Println("Registering node modules...")

	// Register core nodes
	if err := coreNodes.RegisterAllNodes(registry); err != nil {
		log.Printf("Warning: Failed to register core nodes: %v", err)
	} else {
		log.Println("✅ Core nodes registered successfully")
	}

	// Register dashboard widgets
	if err := dashboardNodes.RegisterAll(registry); err != nil {
		log.Printf("Warning: Failed to register dashboard widgets: %v", err)
	} else {
		log.Println("✅ Dashboard widgets (14) registered successfully")
	}

	// Register GPIO nodes (52 hardware nodes - stubs on non-Linux)
	if err := gpioNodes.RegisterAllNodes(registry); err != nil {
		log.Printf("Warning: Failed to register GPIO nodes: %v", err)
	} else {
		log.Println("✅ GPIO nodes (52) registered successfully")
	}

	// Register network nodes
	networkNodes.RegisterAllNodes(registry)
	log.Println("✅ Network nodes registered successfully")

	// Register database nodes
	databaseNodes.RegisterAllNodes(registry)
	log.Println("✅ Database nodes registered successfully")

	// Register storage nodes
	storageNodes.RegisterAllNodes(registry)
	log.Println("✅ Storage nodes registered successfully")

	// Register messaging nodes
	if err := messagingNodes.RegisterAllNodes(registry); err != nil {
		log.Printf("Warning: Failed to register messaging nodes: %v", err)
	} else {
		log.Println("✅ Messaging nodes registered successfully")
	}

	// Register AI nodes
	if err := aiNodes.RegisterAllNodes(registry); err != nil {
		log.Printf("Warning: Failed to register AI nodes: %v", err)
	} else {
		log.Println("✅ AI nodes registered successfully")
	}

	// Register parser nodes (HTML parser)
	if err := parserNodes.RegisterNodes(registry); err != nil {
		log.Printf("Warning: Failed to register parser nodes: %v", err)
	} else {
		log.Println("✅ Parser nodes registered successfully")
	}

	// Industrial protocol nodes (Modbus TCP/RTU, OPC-UA) are auto-registered via init()
	// Just importing the package triggers registration
	_ = industrialNodes.RegisterNodes // Ensure import is used
	log.Println("✅ Industrial nodes registered (Modbus TCP/RTU, OPC-UA)")

	// Wireless protocol nodes (BLE, Zigbee, Z-Wave) are auto-registered via init()
	// Just importing the package triggers registration
	_ = wirelessNodes.RegisterNodes // Ensure import is used
	log.Println("✅ Wireless nodes registered (BLE, Zigbee, Z-Wave)")

	log.Println("✅ Node registration complete")
}
