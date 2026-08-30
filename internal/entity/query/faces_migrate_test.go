package query

import (
	"encoding/json"
	"path/filepath"
	"sort"
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

// TestFaceMigrationUnreadableFileCount pins that the plan counts markers whose file cannot be
// re-embedded. Counting only file_uid = ” reported a clean plan for a run that then failed
// five markers and left them with no vector at all, because a soft-deleted file row does not
// resolve and so never looked unlinked.
func TestFaceMigrationUnreadableFileCount(t *testing.T) {
	result, err := FaceMigrationCounts(face.ModelFaceNet)
	require.NoError(t, err)

	// Counted independently, as a left join rather than the NOT IN under test.
	var want int
	row := Db().Raw("SELECT COUNT(*) FROM markers m LEFT JOIN files f"+
		" ON f.file_uid = m.file_uid AND f.deleted_at IS NULL"+
		" WHERE m.marker_type = ? AND m.marker_invalid = 0 AND m.file_uid <> ''"+
		" AND (f.file_uid IS NULL OR f.file_missing = 1 OR f.file_error <> '')", entity.MarkerFace).Row()
	require.NoError(t, row.Scan(&want))

	require.Positive(t, want, "fixtures must include a face marker whose file cannot be read")
	assert.Equal(t, want, result.Unreadable)

	// Strictly fewer than all valid markers, which is what fails if the predicate ever stops
	// restricting the count to files the index cannot offer.
	assert.Less(t, result.Unreadable, result.Valid, "a readable file must not be counted")
}

// TestFaceMigrationSoftDeletedFileCount pins the case that the fixtures do not cover and that
// the live run tripped over: a file row that still exists but is soft-deleted. It resolves to
// nothing, so the marker neither looks unlinked nor looks flagged, and counting it is the only
// thing that lets a plan predict the failure.
func TestFaceMigrationSoftDeletedFileCount(t *testing.T) {
	before, err := FaceMigrationCounts(face.ModelFaceNet)
	require.NoError(t, err)

	file := entity.File{
		FileUID:  rnd.GenerateUID('f'),
		PhotoUID: rnd.GenerateUID('p'),
		FileName: "soft-deleted/migrate-test.jpg",
		FileRoot: entity.RootOriginals,
	}
	require.NoError(t, Db().Create(&file).Error)

	marker := entity.Marker{
		MarkerUID:  rnd.GenerateUID('m'),
		FileUID:    file.FileUID,
		MarkerType: entity.MarkerFace,
	}
	require.NoError(t, Db().Create(&marker).Error)

	t.Cleanup(func() {
		Db().Unscoped().Delete(&marker)
		Db().Unscoped().Delete(&file)
	})

	// A readable file must not be counted, so the marker is invisible while the file is live.
	live, err := FaceMigrationCounts(face.ModelFaceNet)
	require.NoError(t, err)
	assert.Equal(t, before.Unreadable, live.Unreadable, "a live file must not count as unreadable")

	// Soft-deleting leaves the row in place, so nothing about the marker changes.
	require.NoError(t, Db().Delete(&file).Error)

	after, err := FaceMigrationCounts(face.ModelFaceNet)
	require.NoError(t, err)
	assert.Equal(t, before.Unreadable+1, after.Unreadable, "a soft-deleted file must count as unreadable")
	assert.Equal(t, live.Unlinked, after.Unlinked, "the marker still has a file_uid, so it is not unlinked")
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
	newMarker := func(subjUID, subjSrc string, size, score int) *entity.Marker {
		m := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        "fs6sg6bw45bnlqdw",
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			SubjUID:        subjUID,
			SubjSrc:        subjSrc,
			Size:           size,
			Score:          score,
			EmbedModel:     face.ModelFaceNet,
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
			W:              0.1,
			H:              0.1,
		}
		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(m) })

		return m
	}

	good := face.ClusterSizeThreshold + 10
	goodScore := face.ClusterScore("") + 10

	subjUID := rnd.GenerateUID('j')
	manual := newMarker(subjUID, entity.SrcManual, good, goodScore)
	automatic := newMarker(subjUID, entity.SrcAuto, good, goodScore)
	tiny := newMarker(subjUID, entity.SrcManual, face.ClusterSizeThreshold-1, goodScore)
	faint := newMarker(subjUID, entity.SrcManual, good, face.ClusterScore("")-1)
	other := newMarker(rnd.GenerateUID('j'), entity.SrcAuto, good, goodScore)

	result, err := FaceMigrationSubjectMarkers(face.ModelFaceNet, subjUID)
	require.NoError(t, err)

	found := make(map[string]bool, len(result))
	for _, candidate := range result {
		found[candidate.MarkerUID] = true
	}
	assert.True(t, found[manual.MarkerUID], "manual marker must seed the cluster")
	assert.True(t, found[automatic.MarkerUID], "automatic marker must seed the cluster")
	// A face too small or too poorly scored to be clustered cannot define a centroid
	// either, whoever assigned it.
	assert.False(t, found[tiny.MarkerUID], "a face below the cluster size must not seed one")
	assert.False(t, found[faint.MarkerUID], "a face below the cluster score must not seed one")
	assert.False(t, found[other.MarkerUID], "another subject's marker must not be returned")

	_, err = FaceMigrationSubjectMarkers("", subjUID)
	require.Error(t, err)
	_, err = FaceMigrationSubjectMarkers(face.ModelFaceNet, "")
	require.Error(t, err)

	t.Run("SubjectUIDs", func(t *testing.T) {
		uids, uidErr := FaceMigrationSubjectUIDs(face.ModelFaceNet)
		require.NoError(t, uidErr)
		assert.Contains(t, uids, subjUID)
		assert.True(t, sort.StringsAreSorted(uids), "subjects are paged in order")

		_, uidErr = FaceMigrationSubjectUIDs("")
		require.Error(t, uidErr)
	})
	t.Run("LowQualityMarkers", func(t *testing.T) {
		count, countErr := FaceMigrationLowQualityMarkers(face.ModelFaceNet)
		require.NoError(t, countErr)
		assert.GreaterOrEqual(t, count, 2, "the tiny and faint markers are counted")

		// It reports the complement of what seeds a cluster, so the two have to agree on the
		// bars: a count with its own copy of them describes a set the rebuild does not use.
		var assigned int
		require.NoError(t, whereEmbeddingModel(Db().Model(&entity.Marker{}).
			Where("marker_type = ? AND marker_invalid = 0", entity.MarkerFace).
			Where("subj_uid <> ''").
			Where("LENGTH(embeddings_json) > 0"), face.ModelFaceNet).Count(&assigned).Error)

		var seeds int
		require.NoError(t, whereFaceMigrationSamples(Db().Model(&entity.Marker{}), face.ModelFaceNet).Count(&seeds).Error)

		assert.Equal(t, assigned-seeds, count)

		_, countErr = FaceMigrationLowQualityMarkers("")
		require.Error(t, countErr)
	})
}

func TestFaceMigrationRecropMarkers(t *testing.T) {
	// Absolute counts depend on what else the library holds, and every fixture marker records
	// no detector, so the assertions are on what these rows add.
	beforeYuNet, err := FaceMigrationRecropMarkers(face.ModelFaceNet, face.DetectorYuNet)
	require.NoError(t, err)
	beforeSCRFD, err := FaceMigrationRecropMarkers(face.ModelFaceNet, face.DetectorSCRFD)
	require.NoError(t, err)

	newMarker := func(embedModel, detectModel string, vector []byte) *entity.Marker {
		m := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        "fs6sg6bw45bnlqdw",
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			EmbedModel:     embedModel,
			DetectModel:    detectModel,
			EmbeddingsJSON: vector,
			W:              0.1,
			H:              0.1,
		}
		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(m) })

		return m
	}

	vector := face.Embeddings{face.RandomEmbedding()}.JSON()

	newMarker(face.ModelFaceNet, face.DetectorSCRFD, vector)
	newMarker(face.ModelFaceNet, "", vector)
	newMarker(face.ModelFaceNet, face.DetectorYuNet, vector)
	newMarker(face.ModelFaceNet, face.DetectorYuNet, vector)
	newMarker(face.ModelSFace, face.DetectorSCRFD, vector)
	newMarker(face.ModelFaceNet, face.DetectorSCRFD, nil)

	t.Run("CountsTheOtherDetector", func(t *testing.T) {
		// Only a marker holding a target-model vector that another detector cropped: the
		// current detector's is not stale, another model's is stale for a different reason,
		// and one without a vector has nothing to keep.
		count, countErr := FaceMigrationRecropMarkers(face.ModelFaceNet, face.DetectorYuNet)
		require.NoError(t, countErr)
		assert.Equal(t, beforeYuNet+2, count)
	})
	t.Run("OtherDetectorInForce", func(t *testing.T) {
		// The same rows, counted against the other detector: which crop is stale follows the
		// detector in force rather than naming one of them.
		count, countErr := FaceMigrationRecropMarkers(face.ModelFaceNet, face.DetectorSCRFD)
		require.NoError(t, countErr)
		assert.Equal(t, beforeSCRFD+3, count)
	})
	t.Run("NoDetector", func(t *testing.T) {
		// Detection is off, so there is no detector to disagree with and nothing to re-crop.
		for _, detector := range []string{"", face.DetectorNone} {
			count, countErr := FaceMigrationRecropMarkers(face.ModelFaceNet, detector)
			require.NoError(t, countErr)
			assert.Zero(t, count, detector)
		}
	})
	t.Run("ModelRequired", func(t *testing.T) {
		_, countErr := FaceMigrationRecropMarkers("", face.DetectorYuNet)
		require.Error(t, countErr)
	})
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
	detectedPoints := json.RawMessage(`[{"name":"eye_l","x":-0.05},{"name":"eye_r","x":0.05}]`)
	require.NoError(t, SaveFaceMigrationEmbeddings(face.ModelFaceNet, face.DetectorYuNet,
		map[string]face.Embeddings{marker.MarkerUID: embeddings},
		map[string]MigrationDetection{marker.MarkerUID: {Landmarks: detectedPoints, Size: 84, Score: 91}}))

	stored, err := MarkerByUID(marker.MarkerUID)
	require.NoError(t, err)
	assert.Equal(t, face.ModelFaceNet, stored.EmbedModel)
	assert.Equal(t, face.DetectorYuNet, stored.DetectModel, "a run that re-detected records the detector it used")
	assert.Empty(t, stored.FaceID)
	assert.True(t, stored.Embeddings().One())
	assert.Len(t, stored.Embeddings()[0], len(embeddings[0]))
	assert.JSONEq(t, string(detectedPoints), string(stored.LandmarksJSON),
		"the landmarks the detection placed travel with the vector they produced")
	// The clustering bars are looked up by detect_model, so a marker relabelled with a new detector
	// while holding the old one's score would be judged at a calibration it was never scored against.
	assert.Equal(t, 91, stored.Score, "the score of the detection that produced the vector")
	assert.Equal(t, 84, stored.Size, "and its size, in the pixels of the same detection thumbnail")

	t.Run("BlankDetectorKeepsProvenance", func(t *testing.T) {
		// Re-cropping runs no detector, so overwriting either would attribute the crop to one
		// that never saw the image. The vector has to differ from the one already stored:
		// MariaDB reports no affected rows for an update that changes nothing.
		regenerated := face.Embeddings{face.RandomEmbedding()}
		require.NoError(t, SaveFaceMigrationEmbeddings(face.ModelFaceNet, "", map[string]face.Embeddings{marker.MarkerUID: regenerated}, nil))

		kept, keptErr := MarkerByUID(marker.MarkerUID)
		require.NoError(t, keptErr)
		assert.Equal(t, face.DetectorYuNet, kept.DetectModel)
		assert.JSONEq(t, string(detectedPoints), string(kept.LandmarksJSON))
		assert.Equal(t, 91, kept.Score, "no detection ran, so nothing may overwrite what one recorded")
		assert.Equal(t, 84, kept.Size)
	})
	t.Run("MalformedLandmarksClearTheColumn", func(t *testing.T) {
		// The detector and the landmarks are written as a pair. A payload that is not valid JSON
		// would make the column unreadable, and leaving the previous detector's landmarks beside
		// a newly recorded one is the divergence the pairing exists to prevent, so it is cleared.
		regenerated := face.Embeddings{face.RandomEmbedding()}
		require.NoError(t, SaveFaceMigrationEmbeddings(face.ModelFaceNet, face.DetectorSCRFD,
			map[string]face.Embeddings{marker.MarkerUID: regenerated},
			map[string]MigrationDetection{marker.MarkerUID: {Landmarks: json.RawMessage("{"), Size: 60, Score: 55}}))

		kept, keptErr := MarkerByUID(marker.MarkerUID)
		require.NoError(t, keptErr)
		assert.Equal(t, face.DetectorSCRFD, kept.DetectModel)
		assert.Empty(t, kept.LandmarksJSON)
		assert.Equal(t, 55, kept.Score, "unreadable landmarks do not invalidate what the detection scored")
		assert.Equal(t, 60, kept.Size)
	})

	require.Error(t, SaveFaceMigrationEmbeddings("", face.DetectorYuNet, nil, nil))
	require.Error(t, SaveFaceMigrationEmbeddings(face.ModelFaceNet, face.DetectorYuNet, map[string]face.Embeddings{"": nil}, nil))
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
		DetectModel:    face.EngineONNX,
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
		DetectModel:    face.EngineONNX,
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	// Caught by the bulk stale-vector predicate rather than the failed-marker batch, which
	// is a separate statement and would otherwise keep its provenance untested.
	invalid := entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        "file123",
		MarkerType:     entity.MarkerFace,
		MarkerName:     "Alice",
		SubjUID:        subjectUID,
		SubjSrc:        entity.SrcXmp,
		MarkerInvalid:  true,
		EmbedModel:     face.ModelFaceNet,
		DetectModel:    face.EngineONNX,
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	require.NoError(t, tempDb.Create(&manual).Error)
	require.NoError(t, tempDb.Create(&automatic).Error)
	require.NoError(t, tempDb.Create(&imported).Error)
	require.NoError(t, tempDb.Create(&invalid).Error)

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
	// A blanked vector takes both provenance columns with it: a detector recorded next to
	// no embedding would claim a crop that no longer exists.
	assert.Empty(t, storedAuto.DetectModel)
	assert.Equal(t, face.EngineONNX, storedImported.DetectModel, "a spared vector keeps its detector")

	var storedInvalid entity.Marker
	require.NoError(t, tempDb.First(&storedInvalid, "marker_uid = ?", invalid.MarkerUID).Error)
	assert.Empty(t, storedInvalid.EmbeddingsJSON)
	assert.Empty(t, storedInvalid.EmbedModel)
	assert.Empty(t, storedInvalid.DetectModel, "the bulk stale update clears provenance too")
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

// TestCountMarkersWithoutThumbSize covers the audit count, which reports how many embedded markers
// are still judged by their detection size rather than by what their embedding was sampled from.
func TestCountMarkersWithoutThumbSize(t *testing.T) {
	before, err := CountMarkersWithoutThumbSize()
	require.NoError(t, err)

	m := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        "fs6sg6bw45bnlqdw",
		MarkerType:     entity.MarkerFace,
		MarkerSrc:      entity.SrcImage,
		Size:           200,
		ThumbSize:      -1,
		Score:          100,
		EmbedModel:     face.ModelFaceNet,
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	require.NoError(t, entity.Db().Create(m).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(m) })

	after, err := CountMarkersWithoutThumbSize()
	require.NoError(t, err)
	assert.Equal(t, before+1, after, "an embedded marker with no recorded sample size must be counted")

	// A recorded value takes it back out, however small: 1 is a measurement, not a sentinel.
	require.NoError(t, entity.Db().Model(&entity.Marker{}).Where("marker_uid = ?", m.MarkerUID).
		Update("thumb_size", 1).Error)

	after, err = CountMarkersWithoutThumbSize()
	require.NoError(t, err)
	assert.Equal(t, before, after)
}
