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
	t.Run("StampsEchoedModel", func(t *testing.T) {
		faces := face.Faces{{}}
		res := &ApiResponse{
			Model:  &Model{Name: face.ModelArcFaceR50},
			Result: ApiResult{Embeddings: []face.Embeddings{valid()}},
		}

		assert.Equal(t, 1, applyEndpointEmbeddings(faces, res, face.ModelFaceNet))
		assert.Equal(t, face.ModelArcFaceR50, faces[0].EmbedModel)
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
