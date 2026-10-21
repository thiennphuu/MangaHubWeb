package library

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/models"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "View your library",
	Long: `Display your manga library with filtering and sorting options via the API server.

Examples:
  mangahub library list
  mangahub library list --status reading
  mangahub library list --status completed
  mangahub library list --sort-by title
  mangahub library list --sort-by last-updated --order desc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		order, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")

		// Check if user is logged in
		httpClient, session, err := newAuthenticatedHTTPClient()
		if err != nil {
			fmt.Println("You are not logged in.")
			fmt.Println("\nPlease login first:")
			fmt.Println("  mangahub auth login --username <username>")
			return nil
		}

		// Get library from API
		library, err := httpClient.GetLibrary(status, limit, 0)
		if err != nil {
			return fmt.Errorf("failed to get library: %w", err)
		}

		// Display header
		fmt.Printf("📚 %s's Manga Library\n", session.Username)
		if status != "" {
			fmt.Printf("Filter: status=%s\n", status)
		}
		fmt.Printf("Sort: %s (%s)\n\n", sortBy, order)

		if len(library) == 0 {
			fmt.Println("Your library is empty.")
			fmt.Println("\nAdd manga to your library:")
			fmt.Println("  mangahub library add --manga-id <id>")
			return nil
		}

		// Print library table
		printLibraryTable(library)
		fmt.Printf("\nTotal: %d manga in library\n", len(library))

		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("status", "s", "", "Filter by status (reading, completed, plan-to-read, on-hold, dropped)")
	listCmd.Flags().String("sort-by", "updated", "Sort by field (title, status, updated)")
	listCmd.Flags().String("order", "desc", "Sort order (asc, desc)")
	listCmd.Flags().IntP("limit", "l", 50, "Maximum entries to show")
}

// printLibraryTable prints the library in a formatted table
func printLibraryTable(library []models.Progress) {
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
