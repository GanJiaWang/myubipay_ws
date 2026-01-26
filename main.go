package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-ubipay-websocket/config"
	"go-ubipay-websocket/cron"
	"go-ubipay-websocket/database"
	"go-ubipay-websocket/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	fiberwebsocket "github.com/gofiber/websocket/v2"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Println("🚀 Starting Real-Time Point Mining System (MVP)")
	log.Printf("📋 Configuration loaded: Port=%s, MongoDB=%s", cfg.ServerPort, cfg.MongoDBName)
	log.Println("🔧 System will automatically use test mode if MongoDB is not available")

	// Initialize database (MongoDB operations commented out for testing)
	db, err := database.ConnectMongoDB(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer db.Disconnect()

	// Initialize session manager
	sessionManager := websocket.NewSessionManager()

	// Initialize WebSocket handler
	wsHandler := websocket.NewWebSocketHandler(cfg, sessionManager, db)

	// Initialize accrual job
	accrualJob := cron.NewAccrualJob(cfg, sessionManager, db, wsHandler)
	accrualJob.Start()
	defer accrualJob.Stop()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "UbiPay WebSocket Server",
	})

	// Middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		})
	})

	// WebSocket endpoint (authentication removed for testing)
	app.Get("/ws", wsHandler.HandleWebSocket, func(c *fiber.Ctx) error {
		// Store the fiber context in locals for WebSocket handler to access
		c.Locals("fiberCtx", c)
		return fiberwebsocket.New(wsHandler.WebSocketConnection)(c)
	})

	// Manual accrual trigger endpoint (for testing)
	app.Post("/admin/accrual/run", func(c *fiber.Ctx) error {
		accrualJob.RunManualAccrual()
		return c.JSON(fiber.Map{
			"message": "Manual accrual job triggered",
			"status":  "success",
		})
	})

	// Get active sessions endpoint
	app.Get("/admin/sessions", func(c *fiber.Ctx) error {
		activeSessions := sessionManager.GetActiveSessions()
		sessionInfo := make([]fiber.Map, len(activeSessions))

		for i, session := range activeSessions {
			sessionInfo[i] = fiber.Map{
				"user_id":        session.UserID.Hex(),
				"username":       session.Username,
				"connected_at":   session.ConnectedAt,
				"last_accrual":   session.LastAccrualAt,
				"last_heartbeat": session.LastHeartbeat,
				"is_active":      session.IsActive,
			}
		}

		return c.JSON(fiber.Map{
			"total_sessions": len(activeSessions),
			"sessions":       sessionInfo,
		})
	})

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdown
		log.Println("🛑 Shutdown signal received, stopping services...")
		accrualJob.Stop()
		db.Disconnect()
		log.Println("👋 Services stopped, exiting...")
		os.Exit(0)
	}()

	// Start server
	log.Printf("🌐 Server starting on :%s", cfg.ServerPort)
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
