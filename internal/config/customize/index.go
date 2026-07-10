package customize

// IndexSettings represents indexing settings.
type IndexSettings struct {
	Path         string `json:"path" yaml:"Path"`
	Convert      bool   `json:"convert" yaml:"Convert"`
	Rescan       bool   `json:"rescan" yaml:"Rescan"`
	SkipArchived bool   `json:"skipArchived" yaml:"SkipArchived"`
}
