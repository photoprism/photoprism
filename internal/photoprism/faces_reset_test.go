package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
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

	err := m.ResetAndReindex("invalid", nil)

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

	err := m.ResetAndReindex(face.DetectorAuto, nil)
	require.NoError(t, err)
	require.True(t, called)
	require.True(t, received.FacesOnly)
	require.Equal(t, face.EngineONNX, c.FaceEngine())
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

	require.NoError(t, m.ResetAndReindex("", nil))
	require.NoError(t, m.ResetAndReindex(face.DetectorNone, nil))
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

	err := m.ResetAndReindex(face.DetectorYuNet, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be used")
	assert.Equal(t, before, faceAndMarkerCount(t), "nothing may be removed when the request cannot be met")
}
