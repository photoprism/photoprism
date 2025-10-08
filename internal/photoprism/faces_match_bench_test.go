package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// BenchmarkSelectBestFace measures matching performance when the bucketed index is used.
func BenchmarkSelectBestFace(b *testing.B) {
	const candidateCount = 1024

	faces := make(entity.Faces, 0, candidateCount)

	for i := 0; i < candidateCount; i++ {
		f := entity.NewFace("", entity.SrcAuto, face.RandomEmbeddings(5, face.RegularFace))
		faces = append(faces, *f)
	}

	markerEmb := face.RandomEmbeddings(3, face.RegularFace)
	faces[0] = *entity.NewFace("", entity.SrcAuto, markerEmb)

	index := buildFaceIndex(faces)
	hash := embeddingSignHashFromEmbeddings(markerEmb)
	bucketSize := len(index.buckets[hash])

	b.ResetTimer()

	b.ReportMetric(float64(bucketSize), "bucket_candidates")
	b.ReportMetric(float64(len(index.fallback)), "total_candidates")

	for i := 0; i < b.N; i++ {
		selectBestFace(markerEmb, index)
	}
}

// BenchmarkSelectBestFaceLegacy captures the legacy behaviour that scans every face.
func BenchmarkSelectBestFaceLegacy(b *testing.B) {
	const candidateCount = 1024

	faces := make(entity.Faces, 0, candidateCount)

	for i := 0; i < candidateCount; i++ {
		f := entity.NewFace("", entity.SrcAuto, face.RandomEmbeddings(5, face.RegularFace))
		faces = append(faces, *f)
	}

	markerEmb := face.RandomEmbeddings(3, face.RegularFace)
	faces[0] = *entity.NewFace("", entity.SrcAuto, markerEmb)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		legacySelectBestFace(markerEmb, faces)
	}
}

func legacySelectBestFace(marker face.Embeddings, faces entity.Faces) (*entity.Face, float64) {
	var best *entity.Face
	bestDist := -1.0

	for i := range faces {
		if ok, dist := faces[i].Match(marker); ok {
			if best == nil || dist < bestDist {
				best = &faces[i]
				bestDist = dist
			}
		}
	}

	return best, bestDist
}
