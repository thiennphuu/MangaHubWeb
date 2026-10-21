package server

import (
	"fmt"

	"mangahub/pkg/config"
	"mangahub/pkg/utils"

	"github.com/spf13/cobra"
)

// statusCmd summarizes real reachability checks into a compact table.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server status",
	Long:  `Display current status of all MangaHub server components using health checks and port reachability.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.NewLogger()
		cfg, err := config.LoadConfig("config.yaml")
		if err != nil {
			logger.Warn("failed to load config.yaml, using built-in defaults: %v", err)
			cfg = config.DefaultConfig()
		}

		fmt.Println("MangaHub Server Status")
		fmt.Println("──────────────────────")

		httpOK := checkHTTP(fmt.Sprintf("http://%s:%d/health", cfg.HTTP.Host, cfg.HTTP.Port), "HTTP API")
		wsOK := checkHTTP(fmt.Sprintf("http://%s:%d/health", cfg.WebSocket.Host, cfg.WebSocket.Port), "WebSocket")
		tcpOK := checkTCP(fmt.Sprintf("%s:%d", cfg.TCP.Host, cfg.TCP.Port), "TCP Sync")
		udpOK := checkUDP(fmt.Sprintf("%s:%d", cfg.UDP.Host, cfg.UDP.Port), "UDP Notify")
		grpcOK := checkTCP(fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port), "gRPC")

		fmt.Println()
		fmt.Println("Summary:")
		fmt.Printf(" HTTP API:      %s\n", boolToStatus(httpOK))
		fmt.Printf(" WebSocket:     %s\n", boolToStatus(wsOK))
		fmt.Printf(" TCP Sync:      %s\n", boolToStatus(tcpOK))
		fmt.Printf(" UDP Notify:    %s\n", boolToStatus(udpOK))
		fmt.Printf(" gRPC:          %s\n", boolToStatus(grpcOK))

		overall := httpOK && wsOK && tcpOK && udpOK && grpcOK
		fmt.Println()
		if overall {
			fmt.Println("Overall System Health: ✓ Healthy")
		} else {
			fmt.Println("Overall System Health: ✗ Degraded")
			fmt.Println("Run 'mangahub server health' for detailed diagnostics.")
		}

		return nil
	},
}

func init() {
	// Deprecated: status is now based on real checks, so no synthetic degraded mode.
}

func boolToStatus(ok bool) string {
	if ok {
		return "✓ Online"
	}
	return "✗ Offline"
}
