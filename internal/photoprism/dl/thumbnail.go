package dl

// Thumbnail models a thumbnail entry returned by yt-dlp metadata.
type Thumbnail struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	Preference int    `json:"preference"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Resolution string `json:"resolution"`
}
