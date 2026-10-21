package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"mangahub/pkg/models"
	"mangahub/pkg/utils"
)

// TCPClient is a thin wrapper around a TCP connection to the sync server.
// It is intentionally stateless so each CLI command can create its own client.
type TCPClient struct {
	Addr string
}

// NewTCPClient creates a new TCP client pointing to the given host/port.
func NewTCPClient(host string, port int) *TCPClient {
	return &TCPClient{
		Addr: fmt.Sprintf("%s:%d", host, port),
	}
}

// Connect dials the TCP sync server and returns a live connection.
func (c *TCPClient) Connect() (net.Conn, error) {
	dialer := net.Dialer{
		Timeout: 5 * time.Second,
	}
	return dialer.Dial("tcp", c.Addr)
}

// CheckStatus tries to establish a short connection to verify that the
// TCP sync server is reachable.
func (c *TCPClient) CheckStatus() error {
	conn, err := c.Connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

// SendUpdate sends a single progress update to the TCP sync server.
// The server will broadcast this update to all connected clients.
func (c *TCPClient) SendUpdate(update *models.ProgressUpdate) error {
	conn, err := c.Connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(conn, "%s\n", data)
	return err
}

// MonitorUpdates connects to the TCP server and continuously reads
// JSON-encoded ProgressUpdate messages until the connection is closed
// or the stop channel is triggered.
func (c *TCPClient) MonitorUpdates(stop <-chan struct{}, handler func(models.ProgressUpdate)) error {
	conn, err := c.Connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Handle external stop signal
	go func() {
		<-stop
		if err := conn.Close(); err != nil {
			utils.NewLogger().Error("failed to close TCP monitor connection: %v", err)
		}
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()

		var update models.ProgressUpdate
		if err := json.Unmarshal(line, &update); err != nil {
			utils.NewLogger().Error("error parsing sync update: %v", err)
			continue
		}

		if handler != nil {
			handler(update)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
