package manga

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mangahub/pkg/models"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for manga",
	Long: `Search for manga by title, author, or other criteria via the API server.

Examples:
  mangahub manga search "attack on titan"
  mangahub manga search "romance" --genre romance --status completed
  mangahub manga search "naruto" --limit 5`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		genre, _ := cmd.Flags().GetString("genre")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")

		fmt.Printf("Searching for \"%s\" via API server...\n\n", query)

		// Get HTTP client
		httpClient := getHTTPClient()

		// Build filter
		filter := &models.MangaFilter{
			Query: query,
			Limit: limit,
		}
		if genre != "" {
			filter.Genres = []string{genre}
		}
		if status != "" {
			filter.Status = status
		}

		// Search via API
		results, err := httpClient.SearchManga(filter)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results.Manga) == 0 {
			fmt.Println("No manga found matching your search.")
			return nil
		}

		fmt.Printf("Found %d results:\n\n", len(results.Manga))
		printMangaResults(results.Manga)
		fmt.Println("\nUse 'mangahub manga info <id>' to view details")
		fmt.Println("Use 'mangahub library add --manga-id <id>' to add to your library")

		return nil
	},
}

func init() {
	MangaCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("genre", "g", "", "Filter by genre")
	searchCmd.Flags().StringP("status", "s", "", "Filter by status (ongoing, completed)")
	searchCmd.Flags().IntP("limit", "l", 10, "Maximum results to show")
}

// printMangaResults prints manga in a formatted table
func printMangaResults(mangaList []models.Manga) {
	fmt.Println("┌──────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ %-4s │ %-30s │ %-20s │ %-10s │ %-8s │\n", "ID", "TITLE", "AUTHOR", "STATUS", "CHAPTERS")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────────────────────┤")

	for _, m := range mangaList {
		title := truncateString(m.Title, 30)
		author := truncateString(m.Author, 20)
		id := truncateString(m.ID, 4)
		fmt.Printf("│ %-4s │ %-30s │ %-20s │ %-10s │ %8d │\n",
			id, title, author, m.Status, m.TotalChapters)
	}
	fmt.Println("└──────────────────────────────────────────────────────────────────────────────────────────┘")
}

// truncateString truncates a string to max length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
