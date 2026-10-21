package user

import (
	"database/sql"
	"fmt"
	"time"

	"mangahub/pkg/database"
	"mangahub/pkg/models"
)

// Service handles user operations
type Service struct {
	db *database.Database
}

// NewService creates a new user service
func NewService(db *database.Database) *Service {
	return &Service{db: db}
}

// Create creates a new user
func (s *Service) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	_, err := s.db.Exec(query, user.ID, user.Username, user.Email, user.PasswordHash, now, now)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (s *Service) GetByID(id string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = ?`

	var user models.User
	err := s.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (s *Service) GetByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = ?`

	var user models.User
	err := s.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (s *Service) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = ?`

	var user models.User
	err := s.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// Update updates a user
func (s *Service) Update(user *models.User) error {
	query := `
		UPDATE users
		SET username = ?, email = ?, password_hash = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query, user.Username, user.Email, user.PasswordHash, time.Now(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// UpdatePassword updates a user's password
func (s *Service) UpdatePassword(userID string, hashedPassword string) error {
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

// Delete deletes a user
func (s *Service) Delete(id string) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// LibraryService handles user library operations
type LibraryService struct {
	db *database.Database
}

// NewLibraryService creates a new library service
func NewLibraryService(db *database.Database) *LibraryService {
	return &LibraryService{db: db}
}

// AddToLibrary adds a manga to user's library
func (ls *LibraryService) AddToLibrary(userID, mangaID string, status string, rating int, notes string) error {
	query := `
		INSERT INTO user_progress (user_id, manga_id, status, rating, notes, current_chapter, started_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?)
	`
	now := time.Now()
	_, err := ls.db.Exec(query, userID, mangaID, status, rating, notes, now, now)
	if err != nil {
		return fmt.Errorf("failed to add to library: %w", err)
	}
	return nil
}

// GetLibraryEntry retrieves a single library entry
func (ls *LibraryService) GetLibraryEntry(userID, mangaID string) (*models.Progress, error) {
	query := `
		SELECT user_id, manga_id, current_chapter, status, rating, notes, started_at, completed_at, updated_at
		FROM user_progress WHERE user_id = ? AND manga_id = ?
	`
	row := ls.db.QueryRow(query, userID, mangaID)

	var progress models.Progress
	err := row.Scan(&progress.UserID, &progress.MangaID, &progress.CurrentChapter, &progress.Status,
		&progress.Rating, &progress.Notes, &progress.StartedAt, &progress.CompletedAt, &progress.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// RemoveFromLibrary removes a manga from user's library
func (ls *LibraryService) RemoveFromLibrary(userID, mangaID string) error {
	query := `DELETE FROM user_progress WHERE user_id = ? AND manga_id = ?`
	_, err := ls.db.Exec(query, userID, mangaID)
	if err != nil {
		return fmt.Errorf("failed to remove from library: %w", err)
	}
	return nil
}

// GetLibrary retrieves user's library
func (ls *LibraryService) GetLibrary(userID string, limit, offset int) ([]models.Progress, error) {
	query := `
		SELECT user_id, manga_id, current_chapter, status, rating, notes, started_at, completed_at, updated_at
		FROM user_progress WHERE user_id = ? LIMIT ? OFFSET ?
	`

	rows, err := ls.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get library: %w", err)
	}
	defer rows.Close()

	var progressList []models.Progress
	for rows.Next() {
		var progress models.Progress
		err := rows.Scan(&progress.UserID, &progress.MangaID, &progress.CurrentChapter, &progress.Status,
			&progress.Rating, &progress.Notes, &progress.StartedAt, &progress.CompletedAt, &progress.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan progress: %w", err)
		}
		progressList = append(progressList, progress)
	}

	return progressList, nil
}

// GetLibraryByStatus retrieves user's library filtered by status
func (ls *LibraryService) GetLibraryByStatus(userID, status string, limit, offset int) ([]models.Progress, error) {
	query := `
		SELECT user_id, manga_id, current_chapter, status, rating, notes, started_at, completed_at, updated_at
		FROM user_progress WHERE user_id = ? AND status = ? LIMIT ? OFFSET ?
	`

	rows, err := ls.db.Query(query, userID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get library: %w", err)
	}
	defer rows.Close()

	var progressList []models.Progress
	for rows.Next() {
		var progress models.Progress
		err := rows.Scan(&progress.UserID, &progress.MangaID, &progress.CurrentChapter, &progress.Status,
			&progress.Rating, &progress.Notes, &progress.StartedAt, &progress.CompletedAt, &progress.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan progress: %w", err)
		}
		progressList = append(progressList, progress)
	}

	return progressList, nil
}

// UpdateLibraryEntry updates a library entry
func (ls *LibraryService) UpdateLibraryEntry(progress *models.Progress) error {
	query := `
		UPDATE user_progress
		SET current_chapter = ?, status = ?, rating = ?, notes = ?, completed_at = ?, updated_at = ?
		WHERE user_id = ? AND manga_id = ?
	`
	_, err := ls.db.Exec(query, progress.CurrentChapter, progress.Status, progress.Rating, progress.Notes,
		progress.CompletedAt, time.Now(), progress.UserID, progress.MangaID)
	if err != nil {
		return fmt.Errorf("failed to update library entry: %w", err)
	}
	return nil
}
