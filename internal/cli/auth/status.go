package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
	"mangahub/pkg/session"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long: `Check your current authentication status and session information.
This command validates your session with the API server.

Example:
  mangahub auth status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := session.Load()
		if err != nil {
			fmt.Println("Authentication Status: Not logged in")
			fmt.Println("\nUse 'mangahub auth login --username <username>' to login")
			return nil
		}

		// Create HTTP client with token to validate with API server
		httpClient := client.NewHTTPClient(getAPIURL(), sess.Token)

		// Try to get profile from API to validate token
		user, err := httpClient.GetProfile()
		if err != nil {
			fmt.Println("Authentication Status: Session expired or invalid")
			fmt.Println("\nPlease login again with 'mangahub auth login'")
			return nil
		}

		fmt.Println("Authentication Status: ✓ Logged in")
		fmt.Println()
		fmt.Printf("User ID:  %s\n", user.ID)
		fmt.Printf("Username: %s\n", user.Username)
		fmt.Printf("Email:    %s\n", user.Email)
		fmt.Printf("Expires:  %s\n", sess.ExpiresAt)

		return nil
	},
}

func init() {
	AuthCmd.AddCommand(statusCmd)
}
