package manga

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <manga-id>",
	Short: "View detailed manga information",
	Long: `Display detailed information about a specific manga via the API server.
Includes description, genres, chapters, and your current reading status.

Example:
  mangahub manga info one-piece`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID := args[0]

		// Get HTTP client
		httpClient := getHTTPClient()

		// Fetch manga from API
		m, err := httpClient.GetManga(mangaID)
		if err != nil {
			return fmt.Errorf("manga not found: %w", err)
		}

		// Print manga info
		titleLine := fmt.Sprintf(" %s ", strings.ToUpper(m.Title))
		boxWidth := max(len(titleLine)+2, 70)

		fmt.Println("┌" + strings.Repeat("─", boxWidth) + "┐")
		fmt.Printf("│%-*s│\n", boxWidth, titleLine)
		fmt.Println("└" + strings.Repeat("─", boxWidth) + "┘")
		fmt.Println()

		fmt.Println("Basic Information:")
		fmt.Printf("  ID:      %s\n", m.ID)
		fmt.Printf("  Title:   %s\n", m.Title)
		fmt.Printf("  Author:  %s\n", m.Author)
		fmt.Printf("  Genres:  %s\n", strings.Join(m.Genres, ", "))
		fmt.Printf("  Status:  %s\n", m.Status)
		fmt.Println()

		fmt.Println("Progress:")
		fmt.Printf("  Total Chapters: %d\n", m.TotalChapters)
		fmt.Printf("  Created:        %s\n", m.CreatedAt.Format("2006-01-02"))
		fmt.Printf("  Updated:        %s\n", m.UpdatedAt.Format("2006-01-02"))
		fmt.Println()

		if m.Description != "" {
			fmt.Println("Description:")
			// Word wrap description
			words := strings.Fields(m.Description)
			line := " "
			for _, word := range words {
				if len(line)+len(word)+1 > 75 {
					fmt.Println(line)
					line = " " + word
				} else {
					line += " " + word
				}
			}
			if line != " " {
				fmt.Println(line)
			}
			fmt.Println()
		}

		fmt.Println("Actions:")
		fmt.Printf("  Update Progress: mangahub progress update --manga-id %s --chapter <n>\n", m.ID)
		fmt.Printf("  Add to Library:  mangahub library add --manga-id %s\n", m.ID)
		fmt.Printf("  Rate/Review:     mangahub library update --manga-id %s --rating <1-10>\n", m.ID)

		return nil
	},
}

func init() {
	MangaCmd.AddCommand(infoCmd)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
