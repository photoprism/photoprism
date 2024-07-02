package customize

// ImportSettings represents import settings.
type ImportSettings struct {
	Path string `json:"path" yaml:"Path"`
	Move bool   `json:"move" yaml:"Move"`
}
