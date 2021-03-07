package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBirthTime(t *testing.T) {
	t.Run("time now", func(t *testing.T) {
		result := BirthTime("/testdata/Test.jpg")
		assert.NotEmpty(t, result)
	})
	t.Run("mod time", func(t *testing.T) {
		result := BirthTime("./testdata/CATYELLOW.jpg")
		assert.NotEmpty(t, result)
	})
}
