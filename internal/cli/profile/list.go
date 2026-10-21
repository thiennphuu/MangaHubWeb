package profile

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles := []string{"default", "work", "reading"}
		active := "default"

		cmd.Println("Available profiles:")
		for _, p := range profiles {
			marker := " "
			if p == active {
				marker = "*"
			}
			cmd.Printf(" %s %s\n", marker, p)
		}
		cmd.Println("\nUse 'mangahub profile switch --name <profile>' to change active profile")
		return nil
	},
}
