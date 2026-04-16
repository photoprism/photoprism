package customize

// MetadataLayoutSettings represents per-view metadata field layouts.
type MetadataLayoutSettings struct {
	Cards    []string `json:"cards" yaml:"Cards"`
	List     []string `json:"list" yaml:"List"`
	Lightbox []string `json:"lightbox" yaml:"Lightbox"`
}

// NewMetadataLayoutSettings creates metadata layout settings with default fields.
func NewMetadataLayoutSettings() MetadataLayoutSettings {
	return MetadataLayoutSettings{
		Cards: []string{
			"caption",
			"date",
			"keywords",
		},
		List: []string{
			"filename",
			"date",
			"camera",
			"lens",
			"exposure",
		},
		Lightbox: []string{
			"date",
			"caption",
			"keywords",
			"camera",
			"lens",
			"exposure",
			"filename",
			"fileInfo",
		},
	}
}
