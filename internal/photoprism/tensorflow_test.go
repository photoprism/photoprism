package photoprism

import (
	"io/ioutil"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestTensorFlow_LabelsFromFile(t *testing.T) {
	ctx := config.TestConfig()

	ctx.InitializeTestData(t)

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	result, err := tensorFlow.LabelsFromFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG")

	assert.Nil(t, err)

	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	assert.NotNil(t, result)
	assert.IsType(t, Labels{}, result)
	assert.Equal(t, 2, len(result))

	t.Log(result)

	assert.Equal(t, "tabby cat", result[0].Name)
	assert.Equal(t, "tiger cat", result[1].Name)

	assert.Equal(t, 32, result[0].Uncertainty)
	assert.Equal(t, 86, result[1].Uncertainty)
}

func TestTensorFlow_Labels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := config.TestConfig()

	ctx.InitializeTestData(t)

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	if imageBuffer, err := ioutil.ReadFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG"); err != nil {
		t.Error(err)
	} else {
		result, err := tensorFlow.Labels(imageBuffer)

		t.Log(result)

		assert.NotNil(t, result)

		assert.Nil(t, err)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 2, len(result))

		assert.Equal(t, "tabby cat", result[0].Name)
		assert.Equal(t, "tiger cat", result[1].Name)

		assert.Equal(t, 100 - 68, result[0].Uncertainty)
		assert.Equal(t, 100 - 14, result[1].Uncertainty)
	}
}

func TestTensorFlow_Labels_Dog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := config.TestConfig()

	ctx.InitializeTestData(t)

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	if imageBuffer, err := ioutil.ReadFile(ctx.ImportPath() + "/dog.jpg"); err != nil {
		t.Error(err)
	} else {
		result, err := tensorFlow.Labels(imageBuffer)

		t.Log(result)

		assert.NotNil(t, result)

		assert.Nil(t, err)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 3, len(result))

		assert.Equal(t, "beagle dog", result[0].Name)
		assert.Equal(t, "basset dog", result[1].Name)

		assert.Equal(t, 91, result[0].Uncertainty)
		assert.Equal(t, 92, result[1].Uncertainty)
	}
}
