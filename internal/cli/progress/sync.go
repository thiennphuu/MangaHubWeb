package progress

import (
	"fmt"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize reading progress with the server",
	Long:  `Synchronize your local manga reading progress with the server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔄 Synchronizing progress with the server...")

		// Auth
		httpClient, sess, err := newAuthenticatedHTTPClient()
		if err != nil {
			fmt.Println("You are not logged in.")
			fmt.Println("\nPlease login first:")
			fmt.Println("  mangahub auth login --username <username>")
			return nil
		}

		// 1. Fetch remote progress (all entries)
		remoteProgress, err := httpClient.GetLibrary("", 1000, 0)
		if err != nil {
			return fmt.Errorf("failed to fetch remote progress: %w", err)
		}

		// 2. Fetch local progress from SQLite
		dbPath := "./data/mangahub.db"
		db, err := RequireDatabase(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open local database: %w", err)
		}
		defer db.Close()

		localProgress, err := fetchLocalProgress(db, sess.UserID)
		if err != nil {
			return fmt.Errorf("failed to fetch local progress: %w", err)
		}

		// 3. Merge and resolve conflicts (latest UpdatedAt wins)
		_, toUpdateLocal, toUpdateServer := mergeProgress(localProgress, remoteProgress)

		// 4. Update server if needed
		for _, p := range toUpdateServer {
			err := httpClient.UpdateProgress(p.MangaID, p.CurrentChapter, p.Status, p.Rating, p.Notes)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to update server for manga %s: %v\n", p.MangaID, err)
			}
		}

		// 5. Update local if needed
		for _, p := range toUpdateLocal {
			err := updateLocalProgress(db, p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to update local for manga %s: %v\n", p.MangaID, err)
			}
		}

		fmt.Printf("Progress sync complete. Updated local: %d, updated server: %d\n", len(toUpdateLocal), len(toUpdateServer))
		return nil
	},
}

// RequireDatabase opens the SQLite DB and ensures the schema exists
func RequireDatabase(dbPath string) (*database.Database, error) {
	db, err := database.New(dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Init(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// fetchLocalProgress retrieves all progress for a user from SQLite
func fetchLocalProgress(db *database.Database, userID string) ([]models.Progress, error) {
	rows, err := db.Query(`SELECT user_id, manga_id, current_chapter, status, rating, notes, started_at, completed_at, updated_at FROM user_progress WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Progress
	for rows.Next() {
		var p models.Progress
		var startedAt, updatedAt string
		var completedAt *string
		err := rows.Scan(&p.UserID, &p.MangaID, &p.CurrentChapter, &p.Status, &p.Rating, &p.Notes, &startedAt, &completedAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		p.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
		if completedAt != nil {
			t, _ := time.Parse(time.RFC3339, *completedAt)
			p.CompletedAt = &t
		}
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		result = append(result, p)
	}
	return result, nil
}

// updateLocalProgress updates or inserts a progress row in SQLite
func updateLocalProgress(db *database.Database, p models.Progress) error {
	_, err := db.Exec(`INSERT OR REPLACE INTO user_progress (user_id, manga_id, current_chapter, status, rating, notes, started_at, completed_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.UserID, p.MangaID, p.CurrentChapter, p.Status, p.Rating, p.Notes, p.StartedAt.Format(time.RFC3339),
		func() interface{} {
			if p.CompletedAt != nil {
				return p.CompletedAt.Format(time.RFC3339)
			} else {
				return nil
			}
		}(),
		p.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

// mergeProgress merges local and remote progress, returns merged, toUpdateLocal, toUpdateServer
func mergeProgress(local, remote []models.Progress) (merged, toUpdateLocal, toUpdateServer []models.Progress) {
	byID := make(map[string]models.Progress)
	for _, p := range local {
		byID[p.MangaID] = p
	}
	for _, rp := range remote {
		if lp, ok := byID[rp.MangaID]; ok {
			if rp.UpdatedAt.After(lp.UpdatedAt) {
				// Server is newer
				byID[rp.MangaID] = rp
				toUpdateLocal = append(toUpdateLocal, rp)
			} else if lp.UpdatedAt.After(rp.UpdatedAt) {
				// Local is newer
				toUpdateServer = append(toUpdateServer, lp)
			}
		} else {
			// Only on server
			byID[rp.MangaID] = rp
			toUpdateLocal = append(toUpdateLocal, rp)
		}
	}
	for _, lp := range local {
		if _, ok := byID[lp.MangaID]; !ok {
			// Only local
			toUpdateServer = append(toUpdateServer, lp)
			byID[lp.MangaID] = lp
		}
	}
	for _, p := range byID {
		merged = append(merged, p)
	}
	return
}

func init() {
	ProgressCmd.AddCommand(syncCmd)
}
