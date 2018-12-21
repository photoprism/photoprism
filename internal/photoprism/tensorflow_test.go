package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestTensorFlow_GetImageTagsFromFile(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf.TensorFlowModelPath())

	result, err := tensorFlow.GetImageTagsFromFile(conf.ImportPath() + "/iphone/IMG_6788.JPG")

	assert.NotNil(t, result)
	assert.Nil(t, err)
	assert.IsType(t, []TensorFlowLabel{}, result)
	assert.Equal(t, 5, len(result))

	assert.Equal(t, "tabby", result[0].Label)
	assert.Equal(t, "tiger cat", result[1].Label)

	assert.Equal(t, float32(0.1648176), result[1].Probability)
}
