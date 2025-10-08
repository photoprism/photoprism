package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceIndex sync.Once

func initIndex() {
	services.Index = photoprism.NewIndex(Config(), Convert(), Files(), Photos())
}

// Index returns the singleton indexing service, initializing it on first use.
func Index() *photoprism.Index {
	onceIndex.Do(initIndex)

	return services.Index
}
