package download

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestRegister(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fileName := fs.Abs("./testdata/image.jpg")
		uniqueId, err := Register(fileName)
		assert.NoError(t, err)
		assert.True(t, rnd.IsUUID(uniqueId))

		findName, findErr := Find(uniqueId)

		assert.NoError(t, findErr)
		assert.Equal(t, fileName, findName)

		Flush()

		findName, findErr = Find(uniqueId)

		assert.Error(t, findErr)
		assert.Equal(t, "", findName)
	})
	t.Run("NotFound", func(t *testing.T) {
		fileName := fs.Abs("./testdata/invalid.jpg")
		uniqueId, err := Register(fileName)
		assert.Error(t, err)
		assert.Equal(t, "", uniqueId)
	})
}
