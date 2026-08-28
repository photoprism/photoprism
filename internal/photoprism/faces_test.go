package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
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

	require.NoError(t, m.start(FacesOptions{Force: true, Threshold: 1}))

	var invalid *Faces
	require.Error(t, invalid.start(FacesOptions{}))
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

		require.NoError(t, w.start(FacesOptions{}))
		assert.True(t, entity.UpdateFaces.Load(), "a paused run must leave the work outstanding")
	})
	t.Run("RunsWhenNothingIsBlocked", func(t *testing.T) {
		// The positive control: the same call with no block clears the flag, so the case
		// above cannot pass just because clustering found nothing to do.
		face.UnblockEmbeddings()

		entity.UpdateFaces.Store(true)
		t.Cleanup(func() { entity.UpdateFaces.Store(false) })

		w := NewFaces(config.TestConfig())

		require.NoError(t, w.start(FacesOptions{}))
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
		assert.NoError(t, w.start(FacesOptions{Force: true}))
	})
	t.Run("StartRunsOnceReleased", func(t *testing.T) {
		lock.Release()
		assert.NoError(t, w.Start(FacesOptions{Force: true}))
	})
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
}
