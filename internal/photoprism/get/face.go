package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/tensorflow/face"
)

var onceFaceNet sync.Once

func initFaceNet() {
	services.FaceNet = face.NewNet(conf.FaceNetModelPath(), "", conf.DisableFaces())
}

func FaceNet() *face.Net {
	onceFaceNet.Do(initFaceNet)

	return services.FaceNet
}
