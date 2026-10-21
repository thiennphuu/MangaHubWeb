package manga

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/models"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all manga",
	Long: `List manga with optional filtering and pagination via the API server.

Examples:
  mangahub manga list
  mangahub manga list --page 2 --limit 20
  mangahub manga list --genre shounen`,
	RunE: func(cmd *cobra.Command, args []string) error {
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")
		genre, _ := cmd.Flags().GetString("genre")
		status, _ := cmd.Flags().GetString("status")

		// Get HTTP client
		httpClient := getHTTPClient()

		// Calculate offset
		offset := (page - 1) * limit

		// List via API
		mangaList, err := httpClient.ListManga(limit, offset, status, genre)
		if err != nil {
			return fmt.Errorf("failed to list manga: %w", err)
		}

		fmt.Printf("Listing manga (page %d, limit %d)\n", page, limit)
		if genre != "" {
			fmt.Printf("Filter: genre=%s\n", genre)
		}
		if status != "" {
			fmt.Printf("Filter: status=%s\n", status)
		}
		fmt.Println()

		if len(mangaList) == 0 {
			fmt.Println("No manga found.")
			return nil
		}

		printMangaListTable(mangaList)
		fmt.Printf("\nShowing %d manga (page %d)\n", len(mangaList), page)
		fmt.Println("Use --page <n> to see more results")

		return nil
	},
}

func init() {
	MangaCmd.AddCommand(listCmd)
	listCmd.Flags().IntP("page", "p", 1, "Page number")
	listCmd.Flags().IntP("limit", "l", 20, "Results per page")
	listCmd.Flags().StringP("genre", "g", "", "Filter by genre")
	listCmd.Flags().String("status", "", "Filter by status (ongoing, completed)")
}

// printMangaListTable prints manga list in a formatted table
func printMangaListTable(mangaList []models.Manga) {
	fmt.Println("┌────────────────────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ %-6s │ %-35s │ %-25s │ %-10s │ %-8s │\n", "ID", "TITLE", "AUTHOR", "STATUS", "CHAPTERS")
	fmt.Println("├────────────────────────────────────────────────────────────────────────────────────────────────────────┤")

	for _, m := range mangaList {
		title := truncateString(m.Title, 35)
		author := truncateString(m.Author, 25)
		id := truncateString(m.ID, 6)
		fmt.Printf("│ %-6s │ %-35s │ %-25s │ %-10s │ %8d │\n",
			id, title, author, m.Status, m.TotalChapters)
	}
	fmt.Println("└────────────────────────────────────────────────────────────────────────────────────────────────────────┘")
}
