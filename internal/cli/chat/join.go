package chat

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"mangahub/pkg/client"
	"mangahub/pkg/models"
	"mangahub/pkg/session"
	"mangahub/pkg/utils"
)

var wsClient *client.WebSocketClient

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join a chat room",
	Long:  `Join the general chat room or a specific manga discussion room.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.NewLogger()
		mangaID, _ := cmd.Flags().GetString("manga-id")

		// Load session for user info
		sess, err := session.Load()
		if err != nil || sess.Token == "" {
			fmt.Println("⚠ Not logged in. Using guest mode.")
			sess = &session.Session{
				UserID:   "guest-" + fmt.Sprintf("%d", time.Now().Unix()),
				Username: "guest",
			}
		}

		// Determine room ID
		roomID := "general"
		roomName := "General Chat"
		if mangaID != "" {
			roomID = mangaID
			roomName = mangaID + " Discussion"
		}

		fmt.Println("Connecting to WebSocket chat server at ws://10.238.53.72:9093...")

		// Create WebSocket client
		wsClient = client.NewWebSocketClient("ws://10.238.53.72:9093", sess.UserID, sess.Username)

		// Set callbacks
		wsClient.SetCallbacks(
			func(msg models.ChatMessage) {
				timestamp := time.Unix(msg.Timestamp, 0).Format("15:04")
				fmt.Printf("\r[%s] %s: %s\n", timestamp, msg.Username, msg.Message)
				fmt.Printf("%s> ", sess.Username)
			},
			func(err error) {
				logger.Error("chat connection error: %v", err)
				fmt.Print("\r❌ Chat connection error\n")
				fmt.Printf("%s> ", sess.Username)
			},
			func() {
				fmt.Printf("✓ Connected to %s\n", roomName)
			},
			func() {
				fmt.Println("\n✓ Disconnected from chat server")
			},
		)

		// Connect to server
		if err := wsClient.Connect(roomID); err != nil {
			logger.Error("failed to connect to chat: %v", err)
			fmt.Println("❌ Failed to connect")
			fmt.Println("\nMake sure the WebSocket server is running:")
			fmt.Println("  go run ./cmd/websocket-server")
			return nil
		}

		// Display chat room info
		fmt.Printf("Chat Room: #%s\n", roomID)
		fmt.Printf("Connected users: %d\n", wsClient.GetConnectedUsers())
		fmt.Println("Your status: Online")

		// Display recent messages
		fmt.Println("Recent messages:")
		recentMessages := wsClient.GetRecentMessages()
		if len(recentMessages) > 0 {
			for _, msg := range recentMessages {
				timestamp := time.Unix(msg.Timestamp, 0).Format("15:04")
				fmt.Printf("[%s] %s: %s\n", timestamp, msg.Username, msg.Message)
			}
		} else {
			fmt.Println("  (No recent messages)")
		}

		fmt.Println("─────────────────────────────────────────────────────────────")
		fmt.Println("You are now in chat. Type your message and press Enter.")
		fmt.Println("Type /help for commands or /quit to leave.")
		fmt.Printf("%s> ", sess.Username)

		// Interactive mode
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())

			if input == "" {
				fmt.Printf("%s> ", sess.Username)
				continue
			}

			// Handle commands
			if strings.HasPrefix(input, "/") {
				if handleCommand(input, sess.Username) {
					break // /quit command
				}
				fmt.Printf("%s> ", sess.Username)
				continue
			}

			// Send message
			if err := wsClient.SendMessage(input); err != nil {
				logger.Error("failed to send message: %v", err)
				fmt.Println("❌ Failed to send message")
			}
			fmt.Printf("%s> ", sess.Username)
		}

		// Disconnect
		wsClient.Disconnect()
		return nil
	},
}

func handleCommand(input, username string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		fmt.Println("Chat Commands:")
		fmt.Println("  /help     - Show this help")
		fmt.Println("  /users    - List online users")
		fmt.Println("  /quit     - Leave chat")
		fmt.Println("  /pm <user> <msg> - Private message")
		fmt.Println("  /manga <id>      - Switch to manga chat")
		fmt.Println("  /history  - Show recent history")
		fmt.Println("  /status   - Connection status")

	case "/quit", "/exit", "/q":
		fmt.Println("Leaving chat...")
		return true

	case "/users":
		userCount := 1
		if wsClient != nil {
			userCount = wsClient.GetConnectedUsers()
		}
		fmt.Printf("Online Users (%d):\n", userCount)
		fmt.Println("● " + username + " (You)")
		if userCount > 1 {
			fmt.Printf("  ... and %d other user(s)\n", userCount-1)
		}
		fmt.Println("  (Full user list requires server support)")

	case "/pm":
		if len(parts) < 3 {
			fmt.Println("Usage: /pm <username> <message>")
		} else {
			targetUser := parts[1]
			message := strings.Join(parts[2:], " ")
			if wsClient != nil {
				if err := wsClient.SendPrivateMessage(targetUser, message); err != nil {
					utils.NewLogger().Error("failed to send private message: %v", err)
					fmt.Println("❌ Failed to send PM")
				} else {
					fmt.Printf("[PM to %s] %s\n", targetUser, message)
				}
			}
		}

	case "/manga":
		if len(parts) < 2 {
			fmt.Println("Usage: /manga <manga-id>")
		} else {
			mangaID := parts[1]
			fmt.Printf("Switching to %s discussion...\n", mangaID)
			if wsClient != nil {
				if err := wsClient.SwitchRoom(mangaID); err != nil {
					utils.NewLogger().Error("failed to switch chat room: %v", err)
					fmt.Println("❌ Failed to switch room")
				} else {
					fmt.Printf("✓ Joined #%s\n", mangaID)
				}
			}
		}

	case "/history":
		fmt.Println("Recent Chat History:")
		fmt.Println("(History requires server support)")

	case "/status":
		if wsClient != nil && wsClient.IsConnected() {
			fmt.Println("Status: ✓ Connected")
			fmt.Printf("Room: #%s\n", wsClient.GetRoomID())
			fmt.Printf("Username: %s\n", wsClient.GetUsername())
		} else {
			fmt.Println("Status: ❌ Disconnected")
		}

	default:
		fmt.Printf("Unknown command: %s (type /help for commands)\n", cmd)
	}

	return false
}

func init() {
	joinCmd.Flags().StringP("manga-id", "m", "", "Join specific manga discussion room")
	ChatCmd.AddCommand(joinCmd)
}
