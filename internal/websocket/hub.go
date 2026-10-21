package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"mangahub/pkg/models"
)

// Hub represents a WebSocket chat hub
type Hub struct {
	Clients    map[*websocket.Conn]*models.ChatClient
	Broadcast  chan models.ChatMessage
	Register   chan *models.ChatClient
	Unregister chan *websocket.Conn
	mutex      sync.RWMutex
	done       chan bool
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*websocket.Conn]*models.ChatClient),
		Broadcast:  make(chan models.ChatMessage, 100),
		Register:   make(chan *models.ChatClient),
		Unregister: make(chan *websocket.Conn),
		done:       make(chan bool),
	}
}

// Run starts the hub event loop
func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			return
		case client := <-h.Register:
			h.mutex.Lock()
			// Find connection for this client
			for conn, existingClient := range h.Clients {
				if existingClient == nil {
					h.Clients[conn] = client
					break
				}
			}
			h.mutex.Unlock()
			log.Printf("Client registered: %s (%s)", client.UserID, client.Username)

		case conn := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.Clients[conn]; ok {
				delete(h.Clients, conn)
				conn.Close()
			}
			h.mutex.Unlock()
			log.Printf("Client unregistered")

		case message := <-h.Broadcast:
			h.mutex.RLock()
			for conn, client := range h.Clients {
				if client == nil {
					continue
				}
				// Only send to clients in the same room
				if client.RoomID == message.RoomID {
					err := conn.WriteJSON(message)
					if err != nil {
						log.Printf("Error writing message: %v", err)
						go func(c *websocket.Conn) {
							h.Unregister <- c
						}(conn)
					}
				}
			}
			h.mutex.RUnlock()
			log.Printf("Broadcast message from %s: %s", message.Username, message.Message)
		}
	}
}

// HandleConnection handles a new WebSocket connection
func (h *Hub) HandleConnection(conn *websocket.Conn, client *models.ChatClient) {
	h.mutex.Lock()
	h.Clients[conn] = client
	h.mutex.Unlock()

	defer func() {
		h.Unregister <- conn
	}()

	for {
		var msg models.ChatMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		msg.UserID = client.UserID
		msg.Username = client.Username
		msg.RoomID = client.RoomID
		msg.Timestamp = time.Now().Unix()

		h.Broadcast <- msg
	}
}

// SendMessage sends a message to a room
func (h *Hub) SendMessage(message models.ChatMessage) {
	h.Broadcast <- message
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.Clients)
}

// GetRoomClientCount returns the number of clients in a room
func (h *Hub) GetRoomClientCount(roomID string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	count := 0
	for _, client := range h.Clients {
		if client != nil && client.RoomID == roomID {
			count++
		}
	}
	return count
}

// Stop stops the hub
func (h *Hub) Stop() {
	h.mutex.Lock()
	for conn := range h.Clients {
		conn.Close()
	}
	h.Clients = make(map[*websocket.Conn]*models.ChatClient)
	h.mutex.Unlock()
	close(h.done)
}
