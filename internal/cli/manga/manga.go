package manga

import (
	"os"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
)

// MangaCmd is the main manga command
var MangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Manga search and discovery commands",
	Long:  `Search, discover, and view information about manga titles via the API server.`,
}

// getAPIURL returns the API server URL
func getAPIURL() string {
	if url := os.Getenv("MANGAHUB_API_URL"); url != "" {
		return url
	}
	return "http://10.238.53.72:8080"
}

// getHTTPClient returns an HTTP client for manga operations
func getHTTPClient() *client.HTTPClient {
	return client.NewHTTPClient(getAPIURL(), "")
}
