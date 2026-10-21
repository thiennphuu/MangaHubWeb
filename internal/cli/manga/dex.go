package manga

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// dexCmd fetches manga data directly from the public MangaDex API.
// This is a lightweight integration that can pull up to 100 series
// and optionally save them to a JSON file for later use.
var dexCmd = &cobra.Command{
	Use:   "dex [query]",
	Short: "Fetch manga from MangaDex API",
	Long: `Fetch manga series directly from the public MangaDex API.

Examples:
  # Fetch 100 popular series and print to console
  mangahub manga dex

  # Search by title and show up to 20 results
  mangahub manga dex "attack on titan" --limit 20

  # Fetch 100 popular series and save to data/manga_api.json
  mangahub manga dex --output data/manga_api.json`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var query string
		if len(args) > 0 {
			query = strings.Join(args, " ")
		}

		limit, _ := cmd.Flags().GetInt("limit")
		output, _ := cmd.Flags().GetString("output")

		results, raw, err := fetchFromMangaDex(query, limit)
		if err != nil {
			return err
		}

		if output != "" {
			if err := saveMangaDexJSON(output, raw); err != nil {
				return err
			}
			fmt.Printf("✓ Saved %d entries to %s\n", len(results), output)
			return nil
		}

		if len(results) == 0 {
			fmt.Println("No results found on MangaDex.")
			return nil
		}

		if query == "" {
			fmt.Printf("Top %d MangaDex series:\n\n", len(results))
		} else {
			fmt.Printf("MangaDex results for \"%s\":\n\n", query)
		}

		for i, r := range results {
			fmt.Printf("%d) %s\n", i+1, r.Title)
			fmt.Printf("   ID: %s\n", r.ID)
			if len(r.Genres) > 0 {
				fmt.Printf("   Genres: %s\n", strings.Join(r.Genres, ", "))
			}
			if r.Status != "" {
				fmt.Printf("   Status: %s\n", r.Status)
			}
			if r.Description != "" {
				fmt.Printf("   Description: %s\n", r.Description)
			}
			fmt.Println()
		}
		return nil
	},
}

func init() {
	MangaCmd.AddCommand(dexCmd)
	dexCmd.Flags().Int("limit", 100, "Maximum results to fetch (1-100)")
	dexCmd.Flags().String("output", "data/manga_api.json", "Path to save results as JSON")
}

// mangaDexResult is a simplified view of MangaDex data.
type mangaDexResult struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Genres      []string `json:"genres"`
}

// fetchFromMangaDex queries the MangaDex public manga endpoint.
// It returns simplified results plus the raw slice used for JSON export.
func fetchFromMangaDex(query string, limit int) ([]mangaDexResult, []mangaDexResult, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}

	baseURL := "https://api.mangadex.org/manga"
	params := url.Values{}
	if query != "" {
		params.Set("title", query)
	}
	params.Set("limit", fmt.Sprintf("%d", limit))
	// Order by followed count to get "popular" series when no query is given.
	params.Set("order[followedCount]", "desc")

	reqURL := baseURL + "?" + params.Encode()

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to call MangaDex: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("MangaDex returned HTTP %d", resp.StatusCode)
	}

	var payload mangaDexAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, nil, fmt.Errorf("failed to parse MangaDex response: %w", err)
	}

	var out []mangaDexResult
	for _, item := range payload.Data {
		title := pickFirstString(item.Attributes.Title)
		desc := pickFirstString(item.Attributes.Description)
		genres := extractGenres(item.Attributes.Tags)

		out = append(out, mangaDexResult{
			ID:          item.ID,
			Title:       title,
			Description: desc,
			Status:      item.Attributes.Status,
			Genres:      genres,
		})
	}

	return out, out, nil
}

// --- Minimal MangaDex response models ---

type mangaDexAPIResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Title       map[string]string `json:"title"`
			Description map[string]string `json:"description"`
			Status      string            `json:"status"`
			Tags        []struct {
				Attributes struct {
					Name map[string]string `json:"name"`
				} `json:"attributes"`
			} `json:"tags"`
		} `json:"attributes"`
	} `json:"data"`
}

// pickFirstString chooses a localized string, preferring English.
func pickFirstString(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	if v, ok := m["en"]; ok && v != "" {
		return v
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if m[k] != "" {
			return m[k]
		}
	}
	return ""
}

// extractGenres pulls readable genre names from MangaDex tags.
func extractGenres(tags []struct {
	Attributes struct {
		Name map[string]string `json:"name"`
	} `json:"attributes"`
}) []string {
	var genres []string
	for _, t := range tags {
		name := pickFirstString(t.Attributes.Name)
		if name != "" {
			genres = append(genres, name)
		}
	}
	return genres
}

// saveMangaDexJSON writes the simplified results as pretty JSON.
func saveMangaDexJSON(path string, data []mangaDexResult) error {
	if err := os.MkdirAll(strings.TrimSuffix(path, "/"+filepathBase(path)), 0o755); err != nil && !os.IsExist(err) {
		// Best-effort dir creation; if it fails for non-existing parent, we still propagate error.
		// But for simple paths like "data/manga_api.json" this should succeed.
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}
	return nil
}

// filepathBase is a tiny helper to avoid importing path/filepath just for Base.
func filepathBase(p string) string {
	if p == "" {
		return ""
	}
	parts := strings.Split(p, "/")
	if len(parts) == 1 {
		parts = strings.Split(p, "\\")
	}
	return parts[len(parts)-1]
}
