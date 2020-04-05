package service

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceResample sync.Once

func initResample() {
	services.Resample = photoprism.NewResample(Config())
}

func Resample() *photoprism.Resample {
	onceResample.Do(initResample)

	return services.Resample
}
