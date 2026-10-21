package notify

import "github.com/spf13/cobra"

// NotifyCmd is the main notify command (Parent Command)
// All subcommands (subscribe, unsubscribe, preferences, test) are registered
// via init() in their respective files using NotifyCmd.AddCommand()
var NotifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "UDP notification management",
	Long: `Subscribe to and manage manga release notifications via UDP.

This command group provides functionality to:
  - Subscribe to chapter release notifications (connects to UDP server)
  - Unsubscribe from notifications (disconnects from UDP server)
  - View and update notification preferences
  - Test the notification system

The notification system uses UDP protocol on port 9091 for real-time
broadcasting of chapter release notifications to subscribed clients.

Available Commands:
  subscribe     Subscribe to notifications and register with UDP server
  unsubscribe   Unsubscribe from notifications and unregister from UDP server
  preferences   View and update notification preferences
  test          Test the notification system

Examples:
  # Subscribe to all chapter release notifications
  mangahub notify subscribe

  # Subscribe and keep listening for notifications
  mangahub notify subscribe --listen

  # Subscribe to specific manga notifications
  mangahub notify subscribe --manga-id one-piece

  # Unsubscribe from all notifications
  mangahub notify unsubscribe

  # Unsubscribe from specific manga
  mangahub notify unsubscribe --manga-id one-piece

  # View notification preferences
  mangahub notify preferences

  # Test notification system
  mangahub notify test`,
}
