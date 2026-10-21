package notify

import (
	"fmt"

	"github.com/spf13/cobra"
)

var preferencesCmd = &cobra.Command{
	Use:   "preferences",
	Short: "View notification preferences",
	Long: `View and manage your notification preferences.

Examples:
  mangahub notify preferences
  mangahub notify preferences --set email=true
  mangahub notify preferences --set sound=false`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if user is logged in
		session, err := loadSession()
		if err != nil {
			fmt.Println("You are not logged in.")
			fmt.Println("\nPlease login first:")
			fmt.Println("  mangahub auth login --username <username>")
			return nil
		}

		// Get notification service
		notifySvc, err := getNotificationService()
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}

		fmt.Printf("📬 Notification Preferences for %s\n\n", session.Username)

		// Get preferences
		prefs, err := notifySvc.GetPreferences(session.UserID)
		if err != nil {
			// No preferences set, show defaults
			fmt.Println("General Settings:")
			fmt.Println("  Chapter releases: not configured")
			fmt.Println("  Email notifications: not configured")
			fmt.Println("  Sound enabled: not configured")
			fmt.Println("\nUse 'mangahub notify subscribe' to enable notifications")
			return nil
		}

		fmt.Println("General Settings:")
		fmt.Printf("  Chapter releases: %s\n", boolToEnabled(prefs.ChapterReleases))
		fmt.Printf("  Email notifications: %s\n", boolToEnabled(prefs.EmailNotifications))
		fmt.Printf("  Sound enabled: %s\n", boolToEnabled(prefs.SoundEnabled))

		// Get subscriptions
		subscriptions, err := notifySvc.GetSubscriptions(session.UserID)
		if err == nil && len(subscriptions) > 0 {
			fmt.Printf("\nSubscribed Manga (%d):\n", len(subscriptions))
			for _, mangaID := range subscriptions {
				fmt.Printf("  • %s\n", mangaID)
			}
		} else {
			fmt.Println("\nNo manga subscriptions.")
			fmt.Println("Use 'mangahub notify subscribe --manga-id <id>' to subscribe")
		}

		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(preferencesCmd)
}
