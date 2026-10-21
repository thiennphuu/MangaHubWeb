package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
	"mangahub/pkg/models"
	"mangahub/pkg/session"
)

// progressCmd handles `mangahub export progress`.
var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Export progress history",
	RunE:  runExportProgress,
}

func init() {
	progressCmd.Flags().String("format", "csv", "Export format (currently only csv)")
	progressCmd.Flags().String("output", "progress.csv", "Output file path")

	ExportCmd.AddCommand(progressCmd)
}

func runExportProgress(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	return runExportProgressForPath(format, output)
}

// runExportProgressForPath is shared between `export progress` and `export all`.
func runExportProgressForPath(format, output string) error {
	if format == "" {
		format = "csv"
	}
	if format != "csv" {
		return fmt.Errorf("unsupported format %q (only csv is supported for progress)", format)
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil && filepath.Dir(output) != "." {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Load session for authentication
	sess, err := session.Load()
	if err != nil {
		return fmt.Errorf("you are not logged in. Please login first: %w", err)
	}

	// Create HTTP client and fetch from API server
	fmt.Printf("Fetching progress from API server via HTTP...\n")
	httpClient := client.NewHTTPClient(getAPIURL(), sess.Token)

	entries, err := httpClient.GetLibrary("", 10000, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch progress via HTTP API: %w", err)
	}

	fmt.Printf("✓ Retrieved %d entries from server\n", len(entries))

	if err := exportProgressCSV(output, entries); err != nil {
		return err
	}

	fmt.Printf("✓ Successfully exported progress to %s (csv)\n", output)
	return nil
}

// Helper function removed - now using HTTP API instead of direct database access

// exportProgressCSV writes the progress entries as CSV.
func exportProgressCSV(path string, entries []models.Progress) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	_ = w.Write([]string{
		"user_id",
		"manga_id",
		"status",
		"current_chapter",
		"rating",
		"notes",
		"created_at",
		"updated_at",
	})

	for _, p := range entries {
		record := []string{
			p.UserID,
			p.MangaID,
			p.Status,
			fmt.Sprintf("%d", p.CurrentChapter),
			fmt.Sprintf("%d", p.Rating),
			p.Notes,
			p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		_ = w.Write(record)
	}

	return w.Error()
}
