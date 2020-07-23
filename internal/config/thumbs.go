package config

// Thumb represents thumbnail info for use in client apps.
type Thumb struct {
	Size   string `json:"size"`
	Use    string `json:"use"`
	Width  int    `json:"w"`
	Height int    `json:"h"`
}

// Thumbs is a list of thumbnails for use in client apps.
var Thumbs []Thumb
