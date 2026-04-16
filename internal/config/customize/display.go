package customize

// DisplaySettings represents display settings.
type DisplaySettings struct {
	Originals        bool                   `json:"originals" yaml:"Originals"`
	RetinaLightbox   bool                   `json:"retinaLightbox" yaml:"RetinaLightbox"`
	RetinaThumbnails bool                   `json:"retinaThumbnails" yaml:"RetinaThumbnails"`
	Metadata         MetadataLayoutSettings `json:"metadata" yaml:"Metadata"`
}
