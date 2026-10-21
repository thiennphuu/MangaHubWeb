package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mangahub/internal/tcp"
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
		logger.Warn("failed to load config: %v, using defaults", err)
		cfg = config.DefaultConfig()
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("       ⛩️  MangaHub TCP Sync Server ⛩️             ")
	fmt.Println("--------------------------------------------------")
	logger.Info("Starting MangaHub TCP Server on %s:%d", cfg.TCP.Host, cfg.TCP.Port)

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

	// Create and start TCP server
	//server := tcp.NewServer(fmt.Sprintf("%s:%d", cfg.TCP.Host, cfg.TCP.Port), logger, db)
	server := tcp.NewServer(fmt.Sprintf("%d", cfg.TCP.Port), logger, db)
	go func() {
		logger.Info("TCP Server starting...")
		if err := server.Start(); err != nil {
			logger.Error("TCP server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down TCP server...")
	server.Stop()
	logger.Info("TCP Server stopped")
}
