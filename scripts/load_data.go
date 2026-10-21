package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mangahub/pkg/utils"

	_ "modernc.org/sqlite"
)

// MangaData represents manga entry from JSON files
type MangaData struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Artist      string   `json:"artist"`
	Genres      []string `json:"genres"`
	Status      string   `json:"status"`
	Chapters    int      `json:"chapters"`
	Volumes     int      `json:"volumes"`
	Description string   `json:"description"`
	Year        int      `json:"year"`
	Rating      float64  `json:"rating"`
	Source      string   `json:"source"`
}

// ScrapedQuote represents data from quotes.toscrape.com (educational practice)
type ScrapedQuote struct {
	Text   string   `json:"text"`
	Author string   `json:"author"`
	Tags   []string `json:"tags"`
}

// HTTPBinResponse represents data from httpbin.org (educational practice)
type HTTPBinResponse struct {
	Headers map[string]string `json:"headers"`
	Origin  string            `json:"origin"`
	URL     string            `json:"url"`
}

func main() {
	fmt.Println("=== MangaHub Database Loader ===")
	fmt.Println()

	// Get workspace path
	workspacePath, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}

	// Paths
	dataDir := filepath.Join(workspacePath, "data")
	dbPath := filepath.Join(dataDir, "mangahub.db")
	manualFile := filepath.Join(dataDir, "manga_manual.json")
	apiFile := filepath.Join(dataDir, "manga_api.json")
	combinedFile := filepath.Join(dataDir, "manga_combined.json")
	scrapedFile := filepath.Join(dataDir, "scraped_practice.json")

	// Create data directory if not exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Load manual entries
	fmt.Println("📚 Loading manual manga entries...")
	manualData, err := loadMangaFromFile(manualFile)
	if err != nil {
		log.Fatal("Failed to load manual data:", err)
	}
	fmt.Printf("   Loaded %d manual entries\n", len(manualData))

	// Load API entries
	fmt.Println("🌐 Loading API manga entries...")
	apiData, err := loadMangaFromFile(apiFile)
	if err != nil {
		log.Fatal("Failed to load API data:", err)
	}
	fmt.Printf("   Loaded %d API entries\n", len(apiData))

	// Combine all data
	fmt.Println("🔗 Combining manga data...")
	allManga := append(manualData, apiData...)
	fmt.Printf("   Total: %d manga entries\n", len(allManga))

	// Save combined data
	fmt.Println("💾 Saving combined data...")
	if err := saveCombinedData(combinedFile, allManga); err != nil {
		log.Fatal("Failed to save combined data:", err)
	}

	// Educational practice: Scrape quotes.toscrape.com
	fmt.Println("🎓 Educational Practice: Fetching from practice sites...")
	scrapedData := educationalPractice()
	if err := saveScrapedData(scrapedFile, scrapedData); err != nil {
		log.Printf("Warning: Failed to save scraped data: %v", err)
	}

	// Initialize SQLite database
	fmt.Println("🗄️  Initializing SQLite database...")
	db, err := initDatabase(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Load manga into database
	fmt.Println("📥 Loading manga into SQLite...")
	loaded, err := loadMangaToDatabase(db, allManga)
	if err != nil {
		log.Fatal("Failed to load manga to database:", err)
	}
	fmt.Printf("   Loaded %d manga entries to database\n", loaded)

	// Print statistics
	fmt.Println()
	fmt.Println("=== Database Statistics ===")
	printStatistics(db, allManga)

	fmt.Println()
	fmt.Println("✅ Database setup complete!")
}

func loadMangaFromFile(filepath string) ([]MangaData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var manga []MangaData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&manga); err != nil {
		return nil, err
	}

	return manga, nil
}

func saveCombinedData(filepath string, manga []MangaData) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manga)
}

func educationalPractice() map[string]interface{} {
	result := make(map[string]interface{})
	logger := utils.NewLogger()

	// Practice 1: quotes.toscrape.com (a website specifically designed for scraping practice)
	fmt.Println("   Fetching from quotes.toscrape.com...")
	quotes, err := fetchQuotes()
	if err != nil {
		logger.Warn("could not fetch quotes: %v", err)
		fmt.Println("   Warning: Could not fetch quotes")
	} else {
		result["quotes"] = quotes
		fmt.Printf("   Fetched %d quotes\n", len(quotes))
	}

	// Practice 2: httpbin.org (a service for testing HTTP requests)
	fmt.Println("   Testing with httpbin.org...")
	httpbinData, err := fetchHTTPBin()
	if err != nil {
		logger.Warn("could not fetch httpbin data: %v", err)
		fmt.Println("   Warning: Could not fetch httpbin data")
	} else {
		result["httpbin"] = httpbinData
		fmt.Println("   HTTPBin test successful")
	}

	return result
}

func fetchQuotes() ([]ScrapedQuote, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("http://quotes.toscrape.com/api/quotes?page=1")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Quotes []struct {
			Text   string `json:"text"`
			Author struct {
				Name string `json:"name"`
			} `json:"author"`
			Tags []string `json:"tags"`
		} `json:"quotes"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	quotes := make([]ScrapedQuote, len(data.Quotes))
	for i, q := range data.Quotes {
		quotes[i] = ScrapedQuote{
			Text:   q.Text,
			Author: q.Author.Name,
			Tags:   q.Tags,
		}
	}

	return quotes, nil
}

func fetchHTTPBin() (*HTTPBinResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://httpbin.org/get")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data HTTPBinResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func saveScrapedData(filepath string, data map[string]interface{}) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func initDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS manga (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT,
		artist TEXT,
		genres TEXT,
		status TEXT,
		chapters INTEGER,
		volumes INTEGER,
		description TEXT,
		year INTEGER,
		rating REAL,
		source TEXT,
		cover_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS user_progress (
		user_id TEXT NOT NULL,
		manga_id TEXT NOT NULL,
		current_chapter INTEGER DEFAULT 0,
		status TEXT DEFAULT 'plan-to-read',
		rating INTEGER DEFAULT 0,
		notes TEXT,
		started_at TIMESTAMP,
		completed_at TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, manga_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (manga_id) REFERENCES manga(id)
	);

	CREATE TABLE IF NOT EXISTS chat_messages (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		username TEXT NOT NULL,
		room_id TEXT NOT NULL,
		message TEXT NOT NULL,
		timestamp INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		type TEXT NOT NULL,
		manga_id TEXT,
		message TEXT NOT NULL,
		read BOOLEAN DEFAULT 0,
		data TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_manga_title ON manga(title);
	CREATE INDEX IF NOT EXISTS idx_manga_author ON manga(author);
	CREATE INDEX IF NOT EXISTS idx_manga_status ON manga(status);
	CREATE INDEX IF NOT EXISTS idx_manga_genres ON manga(genres);
	CREATE INDEX IF NOT EXISTS idx_user_progress_user ON user_progress(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_progress_manga ON user_progress(manga_id);
	CREATE INDEX IF NOT EXISTS idx_chat_room ON chat_messages(room_id);
	CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func loadMangaToDatabase(db *sql.DB, manga []MangaData) (int, error) {
	// Prepare statement
	stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO manga (id, title, author, artist, genres, status, chapters, volumes, description, year, rating, source, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, m := range manga {
		genres := strings.Join(m.Genres, ",")
		_, err := stmt.Exec(m.ID, m.Title, m.Author, m.Artist, genres, m.Status, m.Chapters, m.Volumes, m.Description, m.Year, m.Rating, m.Source)
		if err != nil {
			log.Printf("Warning: Failed to insert manga %s: %v", m.ID, err)
			continue
		}
		count++
	}

	return count, nil
}

func printStatistics(db *sql.DB, manga []MangaData) {
	// Count by genre
	genreCounts := make(map[string]int)
	for _, m := range manga {
		for _, g := range m.Genres {
			genreCounts[g]++
		}
	}

	// Count by status
	statusCounts := make(map[string]int)
	for _, m := range manga {
		statusCounts[m.Status]++
	}

	// Count by source
	sourceCounts := make(map[string]int)
	for _, m := range manga {
		sourceCounts[m.Source]++
	}

	fmt.Println()
	fmt.Println("📊 Genre Distribution:")
	majorGenres := []string{"shounen", "seinen", "shoujo", "josei"}
	for _, g := range majorGenres {
		if count, ok := genreCounts[g]; ok {
			fmt.Printf("   %s: %d series\n", g, count)
		}
	}

	fmt.Println()
	fmt.Println("📊 Status Distribution:")
	for status, count := range statusCounts {
		fmt.Printf("   %s: %d series\n", status, count)
	}

	fmt.Println()
	fmt.Println("📊 Source Distribution:")
	for source, count := range sourceCounts {
		fmt.Printf("   %s: %d series\n", source, count)
	}

	// Verify database
	var dbCount int
	db.QueryRow("SELECT COUNT(*) FROM manga").Scan(&dbCount)
	fmt.Printf("\n📊 Total manga in database: %d\n", dbCount)
}
