package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mangahub/internal/auth"
	"mangahub/internal/user"
	"mangahub/pkg/database"
)

// Session stores the current user session
type Session struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// getAPIURL returns the API server URL
func getAPIURL() string {
	if url := os.Getenv("MANGAHUB_API_URL"); url != "" {
		return url
	}
	return "http://10.238.53.72:8080" // Server IP
}

// getDBPath returns the path to the database file
func getDBPath() string {
	paths := []string{
		"data/mangahub.db",
		"./data/mangahub.db",
		filepath.Join(os.Getenv("HOME"), ".mangahub", "mangahub.db"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "data/mangahub.db"
}

// getSessionPath returns the path to the session file
func getSessionPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".mangahub_session"
	}
	return filepath.Join(homeDir, ".mangahub", "session.json")
}

// saveSession saves the current session to file
func saveSession(session *Session) error {
	sessionPath := getSessionPath()
	if err := os.MkdirAll(filepath.Dir(sessionPath), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionPath, data, 0600)
}

// loadSession loads the current session from file
func loadSession() (*Session, error) {
	sessionPath := getSessionPath()
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// clearSession removes the session file
func clearSession() error {
	sessionPath := getSessionPath()
	return os.Remove(sessionPath)
}

// getUserService creates and returns a user service connected to the database
func getUserService() (*user.Service, error) {
	dbPath := getDBPath()
	db, err := database.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return user.NewService(db), nil
}

// getAuthService creates and returns an auth service
func getAuthService() *auth.AuthService {
	return auth.NewAuthService("")
}
