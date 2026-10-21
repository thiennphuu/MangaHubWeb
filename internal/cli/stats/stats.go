package stats

import "github.com/spf13/cobra"

// Date range flags shared by stats subcommands.
var (
	fromDate string
	toDate   string
)

// StatsCmd is the main stats command (parent/root for stats subcommands).
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Reading statistics",
	Long:  `View personal manga reading statistics and analytics.`,
}
