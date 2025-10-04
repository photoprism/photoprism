package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// BenchmarkClusterMaterialize compares pre-sized versus legacy cluster materialisation.
func BenchmarkClusterMaterialize(b *testing.B) {
	const (
		clusterCount         = 64
		embeddingsPerCluster = 32
	)

	total := clusterCount * embeddingsPerCluster

	embeddings := make(face.Embeddings, total)
	for i := range embeddings {
		embeddings[i] = face.RandomEmbedding()
	}

	guesses := make([]int, total)
	sizes := make([]int, clusterCount)

	for cluster := 0; cluster < clusterCount; cluster++ {
		for j := 0; j < embeddingsPerCluster; j++ {
			idx := cluster*embeddingsPerCluster + j
			guesses[idx] = cluster + 1
			sizes[cluster]++
		}
	}

	b.Run("preSized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			results := make([]face.Embeddings, clusterCount)
			for k := range sizes {
				results[k] = make(face.Embeddings, 0, sizes[k])
			}
			for idx, n := range guesses {
				if n < 1 {
					continue
				}
				results[n-1] = append(results[n-1], embeddings[idx])
			}
		}
	})

	b.Run("legacy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			results := make([]face.Embeddings, clusterCount)
			for idx, n := range guesses {
				if n < 1 {
					continue
				}
				results[n-1] = append(results[n-1], embeddings[idx])
			}
		}
	})
}
