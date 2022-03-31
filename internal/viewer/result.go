package viewer

import (
	"time"
)

// Thumbs represents photo viewer thumbs in different sizes.
type Thumbs struct {
	Fit720  Thumb `json:"fit_720"`
	Fit1280 Thumb `json:"fit_1280"`
	Fit1920 Thumb `json:"fit_1920"`
	Fit2048 Thumb `json:"fit_2048"`
	Fit2560 Thumb `json:"fit_2560"`
	Fit3840 Thumb `json:"fit_3840"`
	Fit4096 Thumb `json:"fit_4096"`
	Fit7680 Thumb `json:"fit_7680"`
}

// Result represents a photo viewer result.
type Result struct {
	UID          string    `json:"UID"`
	Title        string    `json:"Title"`
	TakenAtLocal time.Time `json:"TakenAtLocal"`
	Description  string    `json:"Description"`
	Favorite     bool      `json:"Favorite"`
	Playable     bool      `json:"Playable"`
	DownloadUrl  string    `json:"DownloadUrl"`
	Width        int       `json:"Width"`
	Height       int       `json:"Height"`
	Thumbs       Thumbs    `json:"Thumbs"`
}

// Results represents a list of viewer search results.
type Results []Result
