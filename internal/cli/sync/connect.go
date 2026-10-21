package sync

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
	"mangahub/pkg/config"
	"mangahub/pkg/utils"
)

// getTCPClient returns a TCP client configured for the sync server.
func getTCPClient() (*client.TCPClient, *config.Config, error) {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config.yaml: %w", err)
	}

	return client.NewTCPClient(cfg.TCP.Host, cfg.TCP.Port), cfg, nil
}

// connectCmd handles `mangahub sync connect`.
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to sync server",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.NewLogger()
		c, cfg, err := getTCPClient()
		if err != nil {
			logger.Error("%v", err)
			return err
		}

		fmt.Printf("Connecting to TCP sync server at %s:%d...\n", cfg.TCP.Host, cfg.TCP.Port)

		conn, err := c.Connect()
		if err != nil {
			logger.Error("failed to connect to sync server at %s:%d: %v", cfg.TCP.Host, cfg.TCP.Port, err)
			return fmt.Errorf("failed to connect to sync server: %w", err)
		}
		defer conn.Close()

		now := time.Now().UTC().Format("2006-01-02 15:04:05 MST")

		fmt.Println("✓ Connected successfully!")
		fmt.Println("\nConnection Details:")
		fmt.Printf(" Server: %s:%d\n", cfg.TCP.Host, cfg.TCP.Port)
		fmt.Println(" Connection: TCP")
		fmt.Printf(" Connected at: %s\n", now)
		return nil
	},
}

func init() {
	SyncCmd.AddCommand(connectCmd)
}
