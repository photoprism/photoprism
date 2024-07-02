package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/tensorflow/nsfw"
)

var onceNsfwDetector sync.Once

func initNsfwDetector() {
	services.Nsfw = nsfw.New(conf.NSFWModelPath())
}

func NsfwDetector() *nsfw.Detector {
	onceNsfwDetector.Do(initNsfwDetector)

	return services.Nsfw
}
