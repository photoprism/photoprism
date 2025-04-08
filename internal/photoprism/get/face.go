package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/ai/face"
)

var onceFaceNet sync.Once

func initFaceNet() {
	services.FaceNet = face.NewModel(conf.FaceNetModelPath(), "", conf.DisableFaces())
}

func FaceNet() *face.Model {
	onceFaceNet.Do(initFaceNet)

	return services.FaceNet
}
