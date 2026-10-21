package db

import (
	"fmt"

	"mangahub/pkg/client"
	"mangahub/pkg/session"
	"mangahub/pkg/utils"

	"github.com/spf13/cobra"
)

// optimizeCmd handles `mangahub db optimize`.
// It runs lightweight SQLite optimizations via HTTP API: ANALYZE, REINDEX, and VACUUM.
var optimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "Optimize database performance",
	Long: `Run SQLite maintenance commands (ANALYZE, REINDEX, VACUUM) on the remote database via HTTP API.

Use this periodically after large imports or many updates to keep the database compact and query plans up to date.`,
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

		// Optimize database on server
		fmt.Printf("Optimizing remote database via HTTP API...\n\n")
		optimizeResp, err := httpClient.OptimizeDatabase()
		if err != nil {
			return fmt.Errorf("failed to optimize database: %w", err)
		}

		// Display results
		if len(optimizeResp.Steps) == 0 && len(optimizeResp.Errors) == 0 {
			logger.Warn("no database optimization steps or errors reported (status: %s)", optimizeResp.Status)
			fmt.Printf("⚠️  No steps or errors reported (status: %s)\n", optimizeResp.Status)
		}

		for _, step := range optimizeResp.Steps {
			fmt.Printf("✓ %s\n", step)
		}

		if len(optimizeResp.Errors) > 0 {
			fmt.Println("\nErrors encountered:")
			for _, errMsg := range optimizeResp.Errors {
				logger.Error("database optimization error: %s", errMsg)
				fmt.Printf("✗ %s\n", errMsg)
			}
		}

		if optimizeResp.Status == "success" || optimizeResp.Status == "optimized" {
			fmt.Println("\n✓ Database optimization completed successfully")
			return nil
		} else if optimizeResp.Status == "partial" {
			logger.Warn("database optimization completed with some errors")
			fmt.Println("\n⚠️  Database optimization completed with some errors")
			return nil
		}
		return fmt.Errorf("database optimization failed (status: %s)", optimizeResp.Status)
	},
}

func init() {
	DBCmd.AddCommand(optimizeCmd)
}
