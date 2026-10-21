package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mangahub/internal/api"
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
		logger.Warn(fmt.Sprintf("failed to load config: %v, using defaults", err))
		cfg = config.DefaultConfig()
	}

	logger.Info(fmt.Sprintf("Starting MangaHub API Server on %s:%d", cfg.HTTP.Host, cfg.HTTP.Port))

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize database: %v", err))
		os.Exit(1)
	}
	defer db.Close()

	// Initialize schema
	if err := db.Init(); err != nil {
		logger.Error(fmt.Sprintf("failed to initialize schema: %v", err))
		os.Exit(1)
	}

	logger.Info("Database initialized successfully")

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// Add CORS middleware
	engine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize API handler and register routes
	handler := api.NewHandler(db, logger)
	handler.RegisterRoutes(engine)

	// Health check endpoint with server configuration
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"http": gin.H{
				"host": cfg.HTTP.Host,
				"port": cfg.HTTP.Port,
			},
			"tcp": gin.H{
				"host": cfg.TCP.Host,
				"port": cfg.TCP.Port,
			},
			"udp": gin.H{
				"host": cfg.UDP.Host,
				"port": cfg.UDP.Port,
			},
			"grpc": gin.H{
				"host": cfg.GRPC.Host,
				"port": cfg.GRPC.Port,
			},
			"websocket": gin.H{
				"host": cfg.WebSocket.Host,
				"port": cfg.WebSocket.Port,
			},
		})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("API Server listening on %s", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("server error: %v", err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.HTTP.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("forced shutdown: %v", err))
	}

	logger.Info("Server stopped")
}
