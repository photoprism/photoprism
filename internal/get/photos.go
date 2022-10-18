package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var oncePhotos sync.Once

func initPhotos() {
	services.Photos = photoprism.NewPhotos()
}

func Photos() *photoprism.Photos {
	oncePhotos.Do(initPhotos)

	return services.Photos
}
