package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestFaces_Reset(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	err := m.Reset()

	if err != nil {
		t.Fatal(err)
	}
}

func TestFaces_ResetAndReindex_InvalidDetector(t *testing.T) {
	c := config.TestConfig()
	m := NewFaces(c)

	before := faceAndMarkerCount(t)

	err := m.ResetAndReindex("invalid", nil, false)

	require.Error(t, err)
	// The message matters: without it a later failure inside the reset reads as validation.
	assert.Contains(t, err.Error(), "unsupported face detector")
	assert.Equal(t, before, faceAndMarkerCount(t), "an unusable detector must be refused before anything is removed")
}

// faceAndMarkerCount returns how many clusters and face markers the index holds, which is what a reset
// removes. Asserting on it is what tells a refusal apart from a reset that failed afterwards.
func faceAndMarkerCount(t *testing.T) [2]int {
	t.Helper()

	var faces, markers int64

	require.NoError(t, entity.Db().Model(&entity.Face{}).Count(&faces).Error)
	require.NoError(t, entity.Db().Model(&entity.Marker{}).Where("marker_type = ?", entity.MarkerFace).Count(&markers).Error)

	return [2]int{int(faces), int(markers)}
}

func TestFaces_ResetAndReindex_Detect(t *testing.T) {
	defer func(prev func(*Index, IndexOptions) (fs.Done, int, error)) {
		runFacesReindex = prev
	}(runFacesReindex)

	called := false
	var received IndexOptions
	runFacesReindex = func(idx *Index, opt IndexOptions) (fs.Done, int, error) {
		called = true
		received = opt
		return fs.Done{}, 0, nil
	}

	c := config.TestConfig()
	m := NewFaces(c)
	writeOriginalsTestFile(t, c)

	err := m.ResetAndReindex(face.DetectorAuto, NewIndex(c, NewConvert(c), NewFiles(), NewPhotos()), false)
	require.NoError(t, err)
	require.True(t, called)
	require.True(t, received.FacesOnly)
	require.True(t, received.RegenerateFaces)
	require.NotNil(t, received.FaceRegeneration)
	require.Equal(t, face.EngineONNX, c.FaceEngine())
}

// writeOriginalsTestFile adds a file to the originals folder until the test ends, since a
// regeneration refuses to start on an empty one.
func writeOriginalsTestFile(t *testing.T, c *config.Config) {
	t.Helper()

	name := filepath.Join(c.OriginalsPath(), "regenerate-test.txt")
	require.NoError(t, fs.WriteString(name, "test"))
	t.Cleanup(func() { _ = os.Remove(name) })
}

func TestFaces_RegenerateRefused(t *testing.T) {
	c := config.TestConfig()
	w := NewFaces(c)
	ind := NewIndex(c, NewConvert(c), NewFiles(), NewPhotos())

	newOpt := func() IndexOptions {
		opt := IndexOptionsFacesOnly(c)
		opt.RegenerateFaces = true
		return opt
	}

	t.Run("EmptyOriginals", func(t *testing.T) {
		err := w.regenerateRefused(ind, newOpt())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "is empty")
	})

	writeOriginalsTestFile(t, c)

	t.Run("Success", func(t *testing.T) {
		if c.FaceDetector() == face.DetectorNone || face.EmbeddingsDisabled() {
			t.Skip("faces: skipping, the detector or the embedder is not available here")
		}

		assert.NoError(t, w.regenerateRefused(ind, newOpt()))
	})
	t.Run("NoDetector", func(t *testing.T) {
		detector := c.Options().FaceDetector
		t.Cleanup(func() { c.Options().FaceDetector = detector })
		c.Options().FaceDetector = face.DetectorNone

		err := w.regenerateRefused(ind, newOpt())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be used")
	})
	t.Run("EmbeddingsDisabled", func(t *testing.T) {
		prev := face.ConfiguredModel()
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelNone}))
		t.Cleanup(func() { require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: prev})) })

		err := w.regenerateRefused(ind, newOpt())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "embeddings are disabled")
	})
	t.Run("MissingFolder", func(t *testing.T) {
		opt := newOpt()
		opt.Path = "missing-regenerate-folder"

		err := w.regenerateRefused(ind, opt)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
	t.Run("DetectionDisabled", func(t *testing.T) {
		opt := newOpt()
		opt.DetectFaces = false

		err := w.regenerateRefused(ind, opt)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "detection is disabled")
	})
	t.Run("EmbeddingsBlocked", func(t *testing.T) {
		t.Cleanup(face.UnblockEmbeddings)
		face.BlockEmbeddings("12 marker(s) use facenet, but this instance is configured for sface")

		err := w.regenerateRefused(ind, newOpt())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "until the library is migrated")
	})
	t.Run("IndexRunning", func(t *testing.T) {
		require.NoError(t, mutex.IndexWorker.Start())
		t.Cleanup(mutex.IndexWorker.Stop)

		err := w.regenerateRefused(ind, newOpt())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already running")
	})
	t.Run("FacesLocked", func(t *testing.T) {
		lock, lockErr := mutex.AcquireFileLock(c.FacesLockFile(), "faces migration")
		require.NoError(t, lockErr)
		t.Cleanup(lock.Release)

		err := w.regenerateRefused(ind, newOpt())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "faces migration")
	})
	t.Run("NoIndex", func(t *testing.T) {
		err := w.regenerateRefused(nil, newOpt())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "index service unavailable")
	})
}

// TestFaces_ResetAndReindex_ResetOnly pins that naming no detector, or naming "none", resets
// without regenerating rather than being rejected.
func TestFaces_ResetAndReindex_ResetOnly(t *testing.T) {
	defer func(prev func(*Index, IndexOptions) (fs.Done, int, error)) {
		runFacesReindex = prev
	}(runFacesReindex)

	runFacesReindex = func(idx *Index, opt IndexOptions) (fs.Done, int, error) {
		t.Fatal("faces: must not reindex when no detector was named")
		return fs.Done{}, 0, nil
	}

	m := NewFaces(config.TestConfig())

	require.NoError(t, m.ResetAndReindex("", nil, false))
	require.NoError(t, m.ResetAndReindex(face.DetectorNone, nil, false))
}

// TestFaces_ResetAndReindex_NoDetector pins the order this runs in: a request to regenerate that
// cannot be met has to be refused before anything is removed, or it deletes every person and face
// and rebuilds nothing.
func TestFaces_ResetAndReindex_NoDetector(t *testing.T) {
	defer func(prev func(*Index, IndexOptions) (fs.Done, int, error)) {
		runFacesReindex = prev
	}(runFacesReindex)

	runFacesReindex = func(idx *Index, opt IndexOptions) (fs.Done, int, error) {
		t.Fatal("faces: must not reindex without a detector")
		return fs.Done{}, 0, nil
	}

	c := config.TestConfig()

	models := c.Options().ModelsPath
	detector := c.Options().FaceDetector
	t.Cleanup(func() { c.Options().ModelsPath, c.Options().FaceDetector = models, detector })

	c.Options().ModelsPath = t.TempDir()
	m := NewFaces(c)

	before := faceAndMarkerCount(t)

	err := m.ResetAndReindex(face.DetectorYuNet, nil, false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be used")
	assert.Equal(t, before, faceAndMarkerCount(t), "nothing may be removed when the request cannot be met")
}

// TestFaces_ResetAll covers the scope that returns a library to its pre-recognition state while
// keeping the markers, which is what makes it usable between clustering runs.
func TestFaces_ResetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Cleanup(entity.ResetTestFixtures)

	c := config.TestConfig()
	m := NewFaces(c)

	named := entity.MarkerFixtures.Get("actress-a-1")
	require.Equal(t, entity.SrcManual, named.SubjSrc)

	require.NoError(t, m.ResetAll())

	var faces int
	require.NoError(t, entity.Db().Model(&entity.Face{}).Count(&faces).Error)
	assert.Zero(t, faces, "no cluster of any source may survive")

	var assigned int
	require.NoError(t, entity.Db().Model(&entity.Marker{}).
		Where("marker_type = ? AND (subj_uid <> '' OR face_id <> '')", entity.MarkerFace).
		Count(&assigned).Error)
	assert.Zero(t, assigned, "no face marker may keep a subject or a cluster")

	var markers int
	require.NoError(t, entity.Db().Model(&entity.Marker{}).
		Where("marker_type = ?", entity.MarkerFace).Count(&markers).Error)
	assert.Positive(t, markers, "the markers themselves must survive, or this costs a reindex")
}

// TestFaces_ResetAll_OrphanPeople pins that the people whose names the reset clears are removed
// whatever their source, while a verified person keeps their row.
func TestFaces_ResetAll_OrphanPeople(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Cleanup(entity.ResetTestFixtures)

	c := config.TestConfig()
	m := NewFaces(c)

	// A person named in the UI is stored with the manual source, which the subject cleanup skips.
	manual := entity.SubjectFixtures.Get("actress-1")
	require.False(t, manual.Verified)
	require.NoError(t, entity.Db().Model(&entity.Subject{}).
		Where("subj_uid = ?", manual.SubjUID).UpdateColumn("subj_src", entity.SrcManual).Error)

	verified := entity.SubjectFixtures.Get("actor-1")
	require.NoError(t, entity.Db().Model(&entity.Subject{}).
		Where("subj_uid = ?", verified.SubjUID).UpdateColumn("verified", true).Error)

	require.NoError(t, m.ResetAll())

	var removed entity.Subject
	require.NoError(t, entity.UnscopedDb().Where("subj_uid = ?", manual.SubjUID).First(&removed).Error)
	assert.True(t, removed.Deleted(), "an unverified person left without a name must not survive")

	var kept entity.Subject
	require.NoError(t, entity.UnscopedDb().Where("subj_uid = ?", verified.SubjUID).First(&kept).Error)
	assert.False(t, kept.Deleted(), "a verified person keeps their row")

	restored := entity.FindSubjectByName(manual.SubjName, true)
	require.NotNil(t, restored, "naming the person again must restore the row")
	assert.Equal(t, manual.SubjUID, restored.SubjUID)
	assert.False(t, restored.Deleted())
}

// TestFaces_Reset_KeepsManualPeople pins that the default reset leaves a hand-named person alone,
// even one no marker references, since it never clears a hand-assigned name.
func TestFaces_Reset_KeepsManualPeople(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Cleanup(entity.ResetTestFixtures)

	c := config.TestConfig()
	m := NewFaces(c)

	orphan := entity.NewSubject("Reset Keeps Me", entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, orphan)
	require.NoError(t, orphan.Create())

	require.NoError(t, m.Reset())

	var kept entity.Subject
	require.NoError(t, entity.UnscopedDb().Where("subj_uid = ?", orphan.SubjUID).First(&kept).Error)
	assert.False(t, kept.Deleted(), "the default reset must not collect people")
}

// TestFaces_Reset_ClearsDanglingFaces pins that a reset leaves no marker pointing at a cluster it
// just deleted.
//
// The marker reset only reaches rows whose subject was assigned automatically, while the cluster
// delete reaches every automatic cluster - so a hand-named marker that sat on one kept a reference
// to a row that no longer existed. On a real library that was 11 of 12 hand-named markers.
func TestFaces_Reset_ClearsDanglingFaces(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Cleanup(entity.ResetTestFixtures)

	c := config.TestConfig()
	m := NewFaces(c)

	named := entity.MarkerFixtures.Get("actress-a-1")
	require.Equal(t, entity.SrcManual, named.SubjSrc, "the fixture this pins must be hand-named")
	require.NotEmpty(t, named.FaceID)

	// The cluster it points at is automatic, so the reset deletes it while keeping the marker.
	var cluster entity.Face
	require.NoError(t, entity.UnscopedDb().Where("id = ?", named.FaceID).First(&cluster).Error)
	require.Equal(t, entity.SrcAuto, cluster.FaceSrc)

	require.NoError(t, m.Reset())

	var marker entity.Marker
	require.NoError(t, entity.UnscopedDb().Where("marker_uid = ?", named.MarkerUID).First(&marker).Error)

	assert.Equal(t, named.SubjUID, marker.SubjUID, "the hand-assigned subject still stands")
	assert.Empty(t, marker.FaceID, "but not a reference to a cluster that was deleted")
	assert.Nil(t, marker.MatchedAt, "and the marker is up for matching again")

	var dangling int
	require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
		Where("marker_type = ? AND face_id <> ''", entity.MarkerFace).
		Where("face_id NOT IN (SELECT id FROM faces)").Count(&dangling).Error)
	assert.Zero(t, dangling, "no marker may point at a cluster that is gone")
}
