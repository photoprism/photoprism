package form

type IndexOptions struct {
	CompleteRescan bool `json:"rescan"`
	CreateThumbs   bool `json:"thumbs"`
	ConvertRaw     bool `json:"raw"`
}
