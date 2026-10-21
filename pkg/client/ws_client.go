package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"mangahub/pkg/models"
)

// WebSocketClient represents a WebSocket client for chat
type WebSocketClient struct {
	conn           *websocket.Conn
	serverURL      string
	userID         string
	username       string
	roomID         string
	connected      bool
	mutex          sync.RWMutex
	done           chan struct{}
	messages       chan models.ChatMessage
	recentMessages []models.ChatMessage
	connectedUsers int
	onMessage      func(models.ChatMessage)
	onError        func(error)
	onConnect      func()
	onDisconnect   func()
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(serverURL, userID, username string) *WebSocketClient {
	return &WebSocketClient{
		serverURL:      serverURL,
		userID:         userID,
		username:       username,
		roomID:         "general",
		done:           make(chan struct{}),
		messages:       make(chan models.ChatMessage, 100),
		recentMessages: make([]models.ChatMessage, 0),
		connectedUsers: 1,
	}
}

// SetCallbacks sets the callback functions
func (c *WebSocketClient) SetCallbacks(onMessage func(models.ChatMessage), onError func(error), onConnect func(), onDisconnect func()) {
	c.onMessage = onMessage
	c.onError = onError
	c.onConnect = onConnect
	c.onDisconnect = onDisconnect
}

// Connect connects to the WebSocket server
func (c *WebSocketClient) Connect(roomID string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return fmt.Errorf("already connected")
	}

	c.roomID = roomID
	if c.roomID == "" {
		c.roomID = "general"
	}

	// Parse and construct WebSocket URL
	u, err := url.Parse(c.serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	// Add room to path (server expects /ws/:room)
	u.Path = fmt.Sprintf("/ws/%s", c.roomID)

	// Add query parameters for auth
	q := u.Query()
	q.Set("user_id", c.userID)
	q.Set("username", c.username)
	u.RawQuery = q.Encode()

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.connected = true
	c.done = make(chan struct{})

	// Start read loop
	go c.readLoop()

	if c.onConnect != nil {
		c.onConnect()
	}

	return nil
}

// Disconnect disconnects from the WebSocket server
func (c *WebSocketClient) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false
	close(c.done)

	if c.conn != nil {
		// Send close message
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
	}

	if c.onDisconnect != nil {
		c.onDisconnect()
	}

	return nil
}

// SendMessage sends a chat message
func (c *WebSocketClient) SendMessage(message string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected")
	}

	msg := models.ChatMessage{
		UserID:    c.userID,
		Username:  c.username,
		RoomID:    c.roomID,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	return c.conn.WriteJSON(msg)
}

// SendPrivateMessage sends a private message to a user
func (c *WebSocketClient) SendPrivateMessage(targetUser, message string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected")
	}

	msg := models.ChatMessage{
		UserID:    c.userID,
		Username:  c.username,
		RoomID:    "private:" + targetUser,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	return c.conn.WriteJSON(msg)
}

// SwitchRoom switches to a different chat room
func (c *WebSocketClient) SwitchRoom(roomID string) error {
	if err := c.Disconnect(); err != nil {
		return err
	}
	return c.Connect(roomID)
}

// IsConnected returns whether the client is connected
func (c *WebSocketClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}

// GetRoomID returns the current room ID
func (c *WebSocketClient) GetRoomID() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.roomID
}

// GetUsername returns the username
func (c *WebSocketClient) GetUsername() string {
	return c.username
}

// GetConnectedUsers returns the number of connected users
func (c *WebSocketClient) GetConnectedUsers() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connectedUsers
}

// GetRecentMessages returns recent messages from the chat room
func (c *WebSocketClient) GetRecentMessages() []models.ChatMessage {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.recentMessages
}

// AddRecentMessage adds a message to the recent messages list
func (c *WebSocketClient) AddRecentMessage(msg models.ChatMessage) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.recentMessages = append(c.recentMessages, msg)
	// Keep only last 50 messages
	if len(c.recentMessages) > 50 {
		c.recentMessages = c.recentMessages[len(c.recentMessages)-50:]
	}
}

// SetConnectedUsers sets the number of connected users
func (c *WebSocketClient) SetConnectedUsers(count int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.connectedUsers = count
}

// readLoop reads messages from the WebSocket connection
func (c *WebSocketClient) readLoop() {
	defer func() {
		c.mutex.Lock()
		c.connected = false
		c.mutex.Unlock()
	}()

	for {
		select {
		case <-c.done:
			return
		default:
			_, messageData, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
					if c.onError != nil {
						c.onError(err)
					}
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}

			var msg models.ChatMessage
			if err := json.Unmarshal(messageData, &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			if c.onMessage != nil {
				c.onMessage(msg)
			}
		}
	}
}

// GetMessages returns the messages channel
func (c *WebSocketClient) GetMessages() <-chan models.ChatMessage {
	return c.messages
}
