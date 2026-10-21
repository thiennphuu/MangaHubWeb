package config

import "github.com/spf13/cobra"

// ConfigCmd is the main config command (parent/root for config subcommands).
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `View and manage MangaHub configuration settings.`,
}
