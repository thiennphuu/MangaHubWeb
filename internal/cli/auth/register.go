package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
	"mangahub/pkg/utils"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new MangaHub account",
	Long: `Register a new account on MangaHub via the API server.

Examples:
  mangahub auth register --username johndoe --email john@example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		email, _ := cmd.Flags().GetString("email")

		if username == "" {
			return fmt.Errorf("please provide --username")
		}
		if email == "" {
			return fmt.Errorf("please provide --email")
		}

		// Prompt for password
		prompt := utils.NewPrompt()
		password, err := prompt.Password("Password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		if len(password) < 6 {
			return fmt.Errorf("password must be at least 6 characters")
		}

		confirmPassword, err := prompt.Password("Confirm Password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		if password != confirmPassword {
			return fmt.Errorf("passwords do not match")
		}

		// Create HTTP client
		httpClient := client.NewHTTPClient(getAPIURL(), "")

		fmt.Printf("Registering user '%s' via API server...\n", username)

		// Call API server to register
		result, err := httpClient.Register(username, email, password)
		if err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}

		fmt.Println("✓ Registration successful!")
		fmt.Println()
		fmt.Printf("Username: %s\n", username)
		fmt.Printf("Email:    %s\n", email)
		fmt.Printf("User ID:  %s\n", result.UserID)
		fmt.Println()
		fmt.Println("You can now login with:")
		fmt.Printf("  mangahub auth login --username %s\n", username)

		return nil
	},
}

func init() {
	AuthCmd.AddCommand(registerCmd)

	registerCmd.Flags().StringP("username", "u", "", "Username (required)")
	registerCmd.Flags().StringP("email", "e", "", "Email address (required)")
}
