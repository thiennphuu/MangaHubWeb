package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
	"mangahub/pkg/session"
	"mangahub/pkg/utils"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to MangaHub",
	Long: `Login to MangaHub via the API server with your username or email.

Examples:
  mangahub auth login --username johndoe
  mangahub auth login --email john@example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		email, _ := cmd.Flags().GetString("email")

		if username == "" && email == "" {
			return fmt.Errorf("please provide --username or --email")
		}

		// Prompt for password
		prompt := utils.NewPrompt()
		password, err := prompt.Password("Password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		// Create HTTP client
		httpClient := client.NewHTTPClient(getAPIURL(), "")

		// Use username or email for login
		loginUser := username
		if loginUser == "" {
			loginUser = email
		}

		fmt.Printf("Logging in as %s via API server...\n", loginUser)

		// Call API server to login
		loginResp, err := httpClient.Login(loginUser, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Save session
		sess := &session.Session{
			UserID:    loginResp.UserID,
			Username:  loginResp.Username,
			Email:     "", // API doesn't return email in login response
			Token:     loginResp.Token,
			ExpiresAt: loginResp.ExpiresAt.Format("2006-01-02 15:04:05 MST"),
		}
		if err := session.Save(sess); err != nil {
			utils.NewLogger().Warn("could not save session: %v", err)
		}

		fmt.Println("✓ Login successful")
		fmt.Println()
		fmt.Printf("User: %s\n", loginResp.Username)
		fmt.Printf("Token expires: %s\n", loginResp.ExpiresAt.Format("2006-01-02 15:04:05 MST"))

		return nil
	},
}

func init() {
	AuthCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("username", "u", "", "Username")
	loginCmd.Flags().StringP("email", "e", "", "Email address")
}
