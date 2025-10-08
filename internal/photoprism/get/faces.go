package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceFaces sync.Once

func initFaces() {
	services.Faces = photoprism.NewFaces(Config())
}

// Faces returns the singleton face-recognition service instance.
func Faces() *photoprism.Faces {
	onceFaces.Do(initFaces)

	return services.Faces
}
