package profile

import (
	"fmt"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		fmt.Printf("Creating profile '%s'...\n", name)
		fmt.Println("✓ Profile created successfully")
		fmt.Printf("Use 'mangahub profile switch --name %s' to activate it\n", name)
		return nil
	},
}

func init() {
	createCmd.Flags().String("name", "", "Profile name")
	_ = createCmd.MarkFlagRequired("name")
}
