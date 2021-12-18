package viewer

import (
	"time"
)

// Results represents a list of viewer search results.
type Results []Result

// Result represents a photo viewer result.
type Result struct {
	UID         string    `json:"uid"`
	Title       string    `json:"title"`
	Taken       time.Time `json:"taken"`
	Description string    `json:"description"`
	Favorite    bool      `json:"favorite"`
	Playable    bool      `json:"playable"`
	DownloadUrl string    `json:"download_url""`
	OriginalW   int       `json:"original_w"`
	OriginalH   int       `json:"original_h"`
	Fit720      Thumb     `json:"fit_720"`
	Fit1280     Thumb     `json:"fit_1280"`
	Fit1920     Thumb     `json:"fit_1920"`
	Fit2048     Thumb     `json:"fit_2048"`
	Fit2560     Thumb     `json:"fit_2560"`
	Fit3840     Thumb     `json:"fit_3840"`
	Fit4096     Thumb     `json:"fit_4096"`
	Fit7680     Thumb     `json:"fit_7680"`
}
