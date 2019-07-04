package photoprism

import (
	"io/ioutil"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestTensorFlow_LabelsFromFile(t *testing.T) {
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
}

func TestTensorFlow_Labels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

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
}

func TestTensorFlow_Labels_Dog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

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
}
