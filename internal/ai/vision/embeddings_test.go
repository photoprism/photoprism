package vision

import (
	"image"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// stubEmbedder records the crops it receives so tests can assert how they were prepared.
type stubEmbedder struct {
	aligned bool
	dims    int
	sizes   []image.Rectangle
}

// ModelName returns a fixed model name.
func (e *stubEmbedder) ModelName() face.ModelName { return "stub" }

// Dims returns the configured embedding length.
func (e *stubEmbedder) Dims() int { return e.dims }

// CropSize returns the ArcFace input size.
func (e *stubEmbedder) CropSize() (int, int) { return 112, 112 }

// Aligned reports whether the stub requires aligned crops.
func (e *stubEmbedder) Aligned() bool { return e.aligned }

// Run records the crop bounds and returns a non-empty embedding.
func (e *stubEmbedder) Run(img image.Image) face.Embeddings {
	if img == nil {
		return nil
	}

	e.sizes = append(e.sizes, img.Bounds())

	values := make([]float32, e.dims)
	values[0] = 1

	return face.NewEmbeddings([][]float32{values})
}

// Close does nothing.
func (e *stubEmbedder) Close() error { return nil }

// testFaces returns a detected face with landmarks and one without.
func testFaces(withLandmarks bool) face.Faces {
	f := face.Face{Rows: 720, Cols: 720, Score: 40, Area: face.NewArea("face", 300, 300, 200)}

	if withLandmarks {
		var kps [face.NumLandmarks * 2]float32

		for i, p := range face.ArcFaceTemplate {
			kps[i*2] = float32(p[0]*1.5 + 220)
			kps[i*2+1] = float32(p[1]*1.5 + 220)
		}

		f.Eyes, f.Landmarks = face.LandmarkAreas(kps, f.Area.Scale)
	}

	return face.Faces{f}
}

func TestGenerateEmbeddings(t *testing.T) {
	fileName, err := filepath.Abs("../face/testdata/1.jpg")
	require.NoError(t, err)

	t.Run("AlignedCrop", func(t *testing.T) {
		embedder := &stubEmbedder{aligned: true, dims: 128}
		faces := testFaces(true)
		GenerateEmbeddings(embedder, fileName, faces, false)

		require.Len(t, embedder.sizes, 1)
		assert.Equal(t, 112, embedder.sizes[0].Dx())
		assert.Equal(t, 112, embedder.sizes[0].Dy())
		assert.False(t, faces[0].Embeddings.Empty())
	})
	t.Run("FallsBackWithoutLandmarks", func(t *testing.T) {
		// Alignment is impossible, so the plain thumb crop is used instead.
		embedder := &stubEmbedder{aligned: true, dims: 128}
		faces := testFaces(false)
		GenerateEmbeddings(embedder, fileName, faces, false)

		require.Len(t, embedder.sizes, 1)
		assert.Equal(t, face.CropSize.Width, embedder.sizes[0].Dx())
		assert.False(t, faces[0].Embeddings.Empty())
	})
	t.Run("UnalignedModelUsesThumbCrop", func(t *testing.T) {
		embedder := &stubEmbedder{aligned: false, dims: 512}
		faces := testFaces(true)
		GenerateEmbeddings(embedder, fileName, faces, false)

		require.Len(t, embedder.sizes, 1)
		assert.Equal(t, face.CropSize.Width, embedder.sizes[0].Dx())
	})
	t.Run("SkipsFacesWithoutArea", func(t *testing.T) {
		embedder := &stubEmbedder{aligned: false, dims: 512}
		faces := face.Faces{{Rows: 720, Cols: 720}}
		GenerateEmbeddings(embedder, fileName, faces, false)

		assert.Empty(t, embedder.sizes)
		assert.True(t, faces[0].Embeddings.Empty())
	})
	t.Run("NilEmbedder", func(t *testing.T) {
		faces := testFaces(true)
		GenerateEmbeddings(nil, fileName, faces, false)
		assert.True(t, faces[0].Embeddings.Empty())
	})
	t.Run("NoFaces", func(t *testing.T) {
		embedder := &stubEmbedder{aligned: true, dims: 128}
		GenerateEmbeddings(embedder, fileName, face.Faces{}, false)
		assert.Empty(t, embedder.sizes)
	})
	t.Run("UndecodableFile", func(t *testing.T) {
		embedder := &stubEmbedder{aligned: true, dims: 128}
		faces := testFaces(true)
		GenerateEmbeddings(embedder, filepath.Join(t.TempDir(), "missing.jpg"), faces, false)
		assert.Empty(t, embedder.sizes)
	})
}

func TestFaceCropImage(t *testing.T) {
	fileName, err := filepath.Abs("../face/testdata/1.jpg")
	require.NoError(t, err)

	src := image.NewRGBA(image.Rect(0, 0, 720, 720))

	t.Run("Aligned", func(t *testing.T) {
		faces := testFaces(true)
		img, cropErr := faceCropImage(&stubEmbedder{aligned: true, dims: 128}, src, fileName, &faces[0], 112, 112, false)
		require.NoError(t, cropErr)
		assert.Equal(t, 112, img.Bounds().Dx())
	})
	t.Run("BoundingBox", func(t *testing.T) {
		faces := testFaces(true)
		img, cropErr := faceCropImage(&stubEmbedder{aligned: false, dims: 512}, src, fileName, &faces[0], 112, 112, false)
		require.NoError(t, cropErr)
		assert.Equal(t, face.CropSize.Width, img.Bounds().Dx())
	})
	t.Run("MissingThumb", func(t *testing.T) {
		faces := testFaces(false)
		_, cropErr := faceCropImage(&stubEmbedder{aligned: true, dims: 128}, src, filepath.Join(t.TempDir(), "missing.jpg"), &faces[0], 112, 112, false)
		require.Error(t, cropErr)
	})
}
