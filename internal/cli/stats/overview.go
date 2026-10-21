package stats

import (
	"fmt"
	"os"
	"time"

	"mangahub/pkg/client"
	"mangahub/pkg/models"
	"mangahub/pkg/session"

	"github.com/spf13/cobra"
)

// StatsData holds calculated statistics
type StatsData struct {
	TotalManga        int
	Reading           int
	Completed         int
	OnHold            int
	Dropped           int
	PlanToRead        int
	TotalChaptersRead int
	AverageRating     float64
	ReadingStreak     int
	LastReadDate      time.Time
	MostActiveDay     string
	GenreBreakdown    map[string]int
	ChaptersByStatus  map[string]int
	FromDate          *time.Time
	ToDate            *time.Time
}

// ProgressWithManga combines progress with manga information
type ProgressWithManga struct {
	Progress models.Progress
	Manga    *models.Manga
}

// overviewCmd handles `mangahub stats overview`.
var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "View reading overview",
	Long:  `Display a summary of your reading statistics including total manga, chapters read, ratings, and streaks.`,
	RunE:  runOverview,
}

func init() {
	// Attach flags and validation to the root stats command.
	StatsCmd.PersistentFlags().StringVar(&fromDate, "from", "", "Start date for statistics (format: YYYY-MM-DD)")
	StatsCmd.PersistentFlags().StringVar(&toDate, "to", "", "End date for statistics (format: YYYY-MM-DD)")

	StatsCmd.PersistentPreRunE = validateDateFlags

	// Register this subcommand.
	StatsCmd.AddCommand(overviewCmd)
}

// validateDateFlags validates the --from/--to date flags.
// It also ensures the profile from root command is properly set.
func validateDateFlags(cmd *cobra.Command, args []string) error {
	// IMPORTANT: Set profile before any command runs
	if profileFlag, err := cmd.Flags().GetString("profile"); err == nil && profileFlag != "" {
		session.SetProfile(profileFlag)
	}

	if fromDate != "" {
		if _, err := time.Parse("2006-01-02", fromDate); err != nil {
			return fmt.Errorf("invalid --from date format: %w (expected YYYY-MM-DD)", err)
		}
	}
	if toDate != "" {
		if _, err := time.Parse("2006-01-02", toDate); err != nil {
			return fmt.Errorf("invalid --to date format: %w (expected YYYY-MM-DD)", err)
		}
	}
	return nil
}

// runOverview displays overview statistics.
func runOverview(cmd *cobra.Command, args []string) error {
	stats, err := getStats()
	if err != nil {
		return err
	}

	fmt.Println("Reading Statistics Overview")
	fmt.Println("═══════════════════════════════════════")

	if stats.FromDate != nil || stats.ToDate != nil {
		fmt.Print("Period: ")
		if stats.FromDate != nil {
			fmt.Print(stats.FromDate.Format("2006-01-02"))
		} else {
			fmt.Print("beginning")
		}
		fmt.Print(" to ")
		if stats.ToDate != nil {
			fmt.Print(stats.ToDate.Format("2006-01-02"))
		} else {
			fmt.Print("today")
		}
		fmt.Println()
		fmt.Println()
	}

	fmt.Printf("Total Manga: %s\n", FormatNumber(stats.TotalManga))
	fmt.Printf("Currently Reading: %s\n", FormatNumber(stats.Reading))
	fmt.Printf("Completed: %s\n", FormatNumber(stats.Completed))
	fmt.Printf("On Hold: %s\n", FormatNumber(stats.OnHold))
	fmt.Printf("Dropped: %s\n", FormatNumber(stats.Dropped))
	fmt.Printf("Plan to Read: %s\n", FormatNumber(stats.PlanToRead))

	fmt.Printf("\nTotal Chapters Read: %s\n", FormatNumber(stats.TotalChaptersRead))

	// Estimate reading hours (assuming ~15 minutes per chapter)
	estimatedHours := float64(stats.TotalChaptersRead) * 0.25
	fmt.Printf("Estimated Reading Time: ~%s hours\n", FormatFloat(estimatedHours, 0))

	if stats.AverageRating > 0 {
		fmt.Printf("Average Rating: %s/10\n", FormatFloat(stats.AverageRating, 1))
	} else {
		fmt.Println("Average Rating: N/A (no ratings yet)")
	}

	fmt.Printf("Reading Streak: %s days\n", FormatNumber(stats.ReadingStreak))

	if stats.MostActiveDay != "" {
		fmt.Printf("Most Active Day: %s\n", stats.MostActiveDay)
	}

	if !stats.LastReadDate.IsZero() {
		fmt.Printf("Last Read: %s\n", stats.LastReadDate.Format("2006-01-02"))
	}

	return nil
}

// getStats retrieves and calculates statistics using services.
// getAPIURL returns the API server URL
func getAPIURL() string {
	if url := os.Getenv("MANGAHUB_API_URL"); url != "" {
		return url
	}
	return "http://10.238.53.72:8080"
}

func getStats() (*StatsData, error) {
	// Load session to get user ID
	sess, err := session.Load()
	if err != nil {
		return nil, fmt.Errorf("you are not logged in. Please login first: %w", err)
	}

	// Fetch library from HTTP API instead of local database
	fmt.Printf("Fetching library data for user: %s (profile: %s)\n", sess.Username, session.GetProfile())
	httpClient := client.NewHTTPClient(getAPIURL(), sess.Token)

	progressList, err := httpClient.GetLibrary("", 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch library via HTTP API: %w", err)
	}

	fmt.Printf("✓ Retrieved %d entries from server\n", len(progressList))

	// Parse date flags
	var fromDatePtr, toDatePtr *time.Time
	if fromDate != "" {
		t, err := time.Parse("2006-01-02", fromDate)
		if err != nil {
			return nil, fmt.Errorf("invalid --from date: %w", err)
		}
		// Set to start of day
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		fromDatePtr = &startOfDay
	}
	if toDate != "" {
		t, err := time.Parse("2006-01-02", toDate)
		if err != nil {
			return nil, fmt.Errorf("invalid --to date: %w", err)
		}
		// Set to end of day
		endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		toDatePtr = &endOfDay
	}

	// Filter by date range if specified
	var filteredProgress []models.Progress
	for _, p := range progressList {
		if fromDatePtr != nil && p.UpdatedAt.Before(*fromDatePtr) {
			continue
		}
		if toDatePtr != nil && p.UpdatedAt.After(*toDatePtr) {
			continue
		}
		filteredProgress = append(filteredProgress, p)
	}

	// Fetch manga details for each progress entry via HTTP API
	progressWithManga := make([]ProgressWithManga, 0, len(filteredProgress))
	for _, p := range filteredProgress {
		// Fetch manga details from HTTP API
		manga, err := httpClient.GetManga(p.MangaID)
		if err != nil {
			// Manga might not exist, continue without it
			progressWithManga = append(progressWithManga, ProgressWithManga{
				Progress: p,
				Manga:    nil,
			})
			continue
		}
		progressWithManga = append(progressWithManga, ProgressWithManga{
			Progress: p,
			Manga:    manga,
		})
	}

	// Calculate statistics
	stats := calculateStats(progressWithManga, fromDatePtr, toDatePtr)
	return stats, nil
}
