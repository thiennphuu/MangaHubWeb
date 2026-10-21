package chat

import (
	"fmt"
	"mangahub/internal/cli/progress"
	"mangahub/pkg/utils"

	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View chat history",
	Long:  `View recent chat messages from a room.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		limit, _ := cmd.Flags().GetInt("limit")

		roomID := "general"
		if mangaID != "" {
			roomID = mangaID
		}
		if limit == 0 {
			limit = 20
		}

		fmt.Printf("Chat History for #%s (last %d messages):\n", roomID, limit)
		fmt.Println("─────────────────────────────────────────────────────")

		// Show messages from local SQLite
		err := showMessagesHistory(roomID, limit)
		if err != nil {
			utils.NewLogger().Error("failed to load chat history: %v", err)
			fmt.Println("Error loading chat history")
		}
		fmt.Println()
		fmt.Println("To view live messages, use:")
		fmt.Printf("  go run ./cmd/cli chat join")
		if mangaID != "" {
			fmt.Printf(" --manga-id %s", mangaID)
		}
		fmt.Println()
		return nil
	},
}

// showMessagesHistory prints recent chat messages from SQLite for a room
func showMessagesHistory(roomID string, limit int) error {
	// Open DB
	dbPath := "./data/mangahub.db"
	db, err := progress.RequireDatabase(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT username, message, created_at FROM chat_messages WHERE room_id = ? ORDER BY created_at DESC LIMIT ?`, roomID, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	type msg struct {
		Username  string
		Message   string
		CreatedAt string
	}
	var messages []msg
	for rows.Next() {
		var m msg
		if err := rows.Scan(&m.Username, &m.Message, &m.CreatedAt); err != nil {
			return err
		}
		messages = append(messages, m)
	}
	if len(messages) == 0 {
		fmt.Println("No messages found.")
		return nil
	}
	// Print in reverse (oldest first)
	for i := len(messages) - 1; i >= 0; i-- {
		fmt.Printf("[%s] %s: %s\n", messages[i].CreatedAt, messages[i].Username, messages[i].Message)
	}
	return nil
}

func init() {
	historyCmd.Flags().StringP("manga-id", "m", "", "View history for specific manga")
	historyCmd.Flags().IntP("limit", "l", 20, "Number of messages to show")
	ChatCmd.AddCommand(historyCmd)
}
