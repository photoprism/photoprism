package face

import (
	"image"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/pkg/fs"
)

var embeddingModelsPath, _ = filepath.Abs("../../../assets/models")

// newTestONNXEmbedder returns an embedder for the named model, or skips the test
// when its weights have not been installed.
func newTestONNXEmbedder(t *testing.T, name ModelName) Embedder {
	t.Helper()

	m := FindEmbeddingModel(name)
	require.NotNil(t, m)

	if !m.Installed(embeddingModelsPath) {
		t.Skipf("faces: %s is not installed", name)
	}

	embedder, err := NewONNXEmbedder(EmbedderSettings{
		Name:      name,
		Model:     m,
		ModelPath: m.FilePath(embeddingModelsPath),
		Threads:   1,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = embedder.Close() })

	return embedder
}

// useTestDetector activates the ONNX detector for the duration of the test and
// restores the previous engine afterwards, so tests that assert on the global
// engine state are not affected by test order.
func useTestDetector(t *testing.T) {
	t.Helper()

	if ActiveEngineName() == EngineONNX {
		return
	}

	prev := UseEngine(nil)

	t.Cleanup(func() {
		if current := UseEngine(prev); current != nil {
			_ = current.Close()
		}
	})

	if err := ConfigureEngine(EngineSettings{
		Name: EngineONNX,
		ONNX: ONNXOptions{ModelPath: detectorModelPath, Threads: 1},
	}); err != nil {
		t.Skipf("faces: skipping detector-dependent test: %s", err)
	}
}

// detectTestFace returns the highest scoring face in an image along with the image itself.
func detectTestFace(t *testing.T, fileName string) (image.Image, *Face) {
	t.Helper()

	useTestDetector(t)

	faces, err := Detect(fileName, 20)
	require.NoError(t, err)
	require.NotEmpty(t, faces)

	img, _, err := fs.DecodeImageFile(fileName)
	require.NoError(t, err)

	return img, &faces[0]
}

// rotateTestImage rotates an image around its center by the given angle in degrees.
func rotateTestImage(src image.Image, deg float64) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	out := image.NewRGBA(image.Rect(0, 0, width, height))

	rad := deg * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)
	cx := float64(width) / 2
	cy := float64(height) / 2

	for y := range height {
		for x := range width {
			dx := float64(x) - cx
			dy := float64(y) - cy
			out.Set(x, y, src.At(bounds.Min.X+int(cos*dx+sin*dy+cx), bounds.Min.Y+int(-sin*dx+cos*dy+cy)))
		}
	}

	return out
}

// writeTestJPEG encodes an image to a temporary JPEG file and returns its path.
func writeTestJPEG(t *testing.T, img image.Image) string {
	t.Helper()

	fileName := filepath.Join(t.TempDir(), "rotated.jpg")

	//nolint:gosec // G304: Test fixture written to a temporary directory.
	file, err := os.Create(fileName)
	require.NoError(t, err)

	require.NoError(t, jpeg.Encode(file, img, &jpeg.Options{Quality: 95}))
	require.NoError(t, file.Close())

	return fileName
}

func TestNewONNXEmbedder(t *testing.T) {
	t.Run("MissingModel", func(t *testing.T) {
		_, err := NewONNXEmbedder(EmbedderSettings{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing embedding model")
	})
	t.Run("NotAnONNXModel", func(t *testing.T) {
		_, err := NewONNXEmbedder(EmbedderSettings{Model: FindEmbeddingModel(ModelFaceNet), ModelPath: "facenet"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not an ONNX embedding model")
	})
	t.Run("EmptyPath", func(t *testing.T) {
		_, err := NewONNXEmbedder(EmbedderSettings{Model: FindEmbeddingModel(ModelSFace)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path is empty")
	})
	t.Run("MissingFile", func(t *testing.T) {
		_, err := NewONNXEmbedder(EmbedderSettings{
			Model:     FindEmbeddingModel(ModelSFace),
			ModelPath: filepath.Join(t.TempDir(), "missing.onnx"),
		})
		require.Error(t, err)
	})
	t.Run("SFace", func(t *testing.T) {
		embedder := newTestONNXEmbedder(t, ModelSFace)
		width, height := embedder.CropSize()
		assert.Equal(t, ModelSFace, embedder.ModelName())
		assert.Equal(t, 128, embedder.Dims())
		assert.Equal(t, 112, width)
		assert.Equal(t, 112, height)
		assert.True(t, embedder.Aligned())
	})
	t.Run("MissingModelInfo", func(t *testing.T) {
		m := *FindEmbeddingModel(ModelSFace)
		m.ONNX = nil

		_, err := NewONNXEmbedder(EmbedderSettings{Model: &m, ModelPath: "sface.onnx"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no ONNX model description")
	})
	t.Run("ChecksumMismatch", func(t *testing.T) {
		// A name match with a different artifact must not load: two unrelated models have
		// shipped under the same file name, and applying one model's preprocessing to the
		// other's weights cannot be detected once the embeddings are written.
		fileName := filepath.Join(t.TempDir(), "face_recognition_sface_2021dec.onnx")
		require.NoError(t, os.WriteFile(fileName, []byte("not the registered artifact"), fs.ModeFile))

		_, err := NewONNXEmbedder(EmbedderSettings{
			Name:      ModelSFace,
			Model:     FindEmbeddingModel(ModelSFace),
			ModelPath: fileName,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checksum")
	})
	t.Run("GraphContradictsRegistry", func(t *testing.T) {
		// A registered shape that disagrees with the graph means one of the two describes
		// a different model, which is a configuration error rather than a difference to
		// reconcile.
		m := FindEmbeddingModel(ModelSFace)

		if !m.Installed(embeddingModelsPath) {
			t.Skipf("faces: %s is not installed", ModelSFace)
		}

		mismatched := *m
		mismatched.ONNX = &onnx.ModelInfo{
			File:    m.ONNX.File,
			SHA256:  m.ONNX.SHA256,
			License: m.ONNX.License,
			Input:   m.ONNX.Input,
			Output:  &onnx.Output{Width: 512},
		}

		_, err := NewONNXEmbedder(EmbedderSettings{
			Name:      ModelSFace,
			Model:     &mismatched,
			ModelPath: m.FilePath(embeddingModelsPath),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "output width")
	})
}

// installedONNXModels returns the ONNX embedding models whose weights are present, so
// the contract tests below cover every model the environment can actually run instead
// of only the one that "make dep-models" installs.
func installedONNXModels(t *testing.T) []ModelName {
	t.Helper()

	var result []ModelName

	for _, name := range EmbeddingModelNames() {
		if m := FindEmbeddingModel(name); m.Runtime == RuntimeONNX && m.Installed(embeddingModelsPath) {
			result = append(result, name)
		}
	}

	if len(result) == 0 {
		t.Skip("faces: no ONNX embedding models are installed")
	}

	return result
}

func TestONNXEmbedder_Run(t *testing.T) {
	for _, name := range installedONNXModels(t) {
		t.Run(name, func(t *testing.T) {
			embedder := newTestONNXEmbedder(t, name)
			img, f := detectTestFace(t, filepath.Join("testdata", "1.jpg"))
			width, height := embedder.CropSize()

			crop, err := AlignedCrop(img, f, width, height)
			require.NoError(t, err)

			t.Run("Dimensions", func(t *testing.T) {
				embeddings := embedder.Run(crop)
				require.False(t, embeddings.Empty())
				assert.Len(t, embeddings[0], embedder.Dims())
			})
			t.Run("Normalized", func(t *testing.T) {
				embeddings := embedder.Run(crop)
				require.False(t, embeddings.Empty())
				assert.InDelta(t, 1, embeddings[0].Magnitude(), 0.0001)
			})
			t.Run("Deterministic", func(t *testing.T) {
				first := embedder.Run(crop)
				second := embedder.Run(crop)
				require.False(t, first.Empty())
				require.False(t, second.Empty())
				assert.InDelta(t, 0, first[0].Dist(second[0]), 0.0001)
			})
			t.Run("UnalignedCropAccepted", func(t *testing.T) {
				// Crops of a different size are resampled to the model input.
				small := image.NewRGBA(image.Rect(0, 0, 64, 64))
				embeddings := embedder.Run(small)
				require.False(t, embeddings.Empty())
				assert.Len(t, embeddings[0], embedder.Dims())
			})
			t.Run("NilImage", func(t *testing.T) {
				assert.True(t, embedder.Run(nil).Empty())
			})
		})
	}
}

func TestONNXEmbedder_AlignmentReducesPoseDistance(t *testing.T) {
	for _, name := range installedONNXModels(t) {
		t.Run(name, func(t *testing.T) {
			embedder := newTestONNXEmbedder(t, name)
			width, height := embedder.CropSize()
			fileName := filepath.Join("testdata", "1.jpg")

			// Embeds the largest face in an image, once landmark-aligned and once from the
			// detected bounding box, so the two preparations can be compared directly.
			embed := func(name string) (aligned, boxed Embedding) {
				img, f := detectTestFace(t, name)

				alignedCrop, err := AlignedCrop(img, f, width, height)
				require.NoError(t, err)

				alignedEmbeddings := embedder.Run(alignedCrop)
				require.False(t, alignedEmbeddings.Empty())

				size := f.Area.Scale
				box := image.NewRGBA(image.Rect(0, 0, size, size))
				originX := f.Area.Col - size/2
				originY := f.Area.Row - size/2

				for y := range size {
					for x := range size {
						box.Set(x, y, img.At(originX+x, originY+y))
					}
				}

				boxedEmbeddings := embedder.Run(box)
				require.False(t, boxedEmbeddings.Empty())

				return alignedEmbeddings[0], boxedEmbeddings[0]
			}

			upright, uprightBox := embed(fileName)

			original, _, err := fs.DecodeImageFile(fileName)
			require.NoError(t, err)
			rotated, rotatedBox := embed(writeTestJPEG(t, rotateTestImage(original, 20)))

			t.Run("AlignmentBeatsBoundingBox", func(t *testing.T) {
				// Warping onto the template is what makes the embedding pose invariant, so the
				// same face rotated in plane must stay closer when aligned than when boxed.
				assert.Less(t, upright.Dist(rotated), uprightBox.Dist(rotatedBox))
			})
			t.Run("SameFaceCloserThanOtherFace", func(t *testing.T) {
				other, _ := embed(filepath.Join("testdata", "6.jpg"))
				assert.Less(t, upright.Dist(rotated), upright.Dist(other))
			})
		})
	}
}

func TestEmbedderInputSize(t *testing.T) {
	m := FindEmbeddingModel(ModelSFace)
	registeredWidth, registeredHeight := m.InputSize()

	t.Run("FromGraph", func(t *testing.T) {
		width, height := embedderInputSize(m, &onnx.ModelInfo{Input: &onnx.Input{Width: 96, Height: 128}})
		assert.Equal(t, 96, width)
		assert.Equal(t, 128, height)
	})
	t.Run("DynamicShape", func(t *testing.T) {
		width, height := embedderInputSize(m, &onnx.ModelInfo{Input: &onnx.Input{}})
		assert.Equal(t, registeredWidth, width)
		assert.Equal(t, registeredHeight, height)
	})
	t.Run("NoInput", func(t *testing.T) {
		width, height := embedderInputSize(m, &onnx.ModelInfo{})
		assert.Equal(t, registeredWidth, width)
		assert.Equal(t, registeredHeight, height)
	})
	t.Run("PartiallyDynamic", func(t *testing.T) {
		width, height := embedderInputSize(m, &onnx.ModelInfo{Input: &onnx.Input{Height: 128}})
		assert.Equal(t, registeredWidth, width)
		assert.Equal(t, 128, height)
	})
}

func TestEmbedderOutputDims(t *testing.T) {
	m := FindEmbeddingModel(ModelSFace)

	t.Run("FromGraph", func(t *testing.T) {
		assert.Equal(t, 512, embedderOutputDims(m, &onnx.ModelInfo{Output: &onnx.Output{Width: 512}}))
	})
	t.Run("DynamicShape", func(t *testing.T) {
		assert.Equal(t, m.Dims, embedderOutputDims(m, &onnx.ModelInfo{Output: &onnx.Output{}}))
	})
	t.Run("NoOutput", func(t *testing.T) {
		assert.Equal(t, m.Dims, embedderOutputDims(m, &onnx.ModelInfo{}))
	})
}

func TestONNXEmbedder_Close(t *testing.T) {
	t.Run("Idempotent", func(t *testing.T) {
		embedder := newTestONNXEmbedder(t, ModelSFace)
		require.NoError(t, embedder.Close())
		require.NoError(t, embedder.Close())
	})
	t.Run("RunAfterClose", func(t *testing.T) {
		// Reconfiguring the embedder closes the session that indexing workers still
		// hold, so a late call has to report no result rather than dereference nil.
		embedder := newTestONNXEmbedder(t, ModelSFace)
		require.NoError(t, embedder.Close())
		assert.True(t, embedder.Run(image.NewRGBA(image.Rect(0, 0, 112, 112))).Empty())
	})
}
