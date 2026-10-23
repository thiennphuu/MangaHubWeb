package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mangahub/internal/websocket"
	"mangahub/pkg/config"
	"mangahub/pkg/database"
	"mangahub/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := utils.NewLogger()

	// Load configuration
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Warn("failed to load config: %v, using defaults", err)
		cfg = config.DefaultConfig()
	}
	if err := logger.Configure(cfg.App.Logging.Path, cfg.App.Logging.Level); err != nil {
		logger.Warn("failed to configure file logging: %v", err)
	}

	logger.Info("Starting MangaHub WebSocket Server on %s:%d", cfg.WebSocket.Host, cfg.WebSocket.Port)

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		logger.Error("failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		logger.Error("failed to initialize schema: %v", err)
		os.Exit(1)
	}

	// Create chat hub
	hub := websocket.NewHub(logger)
	go hub.Run()

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// WebSocket endpoint
	engine.GET("/ws/:room", func(c *gin.Context) {
		room := c.Param("room")
		websocket.HandleConnection(c, hub, room)
	})

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.WebSocket.Host, cfg.WebSocket.Port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("WebSocket Server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down WebSocket server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown: %v", err)
	}

	logger.Info("WebSocket Server stopped")
}
