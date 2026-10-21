package models

import "time"

// Manga represents a manga series in the database
type Manga struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	Genres        []string  `json:"genres"`
	Status        string    `json:"status"` // "ongoing", "completed", "hiatus"
	TotalChapters int       `json:"total_chapters"`
	Description   string    `json:"description"`
	CoverURL      string    `json:"cover_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// MangaFilter represents search filters for manga
type MangaFilter struct {
	Query       string
	Genres      []string
	Status      string
	Author      string
	YearFrom    int
	YearTo      int
	MinChapters int
	SortBy      string // "popularity", "rating", "recent", "title"
	Order       string // "asc", "desc"
	Limit       int
	Offset      int
}

// SearchResult represents search results
type SearchResult struct {
	Total int     `json:"total"`
	Manga []Manga `json:"manga"`
	Page  int     `json:"page"`
	Limit int     `json:"limit"`
}
