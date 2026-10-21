package db

import (
	"fmt"
	"os"

	"mangahub/pkg/client"
	"mangahub/pkg/session"

	"github.com/spf13/cobra"
)

const defaultDBPath = "./data/mangahub.db"

// checkCmd handles `mangahub db check`.
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check database integrity",
	Long:  `Check the remote database integrity via HTTP API and verify core tables exist.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get session for authentication
		sess, err := session.Load()
		if err != nil {
			return fmt.Errorf("not logged in: %w", err)
		}

		// Create HTTP client
		apiURL := getAPIURL()
		httpClient := client.NewHTTPClient(apiURL, sess.Token)

		// Fetch database check results from server
		fmt.Printf("Checking remote database via HTTP API...\n")
		checkResp, err := httpClient.GetDatabaseCheck()
		if err != nil {
			return fmt.Errorf("failed to check database: %w", err)
		}

		// Display results
		fmt.Printf("\n✓ Database Status: %s\n\n", checkResp.Status)

		// Display integrity check results
		fmt.Println("Integrity Check:")
		if checkResp.Integrity.OK {
			fmt.Println("✓ PRAGMA integrity_check: ok")
		} else {
			fmt.Println("✗ Integrity issues found:")
			for _, issue := range checkResp.Integrity.Issues {
				fmt.Printf("  - %s\n", issue)
			}
		}

		// Display table verification results
		fmt.Println("\nTable Verification:")
		for _, table := range checkResp.Tables.Verified {
			fmt.Printf("✓ Table %s exists\n", table)
		}
		for _, table := range checkResp.Tables.Missing {
			fmt.Printf("✗ Missing table: %s\n", table)
		}

		if checkResp.Status == "healthy" {
			fmt.Println("\n✓ Database check completed successfully")
			return nil
		}
		return fmt.Errorf("database check found issues")
	},
}

func init() {
	DBCmd.AddCommand(checkCmd)
}

// getAPIURL returns the API URL from environment or default
func getAPIURL() string {
	if url := os.Getenv("MANGAHUB_API_URL"); url != "" {
		return url
	}
	return "http://10.238.53.72:8080"
}
