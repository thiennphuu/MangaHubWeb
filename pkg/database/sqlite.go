package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// Database represents the SQLite database connection
type Database struct {
	DB     *sql.DB
	Path   string
	Driver string
}

// New creates a new database connection
func New(dbPath string) (*Database, error) {
	driver, dsn := "sqlite", dbPath
	if url := strings.TrimSpace(os.Getenv("SUPABASE_DB_URL")); url != "" {
		driver, dsn = "postgres", url
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Database{DB: db, Path: dsn, Driver: driver}, nil
}

// Init initializes the database schema
func (d *Database) Init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS manga (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT,
		genres TEXT,
		status TEXT,
		chapters INTEGER,
		description TEXT,
		cover_url TEXT,
		rating REAL,
		is_featured BOOLEAN DEFAULT FALSE,
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

	CREATE TABLE IF NOT EXISTS frontend_reading_progress (
		user_id TEXT NOT NULL,
		manga_id TEXT NOT NULL,
		current_chapter INTEGER NOT NULL DEFAULT 0,
		current_page INTEGER NOT NULL DEFAULT 0,
		progress_percent REAL NOT NULL DEFAULT 0,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, manga_id)
	);

	CREATE INDEX IF NOT EXISTS idx_frontend_progress_user ON frontend_reading_progress(user_id);

	CREATE TABLE IF NOT EXISTS reading_activity (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		manga_id TEXT NOT NULL,
		chapter INTEGER NOT NULL,
		minutes INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (manga_id) REFERENCES manga(id)
	);

	CREATE INDEX IF NOT EXISTS idx_reading_activity_user_date ON reading_activity(user_id, created_at);

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
		read BOOLEAN DEFAULT FALSE,
		data TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS notification_subscriptions (
		user_id TEXT NOT NULL,
		manga_id TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, manga_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (manga_id) REFERENCES manga(id)
	);

	CREATE TABLE IF NOT EXISTS notification_preferences (
		user_id TEXT PRIMARY KEY,
		chapter_releases BOOLEAN DEFAULT TRUE,
		email_notifications BOOLEAN DEFAULT TRUE,
		sound_enabled BOOLEAN DEFAULT TRUE,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_user_progress_user ON user_progress(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_progress_manga ON user_progress(manga_id);
	CREATE INDEX IF NOT EXISTS idx_chat_room ON chat_messages(room_id);
	CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
	CREATE INDEX IF NOT EXISTS idx_notification_subs_user ON notification_subscriptions(user_id);
	`
	if d.Driver == "postgres" {
		schema = strings.ReplaceAll(schema, "INTEGER PRIMARY KEY AUTOINCREMENT", "BIGSERIAL PRIMARY KEY")
	}

	_, err := d.DB.Exec(d.rebind(schema))
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	if err := d.ensureMangaChapterColumn(); err != nil {
		return err
	}
	if err := d.ensureMangaFrontendColumns(); err != nil {
		return err
	}
	if err := d.ensureFrontendProgressPercentColumn(); err != nil {
		return err
	}

	log.Println("Database schema initialized successfully")
	return nil
}

func (d *Database) ensureMangaFrontendColumns() error {
	if d.Driver == "postgres" {
		return nil
	}
	rows, err := d.DB.Query("PRAGMA table_info(manga)")
	if err != nil {
		return fmt.Errorf("failed to inspect manga schema: %w", err)
	}

	defer rows.Close()

	hasRating, hasFeatured := false, false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue interface{}
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("failed to inspect manga schema: %w", err)
		}
		if name == "rating" {
			hasRating = true
		}
		if name == "is_featured" {
			hasFeatured = true
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to inspect manga schema: %w", err)
	}
	if !hasRating {
		if _, err := d.DB.Exec("ALTER TABLE manga ADD COLUMN rating REAL"); err != nil {
			return fmt.Errorf("failed to add manga rating column: %w", err)
		}
	}
	if !hasFeatured {
		if _, err := d.DB.Exec("ALTER TABLE manga ADD COLUMN is_featured BOOLEAN DEFAULT 0"); err != nil {
			return fmt.Errorf("failed to add manga featured column: %w", err)
		}
	}
	return nil
}

func (d *Database) ensureFrontendProgressPercentColumn() error {
	if d.Driver == "postgres" {
		return nil
	}
	rows, err := d.DB.Query("PRAGMA table_info(frontend_reading_progress)")
	if err != nil {
		return fmt.Errorf("failed to inspect frontend progress schema: %w", err)
	}
	defer rows.Close()
	hasColumn := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue interface{}
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("failed to inspect frontend progress schema: %w", err)
		}
		if name == "progress_percent" {
			hasColumn = true
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to inspect frontend progress schema: %w", err)
	}
	if !hasColumn {
		if _, err := d.DB.Exec("ALTER TABLE frontend_reading_progress ADD COLUMN progress_percent REAL NOT NULL DEFAULT 0"); err != nil {
			return fmt.Errorf("failed to add frontend progress percentage: %w", err)
		}
	}
	return nil
}

func (d *Database) ensureMangaChapterColumn() error {
	if d.Driver == "postgres" {
		return nil
	}
	rows, err := d.DB.Query("PRAGMA table_info(manga)")
	if err != nil {
		return fmt.Errorf("failed to inspect manga schema: %w", err)
	}
	defer rows.Close()

	hasChapters, hasLegacyTotalChapters := false, false
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue interface{}
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("failed to inspect manga schema: %w", err)
		}
		switch name {
		case "chapters":
			hasChapters = true
		case "total_chapters":
			hasLegacyTotalChapters = true
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to inspect manga schema: %w", err)
	}
	if hasChapters || !hasLegacyTotalChapters {
		return nil
	}

	if _, err := d.DB.Exec("ALTER TABLE manga ADD COLUMN chapters INTEGER"); err != nil {
		return fmt.Errorf("failed to migrate manga chapters column: %w", err)
	}
	if _, err := d.DB.Exec("UPDATE manga SET chapters = total_chapters WHERE chapters IS NULL"); err != nil {
		return fmt.Errorf("failed to migrate manga chapter values: %w", err)
	}
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// Query executes a SELECT query
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.Query(d.rebind(query), args...)
}

// QueryRow executes a SELECT query that returns a single row
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.DB.QueryRow(d.rebind(query), args...)
}

// Exec executes an INSERT, UPDATE, or DELETE query
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.DB.Exec(d.rebind(query), args...)
}

// BeginTx begins a transaction using the configured database driver.
func (d *Database) BeginTx() (*Tx, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, driver: d.Driver}, nil
}

type Tx struct {
	tx     *sql.Tx
	driver string
}

func (t *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(rebind(query, t.driver), args...)
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

func (d *Database) rebind(query string) string {
	return rebind(query, d.Driver)
}

func rebind(query, driver string) string {
	if driver != "postgres" {
		return query
	}
	var builder strings.Builder
	index := 1
	for _, character := range query {
		if character == '?' {
			fmt.Fprintf(&builder, "$%d", index)
			index++
			continue
		}
		builder.WriteRune(character)
	}
	return builder.String()
}
