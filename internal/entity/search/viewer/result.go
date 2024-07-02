package viewer

import (
	"time"

	"github.com/photoprism/photoprism/internal/thumb"
)

// Result represents a photo viewer result.
type Result struct {
	UID          string       `json:"UID"`
	Title        string       `json:"Title"`
	TakenAtLocal time.Time    `json:"TakenAtLocal"`
	Description  string       `json:"Description"`
	Favorite     bool         `json:"Favorite"`
	Playable     bool         `json:"Playable"`
	DownloadUrl  string       `json:"DownloadUrl"`
	Width        int          `json:"Width"`
	Height       int          `json:"Height"`
	Thumbs       thumb.Public `json:"Thumbs"`
}

// Results represents a list of viewer search results.
type Results []Result
