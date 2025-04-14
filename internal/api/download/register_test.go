package download

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestRegister(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Generate a random token for the remote service to download the file.
		fileUuid := rnd.UUID()
		fileName := fs.Abs("./testdata/image.jpg")
		err := Register(fileUuid, fileName)
		assert.NoError(t, err)
		assert.True(t, rnd.IsUUID(fileUuid))

		findName, findErr := Find(fileUuid)

		assert.NoError(t, findErr)
		assert.Equal(t, fileName, findName)

		Flush()

		findName, findErr = Find(fileUuid)

		assert.Error(t, findErr)
		assert.Equal(t, "", findName)
	})
	t.Run("NotFound", func(t *testing.T) {
		// Generate a random token for the remote service to download the file.
		fileUuid := rnd.UUID()
		fileName := fs.Abs("./testdata/invalid.jpg")
		err := Register(fileUuid, fileName)
		assert.Error(t, err)
		assert.True(t, rnd.IsUUID(fileUuid))

		findName, findErr := Find(fileUuid)

		assert.Error(t, findErr)
		assert.Equal(t, "", findName)
	})
}
