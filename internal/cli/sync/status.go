package sync

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/utils"
)

// statusCmd handles `mangahub sync status`.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check sync status",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.NewLogger()
		c, cfg, err := getTCPClient()
		if err != nil {
			logger.Error("%v", err)
			return err
		}

		fmt.Println("Checking TCP sync server status...")
		if err := c.CheckStatus(); err != nil {
			fmt.Println("TCP Sync Status:")
			fmt.Println(" Connection: ✗ Inactive")
			logger.Error("TCP sync server %s:%d is inactive: %v", cfg.TCP.Host, cfg.TCP.Port, err)
			return nil
		}

		fmt.Println("TCP Sync Status:")
		fmt.Println(" Connection: ✓ Active")
		fmt.Printf(" Server: %s:%d\n", cfg.TCP.Host, cfg.TCP.Port)
		fmt.Println(" Mode: Progress broadcast (multi-device)")
		return nil
	},
}

func init() {
	SyncCmd.AddCommand(statusCmd)
}
