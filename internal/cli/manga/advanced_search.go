package manga

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mangahub/pkg/models"
)

var advancedSearchCmd = &cobra.Command{
	Use:   "advanced-search",
	Short: "Advanced manga search with filters",
	Long: `Perform advanced search with multiple filter options via the API server.

Examples:
  mangahub manga advanced-search "keyword" --genre "action,adventure" --status "ongoing" --author "author name" --year-from 2020 --year-to 2024 --min-chapters 50 --sort-by "popularity" --order "desc"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) > 0 {
			query = strings.Join(args, " ")
		}

		genres, _ := cmd.Flags().GetString("genre")
		status, _ := cmd.Flags().GetString("status")
		author, _ := cmd.Flags().GetString("author")
		minChapters, _ := cmd.Flags().GetInt("min-chapters")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		order, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")

		fmt.Println("Advanced Search Parameters:")
		if query != "" {
			fmt.Printf("  Query: %s\n", query)
		}
		if genres != "" {
			fmt.Printf("  Genres: %s\n", genres)
		}
		if status != "" {
			fmt.Printf("  Status: %s\n", status)
		}
		if author != "" {
			fmt.Printf("  Author: %s\n", author)
		}
		if minChapters > 0 {
			fmt.Printf("  Minimum Chapters: %d\n", minChapters)
		}
		fmt.Printf("  Sort: %s (%s)\n", sortBy, order)
		fmt.Println()

		// Get HTTP client
		httpClient := getHTTPClient()

		// Build filter
		filter := &models.MangaFilter{
			Query:       query,
			Status:      status,
			Author:      author,
			MinChapters: minChapters,
			SortBy:      sortBy,
			Order:       order,
			Limit:       limit,
		}
		if genres != "" {
			filter.Genres = strings.Split(genres, ",")
		}

		// Search via API
		results, err := httpClient.SearchManga(filter)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results.Manga) == 0 {
			fmt.Println("No manga found matching your criteria.")
			return nil
		}

		fmt.Printf("Found %d results:\n\n", len(results.Manga))
		printMangaResults(results.Manga)

		return nil
	},
}

func init() {
	MangaCmd.AddCommand(advancedSearchCmd)
	advancedSearchCmd.Flags().StringP("genre", "g", "", "Genres (comma-separated)")
	advancedSearchCmd.Flags().StringP("status", "s", "", "Status filter")
	advancedSearchCmd.Flags().StringP("author", "a", "", "Author name")
	advancedSearchCmd.Flags().Int("min-chapters", 0, "Minimum chapter count")
	advancedSearchCmd.Flags().String("sort-by", "title", "Sort field (title, total_chapters)")
	advancedSearchCmd.Flags().String("order", "asc", "Sort order (asc/desc)")
	advancedSearchCmd.Flags().IntP("limit", "l", 20, "Maximum results")
}
