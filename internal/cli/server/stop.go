package server

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"mangahub/pkg/config"
	"mangahub/pkg/utils"

	"github.com/spf13/cobra"
)

// stopCmd now identifies and terminates running MangaHub server components.
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop server components",
	Long:  `Identify and terminate running MangaHub server components by closing processes on their ports.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.NewLogger()
		component, _ := cmd.Flags().GetString("component")
		all, _ := cmd.Flags().GetBool("all")

		cfg, err := config.LoadConfig("config.yaml")
		if err != nil {
			logger.Warn("failed to load config.yaml, using built-in defaults: %v", err)
			cfg = config.DefaultConfig()
		}

		fmt.Println("Stopping MangaHub server components...")
		fmt.Println()

		stopFunc := func(port int, name string) {
			fmt.Printf("   Stopping %s (port %d)...\n", name, port)
			if err := killProcessByPort(port); err != nil {
				logger.Error("could not stop %s: %v", name, err)
				fmt.Printf("   ✗ Could not stop %s\n", name)
			} else {
				fmt.Printf("   ✓ %s stopped\n", name)
			}
			fmt.Println()
		}

		if all || component == "" || component == "http" {
			stopFunc(cfg.HTTP.Port, "HTTP API")
		}

		if all || component == "" || component == "tcp" {
			stopFunc(cfg.TCP.Port, "TCP Sync")
		}

		if all || component == "" || component == "udp" {
			stopFunc(cfg.UDP.Port, "UDP Notify")
		}

		if all || component == "" || component == "grpc" {
			stopFunc(cfg.GRPC.Port, "gRPC")
		}

		if all || component == "" || component == "ws" {
			stopFunc(cfg.WebSocket.Port, "WebSocket")
		}

		return nil
	},
}

func killProcessByPort(port int) error {
	if runtime.GOOS == "windows" {
		// Find PID using netstat
		out, err := exec.Command("cmd", "/c", fmt.Sprintf("netstat -ano | findstr :%d", port)).Output()
		if err != nil {
			return fmt.Errorf("process not found or port not in use")
		}

		lines := strings.Split(string(out), "\n")
		pids := make(map[string]bool)
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 5 && strings.Contains(fields[1], fmt.Sprintf(":%d", port)) {
				// The PID is the last column (or second to last if there's no state)
				pid := fields[len(fields)-1]
				if pid != "0" {
					pids[pid] = true
				}
			}
		}

		if len(pids) == 0 {
			return fmt.Errorf("no active PID found for port %d", port)
		}

		for pid := range pids {
			// Kill the process and its children (/T) forcefully (/F)
			if err := exec.Command("taskkill", "/F", "/T", "/PID", pid).Run(); err != nil {
				return fmt.Errorf("failed to kill PID %s: %w", pid, err)
			}
		}
		return nil
	}

	// Unix-like systems (lsof/fuser)
	out, err := exec.Command("sh", "-c", fmt.Sprintf("lsof -t -i:%d", port)).Output()
	if err != nil {
		return fmt.Errorf("process not found")
	}
	pids := strings.Fields(string(out))
	for _, pid := range pids {
		if err := exec.Command("kill", "-9", pid).Run(); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	stopCmd.Flags().StringP("component", "c", "", "Specific component to stop (http, tcp, udp, grpc, ws)")
	stopCmd.Flags().Bool("all", false, "Stop all server components (default)")
}
