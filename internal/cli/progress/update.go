package progress

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update --manga-id <id> --chapter <number>",
	Short: "Update reading progress",
	Long: `Update your reading progress for a manga via the API server.

Examples:
  mangahub progress update --manga-id one-piece --chapter 1095
  mangahub progress update --manga-id naruto --chapter 700 --notes "Great ending!"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		chapter, _ := cmd.Flags().GetInt("chapter")
		notes, _ := cmd.Flags().GetString("notes")

		if mangaID == "" || chapter <= 0 {
			return fmt.Errorf("--manga-id and --chapter are required")
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
		mangaInfo, err := httpClient.GetManga(mangaID)
		if err != nil {
			return fmt.Errorf("manga '%s' not found. Use 'mangahub manga search' to find valid manga IDs", mangaID)
		}

		fmt.Printf("Updating progress for '%s' via API server...\n", mangaInfo.Title)

		// Determine status based on chapter
		status := "reading"
		if mangaInfo.TotalChapters > 0 && chapter >= mangaInfo.TotalChapters {
			status = "completed"
		}

		// Update progress via API
		err = httpClient.UpdateProgress(mangaID, chapter, status, 0, notes)
		if err != nil {
			// If entry doesn't exist, try adding to library first
			addErr := httpClient.AddToLibrary(mangaID, "reading", 0, "")
			if addErr == nil {
				// Try update again
				err = httpClient.UpdateProgress(mangaID, chapter, status, 0, notes)
			}
			if err != nil {
				return fmt.Errorf("failed to update progress: %w", err)
			}
		}

		fmt.Println()
		fmt.Printf("✓ Progress updated for '%s'!\n", mangaInfo.Title)
		fmt.Printf("  Chapter: %d", chapter)
		if mangaInfo.TotalChapters > 0 {
			fmt.Printf(" / %d", mangaInfo.TotalChapters)
		}
		fmt.Println()
		fmt.Printf("  Status: %s\n", status)
		fmt.Printf("  Updated: %s\n", time.Now().Format("2006-01-02 15:04:05"))

		if notes != "" {
			fmt.Printf("  Notes: %s\n", notes)
		}

		if status == "completed" {
			fmt.Println("\n🎉 Congratulations! You've completed this manga!")
		}

		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("manga-id", "m", "", "Manga ID (required)")
	updateCmd.Flags().IntP("chapter", "c", 0, "Chapter number (required)")
	updateCmd.Flags().StringP("notes", "n", "", "Reading notes (optional)")
	updateCmd.MarkFlagRequired("manga-id")
	updateCmd.MarkFlagRequired("chapter")
}
