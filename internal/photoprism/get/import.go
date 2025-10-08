package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceImport sync.Once

func initImport() {
	services.Import = photoprism.NewImport(Config(), Index(), Convert())
}

// Import returns the singleton import service instance.
func Import() *photoprism.Import {
	onceImport.Do(initImport)

	return services.Import
}
