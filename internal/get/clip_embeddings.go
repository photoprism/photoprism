package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/clip_embeddings"
)

var onceSimilarity sync.Once

func initSimilarity() {
	services.ClipEmbeddings = clip_embeddings.New(Config().AssetsPath(), Config().DisableClip())
}

func ClipEmbeddings() *clip_embeddings.TensorFlow {
	onceSimilarity.Do(initSimilarity)

	return services.ClipEmbeddings
}
