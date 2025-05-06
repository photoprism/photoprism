package ytdl

// Subtitle youtube-dl subtitle
type Subtitle struct {
	URL      string `json:"url"`
	Ext      string `json:"ext"`
	Language string `json:"-"`
	// don't unmarshal, populated from subtitle file
	Bytes []byte `json:"-"`
}
