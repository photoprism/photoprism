package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
)

func TestClusterScoreCond(t *testing.T) {
	restore := face.ClusterScoreThreshold
	t.Cleanup(func() { face.ClusterScoreThreshold = restore })

	t.Run("ExplicitFloor", func(t *testing.T) {
		cond, args := ClusterScoreCond("", 42)
		assert.Equal(t, "score >= ?", cond)
		assert.Equal(t, []any{42}, args)
	})
	t.Run("NoFloor", func(t *testing.T) {
		cond, args := ClusterScoreCond("", 0)
		assert.Equal(t, "1 = 1", cond)
		assert.Empty(t, args)
	})
	t.Run("PerDetector", func(t *testing.T) {
		face.ClusterScoreThreshold = 0

		cond, args := ClusterScoreCond("", face.ClusterScoreAuto)

		assert.Contains(t, cond, "score >= CASE detect_model")
		assert.Contains(t, cond, "ELSE ? END")
		assert.NotEmpty(t, args)
		assert.Equal(t, strings.Count(cond, "?"), len(args), "every placeholder needs an argument")
		// A detector the registry does not name, and a row with none recorded, take the shared
		// default through ELSE - which is why the fragment needs no COALESCE.
		assert.Equal(t, face.ClusterScoreThresholdDefault, args[len(args)-1])
	})
	// The fragment is nested where another markers row is already in scope, so an unqualified
	// column would bind to the wrong one - or be rejected as ambiguous.
	t.Run("QualifiesEveryColumn", func(t *testing.T) {
		face.ClusterScoreThreshold = 0

		cond, args := ClusterScoreCond("m2", face.ClusterScoreAuto)

		require.Contains(t, cond, "m2.score")
		require.Contains(t, cond, "m2.detect_model")
		assert.Equal(t, strings.Count(cond, "?"), len(args))

		for _, col := range []string{"score", "detect_model"} {
			for _, before := range []string{"(", " ", ","} {
				assert.NotContains(t, cond, before+col, "%s must not appear unqualified", col)
			}
		}
	})
	t.Run("OperatorThresholdOutranksTheDetectors", func(t *testing.T) {
		face.ClusterScoreThreshold = 55

		cond, args := ClusterScoreCond("m2", face.ClusterScoreAuto)

		assert.Equal(t, "m2.score >= ?", cond)
		assert.Equal(t, []any{55}, args)
	})
}

func TestEmbeddingModelCond(t *testing.T) {
	t.Run("NoModel", func(t *testing.T) {
		cond, args := EmbeddingModelCond("")
		assert.Empty(t, cond, "an unset model restricts nothing")
		assert.Empty(t, args)
	})
	// A row with no recorded model is FaceNet's, so it has to stay selectable under that name.
	t.Run("FaceNet", func(t *testing.T) {
		cond, args := EmbeddingModelCond(face.ModelFaceNet)
		assert.Equal(t, strings.Count(cond, "?"), len(args))
		assert.Contains(t, cond, "embed_model = ''")
		assert.Equal(t, []any{face.ModelFaceNet, face.ModelFaceNet, face.ModelFaceNet}, args)
	})
	t.Run("OtherModel", func(t *testing.T) {
		cond, args := EmbeddingModelCond(face.ModelSFace)
		assert.Equal(t, strings.Count(cond, "?"), len(args))
		assert.Equal(t, []any{face.ModelSFace, face.ModelSFace, face.ModelFaceNet}, args)
	})
}

func TestFaceMemberCond(t *testing.T) {
	cond, args := FaceMemberCond()

	assert.Contains(t, cond, "marker_invalid = FALSE")
	assert.Contains(t, cond, "face_id <> ''")
	assert.Contains(t, cond, "LENGTH(embeddings_json) > 0")
	assert.Equal(t, []any{MarkerFace}, args)
}
