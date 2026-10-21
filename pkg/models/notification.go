package models

import "time"

// Notification represents a notification
type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"` // "chapter_release", "friend_activity", "message"
	MangaID   string                 `json:"manga_id,omitempty"`
	Message   string                 `json:"message"`
	Read      bool                   `json:"read"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// NotificationPayload represents notification to be sent
type NotificationPayload struct {
	Type      string `json:"type"`
	MangaID   string `json:"manga_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// NotificationPreferences represents user notification settings
type NotificationPreferences struct {
	UserID             string `json:"user_id"`
	ChapterReleases    bool   `json:"chapter_releases"`
	FriendActivity     bool   `json:"friend_activity"`
	Messages           bool   `json:"messages"`
	SoundEnabled       bool   `json:"sound_enabled"`
	EmailNotifications bool   `json:"email_notifications"`
	QuietHours         string `json:"quiet_hours"`
}
