package form

type IndexOptions struct {
	Path    string `json:"path"`
	Rescan  bool   `json:"rescan"`
	Cleanup bool   `json:"cleanup"`
}
