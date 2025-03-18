package viewer

import (
	"time"

	"github.com/photoprism/photoprism/internal/thumb"
)

// Result represents a photo viewer result.
type Result struct {
	UID          string        `json:"UID"`
	Type         string        `json:"Type,omitempty"`
	Title        string        `json:"Title,omitempty"`
	Caption      string        `json:"Caption,omitempty"`
	Lat          float64       `json:"Lat,omitempty"`
	Lng          float64       `json:"Lng,omitempty"`
	TakenAtLocal time.Time     `json:"TakenAtLocal"`
	Favorite     bool          `json:"Favorite"`
	Playable     bool          `json:"Playable"`
	Duration     time.Duration `json:"Duration,omitempty"`
	Width        int           `json:"Width"`
	Height       int           `json:"Height"`
	Hash         string        `json:"Hash"`
	Codec        string        `json:"Codec,omitempty"`
	Mime         string        `json:"Mime,omitempty"`
	Thumbs       *thumb.Viewer `json:"Thumbs,omitempty"`
	DownloadUrl  string        `json:"DownloadUrl,omitempty"`
}

// Results represents a list of viewer search results.
type Results []Result
