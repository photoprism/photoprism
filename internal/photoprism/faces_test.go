package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

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
