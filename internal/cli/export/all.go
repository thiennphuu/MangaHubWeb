package export

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// allCmd handles `mangahub export all`.
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Export all data",
	RunE:  runExportAll,
}

func init() {
	allCmd.Flags().String("output", "mangahub-backup.tar.gz", "Output archive path")

	ExportCmd.AddCommand(allCmd)
}

// runExportAll performs a full export by calling the other exporters and tarring the results.
func runExportAll(cmd *cobra.Command, args []string) error {
	output, _ := cmd.Flags().GetString("output")
	tmpDir := filepath.Join(os.TempDir(), "mangahub-export")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	libraryPath := filepath.Join(tmpDir, "library.json")
	progressPath := filepath.Join(tmpDir, "progress.csv")

	// Reuse subcommand logic for consistency.
	if err := runExportLibraryForPath("json", libraryPath); err != nil {
		return err
	}
	if err := runExportProgressForPath("csv", progressPath); err != nil {
		return err
	}

	if err := createTarGz(output, []string{libraryPath, progressPath}); err != nil {
		return err
	}

	fmt.Printf("✓ Full backup written to %s\n", output)
	return nil
}

// createTarGz creates a .tar.gz archive of the given files.
func createTarGz(dest string, files []string) error {
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create header for %s: %w", path, err)
		}
		hdr.Name = filepath.Base(path)

		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("failed to write header for %s: %w", path, err)
		}

		fh, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", path, err)
		}

		if _, err := io.Copy(tw, fh); err != nil {
			fh.Close()
			return fmt.Errorf("failed to copy %s into archive: %w", path, err)
		}
		fh.Close()
	}

	return nil
}
