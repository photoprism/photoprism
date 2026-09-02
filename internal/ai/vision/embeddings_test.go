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
	empty   bool
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

	if e.empty {
		return nil
	}

	values := make([]float32, e.dims)
	values[0] = 1

	return face.NewEmbeddings([][]float32{values})
}

// Close does nothing.
func (e *stubEmbedder) Close() error { return nil }

// testFaces returns a detected face with landmarks and one without.
func testFaces(withLandmarks bool) face.Faces {
	return testFacesOfSize(withLandmarks, 200)
}

// testFacesOfSize returns a detected face of the given size in pixels of a 720 px detection image,
// which is what decides how much of the crop the source can supply.
func testFacesOfSize(withLandmarks bool, size int) face.Faces {
	f := face.Face{Rows: 720, Cols: 720, Score: 40, Area: face.NewArea("face", 300, 300, size)}

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
		assert.Zero(t, GenerateEmbeddings(embedder, fileName, faces, false))

		require.Len(t, embedder.sizes, 1)
		assert.Equal(t, 112, embedder.sizes[0].Dx())
		assert.Equal(t, 112, embedder.sizes[0].Dy())
		assert.False(t, faces[0].Embeddings.Empty())
		// Recorded beside the vector, in the pixels of the image it was sampled from.
		assert.Positive(t, faces[0].ThumbSize, "an embedded face records what it was sampled at")
	})
	t.Run("FallsBackWithoutLandmarks", func(t *testing.T) {
		// Alignment is impossible, so the plain thumb crop is used instead. It is counted because
		// nothing in the resulting vector says it was produced without pose normalization.
		embedder := &stubEmbedder{aligned: true, dims: 128}
		faces := testFaces(false)
		assert.Equal(t, 1, GenerateEmbeddings(embedder, fileName, faces, false))

		require.Len(t, embedder.sizes, 1)
		assert.Equal(t, face.CropSize.Width, embedder.sizes[0].Dx())
		assert.False(t, faces[0].Embeddings.Empty())
		assert.Positive(t, faces[0].ThumbSize, "the unaligned branch records it too")
	})
	t.Run("RecordsFullDetailForALargeFace", func(t *testing.T) {
		// 1.jpg is 1024 px wide, so a face 200 px across in the 720 px detection image covers 284
		// px there - more than the 112 px template needs, and the surplus is not detail either.
		embedder := &stubEmbedder{aligned: true, dims: 128}
		faces := testFaces(true)
		GenerateEmbeddings(embedder, fileName, faces, false)

		require.False(t, faces[0].Embeddings.Empty())
		assert.Equal(t, 284, faces[0].ThumbSize)
		assert.Equal(t, 100, faces[0].EmbedDetail)
	})
	t.Run("RecordsTheShortfallOfASmallFace", func(t *testing.T) {
		// The same picture and a 40 px face, which covers 57 px of it: the crop the model reads
		// is interpolated from half the pixels it asks for, and nothing in the vector says so.
		embedder := &stubEmbedder{aligned: true, dims: 128}
		faces := testFacesOfSize(true, 40)
		GenerateEmbeddings(embedder, fileName, faces, false)

		require.False(t, faces[0].Embeddings.Empty())
		assert.Equal(t, 57, faces[0].ThumbSize)
		assert.Equal(t, 51, faces[0].EmbedDetail)
	})
	t.Run("MeasuredAgainstTheFallbackCrop", func(t *testing.T) {
		// The unaligned branch resamples a 160 px box rather than the 112 px template, so the
		// same face is short by more there. Reading one crop width for both would misstate it.
		embedder := &stubEmbedder{aligned: true, dims: 128}
		faces := testFacesOfSize(false, 40)
		require.Equal(t, 1, GenerateEmbeddings(embedder, fileName, faces, false))

		assert.Equal(t, 57, faces[0].ThumbSize)
		assert.Equal(t, 36, faces[0].EmbedDetail)
	})
	t.Run("EmptyEmbeddingIsNotCounted", func(t *testing.T) {
		// The count describes vectors, so a crop the model returned nothing for is not one.
		embedder := &stubEmbedder{aligned: true, empty: true, dims: 128}
		faces := testFaces(false)
		assert.Zero(t, GenerateEmbeddings(embedder, fileName, faces, false))

		require.Len(t, embedder.sizes, 1, "the crop was still taken")
		assert.True(t, faces[0].Embeddings.Empty())
		assert.Zero(t, faces[0].EmbedDetail, "a crop no vector came out of describes nothing")
	})
	t.Run("UnalignedModelUsesThumbCrop", func(t *testing.T) {
		embedder := &stubEmbedder{aligned: false, dims: 512}
		faces := testFaces(true)
		// A box crop is what this model expects, so it is not a fallback and is not counted.
		assert.Zero(t, GenerateEmbeddings(embedder, fileName, faces, false))

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
		// No crop was produced at all, so there is no vector to describe as unaligned.
		assert.Zero(t, GenerateEmbeddings(embedder, filepath.Join(t.TempDir(), "missing.jpg"), faces, false))
		assert.Empty(t, embedder.sizes)
	})
}

func TestSmallestFaceArea(t *testing.T) {
	t.Run("SmallestWins", func(t *testing.T) {
		// The smallest face needs the widest source to fill a template, so it is the one
		// that decides which rendition is decoded for the whole set.
		faces := face.Faces{
			{Rows: 720, Cols: 720, Area: face.NewArea("face", 300, 300, 400)},
			{Rows: 720, Cols: 720, Area: face.NewArea("face", 300, 300, 60)},
			{Rows: 720, Cols: 720, Area: face.NewArea("face", 300, 300, 200)},
		}

		assert.InDelta(t, faces[1].CropArea().W, smallestFaceArea(faces).W, 1e-6)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Zero(t, smallestFaceArea(nil).W)
	})
}

func TestFaceCropImage(t *testing.T) {
	fileName, err := filepath.Abs("../face/testdata/1.jpg")
	require.NoError(t, err)

	src := image.NewRGBA(image.Rect(0, 0, 720, 720))

	t.Run("Aligned", func(t *testing.T) {
		faces := testFaces(true)
		img, srcWidth, aligned, cropErr := faceCropImage(&stubEmbedder{aligned: true, dims: 128}, src, fileName, &faces[0], 112, 112, false)
		require.NoError(t, cropErr)
		assert.True(t, aligned)
		assert.Equal(t, 112, img.Bounds().Dx())
		// The width reported is the source the crop was warped from, not the crop's own: the crop
		// is always the model input size, which would record a constant.
		assert.Equal(t, src.Bounds().Dx(), srcWidth)
	})
	t.Run("BoundingBox", func(t *testing.T) {
		faces := testFaces(true)
		img, srcWidth, aligned, cropErr := faceCropImage(&stubEmbedder{aligned: false, dims: 512}, src, fileName, &faces[0], 112, 112, false)
		require.NoError(t, cropErr)
		assert.False(t, aligned)
		assert.Equal(t, face.CropSize.Width, img.Bounds().Dx())
		// The unaligned branch selects its own rendition, so the width comes back from there
		// rather than from the image decoded for the aligned models.
		assert.NotEqual(t, face.CropSize.Width, srcWidth, "the crop size must not be recorded as the source")
	})
	t.Run("MissingThumb", func(t *testing.T) {
		faces := testFaces(false)
		_, _, aligned, cropErr := faceCropImage(&stubEmbedder{aligned: true, dims: 128}, src, filepath.Join(t.TempDir(), "missing.jpg"), &faces[0], 112, 112, false)
		require.Error(t, cropErr)
		assert.False(t, aligned, "a crop that failed was not aligned either")
	})
}
