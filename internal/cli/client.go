package cli

import (
	"os"

	"mangahub/pkg/client"
	"mangahub/pkg/session"
)

const defaultAPIURL = "http://10.238.53.72:8080"

// Re-export session types for convenience
type Session = session.Session

// GetAPIURL returns the API URL from config or default
func GetAPIURL() string {
	// Could load from config file in future
	if url := os.Getenv("MANGAHUB_API_URL"); url != "" {
		return url
	}
	return defaultAPIURL
}

// Re-export session functions for backward compatibility
var (
	GetSessionPath = session.GetPath
	SaveSession    = session.Save
	LoadSession    = session.Load
	ClearSession   = session.Clear
)

// NewHTTPClient creates an HTTP client with optional auth token
func NewHTTPClient() *client.HTTPClient {
	apiURL := GetAPIURL()
	token := ""

	// Try to load session for auth token
	if sess, err := session.Load(); err == nil {
		token = sess.Token
	}

	return client.NewHTTPClient(apiURL, token)
}

// NewAuthenticatedHTTPClient creates an HTTP client and returns error if not logged in
func NewAuthenticatedHTTPClient() (*client.HTTPClient, *session.Session, error) {
	sess, err := session.Load()
	if err != nil {
		return nil, nil, err
	}

	apiURL := GetAPIURL()
	return client.NewHTTPClient(apiURL, sess.Token), sess, nil
}
