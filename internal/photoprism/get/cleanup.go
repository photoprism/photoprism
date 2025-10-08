package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceCleanUp sync.Once

func initCleanUp() {
	services.CleanUp = photoprism.NewCleanUp(Config())
}

// CleanUp returns the singleton cleanup worker service instance.
func CleanUp() *photoprism.CleanUp {
	onceCleanUp.Do(initCleanUp)

	return services.CleanUp
}
