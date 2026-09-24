package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb/crop"
)

// newRedetectTestMarker creates a face marker with a foreign detection and removes it after the test.
func newRedetectTestMarker(t *testing.T, src string) (*Marker, File) {
	t.Helper()

	file := FileFixtures.Get("exampleFileName.jpg")
	m := NewMarker(file, crop.NewArea("face", 0.1, 0.1, 0.2, 0.2), "", src, MarkerFace, 100, 50)
	require.NotNil(t, m)

	m.MarkerName, m.MarkerReview = "Jane", true
	m.FaceID, m.FaceDist = "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS", 0.3
	m.SetEmbeddings(face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName(), "centerface")

	require.NoError(t, m.Create())
	t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

	return m, file
}

// newRedetectTestFace returns a detection close to the test marker.
func newRedetectTestFace() face.Face {
	return face.Face{
		Rows: 100, Cols: 100, Score: 90, Area: face.NewArea("face", 22, 22, 22),
		Embeddings: face.RandomEmbeddings(1, face.RegularFace), EmbedModel: face.EmbeddingModelName(),
		DetectModel: face.DetectorYuNet, ThumbSize: 720, EmbedDetail: 80,
	}
}

func TestMarker_Redetect(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, file := newRedetectTestMarker(t, SrcImage)
		f := newRedetectTestFace()

		changed, err := m.Redetect(f, file, false)

		require.NoError(t, err)
		assert.True(t, changed)

		found := FindMarker(m.MarkerUID)
		require.NotNil(t, found)

		area := f.CropArea()
		assert.Equal(t, area.X, found.X)
		assert.Equal(t, area.W, found.W)
		assert.Equal(t, area.Thumb(file.FileHash), found.Thumb)
		assert.Equal(t, f.Size(), found.Size)
		assert.Equal(t, face.DetectorYuNet, found.DetectModel)
		assert.Equal(t, 90, found.Score)
		assert.Equal(t, 720, found.ThumbSize)
		assert.Equal(t, 80, found.EmbedDetail)
		assert.Equal(t, f.Embeddings.JSON(), []byte(found.EmbeddingsJSON))
		assert.Equal(t, "", found.FaceID)
		assert.Equal(t, -1.0, found.FaceDist)

		// The name and review state are not the detection's to change.
		assert.Equal(t, "Jane", found.MarkerName)
		assert.False(t, found.MarkerInvalid)
		assert.True(t, found.MarkerReview)
	})
	t.Run("Unchanged", func(t *testing.T) {
		m, file := newRedetectTestMarker(t, SrcImage)
		f := newRedetectTestFace()

		changed, err := m.Redetect(f, file, false)
		require.NoError(t, err)
		require.True(t, changed)

		updatedAt := FindMarker(m.MarkerUID).UpdatedAt

		changed, err = m.Redetect(f, file, false)

		require.NoError(t, err)
		assert.False(t, changed)
		assert.Equal(t, updatedAt, FindMarker(m.MarkerUID).UpdatedAt)
	})
	t.Run("ManualKeepsArea", func(t *testing.T) {
		m, file := newRedetectTestMarker(t, SrcManual)
		f := newRedetectTestFace()

		changed, err := m.Redetect(f, file, false)

		require.NoError(t, err)
		assert.True(t, changed)

		found := FindMarker(m.MarkerUID)
		require.NotNil(t, found)
		assert.Equal(t, float32(0.1), found.X)
		assert.Equal(t, float32(0.2), found.W)
		assert.Equal(t, 100, found.Size)
		assert.Equal(t, face.DetectorYuNet, found.DetectModel)
	})
	t.Run("KeepArea", func(t *testing.T) {
		m, file := newRedetectTestMarker(t, SrcImage)

		changed, err := m.Redetect(newRedetectTestFace(), file, true)

		require.NoError(t, err)
		assert.True(t, changed)

		found := FindMarker(m.MarkerUID)
		require.NotNil(t, found)
		assert.Equal(t, float32(0.1), found.X)
		assert.Equal(t, 90, found.Score)
		assert.Equal(t, face.DetectorYuNet, found.DetectModel)
	})
	t.Run("RejectedKeepsArea", func(t *testing.T) {
		m, file := newRedetectTestMarker(t, SrcImage)
		require.NoError(t, m.Updates(Values{"marker_invalid": true}))
		m.MarkerInvalid = true

		changed, err := m.Redetect(newRedetectTestFace(), file, false)

		require.NoError(t, err)
		assert.True(t, changed)

		found := FindMarker(m.MarkerUID)
		require.NotNil(t, found)
		assert.True(t, found.MarkerInvalid)
		assert.Equal(t, float32(0.1), found.X)
		assert.Equal(t, face.DetectorYuNet, found.DetectModel)
	})
	t.Run("SidecarNameKeepsArea", func(t *testing.T) {
		m, file := newRedetectTestMarker(t, SrcImage)
		require.NoError(t, m.Updates(Values{"subj_src": SrcXmp}))
		m.SubjSrc = SrcXmp

		_, err := m.Redetect(newRedetectTestFace(), file, false)

		require.NoError(t, err)
		assert.Equal(t, float32(0.1), FindMarker(m.MarkerUID).X)
	})
	t.Run("SidecarKeepsScore", func(t *testing.T) {
		m, file := newRedetectTestMarker(t, SrcXmp)

		changed, err := m.Redetect(newRedetectTestFace(), file, false)

		require.NoError(t, err)
		assert.True(t, changed)

		found := FindMarker(m.MarkerUID)
		require.NotNil(t, found)
		assert.Equal(t, 50, found.Score)
		assert.Equal(t, float32(0.1), found.X)
		assert.Equal(t, face.DetectorYuNet, found.DetectModel)
	})
	t.Run("InvalidEmbeddings", func(t *testing.T) {
		m, file := newRedetectTestMarker(t, SrcImage)
		f := newRedetectTestFace()
		f.Embeddings = face.Embeddings{}

		changed, err := m.Redetect(f, file, false)

		assert.Error(t, err)
		assert.False(t, changed)
		assert.Equal(t, "centerface", FindMarker(m.MarkerUID).DetectModel)
	})
	t.Run("NotAFace", func(t *testing.T) {
		m := &Marker{MarkerUID: "ms6sg6b1wowuy3c3", MarkerType: MarkerLabel}

		_, err := m.Redetect(newRedetectTestFace(), File{}, false)
		assert.Error(t, err)
	})
	t.Run("NoMarker", func(t *testing.T) {
		var m *Marker

		_, err := m.Redetect(newRedetectTestFace(), File{}, false)
		assert.Error(t, err)
	})
}

func TestValidFaceEmbeddings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.True(t, validFaceEmbeddings(face.Face{Embeddings: face.RandomEmbeddings(1, face.RegularFace), EmbedModel: face.EmbeddingModelName()}))
	})
	t.Run("None", func(t *testing.T) {
		assert.False(t, validFaceEmbeddings(face.Face{}))
	})
	t.Run("Several", func(t *testing.T) {
		assert.False(t, validFaceEmbeddings(face.Face{Embeddings: face.RandomEmbeddings(2, face.RegularFace)}))
	})
	t.Run("WrongWidth", func(t *testing.T) {
		assert.False(t, validFaceEmbeddings(face.Face{Embeddings: face.Embeddings{{0.6, 0.8}}, EmbedModel: face.ModelFaceNet}))
	})
}

func TestMarker_RedetectArea(t *testing.T) {
	assert.True(t, (&Marker{MarkerType: MarkerFace, MarkerSrc: SrcImage}).redetectArea())
	assert.True(t, (&Marker{MarkerType: MarkerFace, MarkerSrc: SrcImage, SubjSrc: SrcManual}).redetectArea())
	assert.False(t, (&Marker{MarkerType: MarkerFace, MarkerSrc: SrcImage, MarkerInvalid: true}).redetectArea())
	assert.False(t, (&Marker{MarkerType: MarkerFace, MarkerSrc: SrcImage, SubjSrc: SrcXmp}).redetectArea())
	assert.False(t, (&Marker{MarkerType: MarkerFace, MarkerSrc: SrcManual}).redetectArea())
	assert.False(t, (&Marker{MarkerType: MarkerFace, MarkerSrc: SrcXmp}).redetectArea())
}
