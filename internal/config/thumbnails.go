package config

// ThumbSize represents thumbnail info for use in client apps.
type ThumbSize struct {
	Size   string `json:"size"`
	Usage  string `json:"usage"`
	Width  int    `json:"w"`
	Height int    `json:"h"`
}

// ThumbSizes represents a list of thumbnail types.
type ThumbSizes []ThumbSize

// Thumbs is a list of thumbnails for use in client apps.
var Thumbs ThumbSizes
