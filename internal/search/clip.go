package search

import (
	"sync"

	"github.com/photoprism/photoprism/internal/clip"
)

var onceClip sync.Once

var clipInstance *clip.Clip

func initClip() {
	clipInstance = clip.New("clip", 512, false) // @TODO: read config
	clipInstance.Db.CreateCollectionIfNotExisting()
}

func Clip() *clip.Clip {
	onceClip.Do(initClip)
	return clipInstance
}
