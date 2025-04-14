package vision

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestFace(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		img, imgErr := os.ReadFile(fs.Abs("./testdata/face_160x160.jpg"))

		if imgErr != nil {
			t.Fatal(imgErr)
		}

		result, err := Face(img)

		assert.NoError(t, err)
		assert.IsType(t, face.Embeddings{}, result)
		assert.Equal(t, 1, len(result))

		// t.Log(result)
	})
	t.Run("InvalidFile", func(t *testing.T) {
		result, err := Face([]byte{})

		assert.Error(t, err)
		assert.IsType(t, face.Embeddings{}, result)
		assert.Equal(t, 0, len(result))

		// t.Log(result)
	})
}
