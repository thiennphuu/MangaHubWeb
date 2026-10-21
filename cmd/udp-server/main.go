package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mangahub/internal/udp"
	"mangahub/pkg/config"
	"mangahub/pkg/database"
	"mangahub/pkg/utils"
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

	logger.Info(fmt.Sprintf("Starting MangaHub UDP Server on %s:%d", cfg.UDP.Host, cfg.UDP.Port))

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize database: %v", err))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		logger.Error(fmt.Sprintf("failed to initialize schema: %v", err))
		os.Exit(1)
	}

	// Create and start UDP server
	server := udp.NewServer(fmt.Sprintf("%s:%d", cfg.UDP.Host, cfg.UDP.Port), logger)

	go func() {
		logger.Info("UDP Server starting...")
		if err := server.Start(); err != nil {
			logger.Error(fmt.Sprintf("UDP server error: %v", err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down UDP server...")
	server.Stop()
	logger.Info("UDP Server stopped")
}
