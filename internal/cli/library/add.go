package library

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
)

// newHTTPClient creates an HTTP client for manga operations
func newHTTPClient() *client.HTTPClient {
	return client.NewHTTPClient(getAPIURL(), "")
}

// Valid status options
var validStatuses = []string{"reading", "completed", "plan-to-read", "on-hold", "dropped"}

func isValidStatus(status string) bool {
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

var addCmd = &cobra.Command{
	Use:   "add --manga-id <id>",
	Short: "Add manga to library",
	Long: `Add a manga to your personal library via the API server with optional status and rating.

Status options: reading, completed, plan-to-read, on-hold, dropped

Examples:
  mangahub library add --manga-id one-piece --status reading
  mangahub library add --manga-id death-note --status completed --rating 9
  mangahub library add -m naruto -s plan-to-read`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		status, _ := cmd.Flags().GetString("status")
		rating, _ := cmd.Flags().GetInt("rating")
		notes, _ := cmd.Flags().GetString("notes")

		if mangaID == "" {
			return fmt.Errorf("--manga-id is required")
		}

		// Validate status
		if !isValidStatus(status) {
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

		// Verify manga exists via API
		mangaClient := newHTTPClient()
		mangaInfo, err := mangaClient.GetManga(mangaID)
		if err != nil {
			return fmt.Errorf("manga '%s' not found. Use 'mangahub manga search' to find valid manga IDs", mangaID)
		}

		// Add to library via API
		fmt.Printf("Adding '%s' to library via API server...\n", mangaInfo.Title)

		err = httpClient.AddToLibrary(mangaID, status, rating, notes)
		if err != nil {
			return fmt.Errorf("failed to add to library: %w", err)
		}

		fmt.Println()
		fmt.Printf("✓ Successfully added '%s' to your library!\n", mangaInfo.Title)
		fmt.Printf("  Status: %s\n", status)
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
	LibraryCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("manga-id", "m", "", "Manga ID (required)")
	addCmd.Flags().StringP("status", "s", "reading", "Initial status (reading, completed, plan-to-read, on-hold, dropped)")
	addCmd.Flags().IntP("rating", "r", 0, "Rating (0-10)")
	addCmd.Flags().StringP("notes", "n", "", "Personal notes about this manga")
	addCmd.MarkFlagRequired("manga-id")
}
