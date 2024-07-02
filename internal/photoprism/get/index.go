package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceIndex sync.Once

func initIndex() {
	services.Index = photoprism.NewIndex(Config(), Classify(), NsfwDetector(), FaceNet(), Convert(), Files(), Photos())
}

func Index() *photoprism.Index {
	onceIndex.Do(initIndex)

	return services.Index
}
