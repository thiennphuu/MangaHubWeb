package db

import (
	"fmt"

	"mangahub/pkg/client"
	"mangahub/pkg/session"
	"mangahub/pkg/utils"

	"github.com/spf13/cobra"
)

// statsCmd handles `mangahub db stats`.
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database statistics",
	Long:  `Display basic statistics for the remote database via HTTP API such as size and row counts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.NewLogger()
		// Get session for authentication
		sess, err := session.Load()
		if err != nil {
			return fmt.Errorf("not logged in: %w", err)
		}

		// Create HTTP client
		apiURL := getAPIURL()
		httpClient := client.NewHTTPClient(apiURL, sess.Token)

		// Fetch database stats from server
		fmt.Printf("Fetching database statistics via HTTP API...\n\n")
		statsResp, err := httpClient.GetDatabaseStats()
		if err != nil {
			return fmt.Errorf("failed to get database stats: %w", err)
		}

		// Display file size
		fmt.Printf("File size: %.2f MB\n", statsResp.FileSizeMB)

		// Display table counts
		fmt.Println("\nRow counts:")
		tableOrder := []string{"users", "manga", "user_progress", "chat_messages", "notifications"}
		for _, name := range tableOrder {
			if count, exists := statsResp.Tables[name]; exists {
				if count >= 0 {
					fmt.Printf("  %s: %d rows\n", name, count)
				} else {
					logger.Error("failed to read row count for table %s", name)
					fmt.Printf("  %s: error\n", name)
				}
			}
		}

		return nil
	},
}

func init() {
	DBCmd.AddCommand(statsCmd)
}
