package sync

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/utils"
)

// disconnectCmd handles `mangahub sync disconnect`.
var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from sync server",
	Long: `Disconnect from the TCP sync server.

Note: Each CLI command opens a short-lived connection, so this command
currently acts as a helper to verify that the server is reachable and
then closes the connection immediately.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.NewLogger()
		c, _, err := getTCPClient()
		if err != nil {
			logger.Error("%v", err)
			return err
		}

		// We simply open a connection and close it to simulate an explicit disconnect.
		conn, err := c.Connect()
		if err != nil {
			logger.Error("failed to connect to sync server for disconnect: %v", err)
			return fmt.Errorf("failed to connect to sync server for disconnect: %w", err)
		}
		if err := conn.Close(); err != nil {
			logger.Error("failed to close sync server connection: %v", err)
			return fmt.Errorf("failed to close sync server connection: %w", err)
		}

		fmt.Println("✓ Disconnected from sync server")
		return nil
	},
}

func init() {
	SyncCmd.AddCommand(disconnectCmd)
}
