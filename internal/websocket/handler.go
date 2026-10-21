package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"mangahub/pkg/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for local development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleConnection handles a new WebSocket connection from Gin
func HandleConnection(c *gin.Context, hub *Hub, room string) {
	// Get user info from query params
	userID := c.Query("user_id")
	username := c.Query("username")

	if userID == "" {
		userID = "anonymous"
	}
	if username == "" {
		username = "guest"
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create chat client
	client := &models.ChatClient{
		UserID:   userID,
		Username: username,
		RoomID:   room,
	}

	log.Printf("New WebSocket connection: user=%s, room=%s", username, room)

	// Register with hub
	hub.Register <- client

	// Handle connection
	hub.HandleConnection(conn, client)
}
