package form

type ImportOptions struct {
	Albums []string `json:"albums"`
	Path   string   `json:"path"`
	Move   bool     `json:"move"`
}
