package photoprism

import (
	tensorflow "github.com/tensorflow/tensorflow/tensorflow/go"
	"io/ioutil"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestTensorFlow_LoadLabelRules(t *testing.T) {
	t.Run("labels.txt exists", func(t *testing.T) {
		conf := config.NewTestConfig()

		tensorFlow := NewTensorFlow(conf)

		result := tensorFlow.loadLabelRules()
		assert.Nil(t, result)
	})
	t.Run("labels.txt not existing in config path", func(t *testing.T) {
		conf := config.NewTestErrorConfig()

		tensorFlow := NewTensorFlow(conf)

		result := tensorFlow.loadLabelRules()
		assert.Contains(t, result.Error(), "label rules file not found")
	})
}

func TestTensorFlow_LabelsFromFile(t *testing.T) {
	t.Run("/chameleon_lime.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		tensorFlow := NewTensorFlow(conf)

		result, err := tensorFlow.LabelsFromFile(conf.ExamplesPath() + "/chameleon_lime.jpg")

		assert.Nil(t, err)

		if err != nil {
			t.Log(err.Error())
			t.Fail()
		}

		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 1, len(result))

		t.Log(result)

		assert.Equal(t, "chameleon", result[0].Name)

		assert.Equal(t, 7, result[0].Uncertainty)
	})
	t.Run("not existing file", func(t *testing.T) {
		conf := config.TestConfig()

		tensorFlow := NewTensorFlow(conf)

		result, err := tensorFlow.LabelsFromFile(conf.ExamplesPath() + "/notexisting.jpg")
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Empty(t, result)
	})
}

func TestTensorFlow_Labels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("/chameleon_lime.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		tensorFlow := NewTensorFlow(conf)

		if imageBuffer, err := ioutil.ReadFile(conf.ExamplesPath() + "/chameleon_lime.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer)

			t.Log(result)

			assert.NotNil(t, result)

			assert.Nil(t, err)
			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 1, len(result))

			assert.Equal(t, "chameleon", result[0].Name)

			assert.Equal(t, 100-93, result[0].Uncertainty)
		}
	})
	t.Run("/dog_orange.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		tensorFlow := NewTensorFlow(conf)

		if imageBuffer, err := ioutil.ReadFile(conf.ExamplesPath() + "/dog_orange.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer)

			t.Log(result)

			assert.NotNil(t, result)

			assert.Nil(t, err)
			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 2, len(result))

			assert.Equal(t, "chihuahua dog", result[0].Name)
			assert.Equal(t, "pembroke dog", result[1].Name)

			assert.Equal(t, 34, result[0].Uncertainty)
			assert.Equal(t, 91, result[1].Uncertainty)
		}
	})
	t.Run("/Random.docx", func(t *testing.T) {
		conf := config.TestConfig()

		tensorFlow := NewTensorFlow(conf)

		if imageBuffer, err := ioutil.ReadFile(conf.ExamplesPath() + "/Random.docx"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer)
			assert.Empty(t, result)
			assert.Contains(t, err.Error(), "invalid image")
		}
	})
	t.Run("/6720px_white.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		tensorFlow := NewTensorFlow(conf)

		if imageBuffer, err := ioutil.ReadFile(conf.ExamplesPath() + "/6720px_white.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Labels(imageBuffer)
			assert.Empty(t, result)
			assert.Nil(t, err)
		}
	})
}

func TestTensorFlow_LoadLabels(t *testing.T) {
	t.Run("labels.txt exists", func(t *testing.T) {
		conf := config.NewTestConfig()

		tensorFlow := NewTensorFlow(conf)
		path := conf.TensorFlowModelPath()

		result := tensorFlow.loadLabels(path)
		assert.Nil(t, result)
	})
	t.Run("label.txt does not exist", func(t *testing.T) {
		conf := config.NewTestErrorConfig()

		tensorFlow := NewTensorFlow(conf)
		path := conf.TensorFlowModelPath()

		result := tensorFlow.loadLabels(path)
		assert.Contains(t, result.Error(), "no such file or directory")
	})
}

func TestTensorFlow_LoadModel(t *testing.T) {
	t.Run("model path exists", func(t *testing.T) {
		conf := config.NewTestConfig()

		tensorFlow := NewTensorFlow(conf)

		result := tensorFlow.loadModel()
		assert.Nil(t, result)
	})
	t.Run("model path does not exist", func(t *testing.T) {
		conf := config.NewTestErrorConfig()

		tensorFlow := NewTensorFlow(conf)

		result := tensorFlow.loadModel()
		assert.Contains(t, result.Error(), "Could not find SavedModel")
	})
}

func TestTensorFlow_LabelRule(t *testing.T) {
	t.Run("label.txt exists", func(t *testing.T) {
		conf := config.NewTestConfig()

		tensorFlow := NewTensorFlow(conf)

		result := tensorFlow.labelRule("cat")
		assert.Equal(t, "tabby cat", result.Label)
		assert.Equal(t, "kitty", result.Categories[2])
		assert.Equal(t, 5, result.Priority)
	})
	t.Run("labels.txt not existing in config path", func(t *testing.T) {
		conf := config.NewTestErrorConfig()

		tensorFlow := NewTensorFlow(conf)

		result := tensorFlow.labelRule("cat")
		assert.Empty(t, result.Categories)
		assert.Equal(t, 0, result.Priority)
	})
}

func TestTensorFlow_BestLabels(t *testing.T) {
	t.Run("labels not loaded", func(t *testing.T) {
		conf := config.NewTestConfig()

		tensorFlow := NewTensorFlow(conf)

		p := make([]float32, 1000)

		p[666] = 0.5

		result := tensorFlow.bestLabels(p)
		assert.Empty(t, result)
	})
	t.Run("labels loaded", func(t *testing.T) {
		conf := config.NewTestConfig()
		path := conf.TensorFlowModelPath()
		tensorFlow := NewTensorFlow(conf)
		tensorFlow.loadLabels(path)

		p := make([]float32, 1000)

		p[8] = 0.7
		p[1] = 0.5

		result := tensorFlow.bestLabels(p)
		assert.Equal(t, "hen", result[0].Name)
		assert.Equal(t, "bird", result[0].Categories[0])
		assert.Equal(t, "image", result[0].Source)
		assert.Equal(t, "goldfish", result[1].Name)
		assert.Equal(t, "animal", result[1].Categories[2])
		assert.Equal(t, "image", result[1].Source)
		t.Log(result)
	})
}

func TestTensorFlow_MakeTensor(t *testing.T) {
	t.Run("/cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		tensorFlow := NewTensorFlow(conf)

		imageBuffer, err := ioutil.ReadFile(conf.ExamplesPath() + "/cat_brown.jpg")
		assert.Nil(t, err)
		result, err := tensorFlow.makeTensor(imageBuffer, "jpeg")
		assert.Equal(t, tensorflow.DataType(0x1), result.DataType())
		assert.Equal(t, int64(1), result.Shape()[0])
		assert.Equal(t, int64(224), result.Shape()[2])
	})
	t.Run("/Random.docx", func(t *testing.T) {
		conf := config.TestConfig()

		tensorFlow := NewTensorFlow(conf)

		imageBuffer, err := ioutil.ReadFile(conf.ExamplesPath() + "/Random.docx")
		assert.Nil(t, err)
		result, err := tensorFlow.makeTensor(imageBuffer, "jpeg")
		assert.Empty(t, result)
		assert.Equal(t, "image: unknown format", err.Error())
	})
}

func Test_ConvertTF(t *testing.T) {
	result := convertTF(uint32(98765432))
	assert.Equal(t, float32(3024.898), result)
}
