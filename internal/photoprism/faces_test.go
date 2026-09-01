package photoprism

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
)

// isolatedTestFaces returns a worker on a database of its own, so one test's clusters cannot
// reach another's. The fixtures are still seeded, so tests scope their assertions by subject.
func isolatedTestFaces(t *testing.T, name string) *Faces {
	t.Helper()

	oldCfg := Config()
	c := config.NewMinimalTestConfigWithDb(name, t.TempDir())

	t.Cleanup(func() {
		_ = c.CloseDb()

		if oldCfg != nil {
			oldCfg.RegisterDb()
		}
	})

	return NewFaces(c)
}

func TestFaces_Start(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	opt := FacesOptions{
		Force:     true,
		Threshold: 1,
	}

	err := m.Start(opt)

	if err != nil {
		t.Fatal(err)
	}
}

func TestFaces_StartDefault(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	err := m.StartDefault()

	if err != nil {
		t.Fatal(err)
	}
}

func TestFaces_start(t *testing.T) {
	c := config.TestConfig()
	m := NewFaces(c)

	_, err := m.start(FacesOptions{Force: true, Threshold: 1})
	require.NoError(t, err)

	var invalid *Faces
	_, err = invalid.start(FacesOptions{})
	require.Error(t, err)
}

// TestFaces_startCountsAssignedMarkers pins that a pass reports the subjects it propagated from
// clusters that already carry one. query.MatchFaceMarkers writes subj_uid without going through
// Updated, so a run whose only work was this would otherwise read as idle and stop the settle loop.
func TestFaces_startCountsAssignedMarkers(t *testing.T) {
	w := isolatedTestFaces(t, "faces_assigned")

	marker := entity.MarkerFixtures.Get("actress-a-2")
	require.NotEmpty(t, marker.FaceID)

	// The one state MatchFaceMarkers acts on: an automatic marker in a named cluster that does
	// not carry that cluster's subject yet.
	require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
		Where("marker_uid = ?", marker.MarkerUID).
		UpdateColumns(entity.Values{"subj_uid": "", "subj_src": entity.SrcAuto}).Error)

	result, err := w.start(FacesOptions{})
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Assigned, 1, "the marker took its cluster's subject")
	assert.True(t, result.Moved(), "an assignment is work a later pass builds on")

	assigned := entity.FindMarker(marker.MarkerUID)
	require.NotNil(t, assigned)
	assert.Equal(t, marker.SubjUID, assigned.SubjUID)
}

func TestFaces_startBlocked(t *testing.T) {
	t.Run("PausedWhileTheLibraryDisagrees", func(t *testing.T) {
		// Clustering and matching compare stored vectors, so both wait for the migration
		// that makes them comparable rather than running and finding nothing. A run that
		// completed would clear the flag below, so that is what tells the two apart.
		t.Cleanup(face.UnblockEmbeddings)
		face.BlockEmbeddings("12 marker(s) use facenet, but this instance is configured for sface")

		entity.UpdateFaces.Store(true)
		t.Cleanup(func() { entity.UpdateFaces.Store(false) })

		w := NewFaces(config.TestConfig())

		_, err := w.start(FacesOptions{})
		require.NoError(t, err)
		assert.True(t, entity.UpdateFaces.Load(), "a paused run must leave the work outstanding")
	})
	t.Run("RunsWhenNothingIsBlocked", func(t *testing.T) {
		// The positive control: the same call with no block clears the flag, so the case
		// above cannot pass just because clustering found nothing to do.
		face.UnblockEmbeddings()

		entity.UpdateFaces.Store(true)
		t.Cleanup(func() { entity.UpdateFaces.Store(false) })

		w := NewFaces(config.TestConfig())

		_, err := w.start(FacesOptions{})
		require.NoError(t, err)
		assert.False(t, entity.UpdateFaces.Load())
	})
}

// TestFaces_StartHoldsOffOnAMigration pins the lock check on both sides. It sits on Start rather
// than on start because a migration calls the unexported one directly to rebuild its clusters:
// checking it there would leave every replacement cluster unbuilt, which is worse than the race
// the lock exists to prevent.
func TestFaces_StartHoldsOffOnAMigration(t *testing.T) {
	c := config.TestConfig()
	w := NewFaces(c)

	lock, err := mutex.AcquireFileLock(c.FacesLockFile(), "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)

	t.Run("StartHoldsOff", func(t *testing.T) {
		require.NoError(t, w.Start(FacesOptions{Force: true}))
		assert.False(t, mutex.FacesWorker.Running(), "the worker must not have been started")
	})
	t.Run("StartRunsWhileTheMigrationHoldsIt", func(t *testing.T) {
		// The migration's own re-clustering goes through this path while it holds the lock.
		_, err := w.start(FacesOptions{Force: true})
		assert.NoError(t, err)
	})
	t.Run("StartRunsOnceReleased", func(t *testing.T) {
		lock.Release()
		assert.NoError(t, w.Start(FacesOptions{Force: true}))
	})
}

// TestFaces_StartAfterAMigrationInAnotherProcess pins what happens once the lock is gone: the
// migration recorded its target in "options.yml" and this process still holds the model it
// replaced, so the pass pauses rather than clustering vectors of two different lengths.
func TestFaces_StartAfterAMigrationInAnotherProcess(t *testing.T) {
	c := newMigrateTestConfig(t, "facessuperseded")
	t.Cleanup(face.UnblockEmbeddings)

	loaded := c.FaceModel()
	require.NotEqual(t, face.ModelNone, loaded)

	migrated := face.ModelFaceNet

	if loaded == migrated {
		migrated = face.ModelSFace
	}

	require.NoError(t, os.WriteFile(c.OptionsYaml(), []byte("FaceModel: "+migrated+"\n"), fs.ModeConfigFile))

	require.NoError(t, NewFaces(c).Start(FacesOptions{Force: true}))

	require.True(t, face.EmbeddingsBlocked())
	assert.Contains(t, face.EmbeddingsBlockedReason(), migrated)
}

func TestFaces_reportOnce(t *testing.T) {
	w := NewFaces(config.TestConfig())

	t.Run("FirstTime", func(t *testing.T) {
		assert.True(t, w.reportOnce("probe", 3))
	})
	t.Run("SameValue", func(t *testing.T) {
		// A worker that wakes every few minutes must not repeat an unchanged condition.
		assert.False(t, w.reportOnce("probe", 3))
	})
	t.Run("ChangedValue", func(t *testing.T) {
		assert.True(t, w.reportOnce("probe", 4))
		assert.False(t, w.reportOnce("probe", 4))
	})
	t.Run("OtherKey", func(t *testing.T) {
		assert.True(t, w.reportOnce("other", 4))
	})
	t.Run("NilWorker", func(t *testing.T) {
		assert.False(t, (*Faces)(nil).reportOnce("probe", 1))
	})
}

// TestFaces_StartRefreshesSubjectCounts covers the counts the people views order and filter on.
//
// Nothing else refreshes them on a CLI-only run, so a person this run assigned correctly would sort
// last or be filtered out - which is how a mistyped person stayed invisible long enough to look as
// though naming had failed. Gated on the run having moved something, so an idle wake pays nothing.
func TestFaces_StartRefreshesSubjectCounts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	w := isolatedTestFaces(t, "faces_counts")

	subj := entity.SubjectFixtures.Get("actress-1")

	// A count nothing could have computed, so a stale read is distinguishable from a fresh one.
	require.NoError(t, entity.UnscopedDb().Model(&entity.Subject{}).
		Where("subj_uid = ?", subj.SubjUID).
		UpdateColumns(entity.Values{"file_count": 4242, "photo_count": 4242}).Error)

	require.NoError(t, w.Start(FacesOptions{Force: true}))

	stored := entity.FindSubject(subj.SubjUID)
	require.NotNil(t, stored)

	assert.NotEqual(t, 4242, stored.FileCount, "a run that moved markers has to refresh the counts")

	// The refreshed value has to be the one UpdateSubjectCounts computes, not merely different.
	var want int
	require.NoError(t, entity.UnscopedDb().Raw(`SELECT COUNT(DISTINCT f.id) FROM files f
		JOIN photos p ON p.id = f.photo_id AND p.deleted_at IS NULL AND p.photo_private = 0
		JOIN markers m ON f.file_uid = m.file_uid AND m.subj_uid = ?
		WHERE m.marker_invalid = 0 AND f.deleted_at IS NULL`, subj.SubjUID).Row().Scan(&want))

	assert.Equal(t, want, stored.FileCount)
}

// TestFaces_StartRefreshesCountsAfterRecognition covers the ordinary case for a library whose
// clusters are already named: nothing detaches or reassigns a marker, so the run's whole output is
// the recognition step writing subj_uid - which the counts are computed from.
//
// Without --force, deliberately: forcing recomputes regardless and would pass either way.
func TestFaces_StartRefreshesCountsAfterRecognition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	w := isolatedTestFaces(t, "faces_counts_recognized")

	// Settle the library first, so the run under test has nothing else left to move.
	require.NoError(t, w.Start(FacesOptions{Force: true}))

	subj := entity.SubjectFixtures.Get("actress-1")

	// A marker a named cluster holds, with its person taken away: that is what an automatic
	// assignment looks like before the recognition step restores it.
	marker := entity.Marker{}
	require.NoError(t, entity.UnscopedDb().
		Joins("JOIN faces ON faces.id = markers.face_id AND faces.subj_uid = ?", subj.SubjUID).
		Where("markers.subj_src = ?", entity.SrcAuto).
		Where("markers.marker_invalid = 0").
		First(&marker).Error)

	require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
		Where("marker_uid = ?", marker.MarkerUID).
		UpdateColumns(entity.Values{"subj_uid": ""}).Error)

	require.NoError(t, entity.UnscopedDb().Model(&entity.Subject{}).
		Where("subj_uid = ?", subj.SubjUID).
		UpdateColumns(entity.Values{"file_count": 4242}).Error)

	require.NoError(t, w.Start(FacesOptions{Force: false}))

	recognized := entity.Marker{}
	require.NoError(t, entity.UnscopedDb().First(&recognized, "marker_uid = ?", marker.MarkerUID).Error)
	require.Equal(t, subj.SubjUID, recognized.SubjUID, "the recognition step has to reassign the marker")

	stored := entity.FindSubject(subj.SubjUID)
	require.NotNil(t, stored)

	assert.NotEqual(t, 4242, stored.FileCount, "a run that recognized markers has to refresh the counts")
}

// TestFaces_StartSkipsCountsWhenNothingMoved pins the gate, so a scheduled worker that finds
// nothing to do does not run the join over markers and files on every wake.
func TestFaces_StartSkipsCountsWhenNothingMoved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	w := isolatedTestFaces(t, "faces_counts_idle")

	// Settle the library first, so the second run has nothing left to move.
	require.NoError(t, w.Start(FacesOptions{Force: true}))

	subj := entity.SubjectFixtures.Get("actress-1")

	require.NoError(t, entity.UnscopedDb().Model(&entity.Subject{}).
		Where("subj_uid = ?", subj.SubjUID).
		UpdateColumns(entity.Values{"file_count": 4242}).Error)

	require.NoError(t, w.Start(FacesOptions{Force: false}))

	stored := entity.FindSubject(subj.SubjUID)
	require.NotNil(t, stored)

	assert.Equal(t, 4242, stored.FileCount, "an idle run must not pay for the refresh")

	// Forcing is the recompute gesture, and the escape hatch for drift that arose outside the
	// worker - an interrupted run, or a photo turned private, moves no marker to notice.
	require.NoError(t, w.Start(FacesOptions{Force: true}))

	stored = entity.FindSubject(subj.SubjUID)
	require.NotNil(t, stored)

	assert.NotEqual(t, 4242, stored.FileCount, "a forced run recomputes even when nothing moved")
}
