package video

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// Type represents a video format type.
type Type struct {
	Codec       Codec
	FileType    fs.Type
	WidthLimit  int
	HeightLimit int
	Public      bool
}
