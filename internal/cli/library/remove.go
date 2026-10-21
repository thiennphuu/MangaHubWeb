package library

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove --manga-id <id>",
	Short: "Remove manga from library",
	Long: `Remove a manga from your personal library via the API server.

Examples:
  mangahub library remove --manga-id completed-series
  mangahub library remove -m one-piece
  mangahub library remove -m naruto --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		force, _ := cmd.Flags().GetBool("force")

		if mangaID == "" {
			return fmt.Errorf("--manga-id is required")
		}

		// Check if user is logged in and get HTTP client
		httpClient, _, err := newAuthenticatedHTTPClient()
		if err != nil {
			fmt.Println("You are not logged in.")
			fmt.Println("\nPlease login first:")
			fmt.Println("  mangahub auth login --username <username>")
			return nil
		}

		// Confirm removal unless --force is used
		if !force {
			fmt.Printf("Are you sure you want to remove '%s' from your library?\n", mangaID)
			fmt.Print("\nType 'yes' to confirm: ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			if input != "yes" && input != "y" {
				fmt.Println("Removal cancelled.")
				return nil
			}
		}

		// Remove from library via API
		err = httpClient.RemoveFromLibrary(mangaID)
		if err != nil {
			return fmt.Errorf("failed to remove from library: %w", err)
		}

		fmt.Printf("\n✓ Successfully removed '%s' from your library.\n", mangaID)

		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(removeCmd)
	removeCmd.Flags().StringP("manga-id", "m", "", "Manga ID (required)")
	removeCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	removeCmd.MarkFlagRequired("manga-id")
}
