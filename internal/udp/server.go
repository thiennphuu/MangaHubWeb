package udp

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"

	"mangahub/pkg/models"
	"mangahub/pkg/utils"
)

// Server represents the UDP notification server
type Server struct {
	Port    string
	Clients map[string]*net.UDPAddr
	Queue   chan models.NotificationPayload
	mutex   sync.RWMutex
	done    chan bool
	logger  *utils.Logger
}

// NewServer creates a new UDP server
func NewServer(port string, logger *utils.Logger) *Server {
	return &Server{
		Port:    port,
		Clients: make(map[string]*net.UDPAddr),
		Queue:   make(chan models.NotificationPayload, 100),
		done:    make(chan bool),
		logger:  logger,
	}
}

// Start starts the UDP server
func (s *Server) Start() error {
	addrStr := s.Port
	if !strings.Contains(addrStr, ":") {
		addrStr = ":" + addrStr
	}

	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to start UDP server: %w", err)
	}
	defer conn.Close()

	s.logger.Info(fmt.Sprintf("UDP server started on port %s", s.Port))

	// Start broadcast handler
	go s.handleBroadcast(conn)

	// Handle incoming messages
	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error reading from UDP: %v", err))
			continue
		}

		clientID := remoteAddr.String()
		message := strings.TrimSpace(string(buffer[:n]))

		// Handle different message types
		switch message {
		case "register":
			// Register new client
			s.mutex.Lock()
			s.Clients[clientID] = remoteAddr
			s.mutex.Unlock()
			s.logger.Info(fmt.Sprintf("Client registered: %s", clientID))

			// Send confirmation back to client
			conn.WriteToUDP([]byte(`{"type":"registered","message":"Successfully registered for notifications"}`), remoteAddr)

		case "unregister":
			// Unregister client
			s.mutex.Lock()
			delete(s.Clients, clientID)
			s.mutex.Unlock()
			s.logger.Info(fmt.Sprintf("Client unregistered: %s", clientID))

			// Send confirmation back to client
			conn.WriteToUDP([]byte(`{"type":"unregistered","message":"Successfully unregistered from notifications"}`), remoteAddr)

		default:
			// Try to parse as notification payload
			var payload models.NotificationPayload
			if err := json.Unmarshal(buffer[:n], &payload); err == nil && payload.Type != "" {
				// Register client if not already registered (auto-registration)
				s.mutex.Lock()
				if _, exists := s.Clients[clientID]; !exists {
					s.Clients[clientID] = remoteAddr
					s.logger.Info(fmt.Sprintf("Client auto-registered: %s", clientID))
				}
				s.mutex.Unlock()

				// Broadcast the notification
				s.SendNotification(payload)
				s.logger.Info(fmt.Sprintf("Notification from %s: %s", clientID, payload.Type))
			} else {
				s.logger.Info(fmt.Sprintf("Unknown message from %s: %s", clientID, message))
			}
		}
	}
}

// handleBroadcast broadcasts notifications to all registered clients
func (s *Server) handleBroadcast(conn *net.UDPConn) {
	for notification := range s.Queue {
		data, err := json.Marshal(notification)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error marshaling notification: %v", err))
			continue
		}

		s.mutex.RLock()
		for _, addr := range s.Clients {
			_, err := conn.WriteToUDP(data, addr)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Error sending notification: %v", err))
			}
		}
		s.mutex.RUnlock()

		s.logger.Info(fmt.Sprintf("Broadcast notification: %s - %s", notification.Type, notification.Message))
	}
}

// SendNotification sends a notification
func (s *Server) SendNotification(notification models.NotificationPayload) {
	s.Queue <- notification
}

// RegisterClient registers a client
func (s *Server) RegisterClient(clientID string, addr *net.UDPAddr) {
	s.mutex.Lock()
	s.Clients[clientID] = addr
	s.mutex.Unlock()
	s.logger.Info(fmt.Sprintf("Client registered: %s at %s", clientID, addr.String()))
}

// UnregisterClient unregisters a client
func (s *Server) UnregisterClient(clientID string) {
	s.mutex.Lock()
	delete(s.Clients, clientID)
	s.mutex.Unlock()
	s.logger.Info(fmt.Sprintf("Client unregistered: %s", clientID))
}

// GetClientCount returns the number of registered clients
func (s *Server) GetClientCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.Clients)
}

// Stop stops the server
func (s *Server) Stop() {
	close(s.done)
}
