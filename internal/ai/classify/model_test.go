package classify

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

var assetsPath = fs.Abs("../../../assets")
var modelPath = assetsPath + "/nasnet"
var examplesPath = assetsPath + "/examples"
var once sync.Once
var testInstance *Model

func NewModelTest(t *testing.T) *Model {
	once.Do(func() {
		testInstance = NewNasnet(assetsPath, false)
		if err := testInstance.loadModel(); err != nil {
			t.Fatal(err)
		}
	})

	return testInstance
}

func TestModel_LabelsFromFile(t *testing.T) {
	t.Run("chameleon_lime.jpg", func(t *testing.T) {
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
	t.Run("cat_224.jpeg", func(t *testing.T) {
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
	t.Run("cat_720.jpeg", func(t *testing.T) {
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
	t.Run("green.jpg", func(t *testing.T) {
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
	t.Run("not existing file", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		result, err := tensorFlow.File(examplesPath+"/notexisting.jpg", 10)
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Empty(t, result)
	})
	t.Run("disabled true", func(t *testing.T) {
		tensorFlow := NewNasnet(assetsPath, true)

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

func TestModel_Labels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("chameleon_lime.jpg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		if imageBuffer, err := os.ReadFile(examplesPath + "/chameleon_lime.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer, 10)

			t.Log(result)

			assert.NotNil(t, result)

			if err != nil {
				t.Fatal(err)
			}

			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 1, len(result))

			assert.Equal(t, "chameleon", result[0].Name)

			assert.Equal(t, 100-93, result[0].Uncertainty)
		}
	})
	t.Run("dog_orange.jpg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		if imageBuffer, err := os.ReadFile(examplesPath + "/dog_orange.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer, 10)

			t.Log(result)

			assert.NotNil(t, result)

			if err != nil {
				t.Fatal(err)
			}

			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 1, len(result))

			assert.Equal(t, "dog", result[0].Name)

			assert.Equal(t, 34, result[0].Uncertainty)
		}
	})
	t.Run("Random.docx", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		if imageBuffer, err := os.ReadFile(examplesPath + "/Random.docx"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer, 10)
			assert.Empty(t, result)
			assert.Error(t, err)
		}
	})
	t.Run("6720px_white.jpg", func(t *testing.T) {
		tensorFlow := NewModelTest(t)

		if imageBuffer, err := os.ReadFile(examplesPath + "/6720px_white.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer, 10)

			if err != nil {
				t.Fatal(err)
			}

			assert.Empty(t, result)
		}
	})
	t.Run("disabled true", func(t *testing.T) {
		tensorFlow := NewNasnet(assetsPath, true)

		if imageBuffer, err := os.ReadFile(examplesPath + "/dog_orange.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer, 10)

			t.Log(result)

			assert.Nil(t, result)

			assert.Nil(t, err)
			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 0, len(result))
		}
	})
}

func TestModel_LoadModel(t *testing.T) {
	t.Run("model loaded", func(t *testing.T) {
		tf := NewModelTest(t)
		assert.True(t, tf.ModelLoaded())
	})
	t.Run("model path does not exist", func(t *testing.T) {
		tensorFlow := NewNasnet(assetsPath+"foo", false)
		err := tensorFlow.loadModel()

		if err != nil {
			assert.Contains(t, err.Error(), "no such file or directory")
		}

		assert.Error(t, err)
	})
}

func TestModel_BestLabels(t *testing.T) {
	t.Run("labels not loaded", func(t *testing.T) {
		tensorFlow := NewNasnet(assetsPath, false)

		p := make([]float32, 1000)

		p[666] = 0.5

		result := tensorFlow.bestLabels(p, 10)
		assert.Empty(t, result)
	})
	t.Run("labels loaded", func(t *testing.T) {
		tensorFlow := NewNasnet(assetsPath, false)

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
}
