package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileFormats_Markdown(t *testing.T) {
	t.Run("Render", func(t *testing.T) {
		f := Extensions.Formats(true)
		result := f.Markdown()

		// fmt.Print(result)
		assert.NotEmpty(t, result)
	})
}
