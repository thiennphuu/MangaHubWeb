package progress

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/models"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View reading history",
	Long: `View your manga reading history and progress via the API server.

Examples:
  mangahub progress history
  mangahub progress history --limit 20`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		// Check if user is logged in and get HTTP client
		httpClient, session, err := newAuthenticatedHTTPClient()
		if err != nil {
			fmt.Println("You are not logged in.")
			fmt.Println("\nPlease login first:")
			fmt.Println("  mangahub auth login --username <username>")
			return nil
		}

		fmt.Printf("📖 Reading History for %s\n\n", session.Username)

		// Get library via API
		library, err := httpClient.GetLibrary("", limit, 0)
		if err != nil {
			return fmt.Errorf("failed to get history: %w", err)
		}

		if len(library) == 0 {
			fmt.Println("No reading history found.")
			fmt.Println("\nStart reading with:")
			fmt.Println("  mangahub progress update --manga-id <id> --chapter <n>")
			return nil
		}

		// Print history table
		printHistoryTable(library)

		fmt.Printf("\nShowing %d entries\n", len(library))

		return nil
	},
}

// printHistoryTable prints reading history in a formatted table
func printHistoryTable(library []models.Progress) {
	fmt.Println("┌──────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ %-20s │ %-12s │ %-10s │ %-8s │ %-19s │\n", "MANGA", "STATUS", "CHAPTER", "RATING", "LAST UPDATED")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────────────────┤")

	for _, p := range library {
		mangaName := truncateString(p.MangaID, 20)
		rating := "-"
		if p.Rating > 0 {
			rating = fmt.Sprintf("%d/10", p.Rating)
		}
		updated := p.UpdatedAt.Format("2006-01-02 15:04")
		fmt.Printf("│ %-20s │ %-12s │ %10d │ %-8s │ %-19s │\n",
			mangaName, p.Status, p.CurrentChapter, rating, updated)
	}
	fmt.Println("└──────────────────────────────────────────────────────────────────────────────────────┘")
}

// truncateString truncates a string to max length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	ProgressCmd.AddCommand(historyCmd)
	historyCmd.Flags().IntP("limit", "l", 50, "Maximum entries to show")
}
