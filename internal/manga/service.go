package manga

import (
	"database/sql"
	"fmt"
	"time"

	"mangahub/pkg/database"
	"mangahub/pkg/models"
)

// Service handles manga operations
type Service struct {
	db *database.Database
}

// NewService creates a new manga service
func NewService(db *database.Database) *Service {
	return &Service{db: db}
}

// Create creates a new manga entry
func (s *Service) Create(manga *models.Manga) error {
	genres, _ := models.MangaToJSON(manga.Genres)
	query := `
		INSERT INTO manga (id, title, author, genres, status, chapters, description, cover_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	_, err := s.db.Exec(query, manga.ID, manga.Title, manga.Author, genres, manga.Status, manga.TotalChapters, manga.Description, manga.CoverURL, now, now)
	if err != nil {
		return fmt.Errorf("failed to create manga: %w", err)
	}
	return nil
}

// GetByID retrieves a manga by ID
func (s *Service) GetByID(id string) (*models.Manga, error) {
	query := `SELECT id, title, author, genres, status, chapters, description, cover_url, created_at, updated_at FROM manga WHERE id = ?`

	var manga models.Manga
	var genresJSON, description, coverURL sql.NullString
	err := s.db.QueryRow(query, id).Scan(
		&manga.ID, &manga.Title, &manga.Author, &genresJSON, &manga.Status,
		&manga.TotalChapters, &description, &coverURL,
		&manga.CreatedAt, &manga.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("manga not found")
		}
		return nil, fmt.Errorf("failed to get manga: %w", err)
	}

	if genresJSON.Valid {
		manga.Genres, _ = models.JSONToManga(genresJSON.String)
	}
	if description.Valid {
		manga.Description = description.String
	}
	if coverURL.Valid {
		manga.CoverURL = coverURL.String
	}
	return &manga, nil
}

// Search searches for manga
func (s *Service) Search(filter *models.MangaFilter) (*models.SearchResult, error) {
	query := "SELECT id, title, author, genres, status, chapters, description, cover_url, created_at, updated_at FROM manga WHERE 1=1"
	var args []interface{}

	if filter.Query != "" {
		query += " AND (title LIKE ? OR author LIKE ?)"
		searchTerm := "%" + filter.Query + "%"
		args = append(args, searchTerm, searchTerm)
	}

	if len(filter.Genres) > 0 {
		for _, genre := range filter.Genres {
			query += " AND genres LIKE ?"
			args = append(args, "%"+genre+"%")
		}
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.MinChapters > 0 {
		query += " AND chapters >= ?"
		args = append(args, filter.MinChapters)
	}

	if filter.SortBy == "" {
		filter.SortBy = "title"
	}
	// Map sortBy to actual column names
	sortColumn := filter.SortBy
	if sortColumn == "total_chapters" {
		sortColumn = "chapters"
	}
	if filter.Order == "" {
		filter.Order = "asc"
	}

	query += " ORDER BY " + sortColumn + " " + filter.Order

	if filter.Limit == 0 {
		filter.Limit = 10
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search manga: %w", err)
	}
	defer rows.Close()

	var mangaList []models.Manga
	for rows.Next() {
		var manga models.Manga
		var genresJSON, description, coverURL sql.NullString
		err := rows.Scan(
			&manga.ID, &manga.Title, &manga.Author, &genresJSON, &manga.Status,
			&manga.TotalChapters, &description, &coverURL,
			&manga.CreatedAt, &manga.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan manga: %w", err)
		}
		if genresJSON.Valid {
			manga.Genres, _ = models.JSONToManga(genresJSON.String)
		}
		if description.Valid {
			manga.Description = description.String
		}
		if coverURL.Valid {
			manga.CoverURL = coverURL.String
		}
		mangaList = append(mangaList, manga)
	}

	result := &models.SearchResult{
		Total: len(mangaList),
		Manga: mangaList,
		Page:  (filter.Offset / filter.Limit) + 1,
		Limit: filter.Limit,
	}

	return result, nil
}

// List lists all manga
func (s *Service) List(limit, offset int) ([]models.Manga, error) {
	query := "SELECT id, title, author, genres, status, chapters, description, cover_url, created_at, updated_at FROM manga LIMIT ? OFFSET ?"

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list manga: %w", err)
	}
	defer rows.Close()

	var mangaList []models.Manga
	for rows.Next() {
		var manga models.Manga
		var genresJSON, description, coverURL sql.NullString
		err := rows.Scan(
			&manga.ID, &manga.Title, &manga.Author, &genresJSON, &manga.Status,
			&manga.TotalChapters, &description, &coverURL,
			&manga.CreatedAt, &manga.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan manga: %w", err)
		}
		if genresJSON.Valid {
			manga.Genres, _ = models.JSONToManga(genresJSON.String)
		}
		if description.Valid {
			manga.Description = description.String
		}
		if coverURL.Valid {
			manga.CoverURL = coverURL.String
		}
		mangaList = append(mangaList, manga)
	}

	return mangaList, nil
}

// Update updates a manga entry
func (s *Service) Update(manga *models.Manga) error {
	genres, _ := models.MangaToJSON(manga.Genres)
	query := `
		UPDATE manga
		SET title = ?, author = ?, genres = ?, status = ?, chapters = ?, description = ?, cover_url = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query, manga.Title, manga.Author, genres, manga.Status, manga.TotalChapters, manga.Description, manga.CoverURL, time.Now(), manga.ID)
	if err != nil {
		return fmt.Errorf("failed to update manga: %w", err)
	}
	return nil
}

// Delete deletes a manga entry
func (s *Service) Delete(id string) error {
	query := "DELETE FROM manga WHERE id = ?"
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete manga: %w", err)
	}
	return nil
}
