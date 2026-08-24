package photoprism

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
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

func TestFaces_ResetAndReindex_InvalidEngine(t *testing.T) {
	c := config.TestConfig()
	m := NewFaces(c)

	err := m.ResetAndReindex("invalid", nil)
	require.Error(t, err)
}

func TestFaces_ResetAndReindex_Auto(t *testing.T) {
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

	err := m.ResetAndReindex(face.EngineAuto, nil)
	require.NoError(t, err)
	require.True(t, called)
	require.True(t, received.FacesOnly)
	require.Equal(t, face.EngineONNX, c.FaceEngine())
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
	c.Options().ModelsPath = t.TempDir()
	m := NewFaces(c)

	err := m.ResetAndReindex(face.EngineONNX, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "no face detector")
}
