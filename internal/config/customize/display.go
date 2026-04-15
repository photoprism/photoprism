package customize

// DisplaySettings represents display settings.
type DisplaySettings struct {
	Originals        bool `json:"originals" yaml:"Originals"`
	ImagePacking     bool `json:"imagePacking" yaml:"ImagePacking"`
	RetinaLightbox   bool `json:"retinaLightbox" yaml:"RetinaLightbox"`
	RetinaThumbnails bool `json:"retinaThumbnails" yaml:"RetinaThumbnails"`
}
