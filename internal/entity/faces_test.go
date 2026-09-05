package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/face"
)

func TestFaces_Embeddings(t *testing.T) {
	ValidateFixtures(t)
	m := FaceFixtures.Get("joe-biden")
	m1 := FaceFixtures.Get("jane-doe")
	r := Faces{m, m1}.Embeddings()
	len1 := len(m.Embedding())
	len2 := len(m1.Embedding())
	assert.Equal(t, len1+len2, len(r[0])+len(r[1]))
}

func TestFaces_IDs(t *testing.T) {
	ValidateFixtures(t)
	m := FaceFixtures.Get("joe-biden")
	m1 := FaceFixtures.Get("jane-doe")
	r := Faces{m, m1}.IDs()
	assert.Equal(t, []string{"VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG6", "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG7"}, r)
}

func TestDeleteOrphanFaces(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Ok", func(t *testing.T) {
		if count, err := DeleteOrphanFaces(); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("deleted %d faces", count)
		}
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Save(FaceFixtures.Pointer("jane-doe")).Error)
		})
	})
}

func TestFaces_EmbedModel(t *testing.T) {
	ValidateFixtures(t)
	t.Run("AllBlank", func(t *testing.T) {
		model, ok := Faces{{EmbedModel: ""}, {EmbedModel: ""}}.EmbedModel()
		assert.True(t, ok)
		assert.Empty(t, model)
	})
	t.Run("BlankAndFaceNet", func(t *testing.T) {
		model, ok := Faces{{EmbedModel: ""}, {EmbedModel: face.ModelFaceNet}}.EmbedModel()
		assert.True(t, ok)
		assert.Equal(t, face.ModelFaceNet, model)
	})
	t.Run("BlankAndSFace", func(t *testing.T) {
		_, ok := Faces{{EmbedModel: ""}, {EmbedModel: face.ModelSFace}}.EmbedModel()
		assert.False(t, ok)
	})
	t.Run("Mixed", func(t *testing.T) {
		_, ok := Faces{{EmbedModel: face.ModelFaceNet}, {EmbedModel: face.ModelSFace}}.EmbedModel()
		assert.False(t, ok)
	})
}
