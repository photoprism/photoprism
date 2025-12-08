package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceConvert sync.Once

func initConvert() {
	services.Convert = photoprism.NewConvert(Config())
}

// Convert returns the singleton media conversion service instance.
func Convert() *photoprism.Convert {
	onceConvert.Do(initConvert)

	return services.Convert
}
