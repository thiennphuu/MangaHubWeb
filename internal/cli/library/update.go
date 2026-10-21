package library

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update --manga-id <id>",
	Short: "Update library entry",
	Long: `Update a manga's status, rating, or other library information via the API server.

Status options: reading, completed, plan-to-read, on-hold, dropped

Example:
  mangahub library update --manga-id one-piece --status completed --rating 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		status, _ := cmd.Flags().GetString("status")
		rating, _ := cmd.Flags().GetInt("rating")
		notes, _ := cmd.Flags().GetString("notes")

		if mangaID == "" {
			return fmt.Errorf("--manga-id is required")
		}

		// Validate status if provided
		if status != "" && !isValidStatus(status) {
			return fmt.Errorf("invalid status '%s'. Valid options: %s", status, strings.Join(validStatuses, ", "))
		}

		// Validate rating
		if rating < 0 || rating > 10 {
			return fmt.Errorf("rating must be between 0 and 10")
		}

		// Check if user is logged in and get HTTP client
		httpClient, _, err := newAuthenticatedHTTPClient()
		if err != nil {
			fmt.Println("You are not logged in.")
			fmt.Println("\nPlease login first:")
			fmt.Println("  mangahub auth login --username <username>")
			return nil
		}

		fmt.Printf("Updating %s via API server...\n", mangaID)

		// Update via API (using UpdateProgress which handles library updates)
		err = httpClient.UpdateProgress(mangaID, 0, status, rating, notes)
		if err != nil {
			return fmt.Errorf("failed to update library entry: %w", err)
		}

		fmt.Println("✓ Successfully updated")
		if status != "" {
			fmt.Printf("  Status: %s\n", status)
		}
		if rating > 0 {
			fmt.Printf("  Rating: %d/10\n", rating)
		}
		if notes != "" {
			fmt.Printf("  Notes: %s\n", notes)
		}

		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("manga-id", "m", "", "Manga ID (required)")
	updateCmd.Flags().StringP("status", "s", "", "New status (reading, completed, plan-to-read, on-hold, dropped)")
	updateCmd.Flags().IntP("rating", "r", 0, "Rating (0-10)")
	updateCmd.Flags().StringP("notes", "n", "", "Personal notes")
	updateCmd.MarkFlagRequired("manga-id")
}
