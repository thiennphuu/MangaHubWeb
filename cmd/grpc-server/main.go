package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"mangahub/internal/grpc/service"
	"mangahub/pkg/config"
	"mangahub/pkg/database"
	"mangahub/pkg/utils"
	pb "mangahub/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(JSONCodec{})
}

// JSONCodec implements grpc encoding.Codec using JSON
type JSONCodec struct{}

func (JSONCodec) Name() string { return "json" }

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

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

	logger.Info(fmt.Sprintf("Starting MangaHub gRPC Server on %s:%d", cfg.GRPC.Host, cfg.GRPC.Port))

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

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to listen: %v", err))
		os.Exit(1)
	}

	// Create gRPC server with logging interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			resp, err := handler(ctx, req)
			if err != nil {
				logger.Error("Method: %s, Error: %v", info.FullMethod, err)
			} else {
				respJSON, _ := json.Marshal(resp)
				logger.Info("Method: %s, Response: %s", info.FullMethod, string(respJSON))
			}
			return resp, err
		}),
	)

	// Register services
	mangaService := service.NewMangaService(db, logger)
	pb.RegisterMangaServiceServer(grpcServer, mangaService)

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("gRPC Server listening on %s", lis.Addr()))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(fmt.Sprintf("gRPC server error: %v", err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	logger.Info("gRPC Server stopped")
}
