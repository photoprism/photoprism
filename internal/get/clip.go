package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/clip"
)

var onceClip sync.Once

func initClip() {
	services.Clip = clip.New("clip", 512, false) // @TODO: read config
	services.Clip.Db.CreateCollectionIfNotExisting()
}

func Clip() *clip.Clip {
	onceClip.Do(initClip)
	return services.Clip
}
