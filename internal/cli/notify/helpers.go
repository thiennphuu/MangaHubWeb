package notify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mangahub/internal/user"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
)

// Session stores the current user session
type Session struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
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

// getLibraryService creates and returns a library service
func getLibraryService() (*user.LibraryService, error) {
	dbPath := getDBPath()
	db, err := database.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return user.NewLibraryService(db), nil
}

// NotificationService handles notification operations
type NotificationService struct {
	db *database.Database
}

// getNotificationService creates and returns a notification service
func getNotificationService() (*NotificationService, error) {
	dbPath := getDBPath()
	db, err := database.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ensure notification tables exist
	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS notification_subscriptions (
			user_id TEXT NOT NULL,
			manga_id TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, manga_id)
		)
	`)
	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS notification_preferences (
			user_id TEXT PRIMARY KEY,
			chapter_releases BOOLEAN DEFAULT 1,
			email_notifications BOOLEAN DEFAULT 1,
			sound_enabled BOOLEAN DEFAULT 1,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	return &NotificationService{db: db}, nil
}

// Subscribe subscribes user to manga notifications
func (ns *NotificationService) Subscribe(userID, mangaID string) error {
	query := `
		INSERT OR REPLACE INTO notification_subscriptions (user_id, manga_id, created_at)
		VALUES (?, ?, ?)
	`
	_, err := ns.db.Exec(query, userID, mangaID, time.Now())
	return err
}

// Unsubscribe unsubscribes user from manga notifications
func (ns *NotificationService) Unsubscribe(userID, mangaID string) error {
	query := `DELETE FROM notification_subscriptions WHERE user_id = ? AND manga_id = ?`
	_, err := ns.db.Exec(query, userID, mangaID)
	return err
}

// UnsubscribeAll unsubscribes user from all notifications
func (ns *NotificationService) UnsubscribeAll(userID string) error {
	query := `DELETE FROM notification_subscriptions WHERE user_id = ?`
	_, err := ns.db.Exec(query, userID)
	return err
}

// GetSubscriptions gets all subscriptions for a user
func (ns *NotificationService) GetSubscriptions(userID string) ([]string, error) {
	query := `SELECT manga_id FROM notification_subscriptions WHERE user_id = ?`
	rows, err := ns.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []string
	for rows.Next() {
		var mangaID string
		if err := rows.Scan(&mangaID); err != nil {
			continue
		}
		subscriptions = append(subscriptions, mangaID)
	}
	return subscriptions, nil
}

// EnableNotifications enables general notifications for a user
func (ns *NotificationService) EnableNotifications(userID string) error {
	query := `
		INSERT OR REPLACE INTO notification_preferences 
		(user_id, chapter_releases, email_notifications, sound_enabled, updated_at)
		VALUES (?, 1, 1, 1, ?)
	`
	_, err := ns.db.Exec(query, userID, time.Now())
	return err
}

// DisableNotifications disables all notifications for a user
func (ns *NotificationService) DisableNotifications(userID string) error {
	query := `
		INSERT OR REPLACE INTO notification_preferences 
		(user_id, chapter_releases, email_notifications, sound_enabled, updated_at)
		VALUES (?, 0, 0, 0, ?)
	`
	_, err := ns.db.Exec(query, userID, time.Now())
	return err
}

// GetPreferences gets notification preferences for a user
func (ns *NotificationService) GetPreferences(userID string) (*models.NotificationPreferences, error) {
	query := `
		SELECT user_id, chapter_releases, email_notifications, sound_enabled 
		FROM notification_preferences WHERE user_id = ?
	`
	row := ns.db.QueryRow(query, userID)

	var prefs models.NotificationPreferences
	err := row.Scan(&prefs.UserID, &prefs.ChapterReleases, &prefs.EmailNotifications, &prefs.SoundEnabled)
	if err != nil {
		return nil, err
	}
	return &prefs, nil
}

// UpdatePreferences updates notification preferences
func (ns *NotificationService) UpdatePreferences(prefs *models.NotificationPreferences) error {
	query := `
		INSERT OR REPLACE INTO notification_preferences 
		(user_id, chapter_releases, email_notifications, sound_enabled, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := ns.db.Exec(query, prefs.UserID, prefs.ChapterReleases, prefs.EmailNotifications, prefs.SoundEnabled, time.Now())
	return err
}
