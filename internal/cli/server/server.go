package server

import (
	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage MangaHub server components",
	Long: `Manage MangaHub server components including:
- HTTP API server
- TCP sync server
- UDP notification server
- gRPC internal service
- WebSocket chat server`,
}

func init() {
	ServerCmd.AddCommand(startCmd)
	ServerCmd.AddCommand(stopCmd)
	ServerCmd.AddCommand(statusCmd)
	ServerCmd.AddCommand(healthCmd)
	ServerCmd.AddCommand(logsCmd)
}
