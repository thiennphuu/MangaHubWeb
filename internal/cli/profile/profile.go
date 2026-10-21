package profile

import "github.com/spf13/cobra"

var ProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage MangaHub profiles",
	Long:  "Create, switch, and list MangaHub CLI profiles.",
}

func init() {
	ProfileCmd.AddCommand(createCmd)
	ProfileCmd.AddCommand(switchCmd)
	ProfileCmd.AddCommand(listCmd)
}
