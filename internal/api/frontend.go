package api

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mangahub/pkg/models"

	"github.com/gin-gonic/gin"
)

type frontendManga struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description,omitempty"`
	Genres        []string `json:"genres"`
	TotalChapters int      `json:"totalChapters"`
	CoverURL      string   `json:"coverUrl"`
	Rating        *float64 `json:"rating,omitempty"`
}

type frontendProgress struct {
	MangaID         string  `json:"mangaId"`
	CurrentChapter  int     `json:"currentChapter"`
	CurrentPage     int     `json:"currentPage"`
	ProgressPercent float64 `json:"progressPercent"`
	LastReadAt      string  `json:"lastReadAt"`
}

type frontendLibraryManga struct {
	frontendManga
	Status   string           `json:"status"`
	Progress frontendProgress `json:"progress"`
}

type frontendProgressRequest struct {
	MangaID         string  `json:"mangaId"`
	CurrentChapter  int     `json:"currentChapter"`
	CurrentPage     int     `json:"currentPage"`
	ProgressPercent float64 `json:"progressPercent"`
}

func (h *Handler) registerFrontendRoutes(engine *gin.Engine) {
	public := engine.Group("/api")
	public.GET("/manga", h.listFrontendManga)
	public.GET("/manga/featured", h.getFrontendFeatured)

	protected := engine.Group("/api")
	protected.Use(h.AuthMiddleware())
	protected.GET("/reading-progress/continue", h.getFrontendContinue)
	protected.GET("/reading-activity", h.getFrontendActivity)
	protected.GET("/library", h.getFrontendLibrary)
	protected.PUT("/reading-progress", h.saveFrontendProgress)
}

func (h *Handler) listFrontendManga(c *gin.Context) {
	page, limit, ok := frontendPage(c, 1, 20, 50)
	if !ok {
		return
	}
	search := strings.TrimSpace(c.Query("search"))
	genre := strings.TrimSpace(c.Query("genre"))
	status := strings.TrimSpace(c.Query("status"))
	sortBy := c.DefaultQuery("sort", "popular")
	if !validValue(status, "", "ongoing", "completed", "hiatus", "cancelled") ||
		!validValue(sortBy, "popular", "rating", "newest", "updated") {
		frontendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid manga filter.")
		return
	}

	where, args := mangaFilters(search, genre, status)
	orderBy := map[string]string{
		"popular": "COALESCE(rating, 0) DESC, updated_at DESC",
		"rating":  "COALESCE(rating, 0) DESC, title COLLATE NOCASE ASC",
		"newest":  "created_at DESC",
		"updated": "updated_at DESC",
	}[sortBy]

	var total int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM manga "+where, args...).Scan(&total); err != nil {
		h.logger.Error("failed to count frontend manga: %v", err)
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to search manga.")
		return
	}
	rows, err := h.db.Query(`SELECT id, title, description, genres, chapters, cover_url, rating
		FROM manga `+where+` ORDER BY `+orderBy+` LIMIT ? OFFSET ?`,
		append(args, limit, (page-1)*limit)...)
	if err != nil {
		h.logger.Error("failed to list frontend manga: %v", err)
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to search manga.")
		return
	}
	items, err := scanFrontendManga(rows)
	if err != nil {
		h.logger.Error("failed to scan frontend manga: %v", err)
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to read manga.")
		return
	}
	c.Header("Cache-Control", "public, max-age=60, s-maxage=300")
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "limit": limit, "total": total,
		"totalPages": int(math.Ceil(float64(total) / float64(limit)))})
}

func (h *Handler) getFrontendFeatured(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, title, description, genres, chapters, cover_url, rating
		FROM manga WHERE is_featured = 1 ORDER BY updated_at DESC LIMIT 1`)
	if err != nil {
		h.logger.Error("failed to get featured manga: %v", err)
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load featured manga.")
		return
	}
	items, err := scanFrontendManga(rows)
	if err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to read featured manga.")
		return
	}
	if len(items) == 0 {
		frontendError(c, http.StatusNotFound, "FEATURED_MANGA_NOT_FOUND", "No featured manga is configured.")
		return
	}
	c.Header("Cache-Control", "public, max-age=60, s-maxage=300")
	c.JSON(http.StatusOK, items[0])
}

func (h *Handler) getFrontendContinue(c *gin.Context) {
	userID := frontendAuthenticatedUser(c)
	rows, err := h.db.Query(frontendLibraryQuery+`
		WHERE p.user_id = ? AND COALESCE(p.status, 'reading') = 'reading'
		ORDER BY COALESCE(fp.updated_at, p.updated_at) DESC LIMIT 6`, userID)
	if err != nil {
		h.logger.Error("failed to get continue reading: %v", err)
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load continue reading.")
		return
	}
	items, err := scanFrontendLibrary(rows)
	if err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to read continue reading.")
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) getFrontendLibrary(c *gin.Context) {
	page, limit, ok := frontendPage(c, 1, 50, 100)
	if !ok {
		return
	}
	userID := frontendAuthenticatedUser(c)
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	if !validValue(status, "", "reading", "completed", "plan_to_read", "on_hold", "dropped") {
		frontendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid library status.")
		return
	}
	where := " WHERE p.user_id = ?"
	args := []interface{}{userID}
	if search != "" {
		where += " AND LOWER(m.title) LIKE LOWER(?)"
		args = append(args, "%"+search+"%")
	}
	if status != "" {
		where += " AND REPLACE(COALESCE(p.status, 'plan_to_read'), '-', '_') = ?"
		args = append(args, status)
	}
	var total int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM user_progress p JOIN manga m ON m.id = p.manga_id"+where, args...).Scan(&total); err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to count library.")
		return
	}
	rows, err := h.db.Query(frontendLibraryQuery+where+` ORDER BY COALESCE(fp.updated_at, p.updated_at) DESC LIMIT ? OFFSET ?`,
		append(args, limit, (page-1)*limit)...)
	if err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load library.")
		return
	}
	items, err := scanFrontendLibrary(rows)
	if err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to read library.")
		return
	}
	var stats struct {
		Titles, Reading, Completed, PlanToRead int
	}
	err = h.db.QueryRow(`SELECT COUNT(*),
		SUM(CASE WHEN REPLACE(COALESCE(status, 'plan_to_read'), '-', '_') = 'reading' THEN 1 ELSE 0 END),
		SUM(CASE WHEN REPLACE(COALESCE(status, 'plan_to_read'), '-', '_') = 'completed' THEN 1 ELSE 0 END),
		SUM(CASE WHEN REPLACE(COALESCE(status, 'plan_to_read'), '-', '_') = 'plan_to_read' THEN 1 ELSE 0 END)
		FROM user_progress WHERE user_id = ?`, userID).Scan(&stats.Titles, &stats.Reading, &stats.Completed, &stats.PlanToRead)
	if err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load library statistics.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "stats": gin.H{
		"titles": stats.Titles, "reading": stats.Reading, "completed": stats.Completed, "planToRead": stats.PlanToRead,
	}})
}

func (h *Handler) getFrontendActivity(c *gin.Context) {
	userID := frontendAuthenticatedUser(c)
	start := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -6)
	var activity struct {
		Chapters, Minutes, Completed int
	}
	if err := h.db.QueryRow(`SELECT COALESCE(SUM(chapter), 0), COALESCE(SUM(minutes), 0)
		FROM reading_activity WHERE user_id = ? AND created_at >= ?`, userID, start).Scan(&activity.Chapters, &activity.Minutes); err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load reading activity.")
		return
	}
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM user_progress WHERE user_id = ? AND status = 'completed'
		AND completed_at >= ?`, userID, time.Now().UTC().Format("2006-01")+"-01").Scan(&activity.Completed); err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load reading activity.")
		return
	}
	var streak int
	if err := h.db.QueryRow(`SELECT COUNT(DISTINCT SUBSTRING(CAST(created_at AS TEXT), 1, 10)) FROM reading_activity
		WHERE user_id = ? AND created_at >= ?`, userID, time.Now().UTC().AddDate(0, 0, -30)); err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load reading streak.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"streakDays": streak, "chaptersThisWeek": activity.Chapters,
		"readingTimeMinutes": activity.Minutes, "completedThisMonth": activity.Completed})
}

func (h *Handler) saveFrontendProgress(c *gin.Context) {
	var request frontendProgressRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.MangaID == "" ||
		request.CurrentChapter < 1 || request.CurrentPage < 1 || request.ProgressPercent < 0 || request.ProgressPercent > 100 {
		frontendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "The request contains invalid parameters.")
		return
	}
	var totalChapters int
	if err := h.db.QueryRow("SELECT chapters FROM manga WHERE id = ?", request.MangaID).Scan(&totalChapters); err == sql.ErrNoRows {
		frontendError(c, http.StatusNotFound, "MANGA_NOT_FOUND", "Manga does not exist.")
		return
	} else if err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to validate manga.")
		return
	}
	if request.CurrentChapter > totalChapters {
		frontendError(c, http.StatusUnprocessableEntity, "INVALID_READING_POSITION", "currentChapter must not exceed the manga chapter count.")
		return
	}
	now := time.Now().UTC()
	userID := frontendAuthenticatedUser(c)
	tx, err := h.db.BeginTx()
	if err != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to save reading progress.")
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO frontend_reading_progress
		(user_id, manga_id, current_chapter, current_page, progress_percent, updated_at) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, manga_id) DO UPDATE SET current_chapter=excluded.current_chapter,
		current_page=excluded.current_page, progress_percent=excluded.progress_percent, updated_at=excluded.updated_at`,
		userID, request.MangaID, request.CurrentChapter, request.CurrentPage, request.ProgressPercent, now)
	if err == nil {
		_, err = tx.Exec(`INSERT INTO user_progress (user_id, manga_id, current_chapter, status, updated_at)
			VALUES (?, ?, ?, 'reading', ?) ON CONFLICT(user_id, manga_id) DO UPDATE SET
			current_chapter=excluded.current_chapter, status='reading', updated_at=excluded.updated_at`,
			userID, request.MangaID, request.CurrentChapter, now)
	}
	if err == nil {
		_, err = tx.Exec(`INSERT INTO reading_activity (user_id, manga_id, chapter, minutes, created_at)
			VALUES (?, ?, ?, 0, ?)`, userID, request.MangaID, request.CurrentChapter, now)
	}
	if err != nil || tx.Commit() != nil {
		frontendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to save reading progress.")
		return
	}
	c.JSON(http.StatusOK, frontendProgress{MangaID: request.MangaID, CurrentChapter: request.CurrentChapter,
		CurrentPage: request.CurrentPage, ProgressPercent: request.ProgressPercent, LastReadAt: now.Format(time.RFC3339Nano)})
}

const frontendLibraryQuery = `SELECT m.id, m.title, m.description, m.genres, m.chapters, m.cover_url, m.rating,
		REPLACE(COALESCE(p.status, 'plan_to_read'), '-', '_'), COALESCE(fp.current_chapter, p.current_chapter, 0),
	COALESCE(fp.current_page, 1), COALESCE(fp.progress_percent, 0), COALESCE(fp.updated_at, p.updated_at)
		FROM user_progress p JOIN manga m ON m.id = p.manga_id
		LEFT JOIN frontend_reading_progress fp ON fp.user_id = p.user_id AND fp.manga_id = p.manga_id`

func mangaFilters(search, genre, status string) (string, []interface{}) {
	where := "WHERE 1=1"
	args := make([]interface{}, 0, 3)
	if search != "" {
		where += " AND LOWER(title) LIKE LOWER(?)"
		args = append(args, "%"+search+"%")
	}
	if genre != "" {
		where += " AND LOWER(genres) LIKE LOWER(?)"
		args = append(args, "%"+genre+"%")
	}
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}
	return where, args
}

func scanFrontendManga(rows *sql.Rows) ([]frontendManga, error) {
	defer rows.Close()
	items := make([]frontendManga, 0)
	for rows.Next() {
		var item frontendManga
		var description, genres, coverURL sql.NullString
		var rating sql.NullFloat64
		if err := rows.Scan(&item.ID, &item.Title, &description, &genres, &item.TotalChapters, &coverURL, &rating); err != nil {
			return nil, err
		}
		if description.Valid {
			item.Description = description.String
		}
		if coverURL.Valid {
			item.CoverURL = coverURL.String
		}
		if genres.Valid {
			item.Genres, _ = models.JSONToManga(genres.String)
		}
		if rating.Valid {
			item.Rating = &rating.Float64
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanFrontendLibrary(rows *sql.Rows) ([]frontendLibraryManga, error) {
	defer rows.Close()
	items := make([]frontendLibraryManga, 0)
	for rows.Next() {
		var item frontendLibraryManga
		var description, genres, coverURL, status string
		var rating sql.NullFloat64
		var updated time.Time
		if err := rows.Scan(&item.ID, &item.Title, &description, &genres, &item.TotalChapters, &coverURL, &rating,
			&status, &item.Progress.CurrentChapter, &item.Progress.CurrentPage, &item.Progress.ProgressPercent, &updated); err != nil {
			return nil, err
		}
		item.Description, item.CoverURL, item.Genres, item.Status = description, coverURL, nil, status
		if genres != "" {
			item.Genres, _ = models.JSONToManga(genres)
		}
		if rating.Valid {
			item.Rating = &rating.Float64
		}
		item.Progress.MangaID = item.ID
		if item.Progress.ProgressPercent == 0 && item.Progress.CurrentChapter > 0 {
			item.Progress.ProgressPercent = math.Min(100, math.Max(0, float64(item.Progress.CurrentChapter*100)/float64(max(1, item.TotalChapters))))
		}
		item.Progress.LastReadAt = updated.UTC().Format(time.RFC3339Nano)
		items = append(items, item)
	}
	return items, rows.Err()
}

func frontendPage(c *gin.Context, defaultPage, defaultLimit, maxLimit int) (int, int, bool) {
	page, limit := defaultPage, defaultLimit
	var err error
	if c.Query("page") != "" {
		page, err = strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			frontendError(c, 400, "VALIDATION_ERROR", "page must be greater than or equal to 1.")
			return 0, 0, false
		}
	}
	if c.Query("limit") != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil || limit < 1 || limit > maxLimit {
			frontendError(c, 400, "VALIDATION_ERROR", "limit is outside the allowed range.")
			return 0, 0, false
		}
	}
	return page, limit, true
}

func frontendAuthenticatedUser(c *gin.Context) string {
	userID, _ := c.Get("user_id")
	return fmt.Sprint(userID)
}

func frontendError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func validValue(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
