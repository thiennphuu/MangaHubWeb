package server

import (
	"fmt"
	"os/exec"
	"runtime"

	"mangahub/pkg/config"
	"mangahub/pkg/utils"

	"github.com/spf13/cobra"
)

// startCmd now verifies whether components are already running and can
// start missing ones in the background.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start server components",
	Long: `Check and start MangaHub server components. 
This command checks if each component is reachable and starts any missing ones in the background:`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.NewLogger()
		httpOnly, _ := cmd.Flags().GetBool("http-only")
		tcpOnly, _ := cmd.Flags().GetBool("tcp-only")
		udpOnly, _ := cmd.Flags().GetBool("udp-only")
		grpcOnly, _ := cmd.Flags().GetBool("grpc-only")
		wsOnly, _ := cmd.Flags().GetBool("ws-only")
		all, _ := cmd.Flags().GetBool("all")

		cfg, err := config.LoadConfig("config.yaml")
		if err != nil {
			logger.Warn("failed to load config.yaml, using built-in defaults: %v", err)
			cfg = config.DefaultConfig()
		}

		startHTTP := all || httpOnly || (!tcpOnly && !udpOnly && !grpcOnly && !wsOnly)
		startTCP := all || tcpOnly || (!httpOnly && !udpOnly && !grpcOnly && !wsOnly)
		startUDP := all || udpOnly || (!httpOnly && !tcpOnly && !grpcOnly && !wsOnly)
		startGRPC := all || grpcOnly || (!httpOnly && !tcpOnly && !udpOnly && !wsOnly)
		startWS := all || wsOnly || (!httpOnly && !tcpOnly && !udpOnly && !grpcOnly)

		fmt.Println("Checking and Starting MangaHub Server Components...")
		fmt.Println()

		if startHTTP {
			url := fmt.Sprintf("http://%s:%d/health", cfg.HTTP.Host, cfg.HTTP.Port)
			if checkHTTP(url, "HTTP API") {
				fmt.Println("   Already running (no action needed).")
			} else {
				fmt.Println("   Starting HTTP API server...")
				if err := runBackgroundTask("go", "run", "./cmd/api-server"); err != nil {
					logger.Error("failed to start HTTP API: %v", err)
					fmt.Println("   ✗ Failed to start HTTP API")
				} else {
					fmt.Println("   ✓ HTTP API started in background")
				}
			}
			fmt.Println()
		}

		if startTCP {
			addr := fmt.Sprintf("%s:%d", cfg.TCP.Host, cfg.TCP.Port)
			if checkTCP(addr, "TCP Sync") {
				fmt.Println("   Already running (no action needed).")
			} else {
				fmt.Println("   Starting TCP sync server...")
				if err := runBackgroundTask("go", "run", "./cmd/tcp-server"); err != nil {
					logger.Error("failed to start TCP Sync: %v", err)
					fmt.Println("   ✗ Failed to start TCP Sync")
				} else {
					fmt.Println("   ✓ TCP Sync started in background")
				}
			}
			fmt.Println()
		}

		if startUDP {
			addr := fmt.Sprintf("%s:%d", cfg.UDP.Host, cfg.UDP.Port)
			if checkUDP(addr, "UDP Notify") {
				fmt.Println("   Already running (no action needed).")
			} else {
				fmt.Println("   Starting UDP notification server...")
				if err := runBackgroundTask("go", "run", "./cmd/udp-server"); err != nil {
					logger.Error("failed to start UDP Notify: %v", err)
					fmt.Println("   ✗ Failed to start UDP Notify")
				} else {
					fmt.Println("   ✓ UDP Notify started in background")
				}
			}
			fmt.Println()
		}

		if startGRPC {
			addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
			if checkTCP(addr, "gRPC") {
				fmt.Println("   Already running (no action needed).")
			} else {
				fmt.Println("   Starting gRPC server...")
				if err := runBackgroundTask("go", "run", "./cmd/grpc-server"); err != nil {
					logger.Error("failed to start gRPC: %v", err)
					fmt.Println("   ✗ Failed to start gRPC")
				} else {
					fmt.Println("   ✓ gRPC started in background")
				}
			}
			fmt.Println()
		}

		if startWS {
			url := fmt.Sprintf("http://%s:%d/health", cfg.WebSocket.Host, cfg.WebSocket.Port)
			if checkHTTP(url, "WebSocket") {
				fmt.Println("   Already running (no action needed).")
			} else {
				fmt.Println("   Starting WebSocket server...")
				if err := runBackgroundTask("go", "run", "./cmd/websocket-server"); err != nil {
					logger.Error("failed to start WebSocket: %v", err)
					fmt.Println("   ✗ Failed to start WebSocket")
				} else {
					fmt.Println("   ✓ WebSocket started in background")
				}
			}
			fmt.Println()
		}

		fmt.Println("Verification: Run 'mangahub server status' in a few seconds.")
		fmt.Println("Logs: mangahub server logs --follow")
		return nil
	},
}

func runBackgroundTask(name string, args ...string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows, use 'start /b' to run in background without a new window
		fullCmd := append([]string{"/c", "start", "/b"}, name)
		fullCmd = append(fullCmd, args...)
		cmd = exec.Command("cmd", fullCmd...)
	} else {
		cmd = exec.Command(name, args...)
	}

	// Create a log file for this component to avoid crowding stdout
	return cmd.Start()
}

func init() {
	startCmd.Flags().Bool("http-only", false, "Start only HTTP API server")
	startCmd.Flags().Bool("tcp-only", false, "Start only TCP sync server")
	startCmd.Flags().Bool("udp-only", false, "Start only UDP notification server")
	startCmd.Flags().Bool("grpc-only", false, "Start only gRPC internal service")
	startCmd.Flags().Bool("ws-only", false, "Start only WebSocket chat server")
	startCmd.Flags().Bool("all", false, "Start all server components (default)")
}
