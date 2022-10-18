package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceMoments sync.Once

func initMoments() {
	services.Moments = photoprism.NewMoments(Config())
}

func Moments() *photoprism.Moments {
	onceMoments.Do(initMoments)

	return services.Moments
}
