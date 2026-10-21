package sync

import "github.com/spf13/cobra"

// SyncCmd is the main sync command (parent/root for TCP sync subcommands).
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "TCP progress synchronization",
	Long:  `Connect and manage TCP synchronization with the TCP progress sync server.`,
}
