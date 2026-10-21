package progress

import "github.com/spf13/cobra"

// ProgressCmd is the main progress command
var ProgressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Manage reading progress",
	Long:  `Track and synchronize your manga reading progress.`,
}
