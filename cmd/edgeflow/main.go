package main

import (
	"fmt"
	stdlog "log"
	"os"

	"github.com/edgeflow/edgeflow/internal/api"
	"github.com/edgeflow/edgeflow/internal/logger"
	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/edgeflow/edgeflow/internal/saas"
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
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

var Version = "0.1.0"

func main() {
	fmt.Println("╔═══════════════════════════════════════╗")
	fmt.Printf("║       EdgeFlow v%-20s ║\n", Version)
	fmt.Println("║ Lightweight Edge & IoT Automation Platform ║")
	fmt.Println("╚═══════════════════════════════════════╝")

	// Initialize structured logger (Zap + Lumberjack file rotation)
	logCfg := logger.DefaultConfig()
	logCfg.LogDir = getEnv("EDGEFLOW_LOG_DIR", "./logs")
	logCfg.Level = getEnv("EDGEFLOW_LOG_LEVEL", "info")
	if err := logger.Init(logCfg); err != nil {
		stdlog.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Redirect stdlib log to structured logger
	stdlog.SetOutput(logger.Writer())
	stdlog.SetFlags(0)

	// Initialize Hardware Abstraction Layer (GPIO, I2C, SPI)
	initHAL()

	// Initialize storage
	storageBackend, err := storage.NewFileStorage("./data")
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}
	defer storageBackend.Close()

	// Initialize node registry and register all modules
	// Use GetGlobalRegistry() so that nodes registered via init() are included
	registry := node.GetGlobalRegistry()
	registerModules(registry)

	// Seed default example flows if data directory is empty
	seedDefaultFlows(storageBackend)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Wire logger's WebSocket broadcaster to send logs to frontend LogPanel
	logger.SetBroadcaster(func(level, message, source string, fields map[string]interface{}) {
		wsHub.Broadcast(websocket.MessageTypeLog, map[string]interface{}{
			"level":   level,
			"message": message,
			"source":  source,
			"fields":  fields,
		})
	})

	// Initialize API service
	service := api.NewService(storageBackend, registry, wsHub)
	handler := api.NewHandler(service)

	// Initialize SaaS client (optional - configured via environment)
	saasConfig := getSaaSConfig()
	saasClient := saas.NewClient(saasConfig, zap.L(), "")
	serviceAdapter := saas.NewServiceAdapter(service)
	if err := saasClient.Initialize(serviceAdapter, service); err != nil {
		logger.Warn("Failed to initialize SaaS client", zap.Error(err))
	} else {
		logger.Info("SaaS client initialized",
			zap.String("server", saasConfig.ServerURL),
			zap.Bool("provisioned", saasConfig.IsProvisioned()))

		// Register SaaS API handler
		saasHandler := api.NewSaaSHandler(saasClient)
		handler.SetSaaSHandler(saasHandler)

		// Start SaaS connection in background if enabled
		if saasConfig.Enabled {
			go func() {
				if err := saasClient.Start(); err != nil {
					logger.Error("SaaS connection failed", zap.Error(err))
				}
			}()

			// Ensure graceful shutdown
			defer saasClient.Stop()
		}
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "EdgeFlow v" + Version,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(fiberLogger.New())
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

	// Serve frontend static files from ./web/dist (production build)
	webDist := getEnv("EDGEFLOW_WEB_DIR", "./web/dist")
	if _, err := os.Stat(webDist); err == nil {
		app.Static("/", webDist, fiber.Static{
			Index:    "index.html",
			Compress: true,
		})
		// SPA fallback: serve index.html for all non-API, non-WS routes
		app.Get("/*", func(c *fiber.Ctx) error {
			return c.SendFile(webDist + "/index.html")
		})
		logger.Info("Serving frontend", zap.String("dir", webDist))
	} else {
		logger.Warn("Frontend directory not found, serving API only", zap.String("dir", webDist))
	}

	port := getEnv("EDGEFLOW_SERVER_PORT", "8080")
	host := getEnv("EDGEFLOW_SERVER_HOST", "0.0.0.0")
	addr := fmt.Sprintf("%s:%s", host, port)

	logger.Info("Server starting", zap.String("addr", fmt.Sprintf("http://%s", addr)))
	logger.Info("Endpoints ready",
		zap.String("health", fmt.Sprintf("http://%s/api/health", addr)),
		zap.String("api", fmt.Sprintf("http://%s/api/v1", addr)),
		zap.String("ws", fmt.Sprintf("ws://%s/ws", addr)),
	)

	if err := app.Listen(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getSaaSConfig() *saas.Config {
	config := saas.DefaultConfig()

	// Read from environment variables
	if enabled := os.Getenv("EDGEFLOW_SAAS_ENABLED"); enabled == "true" || enabled == "1" {
		config.Enabled = true
	}
	if serverURL := os.Getenv("EDGEFLOW_SAAS_URL"); serverURL != "" {
		config.ServerURL = serverURL
	}
	if deviceID := os.Getenv("EDGEFLOW_DEVICE_ID"); deviceID != "" {
		config.DeviceID = deviceID
	}
	if apiKey := os.Getenv("EDGEFLOW_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}
	if provCode := os.Getenv("EDGEFLOW_PROVISIONING_CODE"); provCode != "" {
		config.ProvisioningCode = provCode
	}
	if tls := os.Getenv("EDGEFLOW_SAAS_TLS"); tls == "false" || tls == "0" {
		config.EnableTLS = false
	}

	return config
}

func registerModules(registry *node.Registry) {
	logger.Info("Registering node modules...")

	// Register core nodes
	if err := coreNodes.RegisterAllNodes(registry); err != nil {
		logger.Warn("Failed to register core nodes", zap.Error(err))
	} else {
		logger.Info("Core nodes registered")
	}

	// Register dashboard widgets
	if err := dashboardNodes.RegisterAll(registry); err != nil {
		logger.Warn("Failed to register dashboard widgets", zap.Error(err))
	} else {
		logger.Info("Dashboard widgets registered")
	}

	// Register GPIO nodes (stubs on non-Linux)
	if err := gpioNodes.RegisterAllNodes(registry); err != nil {
		logger.Warn("Failed to register GPIO nodes", zap.Error(err))
	} else {
		logger.Info("GPIO nodes registered")
	}

	// Register network nodes
	networkNodes.RegisterAllNodes(registry)
	logger.Info("Network nodes registered")

	// Register database nodes
	databaseNodes.RegisterAllNodes(registry)
	logger.Info("Database nodes registered")

	// Register storage nodes
	storageNodes.RegisterAllNodes(registry)
	logger.Info("Storage nodes registered")

	// Register messaging nodes
	if err := messagingNodes.RegisterAllNodes(registry); err != nil {
		logger.Warn("Failed to register messaging nodes", zap.Error(err))
	} else {
		logger.Info("Messaging nodes registered")
	}

	// Register AI nodes
	if err := aiNodes.RegisterAllNodes(registry); err != nil {
		logger.Warn("Failed to register AI nodes", zap.Error(err))
	} else {
		logger.Info("AI nodes registered")
	}

	// Register parser nodes (HTML parser)
	if err := parserNodes.RegisterNodes(registry); err != nil {
		logger.Warn("Failed to register parser nodes", zap.Error(err))
	} else {
		logger.Info("Parser nodes registered")
	}

	// Industrial protocol nodes (Modbus TCP/RTU, OPC-UA) are auto-registered via init()
	_ = industrialNodes.RegisterNodes
	logger.Info("Industrial nodes registered (Modbus TCP/RTU, OPC-UA)")

	// Wireless protocol nodes (BLE, Zigbee, Z-Wave) are auto-registered via init()
	_ = wirelessNodes.RegisterNodes
	logger.Info("Wireless nodes registered (BLE, Zigbee, Z-Wave)")

	logger.Info("Node registration complete")
}
