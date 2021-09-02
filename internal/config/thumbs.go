package config

// ThumbType represents thumbnail info for use in client apps.
type ThumbType struct {
	Size   string `json:"size"`
	Use    string `json:"use"`
	Width  int    `json:"w"`
	Height int    `json:"h"`
}

// ThumbTypes represents a list of thumbnail types.
type ThumbTypes []ThumbType

// Thumbs is a list of thumbnails for use in client apps.
var Thumbs ThumbTypes
