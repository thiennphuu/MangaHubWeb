package db

import "github.com/spf13/cobra"

// DBCmd is the main database command (parent/root for db subcommands).
var DBCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management",
	Long:  `Inspect, repair, and view statistics for the local MangaHub SQLite database.`,
}


