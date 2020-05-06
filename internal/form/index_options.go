package form

type IndexOptions struct {
	Path    string `json:"path"`
	Convert bool   `json:"convert"`
	Rescan  bool   `json:"rescan"`
}
