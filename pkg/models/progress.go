package models

import "time"

// Progress represents user's reading progress
type Progress struct {
	UserID         string     `json:"user_id"`
	MangaID        string     `json:"manga_id"`
	CurrentChapter int        `json:"current_chapter"`
	Status         string     `json:"status"` // "reading", "completed", "on-hold", "dropped", "plan-to-read"
	Rating         int        `json:"rating"` // 0-10
	Notes          string     `json:"notes"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ProgressUpdate represents a progress update for sync
type ProgressUpdate struct {
	UserID    string `json:"user_id"`
	MangaID   string `json:"manga_id"`
	Chapter   int    `json:"chapter"`
	Timestamp int64  `json:"timestamp"`
	DeviceID  string `json:"device_id"`
}

// ProgressStats represents user reading statistics
type ProgressStats struct {
	UserID            string    `json:"user_id"`
	TotalMangaRead    int       `json:"total_manga_read"`
	TotalChaptersRead int       `json:"total_chapters_read"`
	TotalCompleted    int       `json:"total_completed"`
	TotalReading      int       `json:"total_reading"`
	FavoriteGenres    []string  `json:"favorite_genres"`
	AverageRating     float64   `json:"average_rating"`
	ReadingStreak     int       `json:"reading_streak"`
	LastReadDate      time.Time `json:"last_read_date"`
}
