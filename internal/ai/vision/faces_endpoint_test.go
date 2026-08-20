package vision

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/face"
)

func TestApplyEndpointEmbeddings(t *testing.T) {
	// FaceNet is 512-wide, so a full-length vector is needed for the accepted cases.
	valid := func() face.Embeddings {
		return face.Embeddings{make(face.Embedding, 512)}
	}

	t.Run("NilResponse", func(t *testing.T) {
		faces := face.Faces{{}}
		assert.Zero(t, applyEndpointEmbeddings(faces, nil, face.ModelFaceNet))
	})
	t.Run("StampsConfiguredModel", func(t *testing.T) {
		faces := face.Faces{{}}
		res := &ApiResponse{Result: ApiResult{Embeddings: []face.Embeddings{valid()}}}

		assert.Equal(t, 1, applyEndpointEmbeddings(faces, res, face.ModelFaceNet))
		assert.Equal(t, face.ModelFaceNet, faces[0].EmbedModel)
	})
	t.Run("AcceptsMatchingEcho", func(t *testing.T) {
		faces := face.Faces{{}}
		res := &ApiResponse{
			Model:  &Model{Name: face.ModelFaceNet},
			Result: ApiResult{Embeddings: []face.Embeddings{valid()}},
		}

		assert.Equal(t, 1, applyEndpointEmbeddings(faces, res, face.ModelFaceNet))
		assert.Equal(t, face.ModelFaceNet, faces[0].EmbedModel)
	})
	t.Run("RefusesEchoedOtherModel", func(t *testing.T) {
		// arcface_r50 is 512-wide like FaceNet, so the length proves nothing here. A
		// service reporting another model produced vectors of another space.
		faces := face.Faces{{}}
		res := &ApiResponse{
			Model:  &Model{Name: face.ModelArcFaceR50},
			Result: ApiResult{Embeddings: []face.Embeddings{valid()}},
		}

		assert.Zero(t, applyEndpointEmbeddings(faces, res, face.ModelFaceNet))
		assert.Empty(t, faces[0].EmbedModel)
		assert.Empty(t, faces[0].Embeddings)
	})
	t.Run("WidthFollowsConfiguredModel", func(t *testing.T) {
		// SFace is 128-wide, so a 512-value vector is not one of its embeddings whatever
		// the response claims about it.
		faces := face.Faces{{}}
		res := &ApiResponse{
			Model:  &Model{Name: face.ModelAuraFace},
			Result: ApiResult{Embeddings: []face.Embeddings{valid()}},
		}

		assert.Zero(t, applyEndpointEmbeddings(faces, res, face.ModelSFace))
		assert.Empty(t, faces[0].EmbedModel)
	})
	t.Run("RejectsWrongLength", func(t *testing.T) {
		faces := face.Faces{{}}
		res := &ApiResponse{Result: ApiResult{Embeddings: []face.Embeddings{{{0.1, 0.2}}}}}

		assert.Zero(t, applyEndpointEmbeddings(faces, res, face.ModelFaceNet))
		assert.Empty(t, faces[0].EmbedModel)
		assert.Empty(t, faces[0].Embeddings)
	})
	t.Run("RejectsNaN", func(t *testing.T) {
		values := valid()
		values[0][0] = math.NaN()

		faces := face.Faces{{}}
		res := &ApiResponse{Result: ApiResult{Embeddings: []face.Embeddings{values}}}

		assert.Zero(t, applyEndpointEmbeddings(faces, res, face.ModelFaceNet))
	})
	t.Run("RejectsMultipleEmbeddings", func(t *testing.T) {
		faces := face.Faces{{}}
		res := &ApiResponse{Result: ApiResult{Embeddings: []face.Embeddings{{make(face.Embedding, 512), make(face.Embedding, 512)}}}}

		assert.Zero(t, applyEndpointEmbeddings(faces, res, face.ModelFaceNet))
	})
	t.Run("DropsWhenModelUnknown", func(t *testing.T) {
		faces := face.Faces{{}}
		res := &ApiResponse{Result: ApiResult{Embeddings: []face.Embeddings{valid()}}}

		assert.Zero(t, applyEndpointEmbeddings(faces, res, ""))
		assert.Empty(t, faces[0].EmbedModel)
	})
}
