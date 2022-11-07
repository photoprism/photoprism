package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var oncePurge sync.Once

func initPurge() {
	services.Purge = photoprism.NewPurge(Config(), Files())
}

func Purge() *photoprism.Purge {
	oncePurge.Do(initPurge)

	return services.Purge
}
