package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/utils"
)

var changePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Change your password",
	Long: `Change your MangaHub account password.

You must be logged in to change your password.

Example:
  mangahub auth change-password`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if user is logged in
		session, err := loadSession()
		if err != nil {
			fmt.Println("You are not logged in.")
			fmt.Println("\nPlease login first:")
			fmt.Println("  mangahub auth login --username <username>")
			return nil
		}

		prompt := utils.NewPrompt()

		// Prompt for current password
		currentPassword, err := prompt.Password("Current password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		// Verify current password
		userSvc, err := getUserService()
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}

		user, err := userSvc.GetByID(session.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		authSvc := getAuthService()
		if err := authSvc.VerifyPassword(currentPassword, user.PasswordHash); err != nil {
			return fmt.Errorf("current password is incorrect")
		}

		// Prompt for new password
		newPassword, err := prompt.Password("New password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		if len(newPassword) < 6 {
			return fmt.Errorf("password must be at least 6 characters")
		}

		// Confirm new password
		confirmPassword, err := prompt.Password("Confirm new password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		if newPassword != confirmPassword {
			return fmt.Errorf("passwords do not match")
		}

		// Hash new password
		hashedPassword, err := authSvc.HashPassword(newPassword)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Update password in database
		err = userSvc.UpdatePassword(session.UserID, hashedPassword)
		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		fmt.Println("\n✓ Password changed successfully!")
		fmt.Println("\nYour session remains active. You don't need to login again.")

		return nil
	},
}

func init() {
	AuthCmd.AddCommand(changePasswordCmd)
}
