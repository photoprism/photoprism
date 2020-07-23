package meta

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestData_CodecAvc1(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		data := Data{
			Codec: "avc1",
		}

		assert.Equal(t, true, data.CodecAvc1())
	})

	t.Run("false", func(t *testing.T) {
		data := Data{
			Codec: "heic",
		}

		assert.Equal(t, false, data.CodecAvc1())
	})
}
