package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"mangahub/pkg/session"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from MangaHub",
	Long: `Logout from MangaHub and clear the local session.

Example:
  mangahub auth logout`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := session.Load()
		if err != nil {
			fmt.Println("✓ Already logged out")
			return nil
		}

		fmt.Printf("Logging out user %s...\n", sess.Username)
		if err := session.Clear(); err != nil {
			return fmt.Errorf("failed to clear session: %w", err)
		}

		fmt.Println("✓ Logged out successfully")
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(logoutCmd)
}
