package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
	"mangahub/pkg/models"
	"mangahub/pkg/session"
)

// libraryCmd handles `mangahub export library`.
var libraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Export library to file",
	RunE:  runExportLibrary,
}

func init() {
	libraryCmd.Flags().String("format", "json", "Export format (json or csv)")
	libraryCmd.Flags().String("output", "library.json", "Output file path")

	ExportCmd.AddCommand(libraryCmd)
}

// runExportLibrary exports the user's library in the requested format.
func runExportLibrary(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	return runExportLibraryForPath(format, output)
}

// getAPIURL returns the API server URL
func getAPIURL() string {
	if url := os.Getenv("MANGAHUB_API_URL"); url != "" {
		return url
	}
	return "http://10.238.53.72:8080"
}

// runExportLibraryForPath is shared between `export library` and `export all`.
func runExportLibraryForPath(format, output string) error {
	if format == "" {
		format = "json"
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil && filepath.Dir(output) != "." {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Load session for authentication
	sess, err := session.Load()
	if err != nil {
		return fmt.Errorf("you are not logged in. Please login first: %w", err)
	}

	// Create HTTP client with authentication token
	fmt.Printf("Fetching library from API server via HTTP...\n")
	httpClient := client.NewHTTPClient(getAPIURL(), sess.Token)

	// Fetch library from HTTP API (using network protocol!)
	entries, err := httpClient.GetLibrary("", 10000, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch library via HTTP API: %w", err)
	}

	fmt.Printf("✓ Retrieved %d entries from server\n", len(entries))

	switch format {
	case "json":
		if err := exportLibraryJSON(output, entries); err != nil {
			return err
		}
	case "csv":
		if err := exportLibraryCSV(output, entries); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported format %q (supported: json, csv)", format)
	}

	fmt.Printf("✓ Successfully exported library to %s (%s)\n", output, format)
	return nil
}

// exportLibraryJSON writes the library entries as JSON.
func exportLibraryJSON(path string, entries interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}
	return nil
}

// exportLibraryCSV writes the library entries as CSV.
func exportLibraryCSV(path string, entries []models.Progress) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	_ = w.Write([]string{"manga_id", "status", "current_chapter", "rating", "updated_at"})

	for _, e := range entries {
		record := []string{
			e.MangaID,
			e.Status,
			fmt.Sprintf("%d", e.CurrentChapter),
			fmt.Sprintf("%d", e.Rating),
			e.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		_ = w.Write(record)
	}

	return w.Error()
}
