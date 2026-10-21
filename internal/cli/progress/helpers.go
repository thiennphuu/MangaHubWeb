package progress

import (
	"fmt"
	"os"

	"mangahub/pkg/client"
	"mangahub/pkg/session"
)

// getAPIURL returns the API server URL
func getAPIURL() string {
	if url := os.Getenv("MANGAHUB_API_URL"); url != "" {
		return url
	}
	return "http://10.238.53.72:8080"
}

// newAuthenticatedHTTPClient creates an HTTP client with auth token from session
func newAuthenticatedHTTPClient() (*client.HTTPClient, *session.Session, error) {
	sess, err := session.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("not logged in: %w", err)
	}

	httpClient := client.NewHTTPClient(getAPIURL(), sess.Token)
	return httpClient, sess, nil
}
