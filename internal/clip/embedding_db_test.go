package clip

import (
	"testing"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/stretchr/testify/assert"
)

func embeddingDbTestEnv(t *testing.T) EmbeddingDB {
	db := EmbeddingDB{
		Url:        "http://embedding-db:6333",
		Collection: "clip_testing",
		VectorSize: 4,
	}
	// ignore err, since it's ok if collection does not exist
	db.DeleteCollection()
	err := db.CreateCollectionIfNotExisting()
	assert.Nil(t, err)
	return db
}

func TestSaveAndLoadEmbedding(t *testing.T) {
	db := embeddingDbTestEnv(t)
	embedding := face.Embedding{0.1, 0.2, 0.3, 0.4}
	var testId uint = 123

	t.Run("id=123 is not in database yet", func(t *testing.T) {
		notExistingEmbedding, err := db.LoadEmbedding(testId)
		assert.Error(t, err)
		assert.Nil(t, notExistingEmbedding)
	})
	t.Run("add id=123 to database", func(t *testing.T) {
		err := db.SaveEmbedding(embedding, testId)
		assert.NoError(t, err)
	})
	t.Run("load id=123 from database", func(t *testing.T) {
		actual, err := db.LoadEmbedding(testId)
		assert.NoError(t, err)
		// Note: we cannot compare embedding and actual directly, since we use Cosine distance.
		// 		 Not the embedding directly is saved, only a represantive resulting in the same cosine distance
		// 		 assert.Equal(t, embedding, actual)
		assert.InDelta(t, 1.0, actual.CosineSimilarity(embedding), 0.0001)
	})
	t.Run("delete id=123 from database", func(t *testing.T) {
		err := db.DeleteEmbedding(testId)
		assert.NoError(t, err)
	})
	t.Run("id=123 should be deleted from database", func(t *testing.T) {
		notExistingEmbedding, err := db.LoadEmbedding(testId)
		assert.Error(t, err, "embedding with id 'id=123' does not exist in database")
		assert.Nil(t, notExistingEmbedding)
	})
	t.Run("delete collection", func(t *testing.T) {
		err := db.DeleteCollection()
		assert.Nil(t, err)
	})
}

func TestKNearestNeighbors(t *testing.T) {
	db := embeddingDbTestEnv(t)
	candidate := face.Embedding{1, 1, 1, 1}
	closeNeighbors := []face.Embedding{{3, 3, 3, 3}, {0.9, 0.99, 0.99, 0.99}, {1, 1, 0.9, 0.8}}
	for idx, embedding := range closeNeighbors {
		err := db.SaveEmbedding(embedding, uint(idx))
		assert.NoError(t, err)
	}
	otherNeighbors := []face.Embedding{{-3, 1.2, 0, 2}, {17, -2, 9, 1}, {0.123, -2, 0, 0}}
	for idx, embedding := range otherNeighbors {
		err := db.SaveEmbedding(embedding, uint(idx)+10)
		assert.NoError(t, err)
	}
	actual, err := db.KNearestNeighbors(candidate, 3, 0.0)
	assert.Nil(t, err)
	assert.Equal(t, []uint64{0, 1, 2}, actual)
}
