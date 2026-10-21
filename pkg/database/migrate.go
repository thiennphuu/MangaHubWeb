package database

import (
	"fmt"
	"log"
)

// Migration represents a database migration
type Migration struct {
	Version string
	Up      func(*Database) error
	Down    func(*Database) error
}

// Migrator handles database migrations
type Migrator struct {
	db         *Database
	migrations []Migration
}

// NewMigrator creates a new migrator
func NewMigrator(db *Database) *Migrator {
	return &Migrator{
		db:         db,
		migrations: []Migration{},
	}
}

// Register registers a migration
func (m *Migrator) Register(version string, up, down func(*Database) error) {
	m.migrations = append(m.migrations, Migration{
		Version: version,
		Up:      up,
		Down:    down,
	})
}

// RunUp runs all pending migrations
func (m *Migrator) RunUp() error {
	log.Println("Running migrations...")

	for _, migration := range m.migrations {
		log.Printf("Running migration: %s", migration.Version)
		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Version, err)
		}
	}

	log.Println("Migrations completed successfully")
	return nil
}

// RunDown rolls back all migrations
func (m *Migrator) RunDown() error {
	log.Println("Rolling back migrations...")

	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		log.Printf("Rolling back migration: %s", migration.Version)
		if err := migration.Down(m.db); err != nil {
			return fmt.Errorf("rollback %s failed: %w", migration.Version, err)
		}
	}

	log.Println("Rollback completed successfully")
	return nil
}
