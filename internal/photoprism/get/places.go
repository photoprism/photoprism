package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var oncePlaces sync.Once

func initPlaces() {
	services.Places = photoprism.NewPlaces(Config())
}

// Places returns the singleton place lookup service instance.
func Places() *photoprism.Places {
	oncePlaces.Do(initPlaces)

	return services.Places
}
