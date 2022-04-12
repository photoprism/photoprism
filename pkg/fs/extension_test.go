package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripExt(t *testing.T) {
	t.Run("Test.jpg", func(t *testing.T) {
		result := StripExt("/testdata/Test.jpg")
		assert.Equal(t, "/testdata/Test", result)
	})

	t.Run("Test.jpg.json", func(t *testing.T) {
		result := StripExt("/testdata/Test.jpg.json")
		assert.Equal(t, "/testdata/Test.jpg", result)
	})

	t.Run("Test copy 3.foo", func(t *testing.T) {
		result := StripExt("/testdata/Test copy 3.foo")
		assert.Equal(t, "/testdata/Test copy 3", result)
	})
}

func TestStripKnownExt(t *testing.T) {
	t.Run("Test.jpg", func(t *testing.T) {
		result := StripKnownExt("/testdata/Test.jpg")
		assert.Equal(t, "/testdata/Test", result)
	})

	t.Run("Test.jpg.json", func(t *testing.T) {
		result := StripKnownExt("/testdata/Test.jpg.json")
		assert.Equal(t, "/testdata/Test", result)
	})

	t.Run("Test copy 3.foo", func(t *testing.T) {
		result := StripKnownExt("/testdata/Test copy 3.foo")
		assert.Equal(t, "/testdata/Test copy 3.foo", result)
	})

	t.Run("my/file.jpg.json.xmp", func(t *testing.T) {
		result := StripKnownExt("my/file.jpg.json.xmp")
		assert.Equal(t, "my/file", result)
	})

	t.Run("my/jpg/avi.foo.bar.baz", func(t *testing.T) {
		result := StripKnownExt("my/jpg/avi.foo.bar.baz")
		assert.Equal(t, "my/jpg/avi.foo.bar.baz", result)
	})
}

func TestExt(t *testing.T) {
	t.Run("Test.jpg", func(t *testing.T) {
		result := Ext("/testdata/Test.jpg")
		assert.Equal(t, ".jpg", result)
	})

	t.Run("Test.jpg.json", func(t *testing.T) {
		result := Ext("/testdata/Test.jpg.json")
		assert.Equal(t, ".jpg.json", result)
	})

	t.Run("Test copy 3.foo", func(t *testing.T) {
		result := Ext("/testdata/Test copy 3.foo")
		assert.Equal(t, ".foo", result)
	})

	t.Run("Test", func(t *testing.T) {
		result := Ext("/testdata/Test")
		assert.Equal(t, "", result)
	})
}
