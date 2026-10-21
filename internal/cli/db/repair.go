package db

import (
	"fmt"

	"mangahub/pkg/client"
	"mangahub/pkg/session"
	"mangahub/pkg/utils"

	"github.com/spf13/cobra"
)

// repairCmd handles `mangahub db repair`.
var repairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Attempt to repair the database",
	Long: `Run lightweight SQLite maintenance commands (VACUUM, integrity check) via HTTP API
and re-run schema initialization to ensure tables and indexes exist.`,
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

		// Repair database on server
		fmt.Printf("Repairing remote database via HTTP API...\n\n")
		repairResp, err := httpClient.RepairDatabase()
		if err != nil {
			return fmt.Errorf("failed to repair database: %w", err)
		}

		// Display results
		if len(repairResp.Steps) == 0 && len(repairResp.Errors) == 0 {
			logger.Warn("no database repair steps or errors reported (status: %s)", repairResp.Status)
			fmt.Printf("⚠️  No steps or errors reported (status: %s)\n", repairResp.Status)
		}

		for _, step := range repairResp.Steps {
			fmt.Printf("✓ %s\n", step)
		}

		if len(repairResp.Errors) > 0 {
			fmt.Println("\nIssues encountered:")
			for _, errMsg := range repairResp.Errors {
				logger.Error("database repair issue: %s", errMsg)
				fmt.Printf("⚠️  %s\n", errMsg)
			}
		}

		if repairResp.Status == "success" || repairResp.Status == "repaired" {
			fmt.Println("\n✓ Database repair/maintenance completed")
			return nil
		} else if repairResp.Status == "partial" {
			logger.Warn("database repair completed with some issues")
			fmt.Println("\n⚠️  Database repair completed with some issues")
			return nil
		}
		return fmt.Errorf("database repair failed (status: %s)", repairResp.Status)
	},
}

func init() {
	DBCmd.AddCommand(repairCmd)
}
