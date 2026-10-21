package models

import "time"

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	RoomID    string    `json:"room_id"`
	Message   string    `json:"message"`
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatRoom represents a chat room
type ChatRoom struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	MangaID   string    `json:"manga_id,omitempty"`
	Topic     string    `json:"topic"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatClient represents a connected chat client
type ChatClient struct {
	UserID   string
	Username string
	RoomID   string
	ConnID   string
}

// ClientConnection represents a new client connection
type ClientConnection struct {
	UserID   string
	Username string
	RoomID   string
}
