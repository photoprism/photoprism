package classify

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/fs"
)

var assetsPath = fs.Abs("../../../assets")
var examplesPath = filepath.Join(assetsPath, "examples")
var modelsPath = filepath.Join(assetsPath, "models")
var modelPath = modelsPath + "/nasnet"
var once sync.Once
var testInstance *Model

func NewModelTest(t *testing.T) *Model {
	once.Do(func() {
		testInstance = NewNasnet(modelsPath, false)
		if err := testInstance.loadModel(); err != nil {
			t.Fatal(err)
		}
	})

	return testInstance
}

func TestModel_CenterCrop(t *testing.T) {
	model := NewNasnet(modelsPath, false)
	if err := model.loadModel(); err != nil {
		t.Fatal(err)
	}

	model.meta.Input.ResizeOperation = tensorflow.CenterCrop

	t.Run("NasnetPadding", func(t *testing.T) {
		runBasicLabelsTest(t, model, 6)
	})
}

func TestModel_Padding(t *testing.T) {
	model := NewNasnet(modelsPath, false)
	if err := model.loadModel(); err != nil {
		t.Fatal(err)
	}

	model.meta.Input.ResizeOperation = tensorflow.Padding

	t.Run("NasnetPadding", func(t *testing.T) {
		runBasicLabelsTest(t, model, 6)
	})
}

func TestModel_ResizeBreakAspectRatio(t *testing.T) {
	model := NewNasnet(modelsPath, false)
	if err := model.loadModel(); err != nil {
		t.Fatal(err)
	}

	model.meta.Input.ResizeOperation = tensorflow.ResizeBreakAspectRatio

	t.Run("NasnetBreakAspectRatio", func(t *testing.T) {
		runBasicLabelsTest(t, model, 4)
	})
}

func runBasicLabelsTest(t *testing.T, model *Model, expectedUncertainty int) {
	result, err := model.File(examplesPath+"/zebra_green_brown.jpg", 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.IsType(t, Labels{}, result)
	assert.Equal(t, 1, len(result))

	if len(result) > 0 {
		assert.Equal(t, "zebra", result[0].Name)

		assert.Equal(t, expectedUncertainty, result[0].Uncertainty)
	}
}

func TestModel_LabelsFromFile(t *testing.T) {
	t.Run("ChameleonLimeJpg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)
		result, err := tensorFlow.File(examplesPath+"/chameleon_lime.jpg", 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 1, len(result))

		if len(result) > 0 {
			t.Logf("result: %#v", result[0])
			assert.Equal(t, "chameleon", result[0].Name)

			assert.Equal(t, 7, result[0].Uncertainty)
		}
	})
	t.Run("CatNum224Jpeg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)
		result, err := tensorFlow.File(examplesPath+"/cat_224.jpeg", 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 1, len(result))

		if len(result) > 0 {
			assert.Equal(t, "cat", result[0].Name)

			assert.Equal(t, 59, result[0].Uncertainty)
		}
	})
	t.Run("CatNum720Jpeg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)
		result, err := tensorFlow.File(examplesPath+"/cat_720.jpeg", 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 3, len(result))

		// t.Logf("labels: %#v", result)

		if len(result) > 0 {
			assert.Equal(t, "cat", result[0].Name)
			assert.Equal(t, 60, result[0].Uncertainty)
		}
	})
	t.Run("GreenJpg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)
		result, err := tensorFlow.File(examplesPath+"/green.jpg", 10)

		t.Logf("labels: %#v", result)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 1, len(result))

		if len(result) > 0 {
			assert.Equal(t, "outdoor", result[0].Name)

			assert.Equal(t, 70, result[0].Uncertainty)
		}
	})
	t.Run("NotExistingFile", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		result, err := tensorFlow.File(examplesPath+"/notexisting.jpg", 10)
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Empty(t, result)
	})
	t.Run("Disabled", func(t *testing.T) {
		tensorFlow := NewNasnet(modelsPath, true)

		result, err := tensorFlow.File(examplesPath+"/chameleon_lime.jpg", 10)
		assert.Nil(t, err)

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 0, len(result))

		t.Log(result)
	})
}

func TestModel_Run(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("ChameleonLimeJpg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		if imageBuffer, err := os.ReadFile(filepath.Join(examplesPath, "chameleon_lime.jpg")); err != nil { //nolint:gosec // reading bundled test fixture
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)

			t.Log(result)

			assert.NotNil(t, result)

			if err != nil {
				t.Fatal(err)
			}

			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 1, len(result))

			if len(result) > 0 {
				assert.Equal(t, "chameleon", result[0].Name)
				assert.Equal(t, 100-93, result[0].Uncertainty)
			}
		}
	})
	t.Run("DogOrangeJpg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		if imageBuffer, err := os.ReadFile(filepath.Join(examplesPath, "dog_orange.jpg")); err != nil { //nolint:gosec // reading bundled test fixture
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)

			t.Log(result)

			assert.NotNil(t, result)

			if err != nil {
				t.Fatal(err)
			}

			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 1, len(result))

			if len(result) > 0 {
				assert.Equal(t, "dog", result[0].Name)
				assert.Equal(t, 34, result[0].Uncertainty)
			}
		}
	})
	t.Run("RandomDocx", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		if imageBuffer, err := os.ReadFile(filepath.Join(examplesPath, "Random.docx")); err != nil { //nolint:gosec // reading bundled test fixture
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)
			assert.Empty(t, result)
			assert.Error(t, err)
		}
	})
	t.Run("Num6720PxWhiteJpg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		if imageBuffer, err := os.ReadFile(filepath.Join(examplesPath, "6720px_white.jpg")); err != nil { //nolint:gosec // reading bundled test fixture
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)

			if err != nil {
				t.Fatal(err)
			}

			assert.Empty(t, result)
		}
	})
	t.Run("Disabled", func(t *testing.T) {
		tensorFlow := NewNasnet(modelsPath, true)

		if imageBuffer, err := os.ReadFile(filepath.Join(examplesPath, "dog_orange.jpg")); err != nil { //nolint:gosec // reading bundled test fixture
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)

			t.Log(result)

			assert.Nil(t, result)

			assert.Nil(t, err)
			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 0, len(result))
		}
	})
}

func TestModel_LoadModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tf := NewModelTest(t)
		assert.True(t, tf.ModelLoaded())
	})
	t.Run("NotFound", func(t *testing.T) {
		tensorFlow := NewNasnet(modelsPath+"foo", false)
		err := tensorFlow.loadModel()

		if err != nil {
			assert.Contains(t, err.Error(), "not find SavedModel")
		}

		assert.Error(t, err)
	})
}

func TestModel_BestLabels(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tensorFlow := NewNasnet(modelsPath, false)

		if err := tensorFlow.loadLabels(modelPath); err != nil {
			t.Fatal(err)
		}

		p := make([]float32, 1000)

		p[8] = 0.7
		p[1] = 0.5

		result := tensorFlow.bestLabels(p, 10)
		assert.Equal(t, "chicken", result[0].Name)
		assert.Equal(t, "bird", result[0].Categories[0])
		assert.Equal(t, "image", result[0].Source)
		t.Log(result)
	})
	t.Run("NotLoaded", func(t *testing.T) {
		tensorFlow := NewNasnet(modelsPath, false)

		p := make([]float32, 1000)

		p[666] = 0.5

		result := tensorFlow.bestLabels(p, 10)
		assert.Empty(t, result)
	})
}

func BenchmarkModel_BestLabelWithOptimization(b *testing.B) {
	model := NewNasnet(assetsPath, false)
	err := model.loadModel()
	if err != nil {
		b.Fatal(err)
	}

	imageBuffer, err := os.ReadFile(filepath.Join(examplesPath, "dog_orange.jpg")) //nolint:gosec // reading bundled test fixture
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_, err := model.Run(imageBuffer, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkModel_BestLabelsNoOptimization(b *testing.B) {
	model := NewNasnet(assetsPath, false)
	err := model.loadModel()
	if err != nil {
		b.Fatal(err)
	}
	model.builder = nil

	imageBuffer, err := os.ReadFile(filepath.Join(examplesPath, "dog_orange.jpg")) //nolint:gosec // reading bundled test fixture
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_, err := model.Run(imageBuffer, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}
