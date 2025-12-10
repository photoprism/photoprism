package form

// IndexOptions configure index paths and maintenance flags.
type IndexOptions struct {
	Path    string `json:"path"`
	Rescan  bool   `json:"rescan"`
	Cleanup bool   `json:"cleanup"`
}
