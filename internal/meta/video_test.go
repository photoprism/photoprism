package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestData_CodecAvc(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		data := Data{
			Codec: "avc1",
		}

		assert.Equal(t, true, data.CodecAvc())
	})
	t.Run("false", func(t *testing.T) {
		data := Data{
			Codec: "heic",
		}

		assert.Equal(t, false, data.CodecAvc())
	})
}
