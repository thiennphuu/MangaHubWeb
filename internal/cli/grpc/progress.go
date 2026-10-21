package grpc

import (
	"fmt"

	"mangahub/pkg/client"

	"github.com/spf13/cobra"
)

// progressCmd is the progress subcommand under grpc
var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Progress operations via gRPC",
	Long:  `Update and query reading progress using gRPC protocol.`,
}

// updateCmd is the update subcommand under progress
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reading progress",
	Long: `Update reading progress for a manga via gRPC server.

Example:
  mangahub grpc progress update --manga-id one-piece --chapter 1095`,
	RunE: runUpdateProgress,
}

func runUpdateProgress(cmd *cobra.Command, args []string) error {
	mangaID, _ := cmd.Flags().GetString("manga-id")
	chapter, _ := cmd.Flags().GetInt("chapter")
	serverAddr, _ := cmd.Flags().GetString("server")
	userID, _ := cmd.Flags().GetString("user-id")

	if mangaID == "" {
		return fmt.Errorf("manga ID is required. Use --manga-id or -m flag")
	}
	if chapter == 0 {
		return fmt.Errorf("chapter number is required. Use --chapter or -c flag")
	}

	fmt.Printf("Connecting to gRPC server at %s...\n", serverAddr)

	// Create gRPC client and connect
	grpcClient := client.NewGRPCClient(serverAddr)
	if err := grpcClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer grpcClient.Close()

	fmt.Printf("Updating progress for: %s\n\n", mangaID)

	// Call gRPC server
	resp, err := grpcClient.UpdateProgress(userID, mangaID, chapter)
	if err != nil {
		return fmt.Errorf("gRPC error: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update progress: %s", resp.Message)
	}

	fmt.Println("✓ Progress updated via gRPC")
	fmt.Println()
	fmt.Printf("  User:    %s\n", userID)
	fmt.Printf("  Manga:   %s\n", mangaID)
	fmt.Printf("  Chapter: %d\n", chapter)
	fmt.Printf("  Message: %s\n", resp.Message)

	return nil
}

func init() {
	GRPCCmd.AddCommand(progressCmd)
	progressCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("manga-id", "m", "", "Manga ID (required)")
	updateCmd.Flags().IntP("chapter", "c", 0, "Chapter number (required)")
	updateCmd.Flags().StringP("user-id", "u", "default-user", "User ID")
	updateCmd.Flags().StringP("server", "s", "10.238.53.72:9092", "gRPC server address")
}
