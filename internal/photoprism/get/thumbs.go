package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceThumbs sync.Once

func initThumbs() {
	services.Thumbs = photoprism.NewThumbs(Config())
}

// Thumbs returns the singleton thumbs service instance.
func Thumbs() *photoprism.Thumbs {
	onceThumbs.Do(initThumbs)

	return services.Thumbs
}
