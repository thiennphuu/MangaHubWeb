package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// Database represents the SQLite database connection
type Database struct {
	DB   *sql.DB
	Path string
}

// New creates a new database connection
func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Database{DB: db, Path: dbPath}, nil
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
		total_chapters INTEGER,
		description TEXT,
		cover_url TEXT,
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
		chapter_releases BOOLEAN DEFAULT 1,
		email_notifications BOOLEAN DEFAULT 1,
		sound_enabled BOOLEAN DEFAULT 1,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_user_progress_user ON user_progress(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_progress_manga ON user_progress(manga_id);
	CREATE INDEX IF NOT EXISTS idx_chat_room ON chat_messages(room_id);
	CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
	CREATE INDEX IF NOT EXISTS idx_notification_subs_user ON notification_subscriptions(user_id);
	`

	_, err := d.DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// Query executes a SELECT query
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.Query(query, args...)
}

// QueryRow executes a SELECT query that returns a single row
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.DB.QueryRow(query, args...)
}

// Exec executes an INSERT, UPDATE, or DELETE query
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.DB.Exec(query, args...)
}

// BeginTx begins a transaction
func (d *Database) BeginTx() (*sql.Tx, error) {
	return d.DB.Begin()
}
