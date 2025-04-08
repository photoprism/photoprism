package vision

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

type ModelType = string

var (
	AssetsPath        = fs.Abs("../../../assets")
	ServiceUri        = ""
	ServiceKey        = ""
	DownloadUrl       = ""
	DefaultResolution = 224
)

// NasnetModel is a standard TensorFlow model used for label generation.
var (
	NasnetModel = &Model{Name: "Nasnet", Version: "Mobile", Resolution: 224, Tags: []string{"photoprism"}}
)
