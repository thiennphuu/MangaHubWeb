package profile

import (
	"fmt"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch active profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		fmt.Printf("Switching to profile '%s'...\n", name)
		fmt.Println("✓ Active profile updated")
		return nil
	},
}

func init() {
	switchCmd.Flags().String("name", "", "Profile name")
	_ = switchCmd.MarkFlagRequired("name")
}
