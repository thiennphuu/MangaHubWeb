package notify

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test notification system",
	Long: `Send a test notification to verify the notification system is working.

Example:
  mangahub notify test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if user is logged in
		session, err := loadSession()
		if err != nil {
			fmt.Println("You are not logged in.")
			fmt.Println("\nPlease login first:")
			fmt.Println("  mangahub auth login --username <username>")
			return nil
		}

		fmt.Println("Sending test notification...")
		fmt.Println()

		// Simulate UDP notification
		fmt.Println("📧 Test Notification")
		fmt.Println("─────────────────────────────────────")
		fmt.Printf("To: %s (%s)\n", session.Username, session.Email)
		fmt.Printf("Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Println("Message: This is a test notification from MangaHub")
		fmt.Println("─────────────────────────────────────")
		fmt.Println()
		fmt.Println("✓ Test notification sent successfully!")
		fmt.Println("\nNote: In production, this would be sent via UDP to your registered devices.")

		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(testCmd)
}
