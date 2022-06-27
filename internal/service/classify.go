package service

import (
	"sync"

	"github.com/photoprism/photoprism/internal/classify"
)

var onceClassifyTensorflow, onceClassifyDeepstack sync.Once

func initClassify() {
	services.Classify = classify.New(Config().AssetsPath(), Config().DisableTensorFlow())
}

func initDeepStackClassify() {
	services.DeepStackClassify = classify.DeepStackNew(Config().DeepStackApiUrl(), Config().DisableDeepStack())
}

func Classify() *classify.TensorFlow {
	onceClassifyTensorflow.Do(initClassify)

	return services.Classify
}

func DeepStackClassify() *classify.DeepStack {
	onceClassifyDeepstack.Do(initDeepStackClassify)

	return services.DeepStackClassify
}
