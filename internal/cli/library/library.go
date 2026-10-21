package library

import "github.com/spf13/cobra"

// LibraryCmd is the main library command
var LibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Manage your manga library",
	Long:  `Add, remove, and manage manga in your personal library.`,
}
