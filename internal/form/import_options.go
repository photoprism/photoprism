package form

// ImportOptions holds import path and album assignment flags.
type ImportOptions struct {
	Albums []string `json:"albums"`
	Path   string   `json:"path"`
	Move   bool     `json:"move"`
}
