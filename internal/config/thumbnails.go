package config

// Thumbnail gives direct access to width and height for a thumbnail setting
type Thumbnail struct {
	Name   string
	Width  int
	Height int
}

// Thumbnails is a list of default thumbnail size available for the app
var Thumbnails []Thumbnail
