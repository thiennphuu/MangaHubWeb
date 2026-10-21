package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
	"mangahub/pkg/config"
	"mangahub/pkg/utils"
)

// getAPIURL returns the API server URL
func getAPIURL() string {
	if url := os.Getenv("MANGAHUB_API_URL"); url != "" {
		return url
	}
	return "http://10.238.53.72:8080"
}

// showCmd handles `mangahub config show`.
var showCmd = &cobra.Command{
	Use:   "show [section]",
	Short: "Show configuration by fetching server info via HTTP API",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.NewLogger()
		// Fetch server health/config info via HTTP API
		fmt.Println("Fetching server configuration via HTTP API...")
		httpClient := client.NewHTTPClient(getAPIURL(), "")

		// Try to get server health endpoint
		resp, err := httpClient.GetServerHealth()
		if err != nil {
			logger.Warn("could not connect to API server: %v", err)
			fmt.Println("Showing local configuration instead...\n")
			if len(args) == 0 {
				printFullConfig()
				return nil
			}
			return showSection(args[0])
		}

		fmt.Printf("✓ Connected to server: %s\n\n", getAPIURL())

		if len(args) == 0 {
			printFullConfigFromServer(resp)
			return nil
		}

		return showSection(args[0])
	},
}

func showSection(section string) error {
	// Load config from file
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		utils.NewLogger().Warn("could not load config.yaml, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}

	switch section {
	case "server", "servers":
		fmt.Println("Server Configuration")
		fmt.Printf(" API Server: http://%s:%d\n", cfg.HTTP.Host, cfg.HTTP.Port)
		fmt.Printf(" TCP Sync: %s:%d\n", cfg.TCP.Host, cfg.TCP.Port)
		fmt.Printf(" UDP Notify: %s:%d\n", cfg.UDP.Host, cfg.UDP.Port)
		fmt.Printf(" WebSocket Chat: ws://%s:%d\n", cfg.WebSocket.Host, cfg.WebSocket.Port)
		fmt.Printf(" gRPC Service: %s:%d\n", cfg.GRPC.Host, cfg.GRPC.Port)
	case "database", "db":
		fmt.Println("Database Configuration")
		fmt.Printf(" Type: %s\n", cfg.Database.Type)
		fmt.Printf(" Path: %s\n", cfg.Database.Path)
		fmt.Printf(" Max Connections: %d\n", cfg.Database.MaxConn)
		fmt.Printf(" Timeout: %d seconds\n", cfg.Database.Timeout)
	case "profile":
		fmt.Println("Profile Configuration")
		fmt.Println(" Active profile: default")
	default:
		fmt.Printf("Unknown section: %s\n", section)
		fmt.Println("Available sections: server, database, profile")
	}
	return nil
}

func init() {
	ConfigCmd.AddCommand(showCmd)
}

func printFullConfig() {
	// Load config from file
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		utils.NewLogger().Warn("could not load config.yaml, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}

	fmt.Println("MangaHub Configuration (Local)")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("API Server: http://%s:%d\n", cfg.HTTP.Host, cfg.HTTP.Port)
	fmt.Printf("TCP Sync: %s:%d\n", cfg.TCP.Host, cfg.TCP.Port)
	fmt.Printf("UDP Notify: %s:%d\n", cfg.UDP.Host, cfg.UDP.Port)
	fmt.Printf("WebSocket Chat: ws://%s:%d\n", cfg.WebSocket.Host, cfg.WebSocket.Port)
	fmt.Printf("gRPC Service: %s:%d\n", cfg.GRPC.Host, cfg.GRPC.Port)
	fmt.Println("\nDatabase:")
	fmt.Printf(" Type: %s\n", cfg.Database.Type)
	fmt.Printf(" Path: %s\n", cfg.Database.Path)
	fmt.Printf(" Max Connections: %d\n", cfg.Database.MaxConn)
	fmt.Println("\nProfile: default")
}

func printFullConfigFromServer(healthData map[string]interface{}) {
	// Load config from file for endpoint info
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		utils.NewLogger().Warn("could not load config.yaml, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}

	fmt.Println("MangaHub Configuration (From Server)")
	fmt.Println("═══════════════════════════════════════")

	// Display server status from health check
	if status, ok := healthData["status"].(string); ok {
		fmt.Printf("Server Status: %s\n", status)
	}

	if version, ok := healthData["version"].(string); ok {
		fmt.Printf("Server Version: %s\n", version)
	}

	fmt.Println("\nEndpoints:")
	fmt.Printf(" API Server: http://%s:%d\n", cfg.HTTP.Host, cfg.HTTP.Port)
	fmt.Printf(" TCP Sync: %s:%d\n", cfg.TCP.Host, cfg.TCP.Port)
	fmt.Printf(" UDP Notify: %s:%d\n", cfg.UDP.Host, cfg.UDP.Port)
	fmt.Printf(" WebSocket Chat: ws://%s:%d\n", cfg.WebSocket.Host, cfg.WebSocket.Port)
	fmt.Printf(" gRPC Service: %s:%d\n", cfg.GRPC.Host, cfg.GRPC.Port)

	fmt.Println("\nProtocols:")
	fmt.Printf(" ✓ HTTP REST API (%d)\n", cfg.HTTP.Port)
	fmt.Printf(" ✓ TCP Sync (%d)\n", cfg.TCP.Port)
	fmt.Printf(" ✓ UDP Notifications (%d)\n", cfg.UDP.Port)
	fmt.Printf(" ✓ gRPC (%d)\n", cfg.GRPC.Port)
	fmt.Printf(" ✓ WebSocket (%d)\n", cfg.WebSocket.Port)
}
