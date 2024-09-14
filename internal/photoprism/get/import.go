package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceImport sync.Once

func initImport() {
	services.Import = photoprism.NewImport(Config(), Index(), Convert())
}

func Import() *photoprism.Import {
	onceImport.Do(initImport)

	return services.Import
}
