package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"mangahub/pkg/models"

	"github.com/gin-gonic/gin"
)

const frontendUserID = "public"

type frontendManga struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Author        string   `json:"author,omitempty"`
	Genres        []string `json:"genres,omitempty"`
	Status        string   `json:"status,omitempty"`
	TotalChapters int      `json:"totalChapters"`
	Description   string   `json:"description,omitempty"`
	CoverURL      string   `json:"coverUrl,omitempty"`
}

type frontendProgress struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	CurrentChapter int    `json:"currentChapter"`
	CurrentPage    int    `json:"currentPage"`
	TotalChapters  int    `json:"totalChapters"`
	Progress       int    `json:"progress"`
}

type frontendProgressRequest struct {
	MangaID        string `json:"mangaId"`
	ID             string `json:"id"`
	CurrentChapter int    `json:"currentChapter"`
	CurrentPage    int    `json:"currentPage"`
}

func (h *Handler) registerFrontendRoutes(engine *gin.Engine) {
	frontend := engine.Group("/api")
	frontend.GET("/manga", h.listFrontendManga)
	frontend.GET("/manga/:id", h.getFrontendManga)
	frontend.GET("/library", h.getFrontendLibrary)
	frontend.GET("/reading-progress", h.getFrontendProgress)
	frontend.POST("/reading-progress", h.saveFrontendProgress)
}

func (h *Handler) listFrontendManga(c *gin.Context) {
	limit, offset := frontendPagination(c)
	items, err := h.mangaService.List(limit, offset)
	if err != nil {
		h.logger.Error("failed to list frontend manga: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list manga"})
		return
	}

	result := make([]frontendManga, 0, len(items))
	for _, item := range items {
		result = append(result, toFrontendManga(item))
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) getFrontendManga(c *gin.Context) {
	item, err := h.mangaService.GetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toFrontendManga(*item)})
}

func (h *Handler) getFrontendLibrary(c *gin.Context) {
	h.getFrontendProgress(c)
}

func (h *Handler) getFrontendProgress(c *gin.Context) {
	userID := frontendUser(c)
	rows, err := h.db.Query(`
		SELECT p.manga_id, m.title, p.current_chapter, p.current_page, m.chapters
		FROM frontend_reading_progress p
		LEFT JOIN manga m ON m.id = p.manga_id
		WHERE p.user_id = ?
		ORDER BY p.updated_at DESC`, userID)
	if err != nil {
		h.logger.Error("failed to get frontend progress: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get reading progress"})
		return
	}
	defer rows.Close()

	result := make([]frontendProgress, 0)
	for rows.Next() {
		var item frontendProgress
		var title sql.NullString
		var totalChapters sql.NullInt64
		if err := rows.Scan(&item.ID, &title, &item.CurrentChapter, &item.CurrentPage, &totalChapters); err != nil {
			h.logger.Error("failed to scan frontend progress: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read reading progress"})
			return
		}
		if title.Valid {
			item.Title = title.String
		}
		if totalChapters.Valid {
			item.TotalChapters = int(totalChapters.Int64)
		}
		item.Progress = calculateProgress(item.CurrentChapter, item.TotalChapters)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		h.logger.Error("failed to iterate frontend progress: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read reading progress"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) saveFrontendProgress(c *gin.Context) {
	var request frontendProgressRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	mangaID := request.MangaID
	if mangaID == "" {
		mangaID = request.ID
	}
	if mangaID == "" || request.CurrentChapter < 0 || request.CurrentPage < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mangaId, currentChapter and currentPage are required"})
		return
	}

	_, err := h.db.Exec(`
		INSERT INTO frontend_reading_progress (user_id, manga_id, current_chapter, current_page, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id, manga_id) DO UPDATE SET
			current_chapter = excluded.current_chapter,
			current_page = excluded.current_page,
			updated_at = excluded.updated_at`,
		frontendUser(c), mangaID, request.CurrentChapter, request.CurrentPage, time.Now())
	if err != nil {
		h.logger.Error("failed to save frontend progress for manga %s: %v", mangaID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save reading progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"id":             mangaID,
		"currentChapter": request.CurrentChapter,
		"currentPage":    request.CurrentPage,
	}})
}

func frontendPagination(c *gin.Context) (int, int) {
	limit, offset := 20, 0
	if value, err := strconv.Atoi(c.Query("limit")); err == nil && value > 0 && value <= 100 {
		limit = value
	}
	if value, err := strconv.Atoi(c.Query("offset")); err == nil && value >= 0 {
		offset = value
	}
	return limit, offset
}

func frontendUser(c *gin.Context) string {
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}
	if userID := c.Query("userId"); userID != "" {
		return userID
	}
	return frontendUserID
}

func toFrontendManga(item models.Manga) frontendManga {
	return frontendManga{
		ID: item.ID, Title: item.Title, Author: item.Author, Genres: item.Genres,
		Status: item.Status, TotalChapters: item.TotalChapters, Description: item.Description,
		CoverURL: item.CoverURL,
	}
}

func calculateProgress(currentChapter, totalChapters int) int {
	if totalChapters <= 0 {
		return 0
	}
	progress := currentChapter * 100 / totalChapters
	if progress > 100 {
		return 100
	}
	return progress
}
