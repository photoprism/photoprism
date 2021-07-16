package service

import (
	"github.com/photoprism/photoprism/internal/face"
	"sync"
)

var onceFaceNet sync.Once

func initFaceNet() {
	services.FaceNet = face.NewNet(conf.FaceNetModelPath(), conf.DisableTensorFlow())
}

func FaceNet() *face.Net {
	onceFaceNet.Do(initFaceNet)

	return services.FaceNet
}
