package classify

import (
	"os"
	"sync"
	"testing"

	"github.com/photoprism/photoprism/pkg/fs"
	tensorflow "github.com/tensorflow/tensorflow/tensorflow/go"

	"github.com/stretchr/testify/assert"
)

var assetsPath = fs.Abs("../../assets")
var modelPath = assetsPath + "/nasnet"
var examplesPath = assetsPath + "/examples"
var once sync.Once
var testInstance *TensorFlow

// NewTest returns a new TensorFlow test instance.
func NewTest(t *testing.T) *TensorFlow {
	once.Do(func() {
		testInstance = New(assetsPath, false)
		if err := testInstance.loadModel(); err != nil {
			t.Fatal(err)
		}
	})

	return testInstance
}

func TestTensorFlow_LabelsFromFile(t *testing.T) {
	t.Run("chameleon_lime.jpg", func(t *testing.T) {
		tensorFlow := NewTest(t)

		result, err := tensorFlow.File(examplesPath + "/chameleon_lime.jpg")

		assert.Nil(t, err)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 1, len(result))

		t.Log(result)

		assert.Equal(t, "chameleon", result[0].Name)

		assert.Equal(t, 7, result[0].Uncertainty)
	})
	t.Run("not existing file", func(t *testing.T) {
		tensorFlow := NewTest(t)

		result, err := tensorFlow.File(examplesPath + "/notexisting.jpg")
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Empty(t, result)
	})
	t.Run("disabled true", func(t *testing.T) {
		tensorFlow := New(assetsPath, true)

		result, err := tensorFlow.File(examplesPath + "/chameleon_lime.jpg")
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

func TestTensorFlow_Labels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("chameleon_lime.jpg", func(t *testing.T) {
		tensorFlow := NewTest(t)

		if imageBuffer, err := os.ReadFile(examplesPath + "/chameleon_lime.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer)

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
		tensorFlow := NewTest(t)

		if imageBuffer, err := os.ReadFile(examplesPath + "/dog_orange.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer)

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
		tensorFlow := NewTest(t)

		if imageBuffer, err := os.ReadFile(examplesPath + "/Random.docx"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer)
			assert.Empty(t, result)
			assert.Error(t, err)
		}
	})
	t.Run("6720px_white.jpg", func(t *testing.T) {
		tensorFlow := NewTest(t)

		if imageBuffer, err := os.ReadFile(examplesPath + "/6720px_white.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer)

			if err != nil {
				t.Fatal(err)
			}

			assert.Empty(t, result)
		}
	})
	t.Run("disabled true", func(t *testing.T) {
		tensorFlow := New(assetsPath, true)

		if imageBuffer, err := os.ReadFile(examplesPath + "/dog_orange.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer)

			t.Log(result)

			assert.Nil(t, result)

			assert.Nil(t, err)
			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 0, len(result))
		}
	})
}

func TestTensorFlow_LoadModel(t *testing.T) {
	t.Run("model loaded", func(t *testing.T) {
		tf := NewTest(t)
		assert.True(t, tf.ModelLoaded())
	})
	t.Run("model path does not exist", func(t *testing.T) {
		tensorFlow := New(assetsPath+"foo", false)
		if err := tensorFlow.loadModel(); err != nil {
			assert.Contains(t, err.Error(), "Could not find SavedModel")
		} else {
			t.Fatal("err should NOT be nil")
		}
	})
}

func TestTensorFlow_BestLabels(t *testing.T) {
	t.Run("labels not loaded", func(t *testing.T) {
		tensorFlow := New(assetsPath, false)

		p := make([]float32, 1000)

		p[666] = 0.5

		result := tensorFlow.bestLabels(p)
		assert.Empty(t, result)
	})
	t.Run("labels loaded", func(t *testing.T) {
		tensorFlow := New(assetsPath, false)

		if err := tensorFlow.loadLabels(modelPath); err != nil {
			t.Fatal(err)
		}

		p := make([]float32, 1000)

		p[8] = 0.7
		p[1] = 0.5

		result := tensorFlow.bestLabels(p)
		assert.Equal(t, "chicken", result[0].Name)
		assert.Equal(t, "bird", result[0].Categories[0])
		assert.Equal(t, "image", result[0].Source)
		t.Log(result)
	})
}

func TestTensorFlow_MakeTensor(t *testing.T) {
	t.Run("cat_brown.jpg", func(t *testing.T) {
		tensorFlow := NewTest(t)

		imageBuffer, err := os.ReadFile(examplesPath + "/cat_brown.jpg")

		if err != nil {
			t.Fatal(err)
		}

		result, err := tensorFlow.createTensor(imageBuffer, "jpeg")
		assert.Equal(t, tensorflow.DataType(0x1), result.DataType())
		assert.Equal(t, int64(1), result.Shape()[0])
		assert.Equal(t, int64(224), result.Shape()[2])
	})
	t.Run("Random.docx", func(t *testing.T) {
		tensorFlow := NewTest(t)

		imageBuffer, err := os.ReadFile(examplesPath + "/Random.docx")
		assert.Nil(t, err)
		result, err := tensorFlow.createTensor(imageBuffer, "jpeg")

		assert.Empty(t, result)
		assert.EqualError(t, err, "image: unknown format")
	})
}

func Test_convertValue(t *testing.T) {
	result := convertValue(uint32(98765432))
	assert.Equal(t, float32(3024.898), result)
}
