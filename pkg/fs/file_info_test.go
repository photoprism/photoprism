package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeInfo(t *testing.T) {
	t.Run("Descriptions", func(t *testing.T) {
		assert.Equal(t, "Kodak Cineon", TypeInfo[ImageCineon])
		assert.Equal(t, "Other", TypeInfo[TypeUnknown])
	})
	t.Run("Complete", func(t *testing.T) {
		// Every type a filename can resolve to needs a description, since the generated
		// file format report is the reference the user-facing documentation is built from.
		for ext, fileType := range Extensions {
			assert.NotEmptyf(t, TypeInfo[fileType], "%s (%s) has no description", fileType, ext)
		}
	})
}
