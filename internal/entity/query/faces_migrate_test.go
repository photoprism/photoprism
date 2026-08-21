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

func TestHiddenFaceMarkers(t *testing.T) {
	hiddenFace := &entity.Face{
		ID:            "HIDDENCLUSTERFORMIGRATIONTEST00000000000",
		FaceSrc:       entity.SrcAuto,
		FaceHidden:    true,
		EmbedModel:    face.ModelFaceNet,
		EmbeddingJSON: []byte("[0.1,0.2]"),
	}
	require.NoError(t, entity.Db().Create(hiddenFace).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(hiddenFace) })

	marker := &entity.Marker{
		MarkerUID:  rnd.GenerateUID('m'),
		FileUID:    "fs6sg6bw45bnlqdw",
		MarkerType: entity.MarkerFace,
		MarkerSrc:  entity.SrcImage,
		FaceID:     hiddenFace.ID,
		W:          0.1,
		H:          0.1,
	}
	require.NoError(t, entity.Db().Create(marker).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(marker) })

	result, err := HiddenFaceMarkers()
	require.NoError(t, err)
	assert.Contains(t, result, marker.MarkerUID)
}

func TestRestoreHiddenFaces(t *testing.T) {
	newCluster := func(id string) *entity.Face {
		f := &entity.Face{ID: id, FaceSrc: entity.SrcAuto, EmbedModel: face.ModelFaceNet, EmbeddingJSON: []byte("[0.1,0.2]")}
		require.NoError(t, entity.Db().Create(f).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(f) })

		return f
	}
	newMarker := func(faceID string) *entity.Marker {
		m := &entity.Marker{MarkerUID: rnd.GenerateUID('m'), FileUID: "fs6sg6bw45bnlqdw",
			MarkerType: entity.MarkerFace, MarkerSrc: entity.SrcImage, FaceID: faceID, W: 0.1, H: 0.1}
		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(m) })

		return m
	}

	t.Run("MajorityHidden", func(t *testing.T) {
		c := newCluster("RESTOREHIDDENMAJORITY0000000000000000000")
		was := []string{newMarker(c.ID).MarkerUID, newMarker(c.ID).MarkerUID}
		newMarker(c.ID)

		hidden, err := RestoreHiddenFaces(was)
		require.NoError(t, err)
		assert.Positive(t, hidden)

		stored := entity.Face{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "id = ?", c.ID).Error)
		assert.True(t, stored.FaceHidden)
	})
	t.Run("MinorityStaysVisible", func(t *testing.T) {
		// A cluster that merely absorbed one hidden face was not the operator's decision.
		c := newCluster("RESTOREHIDDENMINORITY0000000000000000000")
		was := []string{newMarker(c.ID).MarkerUID}
		newMarker(c.ID)
		newMarker(c.ID)

		_, err := RestoreHiddenFaces(was)
		require.NoError(t, err)

		stored := entity.Face{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "id = ?", c.ID).Error)
		assert.False(t, stored.FaceHidden)
	})
	t.Run("Empty", func(t *testing.T) {
		hidden, err := RestoreHiddenFaces(nil)
		require.NoError(t, err)
		assert.Zero(t, hidden)
	})
}

func TestFaceMigrationIdentities(t *testing.T) {
	result, err := FaceMigrationIdentities()
	require.NoError(t, err)
	for _, identity := range result {
		assert.NotEmpty(t, identity.MarkerUID)
		assert.True(t, identity.SubjUID != "" || identity.MarkerName != "",
			"an identity must carry a subject or a name")
	}
}

func TestFaceMigrationSubjectMarkers(t *testing.T) {
	// An automatic assignment is what the old model recorded about a person, not about
	// its own embedding space, so it has to seed the replacement cluster as well.
	newMarker := func(subjUID, subjSrc string) *entity.Marker {
		m := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        "fs6sg6bw45bnlqdw",
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			SubjUID:        subjUID,
			SubjSrc:        subjSrc,
			EmbedModel:     face.ModelFaceNet,
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
			W:              0.1,
			H:              0.1,
		}
		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(m) })

		return m
	}

	subjUID := rnd.GenerateUID('j')
	manual := newMarker(subjUID, entity.SrcManual)
	automatic := newMarker(subjUID, entity.SrcAuto)
	unassigned := newMarker("", entity.SrcAuto)

	result, err := FaceMigrationSubjectMarkers(face.ModelFaceNet)
	require.NoError(t, err)

	found := make(map[string]bool, len(result))
	for _, candidate := range result {
		found[candidate.MarkerUID] = true
	}
	assert.True(t, found[manual.MarkerUID], "manual marker must seed the cluster")
	assert.True(t, found[automatic.MarkerUID], "automatic marker must seed the cluster")
	assert.False(t, found[unassigned.MarkerUID], "marker without a subject must be excluded")

	_, err = FaceMigrationSubjectMarkers("")
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
	// A marker from a third source, to prove the relink is not restricted to manual ones.
	imported := entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        "file123",
		MarkerType:     entity.MarkerFace,
		MarkerName:     "Alice",
		SubjUID:        subjectUID,
		SubjSrc:        entity.SrcXmp,
		EmbedModel:     face.ModelFaceNet,
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	require.NoError(t, tempDb.Create(&manual).Error)
	require.NoError(t, tempDb.Create(&automatic).Error)
	require.NoError(t, tempDb.Create(&imported).Error)

	// A cluster from the previous run, so the delete this function is named for has
	// something to remove rather than passing over an empty table.
	stale := entity.Face{
		ID:            "STALECLUSTERFROMPREVIOUSMODEL00000000000",
		FaceSrc:       entity.SrcAuto,
		SubjUID:       subjectUID,
		EmbedModel:    face.ModelFaceNet,
		EmbeddingJSON: []byte("[0.1,0.2]"),
	}
	require.NoError(t, tempDb.Create(&stale).Error)

	identities, err := FaceMigrationIdentities()
	require.NoError(t, err)
	cluster := entity.NewFace(subjectUID, entity.SrcManual, manual.Embeddings(), face.EmbeddingModelName())
	require.NotNil(t, cluster)

	require.NoError(t, FinalizeFaceMigration(face.ModelFaceNet, identities, []FaceMigrationCluster{{
		Face:            *cluster,
		MarkerDistances: map[string]float64{manual.MarkerUID: 0, imported.MarkerUID: 0.2},
	}}, []string{automatic.MarkerUID}))

	var staleCount int
	require.NoError(t, tempDb.Unscoped().Model(&entity.Face{}).Where("id = ?", stale.ID).Count(&staleCount).Error)
	assert.Zero(t, staleCount, "the previous run's clusters must be replaced")

	var storedManual, storedAuto, storedImported entity.Marker
	require.NoError(t, tempDb.First(&storedManual, "marker_uid = ?", manual.MarkerUID).Error)
	require.NoError(t, tempDb.First(&storedAuto, "marker_uid = ?", automatic.MarkerUID).Error)
	require.NoError(t, tempDb.First(&storedImported, "marker_uid = ?", imported.MarkerUID).Error)
	assert.Equal(t, "Alice", storedManual.MarkerName)
	assert.Equal(t, subjectUID, storedManual.SubjUID)
	assert.Equal(t, cluster.ID, storedManual.FaceID)
	// Who a face shows outlives the vectors, so an automatic assignment survives even
	// when its embedding could not be regenerated.
	assert.Equal(t, "Alice", storedAuto.MarkerName)
	assert.Equal(t, subjectUID, storedAuto.SubjUID)
	assert.Empty(t, storedAuto.EmbeddingsJSON)
	assert.Equal(t, cluster.ID, storedImported.FaceID)
	assert.Equal(t, subjectUID, storedImported.SubjUID)

	var facesBefore, facesAfter int
	require.NoError(t, tempDb.Model(&entity.Face{}).Count(&facesBefore).Error)
	changedErr := FinalizeFaceMigration(face.ModelFaceNet, []FaceMigrationIdentity{{MarkerUID: "changed"}}, nil, nil)
	require.Error(t, changedErr)
	// Callers distinguish this from a storage failure, because it is the one rollback an
	// operator caused and can avoid on the next run.
	assert.ErrorIs(t, changedErr, ErrFaceMigrationIdentitiesChanged)
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
