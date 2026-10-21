package models

import (
	"encoding/json"
)

// MangaToJSON converts genres to JSON string
func MangaToJSON(genres []string) (string, error) {
	data, err := json.Marshal(genres)
	return string(data), err
}

// JSONToManga converts JSON string to genres
func JSONToManga(jsonStr string) ([]string, error) {
	var genres []string
	err := json.Unmarshal([]byte(jsonStr), &genres)
	return genres, err
}
