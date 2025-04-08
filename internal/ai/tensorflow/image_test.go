package tensorflow

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConvertValue(t *testing.T) {
	result := convertValue(uint32(98765432), 127.5)
	assert.Equal(t, float32(3024.898), result)
}

func TestImageFromBytes(t *testing.T) {
	var assetsPath = fs.Abs("../../../assets")
	var examplesPath = assetsPath + "/examples"

	t.Run("CatJpeg", func(t *testing.T) {
		imageBuffer, err := os.ReadFile(examplesPath + "/cat_brown.jpg")

		if err != nil {
			t.Fatal(err)
		}

		result, err := ImageFromBytes(imageBuffer, 224)
		assert.Equal(t, tensorflow.DataType(0x1), result.DataType())
		assert.Equal(t, int64(1), result.Shape()[0])
		assert.Equal(t, int64(224), result.Shape()[2])
	})
	t.Run("Document", func(t *testing.T) {
		imageBuffer, err := os.ReadFile(examplesPath + "/Random.docx")
		assert.Nil(t, err)
		result, err := ImageFromBytes(imageBuffer, 224)

		assert.Empty(t, result)
		assert.EqualError(t, err, "image: unknown format")
	})
}
