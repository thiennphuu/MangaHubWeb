package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setCmd handles `mangahub config set`.
var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// For now we only simulate config updates (same as original behavior).
		// This can later be wired to pkg/config to persist changes.
		fmt.Printf("Setting %s = %s\n", key, value)
		fmt.Println("✓ Configuration updated")
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
}
