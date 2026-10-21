package client

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"mangahub/pkg/models"
)

// UDPClient represents a UDP client for notifications
type UDPClient struct {
	ServerAddr string
	conn       *net.UDPConn
	localAddr  *net.UDPAddr
	Done       chan bool
}

// NewUDPClient creates a new UDP client
func NewUDPClient(serverAddr string) *UDPClient {
	return &UDPClient{
		ServerAddr: serverAddr,
		Done:       make(chan bool),
	}
}

// Connect connects to the UDP server
func (c *UDPClient) Connect() error {
	serverUDPAddr, err := net.ResolveUDPAddr("udp", c.ServerAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve server address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, serverUDPAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to UDP server: %w", err)
	}

	c.conn = conn
	c.localAddr = conn.LocalAddr().(*net.UDPAddr)

	return nil
}

// Register sends a registration message to the server
func (c *UDPClient) Register() error {
	if c.conn == nil {
		return fmt.Errorf("not connected to server")
	}

	_, err := c.conn.Write([]byte("register"))
	if err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}

	return nil
}

// Unregister sends an unregister message to the server
func (c *UDPClient) Unregister() error {
	if c.conn == nil {
		return fmt.Errorf("not connected to server")
	}

	_, err := c.conn.Write([]byte("unregister"))
	if err != nil {
		return fmt.Errorf("failed to send unregistration: %w", err)
	}

	return nil
}

// SendNotification sends a notification payload to the server
func (c *UDPClient) SendNotification(payload models.NotificationPayload) error {
	if c.conn == nil {
		return fmt.Errorf("not connected to server")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	_, err = c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// Listen listens for notifications from the server
func (c *UDPClient) Listen(callback func(models.NotificationPayload)) error {
	if c.conn == nil {
		return fmt.Errorf("not connected to server")
	}

	buffer := make([]byte, 4096)

	for {
		select {
		case <-c.Done:
			return nil
		default:
			c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, err := c.conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				continue
			}

			var payload models.NotificationPayload
			if err := json.Unmarshal(buffer[:n], &payload); err == nil {
				callback(payload)
			}
		}
	}
}

// Close closes the UDP connection
func (c *UDPClient) Close() error {
	if c.conn != nil {
		close(c.Done)
		return c.conn.Close()
	}
	return nil
}

// GetLocalAddr returns the local address of the client
func (c *UDPClient) GetLocalAddr() string {
	if c.localAddr != nil {
		return c.localAddr.String()
	}
	return ""
}
