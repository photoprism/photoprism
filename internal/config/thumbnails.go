package config

// Thumbnail gives direct access to width and height for a thumbnail setting
type Thumbnail struct {
	Size   string `json:"size"`
	Use    string `json:"use"`
	Width  int    `json:"w"`
	Height int    `json:"h"`
}

// Thumbnails is a list of default thumbnail size available for the app
var Thumbnails []Thumbnail
