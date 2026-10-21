package export

import "github.com/spf13/cobra"

// ExportCmd is the main export command (parent/root for export subcommands).
var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data",
	Long:  `Export your library and progress data in various formats.`,
}
