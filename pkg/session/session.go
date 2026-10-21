package session

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// CurrentProfile holds the active profile name
var CurrentProfile = "default"

// Session stores the current user session
type Session struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// SetProfile sets the current profile name
func SetProfile(profile string) {
	if profile != "" {
		CurrentProfile = profile
	}
}

// GetProfile returns the current profile name
func GetProfile() string {
	return CurrentProfile
}

// GetPath returns the path to the session file for the current profile
func GetPath() string {
	return GetPathForProfile(CurrentProfile)
}

// GetPathForProfile returns the path to the session file for a specific profile
func GetPathForProfile(profile string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".mangahub_session_" + profile + ".json"
	}

	if profile == "" || profile == "default" {
		return filepath.Join(homeDir, ".mangahub", "session.json")
	}
	return filepath.Join(homeDir, ".mangahub", "session_"+profile+".json")
}

// Save saves the current session to file
func Save(session *Session) error {
	sessionPath := GetPath()
	if err := os.MkdirAll(filepath.Dir(sessionPath), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionPath, data, 0600)
}

// Load loads the current session from file
func Load() (*Session, error) {
	sessionPath := GetPath()
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, err
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// Clear removes the session file
func Clear() error {
	sessionPath := GetPath()
	return os.Remove(sessionPath)
}

// ListProfiles returns all available profile names
func ListProfiles() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	mangahubDir := filepath.Join(homeDir, ".mangahub")
	entries, err := os.ReadDir(mangahubDir)
	if err != nil {
		return nil, err
	}

	profiles := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if name == "session.json" {
			profiles = append(profiles, "default")
		} else if len(name) > 13 && name[:8] == "session_" && name[len(name)-5:] == ".json" {
			// Extract profile name from session_<profile>.json
			profileName := name[8 : len(name)-5]
			profiles = append(profiles, profileName)
		}
	}

	return profiles, nil
}
