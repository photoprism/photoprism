package service

import (
	"net/http"
	"time"
)

// TestRequest makes a test request to the given URL and returns true if successful.
func (h Heuristic) TestRequest(method, rawUrl string) bool {
	req, err := http.NewRequest(method, rawUrl, nil)

	if err != nil {
		return false
	}

	// Add custom request headers:
	// https://github.com/photoprism/photoprism/pull/4608
	if len(h.Headers) > 0 {
		for key, val := range h.Headers {
			req.Header.Add(key, val)
		}
	}

	// Create new http.Client instance.
	//
	// NOTE: Timeout specifies a time limit for requests made by
	// this Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	client := &http.Client{Timeout: 30 * time.Second}

	// Send request to see if it fails.
	if resp, reqErr := client.Do(req); reqErr != nil {
		return false
	} else if resp.StatusCode < 400 {
		return true
	}

	return false
}
