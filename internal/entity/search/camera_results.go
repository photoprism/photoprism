package search

import (
	"time"
)

// Camera represents a camera search result.
type Camera struct {
	ID                uint      `json:"ID"`
	CameraSlug        string    `json:"Slug"`
	CameraName        string    `json:"Name"`
	CameraMake        string    `json:"Make"`
	CameraModel       string    `json:"Model"`
	CameraType        string    `json:"Type"`
	CameraDescription string    `json:"Description"`
	CameraNotes       string    `json:"Notes"`
	CreatedAt         time.Time `json:"CreatedAt"`
	UpdatedAt         time.Time `json:"UpdatedAt"`
	DeletedAt         time.Time `json:"DeletedAt"`
}
