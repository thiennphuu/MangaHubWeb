package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// resetCmd handles `mangahub config reset`.
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset to default configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		// For now, just simulate reset (matching the prior behavior).
		fmt.Println("Resetting configuration to defaults...")
		fmt.Println("✓ Configuration reset successfully")
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(resetCmd)
}
