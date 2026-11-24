package dl

// Subtitle youtube-dl subtitle
type Subtitle struct {
	URL      string `json:"url"`
	Ext      string `json:"ext"`
	Language string `json:"-"`
}
