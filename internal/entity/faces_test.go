package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/face"
)

func TestFaces_Embeddings(t *testing.T) {
	m := FaceFixtures.Get("joe-biden")
	m1 := FaceFixtures.Get("jane-doe")
	r := Faces{m, m1}.Embeddings()
	len1 := len(m.Embedding())
	len2 := len(m1.Embedding())
	assert.Equal(t, len1+len2, len(r[0])+len(r[1]))
}

func TestFaces_IDs(t *testing.T) {
	m := FaceFixtures.Get("joe-biden")
	m1 := FaceFixtures.Get("jane-doe")
	r := Faces{m, m1}.IDs()
	assert.Equal(t, []string{"VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG6", "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG7"}, r)
}

func TestDeleteOrphanFaces(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		if count, err := DeleteOrphanFaces(); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("deleted %d faces", count)
		}
	})
}

func TestFaces_EmbedModel(t *testing.T) {
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

func TestFaces_CollisionBound(t *testing.T) {
	// Comfortably above CollisionDist, so the floor is not what any of these cases turn on.
	const near, far = 0.30, 0.50

	t.Run("None", func(t *testing.T) {
		radius, collisions := Faces{{ID: "A"}, {ID: "B"}}.CollisionBound(0.1)

		assert.Zero(t, radius)
		assert.Zero(t, collisions)
	})
	t.Run("Tightest", func(t *testing.T) {
		// The count travels with the radius it belongs to, so the row reports one bound and
		// the number of collisions that produced it rather than a mixture of both sources.
		radius, collisions := Faces{
			{ID: "A", CollisionRadius: far, Collisions: 1},
			{ID: "B", CollisionRadius: near, Collisions: 3},
		}.CollisionBound(0.1)

		assert.InDelta(t, near, radius, 1e-9)
		assert.Equal(t, 3, collisions)
	})
	t.Run("OrderIndependent", func(t *testing.T) {
		a := Faces{{ID: "A", CollisionRadius: far, Collisions: 1}, {ID: "B", CollisionRadius: near, Collisions: 3}}
		b := Faces{a[1], a[0]}

		forward, forwardCount := a.CollisionBound(0.1)
		reverse, reverseCount := b.CollisionBound(0.1)

		assert.InDelta(t, forward, reverse, 1e-9)
		assert.Equal(t, forwardCount, reverseCount)
	})
	t.Run("IgnoresInactive", func(t *testing.T) {
		// At or below CollisionDist nothing enforces the radius, so inheriting one would put a
		// bound on the merged cluster that its sources never had.
		radius, collisions := Faces{{ID: "A", CollisionRadius: face.CollisionDist, Collisions: 2}}.CollisionBound(0.1)

		assert.Zero(t, radius)
		assert.Zero(t, collisions)
	})
	t.Run("DroppedBelowExtent", func(t *testing.T) {
		radius, collisions := Faces{{ID: "A", CollisionRadius: near, Collisions: 2}}.CollisionBound(near + 0.01)

		assert.Zero(t, radius)
		assert.Zero(t, collisions)
	})
	t.Run("KeptAtExtent", func(t *testing.T) {
		// Match refuses beyond the radius rather than at it, so a bound exactly on the extent
		// still holds every member.
		radius, _ := Faces{{ID: "A", CollisionRadius: near, Collisions: 2}}.CollisionBound(near)

		assert.InDelta(t, near, radius, 1e-9)
	})
}
