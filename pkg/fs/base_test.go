package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	t.Run("Test.jpg", func(t *testing.T) {
		result := Base("/testdata/Test.jpg", true)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test.3453453.jpg", func(t *testing.T) {
		result := Base("/testdata/Test.3453453.jpg", true)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test copy 3.jpg", func(t *testing.T) {
		result := Base("/testdata/Test copy 3.jpg", true)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test (3).jpg", func(t *testing.T) {
		result := Base("/testdata/Test (3).jpg", true)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test.jpg", func(t *testing.T) {
		result := Base("/testdata/Test.jpg", false)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test.3453453.jpg", func(t *testing.T) {
		result := Base("/testdata/Test.3453453.jpg", false)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test copy 3.jpg", func(t *testing.T) {
		result := Base("/testdata/Test copy 3.jpg", false)
		assert.Equal(t, "Test copy 3", result)
	})

	t.Run("Test (3).jpg", func(t *testing.T) {
		result := Base("/testdata/Test (3).jpg", false)
		assert.Equal(t, "Test (3)", result)
	})
}

func TestBaseAbs(t *testing.T) {
	t.Run("Test copy 3.jpg", func(t *testing.T) {
		result := AbsBase("/testdata/Test (4).jpg", true)

		assert.Equal(t, "/testdata/Test", result)
	})

	t.Run("Test (3).jpg", func(t *testing.T) {
		result := AbsBase("/testdata/Test (4).jpg", false)

		assert.Equal(t, "/testdata/Test (4)", result)
	})


}
