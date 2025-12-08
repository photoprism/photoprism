//go:build debug

package hub

import (
	"os"
)

// init lets debug builds override the Hub base URL via PHOTOPRISM_HUB_URL so
// developers can point tests at staging services without code changes.
func init() {
	if debugUrl := os.Getenv("PHOTOPRISM_HUB_URL"); debugUrl != "" {
		SetBaseURL(debugUrl)
	}
}
