package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var oncePlaces sync.Once

func initPlaces() {
	services.Places = photoprism.NewPlaces(Config())
}

func Places() *photoprism.Places {
	oncePlaces.Do(initPlaces)

	return services.Places
}
