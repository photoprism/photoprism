package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceThumbs sync.Once

func initThumbs() {
	services.Thumbs = photoprism.NewThumbs(Config())
}

func Thumbs() *photoprism.Thumbs {
	onceThumbs.Do(initThumbs)

	return services.Thumbs
}
