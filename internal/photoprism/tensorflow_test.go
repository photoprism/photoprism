package photoprism

import (
	"io/ioutil"
	"testing"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/stretchr/testify/assert"
)

func TestTensorFlow_GetImageTagsFromFile(t *testing.T) {
	ctx := context.TestContext()

	ctx.InitializeTestData(t)

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	result, err := tensorFlow.GetImageTagsFromFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG")

	assert.Nil(t, err)

	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	assert.NotNil(t, result)
	assert.IsType(t, []TensorFlowLabel{}, result)
	assert.Equal(t, 2, len(result))

	t.Log(result)

	assert.Equal(t, "tabby cat", result[0].Label)
	assert.Equal(t, "tiger cat", result[1].Label)

	assert.Equal(t, 68, result[0].Percent())
	assert.Equal(t, 13, result[1].Percent())
}

func TestTensorFlow_GetImageTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.TestContext()

	ctx.InitializeTestData(t)

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	if imageBuffer, err := ioutil.ReadFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG"); err != nil {
		t.Error(err)
	} else {
		result, err := tensorFlow.GetImageTags(imageBuffer)

		t.Log(result)

		assert.NotNil(t, result)

		assert.Nil(t, err)
		assert.IsType(t, []TensorFlowLabel{}, result)
		assert.Equal(t, 2, len(result))

		assert.Equal(t, "tabby cat", result[0].Label)
		assert.Equal(t, "tiger cat", result[1].Label)

		assert.Equal(t, 68, result[0].Percent())
		assert.Equal(t, 13, result[1].Percent())
	}
}

func TestTensorFlow_GetImageTags_Dog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.TestContext()

	ctx.InitializeTestData(t)

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	if imageBuffer, err := ioutil.ReadFile(ctx.ImportPath() + "/dog.jpg"); err != nil {
		t.Error(err)
	} else {
		result, err := tensorFlow.GetImageTags(imageBuffer)

		t.Log(result)

		assert.NotNil(t, result)

		assert.Nil(t, err)
		assert.IsType(t, []TensorFlowLabel{}, result)
		assert.Equal(t, 3, len(result))

		assert.Equal(t, "belt", result[0].Label)
		assert.Equal(t, "beagle dog", result[1].Label)

		assert.Equal(t, 10, result[0].Percent())
		assert.Equal(t, 9, result[1].Percent())
	}
}
