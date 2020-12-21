package service

import (
	"sync"

	"github.com/photoprism/photoprism/internal/classify"
)

var onceClassify sync.Once

func initClassify() {
	services.Classify = classify.New(Config().AssetsPath(), Config().DisableTensorFlow())
}

func Classify() *classify.TensorFlow {
	onceClassify.Do(initClassify)

	return services.Classify
}
