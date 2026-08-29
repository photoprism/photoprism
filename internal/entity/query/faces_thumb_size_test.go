package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestClusterSizeGateReadSites pins that every query gating on the clustering size reads the extent
// an embedding was sampled at rather than the detection size.
//
// Each marker's effect is measured on its own against a baseline. Counting both at once cannot tell
// the two gates apart, because the pair is a mirror: whichever column is read, exactly one passes.
func TestClusterSizeGateReadSites(t *testing.T) {
	const floor = 112

	model := face.EmbeddingModelName()

	addMarker := func(t *testing.T, size, thumbSize int) {
		t.Helper()

		m := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        "fs6sg6bw45bnlqdw",
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			Size:           size,
			ThumbSize:      thumbSize,
			Score:          100,
			EmbedModel:     model,
			DetectModel:    face.DetectorYuNet,
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
			W:              0.1,
			H:              0.1,
		}

		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(m) })
	}

	counts := func(t *testing.T) (embeddings, clusterable, newMarkers int) {
		t.Helper()

		e, err := Embeddings(false, false, floor, face.ClusterScoreAuto, model)
		require.NoError(t, err)

		return len(e), CountFaceClusterGates(model, floor, face.ClusterScoreAuto).Clusterable,
			CountNewFaceMarkers(floor, face.ClusterScoreAuto)
	}

	baseEmb, baseClusterable, baseNew := counts(t)

	t.Run("AdmittedBySampledExtent", func(t *testing.T) {
		// Small in the detection thumbnail, sampled well above the bar from a wider rendition.
		// A site still reading markers.size refuses it and every count stays flat.
		addMarker(t, floor-52, floor+38)

		emb, clusterable, newMarkers := counts(t)

		assert.Equal(t, baseEmb+1, emb, "query.Embeddings must select it")
		assert.Equal(t, baseClusterable+1, clusterable, "CountFaceClusterGates must count it")
		assert.Equal(t, baseNew+1, newMarkers, "CountNewFaceMarkers must count it")
	})
	t.Run("RefusedDespiteDetectionSize", func(t *testing.T) {
		// The mirror: prominent in the detection thumbnail, sampled below the bar. A site reading
		// markers.size admits it, so every count moves where it must not.
		addMarker(t, floor+38, floor-52)

		emb, clusterable, newMarkers := counts(t)

		assert.Equal(t, baseEmb, emb, "query.Embeddings must skip it")
		assert.Equal(t, baseClusterable, clusterable, "CountFaceClusterGates must skip it")
		assert.Equal(t, baseNew, newMarkers, "CountNewFaceMarkers must skip it")
	})
	t.Run("MarkerClusterable", func(t *testing.T) {
		restore := face.ClusterSizeThreshold
		t.Cleanup(func() { face.ClusterSizeThreshold = restore })
		face.ClusterSizeThreshold = floor

		assert.True(t, (&entity.Marker{MarkerType: entity.MarkerFace, Size: floor - 52, ThumbSize: floor + 38, Score: 100}).Clusterable())
		assert.False(t, (&entity.Marker{MarkerType: entity.MarkerFace, Size: floor + 38, ThumbSize: floor - 52, Score: 100}).Clusterable())
	})
}
