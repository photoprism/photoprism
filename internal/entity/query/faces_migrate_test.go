package query

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestFaceMigrationCounts(t *testing.T) {
	result, err := FaceMigrationCounts(face.ModelFaceNet)
	require.NoError(t, err)
	assert.Positive(t, result.Total)
	assert.Equal(t, result.Total, result.Valid+result.Invalid)
	assert.LessOrEqual(t, result.Ready, result.Valid)

	_, err = FaceMigrationCounts("")
	require.Error(t, err)
}

func TestFaceMigrationFileUIDs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := FaceMigrationFileUIDs("", 2)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(result), 2)
		assert.NotEmpty(t, result)
	})
	t.Run("InvalidLimit", func(t *testing.T) {
		_, err := FaceMigrationFileUIDs("", 0)
		require.Error(t, err)
	})
}

func TestFaceMigrationMarkers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := FaceMigrationMarkers("fs6sg6bw45bnlqdw")
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})
	t.Run("MissingUID", func(t *testing.T) {
		_, err := FaceMigrationMarkers("")
		require.Error(t, err)
	})
}

func TestFaceMigrationManualIdentities(t *testing.T) {
	result, err := FaceMigrationManualIdentities()
	require.NoError(t, err)
	for _, identity := range result {
		assert.Equal(t, entity.SrcManual, identity.SubjSrc)
		assert.NotEmpty(t, identity.MarkerUID)
	}
}

func TestFaceMigrationManualMarkers(t *testing.T) {
	marker := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        "fs6sg6bw45bnlqdw",
		MarkerType:     entity.MarkerFace,
		MarkerSrc:      entity.SrcImage,
		SubjUID:        rnd.GenerateUID('j'),
		SubjSrc:        entity.SrcManual,
		EmbedModel:     face.ModelFaceNet,
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	require.NoError(t, entity.Db().Create(marker).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(marker) })

	result, err := FaceMigrationManualMarkers(face.ModelFaceNet)
	require.NoError(t, err)
	found := false
	for _, candidate := range result {
		found = found || candidate.MarkerUID == marker.MarkerUID
	}
	assert.True(t, found)

	_, err = FaceMigrationManualMarkers("")
	require.Error(t, err)
}

func TestSaveFaceMigrationEmbeddings(t *testing.T) {
	marker := &entity.Marker{
		MarkerUID:  rnd.GenerateUID('m'),
		FileUID:    "fs6sg6bw45bnlqdw",
		MarkerType: entity.MarkerFace,
		MarkerSrc:  entity.SrcImage,
		FaceID:     "old-face",
		FaceDist:   0.2,
		W:          0.1,
		H:          0.1,
	}
	require.NoError(t, entity.Db().Create(marker).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(marker) })

	embeddings := face.Embeddings{face.RandomEmbedding()}
	require.NoError(t, SaveFaceMigrationEmbeddings(face.ModelFaceNet, map[string]face.Embeddings{marker.MarkerUID: embeddings}))

	stored, err := MarkerByUID(marker.MarkerUID)
	require.NoError(t, err)
	assert.Equal(t, face.ModelFaceNet, stored.EmbedModel)
	assert.Empty(t, stored.FaceID)
	assert.True(t, stored.Embeddings().One())
	assert.Len(t, stored.Embeddings()[0], len(embeddings[0]))

	require.Error(t, SaveFaceMigrationEmbeddings("", nil))
	require.Error(t, SaveFaceMigrationEmbeddings(face.ModelFaceNet, map[string]face.Embeddings{"": nil}))
}

func TestFinalizeFaceMigration(t *testing.T) {
	restore := face.ConfiguredModel()
	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelFaceNet, Model: face.FindEmbeddingModel(face.ModelFaceNet)}))
	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	originalDb := entity.Db()
	require.NotNil(t, originalDb)

	tempConn := &entity.DbConn{Driver: dsn.DriverSQLite3, Dsn: filepath.Join(t.TempDir(), "faces-migrate.db")}
	tempDb := tempConn.Db()
	require.NotNil(t, tempDb)
	require.NoError(t, tempDb.AutoMigrate(&entity.Face{}, &entity.Marker{}).Error)

	entity.SetDbProvider(tempConn)
	t.Cleanup(func() {
		entity.SetDbProvider(staticDbProvider{db: originalDb})
		tempConn.Close()
	})

	subjectUID := rnd.GenerateUID('j')
	manual := entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		MarkerType:     entity.MarkerFace,
		MarkerName:     "Alice",
		SubjUID:        subjectUID,
		SubjSrc:        entity.SrcManual,
		EmbedModel:     face.ModelFaceNet,
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	automatic := entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        "file123",
		MarkerType:     entity.MarkerFace,
		MarkerName:     "Alice",
		SubjUID:        subjectUID,
		SubjSrc:        entity.SrcAuto,
		EmbedModel:     face.ModelFaceNet,
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	require.NoError(t, tempDb.Create(&manual).Error)
	require.NoError(t, tempDb.Create(&automatic).Error)

	identities, err := FaceMigrationManualIdentities()
	require.NoError(t, err)
	cluster := entity.NewFace(subjectUID, entity.SrcManual, manual.Embeddings(), face.EmbeddingModelName())
	require.NotNil(t, cluster)

	require.NoError(t, FinalizeFaceMigration(face.ModelFaceNet, identities, []FaceMigrationCluster{{
		Face:            *cluster,
		MarkerDistances: map[string]float64{manual.MarkerUID: 0},
	}}, []string{automatic.MarkerUID}))

	var storedManual, storedAuto entity.Marker
	require.NoError(t, tempDb.First(&storedManual, "marker_uid = ?", manual.MarkerUID).Error)
	require.NoError(t, tempDb.First(&storedAuto, "marker_uid = ?", automatic.MarkerUID).Error)
	assert.Equal(t, "Alice", storedManual.MarkerName)
	assert.Equal(t, subjectUID, storedManual.SubjUID)
	assert.Equal(t, cluster.ID, storedManual.FaceID)
	assert.Empty(t, storedAuto.MarkerName)
	assert.Empty(t, storedAuto.SubjUID)
	assert.Empty(t, storedAuto.EmbeddingsJSON)

	var facesBefore, facesAfter int
	require.NoError(t, tempDb.Model(&entity.Face{}).Count(&facesBefore).Error)
	require.Error(t, FinalizeFaceMigration(face.ModelFaceNet, []FaceMigrationIdentity{{MarkerUID: "changed"}}, nil, nil))
	require.NoError(t, tempDb.Model(&entity.Face{}).Count(&facesAfter).Error)
	assert.Equal(t, facesBefore, facesAfter)

	require.Error(t, FinalizeFaceMigration("", nil, nil, nil))
}

func TestSameFaceMigrationIdentities(t *testing.T) {
	identities := []FaceMigrationIdentity{{MarkerUID: "m1", SubjUID: "s1", MarkerName: "Alice", SubjSrc: entity.SrcManual}}
	assert.True(t, sameFaceMigrationIdentities(identities, identities))
	assert.False(t, sameFaceMigrationIdentities(identities, nil))
	assert.False(t, sameFaceMigrationIdentities(identities, []FaceMigrationIdentity{{MarkerUID: "m2"}}))
}
