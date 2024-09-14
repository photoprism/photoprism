package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/ai/classify"
)

var onceClassify sync.Once

func initClassify() {
	services.Classify = classify.New(Config().AssetsPath(), Config().DisableClassification())
}

func Classify() *classify.TensorFlow {
	onceClassify.Do(initClassify)

	return services.Classify
}
