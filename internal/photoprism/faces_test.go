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
	c := config.NewMinimalTestConfigWithDbTTest(name, t.TempDir(), t)

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
