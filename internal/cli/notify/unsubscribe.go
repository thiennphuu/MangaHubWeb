package notify

import (
	"fmt"

	"mangahub/pkg/client"
	"mangahub/pkg/utils"

	"github.com/spf13/cobra"
)

var unsubscribeCmd = &cobra.Command{
	Use:   "unsubscribe",
	Short: "Unsubscribe from notifications",
	Long: `Unsubscribe from manga chapter release notifications.

This command disconnects from the UDP notification server and
unregisters the client from receiving notifications.

Examples:
  mangahub notify unsubscribe
  mangahub notify unsubscribe --manga-id one-piece
  mangahub notify unsubscribe --all`,
	RunE: runUnsubscribe,
}

func runUnsubscribe(cmd *cobra.Command, args []string) error {
	logger := utils.NewLogger()
	mangaID, _ := cmd.Flags().GetString("manga-id")
	all, _ := cmd.Flags().GetBool("all")
	serverAddr, _ := cmd.Flags().GetString("server")

	// Check if user is logged in
	session, err := loadSession()
	if err != nil {
		logger.Warn("user is not logged in")
		fmt.Println("You are not logged in.")
		fmt.Println("\nPlease login first:")
		fmt.Println("  mangahub auth login --username <username>")
		return nil
	}

	// Get notification service for database operations
	notifySvc, err := getNotificationService()
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Connect to UDP server and unregister
	fmt.Printf("Connecting to UDP notification server at %s...\n", serverAddr)
	udpClient := client.NewUDPClient(serverAddr)
	err = udpClient.Connect()
	if err != nil {
		logger.Warn("could not connect to UDP server: %v", err)
		fmt.Println("  Continuing with local database unsubscription only.")
	} else {
		defer udpClient.Close()

		// Unregister from UDP server
		err = udpClient.Unregister()
		if err != nil {
			logger.Warn("failed to unregister from UDP server: %v", err)
		} else {
			fmt.Println("✓ Unregistered from UDP notification server")
		}
	}

	// Handle unsubscription logic
	if mangaID != "" {
		// Unsubscribe from specific manga
		err = notifySvc.Unsubscribe(session.UserID, mangaID)
		if err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
		fmt.Printf("✓ Unsubscribed from notifications for '%s'\n", mangaID)
	} else if all {
		// Unsubscribe from all
		err = notifySvc.UnsubscribeAll(session.UserID)
		if err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
		err = notifySvc.DisableNotifications(session.UserID)
		if err != nil {
			return fmt.Errorf("failed to disable notifications: %w", err)
		}
		fmt.Println("✓ Unsubscribed from all notifications")
	} else {
		// Disable general notifications
		err = notifySvc.DisableNotifications(session.UserID)
		if err != nil {
			return fmt.Errorf("failed to disable notifications: %w", err)
		}
		fmt.Println("✓ Disabled chapter release notifications")
	}

	// Show current subscription status
	fmt.Println("\nCurrent Status:")
	prefs, _ := notifySvc.GetPreferences(session.UserID)
	if prefs != nil {
		fmt.Printf("  Chapter releases: %s\n", boolToEnabled(prefs.ChapterReleases))
		fmt.Printf("  Email notifications: %s\n", boolToEnabled(prefs.EmailNotifications))
	} else {
		fmt.Println("  Notifications: disabled")
	}

	// Show remaining subscriptions
	subs, _ := notifySvc.GetSubscriptions(session.UserID)
	if len(subs) > 0 {
		fmt.Printf("\nRemaining manga subscriptions: %d\n", len(subs))
		for _, s := range subs {
			fmt.Printf("  - %s\n", s)
		}
	}

	return nil
}

func init() {
	NotifyCmd.AddCommand(unsubscribeCmd)
	unsubscribeCmd.Flags().StringP("manga-id", "m", "", "Unsubscribe from specific manga")
	unsubscribeCmd.Flags().BoolP("all", "a", false, "Unsubscribe from all notifications")
	unsubscribeCmd.Flags().StringP("server", "s", "10.238.53.72:9091", "UDP server address")
}
