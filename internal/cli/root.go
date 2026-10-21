package cli

import (
	"mangahub/internal/cli/auth"
	"mangahub/internal/cli/chat"
	"mangahub/internal/cli/config"
	"mangahub/internal/cli/db"
	"mangahub/internal/cli/export"
	"mangahub/internal/cli/grpc"
	"mangahub/internal/cli/library"
	"mangahub/internal/cli/manga"
	"mangahub/internal/cli/notify"
	"mangahub/internal/cli/profile"
	"mangahub/internal/cli/progress"
	"mangahub/internal/cli/server"
	"mangahub/internal/cli/stats"
	"mangahub/internal/cli/sync"
	"mangahub/pkg/session"

	"github.com/spf13/cobra"
)

var (
	token       string
	apiURL      string = "http://10.238.53.72:8080"
	verbose     bool
	profileName string
)

var rootCmd = &cobra.Command{
	Use:   "mangahub",
	Short: "MangaHub - Manga tracking and discovery CLI",
	Long: `MangaHub is a comprehensive manga tracking system with support for:
- Searching and discovering manga
- Managing your personal library
- Tracking reading progress
- Real-time chat and notifications
- Cross-device synchronization`,
	Version: "3.0.0",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set the profile before any command runs
		if profileName != "" {
			session.SetProfile(profileName)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Authentication token")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api", apiURL, "API server URL")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&profileName, "profile", "", "User profile name (allows multiple users in different terminals)")

	// Add subcommands
	rootCmd.AddCommand(manga.MangaCmd)
	rootCmd.AddCommand(library.LibraryCmd)
	rootCmd.AddCommand(progress.ProgressCmd)
	rootCmd.AddCommand(sync.SyncCmd)
	rootCmd.AddCommand(notify.NotifyCmd)
	rootCmd.AddCommand(chat.ChatCmd)
	rootCmd.AddCommand(grpc.GRPCCmd)
	rootCmd.AddCommand(stats.StatsCmd)
	rootCmd.AddCommand(export.ExportCmd)
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(server.ServerCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(db.DBCmd)
	rootCmd.AddCommand(profile.ProfileCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
