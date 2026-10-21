package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"mangahub/pkg/models"
)

// HTTPClient represents an HTTP client for API calls
type HTTPClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL, token string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Token:   token,
		Client:  &http.Client{},
	}
}

// SetToken sets the authentication token
func (c *HTTPClient) SetToken(token string) {
	c.Token = token
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

// Register registers a new user
func (c *HTTPClient) Register(username, email, password string) (*RegisterResponse, error) {
	req := models.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.post("/auth/register", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	var result RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Login logs in a user
func (c *HTTPClient) Login(username, password string) (*models.LoginResponse, error) {
	req := models.LoginRequest{
		Username: username,
		Password: password,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.post("/auth/login", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	var loginResp models.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, err
	}

	c.Token = loginResp.Token
	return &loginResp, nil
}

// GetProfile retrieves the current user's profile
func (c *HTTPClient) GetProfile() (*models.User, error) {
	resp, err := c.get("/users/profile")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("session expired or invalid")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get profile with status %d", resp.StatusCode)
	}

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// ListManga lists manga with pagination and optional filters
func (c *HTTPClient) ListManga(limit, offset int, status, genre string) ([]models.Manga, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	if status != "" {
		params.Set("status", status)
	}
	if genre != "" {
		params.Set("genre", genre)
	}

	endpoint := "/manga?" + params.Encode()
	resp, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list manga with status %d", resp.StatusCode)
	}

	var mangaList []models.Manga
	if err := json.NewDecoder(resp.Body).Decode(&mangaList); err != nil {
		return nil, err
	}

	return mangaList, nil
}

// SearchManga searches for manga using POST with filters
func (c *HTTPClient) SearchManga(filter *models.MangaFilter) (*models.SearchResult, error) {
	data, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	resp, err := c.post("/manga/search", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status %d", resp.StatusCode)
	}

	var result models.SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetManga gets manga details
func (c *HTTPClient) GetManga(id string) (*models.Manga, error) {
	resp, err := c.get("/manga/" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("manga not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get manga failed with status %d", resp.StatusCode)
	}

	var manga models.Manga
	if err := json.NewDecoder(resp.Body).Decode(&manga); err != nil {
		return nil, err
	}

	return &manga, nil
}

// Helper methods

func (c *HTTPClient) post(endpoint string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, io.NopCloser(bytes.NewBuffer(data)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return c.Client.Do(req)
}

func (c *HTTPClient) get(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return c.Client.Do(req)
}

func (c *HTTPClient) put(endpoint string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest("PUT", c.BaseURL+endpoint, io.NopCloser(bytes.NewBuffer(data)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return c.Client.Do(req)
}

func (c *HTTPClient) delete(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return c.Client.Do(req)
}

// GetLibrary retrieves user's library
func (c *HTTPClient) GetLibrary(status string, limit, offset int) ([]models.Progress, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))

	endpoint := "/users/library"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized: please login first")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get library: status %d", resp.StatusCode)
	}

	var progressList []models.Progress
	if err := json.NewDecoder(resp.Body).Decode(&progressList); err != nil {
		return nil, err
	}

	return progressList, nil
}

// AddToLibrary adds a manga to the user's library
func (c *HTTPClient) AddToLibrary(mangaID, status string, rating int, notes string) error {
	payload := map[string]interface{}{
		"manga_id": mangaID,
		"status":   status,
		"rating":   rating,
		"notes":    notes,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.post("/users/library", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: please login first")
	}

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			return fmt.Errorf("%s", msg)
		}
		return fmt.Errorf("failed to add to library: status %d", resp.StatusCode)
	}

	return nil
}

// RemoveFromLibrary removes a manga from the user's library
func (c *HTTPClient) RemoveFromLibrary(mangaID string) error {
	resp, err := c.delete("/users/library/" + mangaID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: please login first")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to remove from library: status %d", resp.StatusCode)
	}

	return nil
}

// UpdateProgress updates reading progress for a manga
func (c *HTTPClient) UpdateProgress(mangaID string, chapter int, status string, rating int, notes string) error {
	payload := map[string]interface{}{
		"current_chapter": chapter,
		"status":          status,
		"rating":          rating,
		"notes":           notes,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.put("/users/library/"+mangaID+"/progress", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: please login first")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			return fmt.Errorf("%s", msg)
		}
		return fmt.Errorf("failed to update progress: status %d", resp.StatusCode)
	}

	return nil
}

// GetServerHealth fetches server health/status information
func (c *HTTPClient) GetServerHealth() (map[string]interface{}, error) {
	resp, err := c.get("/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, err
	}

	return health, nil
}

// ServerLogsResponse represents the server logs API response
type ServerLogsResponse struct {
	Logs     []string `json:"logs"`
	Count    int      `json:"count"`
	MaxLines int      `json:"max_lines"`
	Level    string   `json:"level"`
}

// GetServerLogs fetches server logs from the API
func (c *HTTPClient) GetServerLogs(maxLines int, level string) (*ServerLogsResponse, error) {
	// Build query parameters
	endpoint := fmt.Sprintf("/server/logs?max_lines=%d", maxLines)
	if level != "" {
		endpoint += fmt.Sprintf("&level=%s", level)
	}

	resp, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch logs with status %d: %s", resp.StatusCode, string(body))
	}

	var logsResp ServerLogsResponse
	if err := json.NewDecoder(resp.Body).Decode(&logsResp); err != nil {
		return nil, fmt.Errorf("failed to decode logs response: %w", err)
	}

	return &logsResp, nil
}

// DatabaseCheckResponse represents the database check API response
type DatabaseCheckResponse struct {
	Status    string `json:"status"`
	Integrity struct {
		OK     bool     `json:"ok"`
		Issues []string `json:"issues"`
	} `json:"integrity"`
	Tables struct {
		Verified []string `json:"verified"`
		Missing  []string `json:"missing"`
	} `json:"tables"`
}

// GetDatabaseCheck fetches database integrity check results from the API
func (c *HTTPClient) GetDatabaseCheck() (*DatabaseCheckResponse, error) {
	resp, err := c.get("/server/database/check")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to check database with status %d: %s", resp.StatusCode, string(body))
	}

	var checkResp DatabaseCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
		return nil, fmt.Errorf("failed to decode database check response: %w", err)
	}

	return &checkResp, nil
}

// DatabaseOptimizeResponse represents the optimize API response
type DatabaseOptimizeResponse struct {
	Status string   `json:"status"`
	Steps  []string `json:"steps"`
	Errors []string `json:"errors"`
}

// OptimizeDatabase triggers database optimization on the server
func (c *HTTPClient) OptimizeDatabase() (*DatabaseOptimizeResponse, error) {
	resp, err := c.post("/server/database/optimize", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to optimize database with status %d: %s", resp.StatusCode, string(body))
	}

	var optimizeResp DatabaseOptimizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&optimizeResp); err != nil {
		return nil, fmt.Errorf("failed to decode optimize response: %w", err)
	}

	return &optimizeResp, nil
}

// DatabaseStatsResponse represents the stats API response
type DatabaseStatsResponse struct {
	FileSizeBytes int64          `json:"file_size_bytes"`
	FileSizeMB    float64        `json:"file_size_mb"`
	Tables        map[string]int `json:"tables"`
}

// GetDatabaseStats fetches database statistics from the API
func (c *HTTPClient) GetDatabaseStats() (*DatabaseStatsResponse, error) {
	resp, err := c.get("/server/database/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get database stats with status %d: %s", resp.StatusCode, string(body))
	}

	var statsResp DatabaseStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&statsResp); err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %w", err)
	}

	return &statsResp, nil
}

// DatabaseRepairResponse represents the repair API response
type DatabaseRepairResponse struct {
	Status string   `json:"status"`
	Steps  []string `json:"steps"`
	Errors []string `json:"errors"`
}

// RepairDatabase triggers database repair on the server
func (c *HTTPClient) RepairDatabase() (*DatabaseRepairResponse, error) {
	resp, err := c.post("/server/database/repair", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to repair database with status %d: %s", resp.StatusCode, string(body))
	}

	var repairResp DatabaseRepairResponse
	if err := json.NewDecoder(resp.Body).Decode(&repairResp); err != nil {
		return nil, fmt.Errorf("failed to decode repair response: %w", err)
	}

	return &repairResp, nil
}
