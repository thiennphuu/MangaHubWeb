package auth

import "github.com/spf13/cobra"

// AuthCmd is the main auth command
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  `Login, logout, and manage authentication tokens.`,
}
