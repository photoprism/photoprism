package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceCleanUp sync.Once

func initCleanUp() {
	services.CleanUp = photoprism.NewCleanUp(Config())
}

func CleanUp() *photoprism.CleanUp {
	onceCleanUp.Do(initCleanUp)

	return services.CleanUp
}
