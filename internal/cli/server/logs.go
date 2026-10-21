package server

import (
	"fmt"
	"os"

	"mangahub/pkg/client"
	"mangahub/pkg/session"

	"github.com/spf13/cobra"
)

// logsCmd fetches server logs via HTTP API
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View server logs",
	Long:  `View server logs from the remote server via HTTP API with optional filtering.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		maxLines, _ := cmd.Flags().GetInt("max-lines")
		level, _ := cmd.Flags().GetString("level")

		// Get session for authentication
		sess, err := session.Load()
		if err != nil {
			return fmt.Errorf("not logged in: %w", err)
		}

		// Create HTTP client
		apiURL := getAPIURL()
		httpClient := client.NewHTTPClient(apiURL, sess.Token)

		// Fetch logs from server
		fmt.Printf("Fetching server logs via HTTP API...\n")
		logsResp, err := httpClient.GetServerLogs(maxLines, level)
		if err != nil {
			return fmt.Errorf("failed to fetch logs from server: %w", err)
		}

		// Display logs
		fmt.Printf("✓ Retrieved %d log entries from server\n\n", logsResp.Count)
		
		if logsResp.Count == 0 {
			fmt.Println("No logs available on the server.")
			return nil
		}

		fmt.Printf("Recent server logs (max: %d, level: %s):\n\n", logsResp.MaxLines, 
			func() string {
				if logsResp.Level == "" {
					return "all"
				}
				return logsResp.Level
			}())

		for _, line := range logsResp.Logs {
			fmt.Println(line)
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().IntP("max-lines", "n", 100, "Maximum number of log lines to retrieve")
	logsCmd.Flags().StringP("level", "l", "", "Filter by log level (debug, info, warn, error)")
}

// getAPIURL returns the API URL from environment or default
func getAPIURL() string {
	if url := os.Getenv("MANGAHUB_API_URL"); url != "" {
		return url
	}
	return "http://10.238.53.72:8080"
}
