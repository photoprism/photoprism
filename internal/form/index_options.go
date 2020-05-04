package form

type IndexOptions struct {
	Convert  bool `json:"convert"`
	Resample bool `json:"resample"`
	Rescan   bool `json:"rescan"`
}
