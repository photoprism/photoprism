package photoprism

import (
	"crypto/sha1" //nolint:gosec // SHA1 retained for legacy audit IDs.
	"encoding/base32"
	"math"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestFaces_Audit(t *testing.T) {
	t.Run("FixEqualTrue", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		err := m.Audit(true, "")

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("FixeEqualFalse", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		err := m.Audit(false, "")

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("SubjectFilter", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		require.NoError(t, m.Audit(false, "jr0ncy131y7igds8"))
	})
}

func TestFaces_AuditNormalizesEmbeddings(t *testing.T) {
	t.Helper()

	oldCfg := Config()
	c := config.NewMinimalTestConfigWithDb("faces-audit-normalize", t.TempDir())
	t.Cleanup(func() {
		_ = c.CloseDb()

		if oldCfg != nil {
			oldCfg.RegisterDb()
		}
	})

	m := NewFaces(c)

	raw := make(face.Embedding, len(face.NullEmbedding))
	raw[0] = 2
	raw[1] = 1

	rawJSON := raw.JSON()

	//nolint:gosec // legacy ID compatibility relies on SHA1.
	original := sha1.Sum(rawJSON)
	oldID := base32.StdEncoding.EncodeToString(original[:])

	now := entity.Now()

	faceRow := &entity.Face{
		ID:            oldID,
		FaceSrc:       entity.SrcAuto,
		EmbeddingJSON: rawJSON,
		Samples:       5,
		SampleRadius:  0.12,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	require.NoError(t, entity.Db().Create(faceRow).Error)

	markerEmbJSON := (face.Embeddings{raw}).JSON()

	marker := &entity.Marker{
		MarkerType:     entity.MarkerFace,
		MarkerSrc:      entity.SrcAuto,
		FaceID:         oldID,
		EmbeddingsJSON: markerEmbJSON,
		FaceDist:       0.5,
	}

	require.NoError(t, entity.Db().Create(marker).Error)

	//nolint:gosec // legacy ID compatibility relies on SHA1.
	hashNorm := sha1.Sum(normalizeEmbeddingCopy(raw).JSON())
	expectedID := base32.StdEncoding.EncodeToString(hashNorm[:])

	require.NoError(t, m.Audit(true, ""))

	var updated entity.Face
	require.NoError(t, entity.Db().Where("id = ?", expectedID).First(&updated).Error)
	require.NotEqual(t, oldID, updated.ID)

	updatedEmbedding := updated.Embedding()
	require.InDelta(t, 1.0, updatedEmbedding.Magnitude(), 1e-9)
	normalized := normalizeEmbeddingCopy(raw)
	require.InDelta(t, normalized[0], updatedEmbedding[0], 1e-9)

	var updatedMarker entity.Marker
	require.NoError(t, entity.Db().Where("marker_uid = ?", marker.MarkerUID).First(&updatedMarker).Error)

	expectedDist := minEmbeddingDistance(normalized, updatedMarker.Embeddings())
	require.InDelta(t, expectedDist, updatedMarker.FaceDist, 1e-9)
	require.Equal(t, expectedID, updatedMarker.FaceID)
}

// normalizeEmbeddingCopy returns a unit-length copy of an embedding, leaving the original untouched
// so a test can compare what it passed in against what was stored.
func normalizeEmbeddingCopy(src face.Embedding) face.Embedding {
	copyEmb := make(face.Embedding, len(src))
	copy(copyEmb, src)

	var sum float64

	for _, v := range copyEmb {
		sum += v * v
	}

	length := math.Sqrt(sum)

	if length == 0 {
		return copyEmb
	}

	inv := 1 / length

	for i := range copyEmb {
		copyEmb[i] *= inv
	}

	return copyEmb
}

// captureLog records what the shared logger emits for the duration of a test, so the
// audit functions can be checked on what they report rather than only on not panicking.
func captureLog(t *testing.T) *test.Hook {
	t.Helper()

	logger, ok := log.(*logrus.Logger)
	require.True(t, ok)

	hook := test.NewLocal(logger)
	t.Cleanup(hook.Reset)

	return hook
}

// loggedMessages returns the recorded messages at or above the given level.
func loggedMessages(hook *test.Hook, level logrus.Level) []string {
	var result []string

	for _, entry := range hook.AllEntries() {
		if entry.Level <= level {
			result = append(result, entry.Message)
		}
	}

	return result
}

// TestFaces_recomputeMissingRadius covers the repair audit --fix applies to a cluster stored without
// an extent. It measures over the members rather than writing a constant, because a width nothing
// observed is what the singleton rule exists to remove - and a singleton has nothing to measure.
func TestFaces_recomputeMissingRadius(t *testing.T) {
	const subjUID = "js6sg6b1qekk9jy2"

	// The row shape an upgraded library holds: SetEmbeddings has written at least the floor since.
	storedClusterFor := func(t *testing.T, subj string, seed uint64, dists ...float64) *entity.Face {
		t.Helper()

		f, _ := recomputeTestCluster(t, subj, seed, dists...)
		require.NoError(t, entity.Db().Model(&entity.Face{}).Where("id = ?", f.ID).
			UpdateColumns(entity.Values{"sample_radius": 0, "matched_at": entity.Now()}).Error)

		return f
	}

	storedCluster := func(t *testing.T, seed uint64, dists ...float64) *entity.Face {
		t.Helper()
		return storedClusterFor(t, subjUID, seed, dists...)
	}

	t.Run("MeasuresTheMembers", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-audit-measure-fix")
		f := storedCluster(t, 8901, 0.05, 0.08, 0.11)

		n, err := w.recomputeMissingRadius(true, subjUID)
		require.NoError(t, err)
		assert.Equal(t, 1, n)

		after := entity.FindFace(f.ID)
		require.NotNil(t, after)
		assert.GreaterOrEqual(t, after.SampleRadius, 0.11, "the extent covers the furthest member")
		assert.Less(t, after.SampleRadius, face.ClusterRadius, "and is measured, not a constant")
		assert.Nil(t, after.MatchedAt, "so the cluster is compared against the markers again")
	})
	t.Run("ReportsWithoutFixing", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-audit-measure-report")
		f := storedCluster(t, 8911, 0.05, 0.09)

		n, err := w.recomputeMissingRadius(false, subjUID)
		require.NoError(t, err)
		assert.Equal(t, 1, n)

		after := entity.FindFace(f.ID)
		require.NotNil(t, after)
		assert.Zero(t, after.SampleRadius, "a dry run must not write")
		assert.NotNil(t, after.MatchedAt)
	})
	t.Run("LeavesASingletonAlone", func(t *testing.T) {
		// One embedding has no extent, so there is nothing to repair and nothing to measure. It is
		// given a marker deliberately: with none, an empty member set would decide the case and the
		// core predicate this covers would never be what refused it.
		w := isolatedTestFaces(t, "faces-audit-measure-singleton")

		base := face.FixtureEmbedding(8921)
		f := entity.NewFace(subjUID, entity.SrcManual, face.Embeddings{base}, face.EmbeddingModelName())
		require.NotNil(t, f)
		require.NoError(t, f.Create())
		require.Equal(t, 1, f.Samples)
		require.NoError(t, entity.Db().Model(&entity.Face{}).Where("id = ?", f.ID).
			UpdateColumn("sample_radius", 0).Error)

		marker := &entity.Marker{
			FileUID: "fs6sg6bw45bnlqdw", MarkerType: entity.MarkerFace, MarkerSrc: entity.SrcImage,
			FaceID: f.ID, SubjUID: subjUID, SubjSrc: entity.SrcManual,
			Size: face.SizeThreshold, Score: 50, X: 0.2, Y: 0.2, W: 0.1, H: 0.1,
		}
		marker.SetEmbeddings(face.Embeddings{face.FixtureEmbeddingAt(base, 0.05, 8922)},
			f.EmbedModel, face.DetectorYuNet)
		require.NoError(t, entity.Db().Create(marker).Error)

		t.Cleanup(func() {
			entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", marker.MarkerUID)
			entity.UnscopedDb().Delete(&entity.Face{}, "id = ?", f.ID)
		})

		members, err := query.FaceMembers(f.ID)
		require.NoError(t, err)
		require.Len(t, members, 1, "so a measurement is available and only the core refuses it")

		n, err := w.recomputeMissingRadius(true, subjUID)
		require.NoError(t, err)
		assert.Zero(t, n)

		after := entity.FindFace(f.ID)
		require.NotNil(t, after)
		assert.Zero(t, after.SampleRadius)
	})
	t.Run("SkipsRowsTheMatcherExcludes", func(t *testing.T) {
		// Nothing narrows a radius again, so a row the matcher would not read must not be measured:
		// un-hiding it later would bring it back with its own extent already gone.
		w := isolatedTestFaces(t, "faces-audit-measure-excluded")

		hidden := storedCluster(t, 8931, 0.05, 0.09, 0.11)
		require.NoError(t, entity.Db().Model(&entity.Face{}).Where("id = ?", hidden.ID).
			UpdateColumn("face_hidden", true).Error)

		ignored := storedCluster(t, 8941, 0.05, 0.09, 0.11)
		require.NoError(t, entity.Db().Model(&entity.Face{}).Where("id = ?", ignored.ID).
			UpdateColumn("face_kind", int(face.AmbiguousFace)).Error)

		foreign := storedCluster(t, 8951, 0.05, 0.09, 0.11)
		require.NoError(t, entity.Db().Model(&entity.Face{}).Where("id = ?", foreign.ID).
			UpdateColumn("embed_model", "not-the-configured-model").Error)

		n, err := w.recomputeMissingRadius(true, subjUID)
		require.NoError(t, err)
		assert.Zero(t, n)

		for name, id := range map[string]string{"hidden": hidden.ID, "ignored": ignored.ID, "foreign": foreign.ID} {
			after := entity.FindFace(id)
			require.NotNil(t, after, name)
			assert.Zero(t, after.SampleRadius, name)
		}
	})
	t.Run("KeepsARowAtTheFloor", func(t *testing.T) {
		// A row already at the floor was written by a current path, so it is current rather than
		// degenerate. Flipping the predicate to <= would re-measure every singleton in the library.
		w := isolatedTestFaces(t, "faces-audit-measure-floor")

		f := storedCluster(t, 8961, 0.05, 0.09, 0.11)
		require.NoError(t, entity.Db().Model(&entity.Face{}).Where("id = ?", f.ID).
			UpdateColumn("sample_radius", face.Epsilon).Error)

		n, err := w.recomputeMissingRadius(true, subjUID)
		require.NoError(t, err)
		assert.Zero(t, n)

		after := entity.FindFace(f.ID)
		require.NotNil(t, after)
		assert.InDelta(t, face.Epsilon, after.SampleRadius, 1e-9)
	})
	t.Run("AuditAppliesTheMeasurement", func(t *testing.T) {
		// Pins the wiring rather than the measurement: called directly, every subtest above stays
		// green while nothing in Audit reaches it at all.
		// Its own subject, so the audit's other passes see one cluster rather than the several the
		// subtests above leave behind - their cleanups run when the parent test ends, not sooner.
		w := isolatedTestFaces(t, "faces-audit-measure-wiring")

		// A real subject row, because Audit --fix clears references to subjects that do not exist -
		// and its own subject, so the other passes see one cluster rather than the several the
		// subtests above leave behind (their cleanups run when the parent test ends, not sooner).
		subj := entity.NewSubject("Audit Wiring Person", entity.SubjPerson, entity.SrcManual)
		require.NotNil(t, subj)
		require.NoError(t, subj.Create())
		t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", subj.SubjUID) })

		wiringUID := subj.SubjUID

		storedClusterFor(t, wiringUID, 8971, 0.05, 0.09, 0.11)

		require.NoError(t, w.Audit(true, wiringUID))

		// Looked up by subject rather than by id: the same run normalizes stored embeddings, which
		// re-keys a row whose vector was not unit length.
		var after entity.Faces
		require.NoError(t, entity.UnscopedDb().Where("subj_uid = ?", wiringUID).Find(&after).Error)
		require.Len(t, after, 1)
		assert.GreaterOrEqual(t, after[0].SampleRadius, 0.11, "the audit measured it")
	})
}

func TestFaces_auditEmbeddingModels(t *testing.T) {
	w := NewFaces(config.TestConfig())

	t.Cleanup(func() {
		restoreEmbedder(t)
	})

	t.Run("ConfiguredModel", func(t *testing.T) {
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
			Name:  face.ModelFaceNet,
			Model: face.FindEmbeddingModel(face.ModelFaceNet),
		}))
		hook := captureLog(t)
		w.auditEmbeddingModels()

		assert.NotEmpty(t, hook.AllEntries(), "the audit reports what the library holds")
	})
	t.Run("StaleClusters", func(t *testing.T) {
		// A cluster from another model must be reported as stale.
		m := entity.NewFace("", entity.SrcAuto, face.Embeddings{face.RandomEmbedding()}, face.EmbeddingModelName())
		require.NotNil(t, m)
		m.EmbedModel = face.ModelArcFaceR50
		require.NoError(t, entity.Db().Create(m).Error)

		t.Cleanup(func() { entity.Db().Delete(m) })

		count, err := query.FacesFromOtherModels()
		require.NoError(t, err)
		require.Positive(t, count)

		hook := captureLog(t)
		w.auditEmbeddingModels()

		warnings := loggedMessages(hook, logrus.WarnLevel)
		require.NotEmpty(t, warnings, "a cluster from another model is a warning")
		assert.Contains(t, strings.Join(warnings, "\n"), face.ModelArcFaceR50)
		assert.Contains(t, strings.Join(warnings, "\n"), "faces migrate",
			"the operator is told how to resolve it")
	})
	t.Run("NoModelConfigured", func(t *testing.T) {
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelNone}))

		hook := captureLog(t)
		w.auditEmbeddingModels()

		// Without a configured model nothing can be called incompatible with it, so the
		// counts are reported without telling the operator to migrate.
		assert.NotContains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"), "faces migrate")
	})
}

func TestFaces_auditMarkerEmbeddingModels(t *testing.T) {
	w := NewFaces(config.TestConfig())

	t.Run("StaleMarkers", func(t *testing.T) {
		m := &entity.Marker{
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			EmbedModel:     face.ModelArcFaceR50,
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		}

		require.NoError(t, entity.Db().Create(m).Error)

		t.Cleanup(func() { entity.Db().Delete(m) })

		counts, err := query.MarkerEmbeddingModels()
		require.NoError(t, err)
		require.NotEmpty(t, counts)

		hook := captureLog(t)
		w.auditMarkerEmbeddingModels(face.ModelFaceNet)

		warnings := strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n")
		assert.Contains(t, warnings, face.ModelArcFaceR50)
		assert.Contains(t, warnings, face.ModelFaceNet, "the configured model is named alongside it")
	})
	t.Run("NoModelConfigured", func(t *testing.T) {
		hook := captureLog(t)
		w.auditMarkerEmbeddingModels("")

		// A blank stored model is legacy FaceNet, which nothing contradicts when no model
		// is configured, so those markers are counted rather than flagged.
		assert.NotContains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"), "not compatible")
	})
}

func TestFaces_auditMarkerDetectModels(t *testing.T) {
	w := NewFaces(config.TestConfig())

	// newDetectedMarker persists a face marker attributed to the specified detector.
	newDetectedMarker := func(t *testing.T, detector string) {
		t.Helper()

		m := &entity.Marker{
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			EmbedModel:     face.ModelSFace,
			DetectModel:    detector,
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		}

		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { entity.Db().Delete(m) })
	}

	t.Run("WarnsAboutAnotherDetector", func(t *testing.T) {
		// A detector nobody is running now placed landmarks the active one would not
		// reproduce. Telling the operator so is the only actionable output of the report,
		// so the level is asserted: an Info line here would read as normal.
		require.NotEqual(t, "retired-detector", face.ActiveDetector())
		newDetectedMarker(t, "retired-detector")

		hook := captureLog(t)
		w.auditMarkerDetectModels()

		warnings := strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n")
		assert.Contains(t, warnings, "retired-detector")
		assert.Contains(t, warnings, "not the active")
	})
	t.Run("ActiveDetectorIsNotAWarning", func(t *testing.T) {
		// The detector rather than the engine that runs it. Comparing against the engine name
		// would report every marker a real library holds as produced by a foreign detector,
		// because no detector is called "onnx".
		detector := face.ActiveDetector()
		require.NotEqual(t, face.EngineONNX, detector, "the active detector must not be the engine name")
		newDetectedMarker(t, detector)

		hook := captureLog(t)
		w.auditMarkerDetectModels()

		assert.NotContains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"), "not the active")
		assert.Contains(t, strings.Join(loggedMessages(hook, logrus.InfoLevel), "\n"), detector)
	})
	t.Run("UnrecordedDetector", func(t *testing.T) {
		// A marker written before the column existed records nothing, which is the state an
		// operator has to be able to see before anything relies on the column.
		newDetectedMarker(t, "")

		hook := captureLog(t)
		w.auditMarkerDetectModels()

		assert.Contains(t, strings.Join(loggedMessages(hook, logrus.InfoLevel), "\n"), "without a recorded detector")
	})
}

func TestFaces_auditProvenance(t *testing.T) {
	w := NewFaces(config.TestConfig())

	t.Run("ReportsDetectorsAsWellAsModels", func(t *testing.T) {
		// The three reports read separate tables, and the detector one is what the README
		// promises photoprism faces audit prints, so its wiring is pinned rather than
		// left to whichever reporter happens to call the next.
		m := &entity.Marker{
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			EmbedModel:     face.ModelSFace,
			DetectModel:    "wired-detector",
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		}

		require.NoError(t, entity.Db().Create(m).Error)

		t.Cleanup(func() { entity.Db().Delete(m) })

		hook := captureLog(t)
		w.auditProvenance()

		logged := strings.Join(loggedMessages(hook, logrus.InfoLevel), "\n")
		assert.Contains(t, logged, "wired-detector", "the detector report runs")
		assert.Contains(t, logged, "embedding model", "the model reports still run")
	})
}

// TestFaces_auditMarkerThumbSizes pins that the audit reports markers judged by their detection
// size rather than by what their embedding was sampled from. It only counts: synthesizing the
// value would write today's rendition choice against a vector produced under another one.
func TestFaces_auditMarkerThumbSizes(t *testing.T) {
	w := NewFaces(config.TestConfig())

	hook := test.NewGlobal()
	t.Cleanup(hook.Reset)

	m := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        "fs6sg6bw45bnlqdw",
		MarkerType:     entity.MarkerFace,
		MarkerSrc:      entity.SrcImage,
		Size:           200,
		ThumbSize:      -1,
		Score:          100,
		EmbedModel:     face.EmbeddingModelName(),
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	require.NoError(t, entity.Db().Create(m).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(m) })

	w.auditMarkerThumbSizes()

	var reported int

	for _, entry := range hook.AllEntries() {
		if strings.Contains(entry.Message, "no recorded sample size") {
			reported++
			assert.Equal(t, logrus.InfoLevel, entry.Level)
		}
	}

	assert.Equal(t, 1, reported, "a marker the bar falls back on must be named")
	assert.NotPanics(t, func() { (*Faces)(nil).auditMarkerThumbSizes() })
}

// TestFaces_auditMarkerSampleShortfall covers the one state that leaves no trace in the vectors: a
// marker embedded from an upscaled crop is indistinguishable from one that was not, so the audit is
// where an operator finds out that their thumbnails, or a migration that could not write one, cost
// them recognition.
func TestFaces_auditMarkerSampleShortfall(t *testing.T) {
	w := NewFaces(config.TestConfig())

	restore := face.ClusterSizeThreshold
	t.Cleanup(func() { face.ClusterSizeThreshold = restore })

	file := &entity.File{
		FileUID:   rnd.GenerateUID('f'),
		PhotoUID:  rnd.GenerateUID('p'),
		FileName:  "audit-shortfall/large.jpg",
		FileRoot:  entity.RootOriginals,
		FileWidth: 4000,
	}
	require.NoError(t, entity.Db().Create(file).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(file) })

	// Sampled at 60 px while its original holds 400, so a re-sampling clears a bar of 112.
	marker := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        file.FileUID,
		MarkerType:     entity.MarkerFace,
		W:              0.1,
		H:              0.1,
		ThumbSize:      60,
		EmbeddingsJSON: []byte("[[0.1,0.2]]"),
	}
	require.NoError(t, entity.Db().Create(marker).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(marker) })

	t.Run("NamesWhatAMigrationWouldRecover", func(t *testing.T) {
		face.ClusterSizeThreshold = 112

		hook := captureLog(t)
		w.auditMarkerSampleShortfall()

		reported := strings.Join(loggedMessages(hook, logrus.InfoLevel), "\n")
		assert.Contains(t, reported, "sampled below the 112 px clustering size")
		assert.Contains(t, reported, "faces migrate")
	})
	t.Run("NothingToReport", func(t *testing.T) {
		// A bar every measured marker clears is not a finding, so it stays out of the report.
		face.ClusterSizeThreshold = 1

		hook := captureLog(t)
		w.auditMarkerSampleShortfall()

		assert.Empty(t, loggedMessages(hook, logrus.InfoLevel))
	})
	t.Run("NilWorker", func(t *testing.T) {
		assert.NotPanics(t, func() { (*Faces)(nil).auditMarkerSampleShortfall() })
	})
}
