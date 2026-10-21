package stats

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// detailedCmd handles `mangahub stats detailed`.
var detailedCmd = &cobra.Command{
	Use:   "detailed",
	Short: "Detailed statistics",
	Long:  `Display detailed breakdown of reading statistics including chapters by status and genre distribution.`,
	RunE:  runDetailed,
}

func init() {
	// Register this subcommand with the root stats command.
	StatsCmd.AddCommand(detailedCmd)
}

// runDetailed displays detailed statistics.
func runDetailed(cmd *cobra.Command, args []string) error {
	stats, err := getStats()
	if err != nil {
		return err
	}

	fmt.Println("Detailed Reading Statistics")
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

	// Chapters by status
	fmt.Println("Chapters by Status:")
	totalChapters := 0
	for status, chapters := range stats.ChaptersByStatus {
		if chapters > 0 {
			fmt.Printf("  %s: %s chapters\n", status, FormatNumber(chapters))
			totalChapters += chapters
		}
	}
	fmt.Printf("  Total: %s chapters\n", FormatNumber(totalChapters))

	// Genre breakdown
	if len(stats.GenreBreakdown) > 0 {
		fmt.Println("\nTop Genres:")
		topGenres := GetTopGenres(stats.GenreBreakdown, 10)
		for _, genre := range topGenres {
			fmt.Printf("  %s: %s (%.1f%%)\n", genre.Genre, FormatNumber(genre.Count), genre.Pct)
		}
	} else {
		fmt.Println("\nTop Genres: No genre data available")
	}

	// Status breakdown
	fmt.Println("\nManga by Status:")
	fmt.Printf("  Reading: %s\n", FormatNumber(stats.Reading))
	fmt.Printf("  Completed: %s\n", FormatNumber(stats.Completed))
	fmt.Printf("  On Hold: %s\n", FormatNumber(stats.OnHold))
	fmt.Printf("  Dropped: %s\n", FormatNumber(stats.Dropped))
	fmt.Printf("  Plan to Read: %s\n", FormatNumber(stats.PlanToRead))
	fmt.Printf("  Total: %s\n", FormatNumber(stats.TotalManga))

	return nil
}

// calculateStats calculates statistics from progress and manga data.
func calculateStats(progressList []ProgressWithManga, fromDate, toDate *time.Time) *StatsData {
	stats := &StatsData{
		GenreBreakdown:   make(map[string]int),
		ChaptersByStatus: make(map[string]int),
		FromDate:         fromDate,
		ToDate:           toDate,
	}

	var totalRating int
	var ratingCount int
	var lastReadDate time.Time
	dayActivity := make(map[string]int)

	for _, pm := range progressList {
		// Count by status
		switch pm.Progress.Status {
		case "reading":
			stats.Reading++
		case "completed":
			stats.Completed++
		case "on-hold":
			stats.OnHold++
		case "dropped":
			stats.Dropped++
		case "plan-to-read":
			stats.PlanToRead++
		}

		// Count chapters by status
		stats.ChaptersByStatus[pm.Progress.Status] += pm.Progress.CurrentChapter

		// Total chapters read (only for reading/completed)
		if pm.Progress.Status == "reading" || pm.Progress.Status == "completed" {
			stats.TotalChaptersRead += pm.Progress.CurrentChapter
		}

		// Average rating (only count rated items)
		if pm.Progress.Rating > 0 {
			totalRating += pm.Progress.Rating
			ratingCount++
		}

		// Track last read date
		if pm.Progress.UpdatedAt.After(lastReadDate) {
			lastReadDate = pm.Progress.UpdatedAt
		}

		// Track day activity
		day := pm.Progress.UpdatedAt.Weekday().String()
		dayActivity[day]++

		// Genre breakdown (from manga data)
		if pm.Manga != nil {
			for _, genre := range pm.Manga.Genres {
				stats.GenreBreakdown[genre]++
			}
		}
	}

	stats.TotalManga = len(progressList)
	stats.LastReadDate = lastReadDate

	// Calculate average rating
	if ratingCount > 0 {
		stats.AverageRating = float64(totalRating) / float64(ratingCount)
	}

	// Find most active day
	if len(dayActivity) > 0 {
		maxCount := 0
		for day, count := range dayActivity {
			if count > maxCount {
				maxCount = count
				stats.MostActiveDay = day
			}
		}
	}

	// Calculate reading streak
	stats.ReadingStreak = calculateReadingStreak(progressList)

	return stats
}

// calculateReadingStreak calculates consecutive days with reading activity.
func calculateReadingStreak(progressList []ProgressWithManga) int {
	if len(progressList) == 0 {
		return 0
	}

	// Get all unique dates with activity
	activityDates := make(map[string]bool)
	for _, pm := range progressList {
		if pm.Progress.Status == "reading" || pm.Progress.Status == "completed" {
			date := pm.Progress.UpdatedAt.Format("2006-01-02")
			activityDates[date] = true
		}
	}

	if len(activityDates) == 0 {
		return 0
	}

	// Convert to sorted slice
	var dates []time.Time
	for dateStr := range activityDates {
		t, _ := time.Parse("2006-01-02", dateStr)
		dates = append(dates, t)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].After(dates[j])
	})

	// Calculate streak from today backwards
	today := time.Now()
	currentDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	streak := 0

	// Check if there's activity today or yesterday (allow for timezone differences)
	foundToday := false
	for _, date := range dates {
		dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		daysDiff := int(currentDate.Sub(dateOnly).Hours() / 24)
		if daysDiff == 0 || daysDiff == 1 {
			foundToday = true
			break
		}
	}

	if !foundToday {
		return 0
	}

	// Count consecutive days
	checkDate := currentDate
	for {
		dateStr := checkDate.Format("2006-01-02")
		if activityDates[dateStr] {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak
}

// FormatNumber formats numbers with commas.
func FormatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	var result strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(r)
	}
	return result.String()
}

// FormatFloat formats float with specified decimal places.
func FormatFloat(f float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, f)
}

// GetTopGenres returns top N genres sorted by count.
func GetTopGenres(genreBreakdown map[string]int, limit int) []struct {
	Genre string
	Count int
	Pct   float64
} {
	type genreCount struct {
		genre string
		count int
	}

	var genres []genreCount
	total := 0
	for genre, count := range genreBreakdown {
		genres = append(genres, genreCount{genre: genre, count: count})
		total += count
	}

	// Sort by count descending
	sort.Slice(genres, func(i, j int) bool {
		return genres[i].count > genres[j].count
	})

	// Take top N
	if limit > len(genres) {
		limit = len(genres)
	}

	result := make([]struct {
		Genre string
		Count int
		Pct   float64
	}, limit)

	for i := 0; i < limit; i++ {
		pct := 0.0
		if total > 0 {
			pct = (float64(genres[i].count) / float64(total)) * 100
		}
		result[i] = struct {
			Genre string
			Count int
			Pct   float64
		}{
			Genre: genres[i].genre,
			Count: genres[i].count,
			Pct:   pct,
		}
	}

	return result
}
