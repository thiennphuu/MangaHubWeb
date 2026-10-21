package api

import (
	"fmt"
	"net/http"
	"strconv"

	"mangahub/internal/auth"
	"mangahub/internal/manga"
	"mangahub/internal/user"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"mangahub/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Handler holds API handlers
type Handler struct {
	db             *database.Database
	authService    *auth.AuthService
	userService    *user.Service
	libraryService *user.LibraryService
	mangaService   *manga.Service
	logger         *utils.Logger
}

// NewHandler creates a new API handler
func NewHandler(db *database.Database, logger *utils.Logger) *Handler {
	return &Handler{
		db:             db,
		authService:    auth.NewAuthService("your-secret-key"),
		userService:    user.NewService(db),
		libraryService: user.NewLibraryService(db),
		mangaService:   manga.NewService(db),
		logger:         logger,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(engine *gin.Engine) {
	// Auth routes
	auth := engine.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}

	// Public manga routes
	mangaGroup := engine.Group("/manga")
	{
		mangaGroup.GET("", h.ListManga)
		mangaGroup.GET("/:id", h.GetManga)
		mangaGroup.POST("/search", h.SearchManga)
	}

	// Protected routes
	protected := engine.Group("")
	protected.Use(h.AuthMiddleware())
	{
		// User routes
		user := protected.Group("/users")
		{
			user.GET("/profile", h.GetProfile)
			user.PUT("/profile", h.UpdateProfile)
		}

		// Library routes
		library := protected.Group("/users/library")
		{
			library.GET("", h.GetLibrary)
			library.POST("", h.AddToLibrary)
			library.DELETE("/:mangaId", h.RemoveFromLibrary)
			library.PUT("/:mangaId/progress", h.UpdateProgress)
		}

		// Server management routes
		server := protected.Group("/server")
		{
			server.GET("/logs", h.GetServerLogs)
			server.GET("/database/check", h.GetDatabaseCheck)
			server.POST("/database/optimize", h.OptimizeDatabase)
			server.GET("/database/stats", h.GetDatabaseStats)
			server.POST("/database/repair", h.RepairDatabase)
		}

		// Admin routes (placeholder)
		admin := protected.Group("/admin")
		admin.Use(h.AdminMiddleware())
		{
			admin.POST("/manga", h.CreateManga)
			admin.PUT("/manga/:id", h.UpdateManga)
			admin.DELETE("/manga/:id", h.DeleteManga)
		}
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Validate input
	if err := utils.ValidateUsername(req.Username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := utils.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	if _, err := h.userService.GetByUsername(req.Username); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Create user
	user := &models.User{
		ID:           h.authService.GenerateUserID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := h.userService.Create(user); err != nil {
		h.logger.Error("failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully", "user_id": user.ID})
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get user
	user, err := h.userService.GetByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Verify password
	if err := h.authService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate token
	token, expiresAt, err := h.authService.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		UserID:    user.ID,
		Username:  user.Username,
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

// GetProfile retrieves user profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.userService.GetByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates user profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.User
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.ID = userID.(string)
	if err := h.userService.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}

// ListManga lists all manga
func (h *Handler) ListManga(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	mangaList, err := h.mangaService.List(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list manga"})
		return
	}

	c.JSON(http.StatusOK, mangaList)
}

// GetManga retrieves a manga by ID
func (h *Handler) GetManga(c *gin.Context) {
	id := c.Param("id")

	manga, err := h.mangaService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
		return
	}

	c.JSON(http.StatusOK, manga)
}

// SearchManga searches for manga
func (h *Handler) SearchManga(c *gin.Context) {
	var filter models.MangaFilter
	if err := c.BindJSON(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	results, err := h.mangaService.Search(&filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search manga"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// CreateManga creates a new manga (admin)
func (h *Handler) CreateManga(c *gin.Context) {
	var manga models.Manga
	if err := c.BindJSON(&manga); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.mangaService.Create(&manga); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create manga"})
		return
	}

	c.JSON(http.StatusCreated, manga)
}

// UpdateManga updates a manga (admin)
func (h *Handler) UpdateManga(c *gin.Context) {
	id := c.Param("id")

	var manga models.Manga
	if err := c.BindJSON(&manga); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	manga.ID = id
	if err := h.mangaService.Update(&manga); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update manga"})
		return
	}

	c.JSON(http.StatusOK, manga)
}

// DeleteManga deletes a manga (admin)
func (h *Handler) DeleteManga(c *gin.Context) {
	id := c.Param("id")

	if err := h.mangaService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete manga"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "manga deleted successfully"})
}

// GetLibrary retrieves user's library
func (h *Handler) GetLibrary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	var progressList []models.Progress
	var err error

	if status := c.Query("status"); status != "" {
		progressList, err = h.libraryService.GetLibraryByStatus(userID.(string), status, limit, offset)
	} else {
		progressList, err = h.libraryService.GetLibrary(userID.(string), limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get library"})
		return
	}

	c.JSON(http.StatusOK, progressList)
}

// AddToLibrary adds manga to user's library
func (h *Handler) AddToLibrary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		MangaID string `json:"manga_id"`
		Status  string `json:"status"`
		Rating  int    `json:"rating"`
		Notes   string `json:"notes"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.libraryService.AddToLibrary(userID.(string), req.MangaID, req.Status, req.Rating, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add to library"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "manga added to library"})
}

// RemoveFromLibrary removes manga from user's library
func (h *Handler) RemoveFromLibrary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	mangaID := c.Param("mangaId")

	if err := h.libraryService.RemoveFromLibrary(userID.(string), mangaID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove from library"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "manga removed from library"})
}

// UpdateProgress updates reading progress
func (h *Handler) UpdateProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	mangaID := c.Param("mangaId")

	var req models.Progress
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.UserID = userID.(string)
	req.MangaID = mangaID

	if err := h.libraryService.UpdateLibraryEntry(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "progress updated successfully"})
}

// AuthMiddleware checks JWT token
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if len(token) > 7 {
			token = token[7:]
		}

		claims, err := h.authService.VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// AdminMiddleware checks admin privileges (placeholder)
func (h *Handler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement admin role checking
		c.Next()
	}
}

// GetServerLogs returns recent server logs
func (h *Handler) GetServerLogs(c *gin.Context) {
	// Parse query parameters
	maxLines := 100
	if maxLinesStr := c.Query("max_lines"); maxLinesStr != "" {
		if parsed, err := strconv.Atoi(maxLinesStr); err == nil && parsed > 0 {
			maxLines = parsed
		}
	}

	level := c.Query("level") // filter by level (debug, info, warn, error)

	// Get log file path
	logPath, err := h.getLogFilePath()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to determine log path"})
		return
	}

	// Read logs
	logs, err := h.readLogFile(logPath, level, maxLines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to read logs: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"count":     len(logs),
		"max_lines": maxLines,
		"level":     level,
	})
}

func (h *Handler) getLogFilePath() (string, error) {
	// In production, this would read from config
	// For now, use the standard log path
	return utils.GetLogFilePath()
}

func (h *Handler) readLogFile(logPath, level string, maxLines int) ([]string, error) {
	file, err := utils.OpenLogFile(logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return utils.ReadLogLines(file, level, maxLines)
}

// GetDatabaseCheck performs database integrity checks
func (h *Handler) GetDatabaseCheck(c *gin.Context) {
	// Run integrity check
	integrityOK, integrityIssues, err := h.checkDatabaseIntegrity()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("integrity check failed: %v", err)})
		return
	}

	// Verify core tables
	tables, missingTables, err := h.verifyDatabaseTables()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("table verification failed: %v", err)})
		return
	}

	// Determine overall status
	status := "healthy"
	if !integrityOK || len(missingTables) > 0 {
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"integrity": gin.H{
			"ok":     integrityOK,
			"issues": integrityIssues,
		},
		"tables": gin.H{
			"verified": tables,
			"missing":  missingTables,
		},
	})
}

func (h *Handler) checkDatabaseIntegrity() (bool, []string, error) {
	rows, err := h.db.DB.Query(`PRAGMA integrity_check;`)
	if err != nil {
		return false, nil, err
	}
	defer rows.Close()

	var issues []string
	ok := true
	for rows.Next() {
		var res string
		if err := rows.Scan(&res); err != nil {
			return false, nil, err
		}
		if res != "ok" {
			ok = false
			issues = append(issues, res)
		}
	}
	if err := rows.Err(); err != nil {
		return false, nil, err
	}

	return ok, issues, nil
}

func (h *Handler) verifyDatabaseTables() ([]string, []string, error) {
	required := []string{"users", "manga", "user_progress"}
	var verified []string
	var missing []string

	for _, tbl := range required {
		var name string
		err := h.db.DB.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, tbl).Scan(&name)
		if err != nil {
			missing = append(missing, tbl)
		} else {
			verified = append(verified, tbl)
		}
	}

	return verified, missing, nil
}

// OptimizeDatabase runs database optimization commands
func (h *Handler) OptimizeDatabase(c *gin.Context) {
	var steps []string
	var errors []string

	// Run ANALYZE
	if _, err := h.db.Exec("ANALYZE;"); err != nil {
		errors = append(errors, fmt.Sprintf("ANALYZE failed: %v", err))
	} else {
		steps = append(steps, "ANALYZE completed")
	}

	// Run REINDEX
	if _, err := h.db.Exec("REINDEX;"); err != nil {
		errors = append(errors, fmt.Sprintf("REINDEX failed: %v", err))
	} else {
		steps = append(steps, "REINDEX completed")
	}

	// Run VACUUM
	if _, err := h.db.Exec("VACUUM;"); err != nil {
		errors = append(errors, fmt.Sprintf("VACUUM failed: %v", err))
	} else {
		steps = append(steps, "VACUUM completed")
	}

	status := "success"
	if len(errors) > 0 {
		status = "partial"
		if len(errors) == 3 {
			status = "failed"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"steps":  steps,
		"errors": errors,
	})
}

// GetDatabaseStats returns database statistics
func (h *Handler) GetDatabaseStats(c *gin.Context) {
	// Get table counts
	tables := map[string]int{}
	tableNames := []string{"users", "manga", "user_progress", "chat_messages", "notifications"}

	for _, name := range tableNames {
		var count int
		err := h.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", name)).Scan(&count)
		if err != nil {
			tables[name] = -1 // Indicate error
		} else {
			tables[name] = count
		}
	}

	// Get database file size using PRAGMA
	var fileSize int64
	var pageCount int
	var pageSize int

	if err := h.db.DB.QueryRow("PRAGMA page_count").Scan(&pageCount); err == nil {
		if err := h.db.DB.QueryRow("PRAGMA page_size").Scan(&pageSize); err == nil {
			fileSize = int64(pageCount * pageSize)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"file_size_bytes": fileSize,
		"file_size_mb":    float64(fileSize) / 1024.0 / 1024.0,
		"tables":          tables,
	})
}

// RepairDatabase performs database repair operations
func (h *Handler) RepairDatabase(c *gin.Context) {
	var steps []string
	var errors []string

	// Run quick_check
	quickCheckOK := true
	rows, err := h.db.DB.Query(`PRAGMA quick_check;`)
	if err != nil {
		errors = append(errors, fmt.Sprintf("quick_check failed: %v", err))
		quickCheckOK = false
	} else {
		defer rows.Close()
		for rows.Next() {
			var res string
			if err := rows.Scan(&res); err != nil {
				errors = append(errors, fmt.Sprintf("quick_check scan error: %v", err))
				quickCheckOK = false
				break
			}
			if res != "ok" {
				errors = append(errors, fmt.Sprintf("quick_check issue: %s", res))
				quickCheckOK = false
			}
		}
		if quickCheckOK {
			steps = append(steps, "PRAGMA quick_check: ok")
		}
	}

	// Run VACUUM
	if _, err := h.db.Exec("VACUUM;"); err != nil {
		errors = append(errors, fmt.Sprintf("VACUUM failed: %v", err))
	} else {
		steps = append(steps, "VACUUM completed")
	}

	// Re-initialize schema
	if err := h.db.Init(); err != nil {
		errors = append(errors, fmt.Sprintf("schema initialization failed: %v", err))
	} else {
		steps = append(steps, "Schema re-initialized")
	}

	status := "success"
	if len(errors) > 0 {
		status = "partial"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"steps":  steps,
		"errors": errors,
	})
}
