package chat

import "github.com/spf13/cobra"

// ChatCmd is the main chat command
var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "WebSocket chat system",
	Long:  `Join chat rooms and communicate with other manga fans in real-time.`,
}
