package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeExt(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		result := NormalizedExt("testdata/test")
		assert.Equal(t, "", result)
	})

	t.Run("dot", func(t *testing.T) {
		result := NormalizedExt("testdata/test.")
		assert.Equal(t, "", result)
	})

	t.Run("test.z", func(t *testing.T) {
		result := NormalizedExt("testdata/test.z")
		assert.Equal(t, "z", result)
	})

	t.Run("test.jpg", func(t *testing.T) {
		result := NormalizedExt("testdata/test.jpg")
		assert.Equal(t, "jpg", result)
	})

	t.Run("test.PNG", func(t *testing.T) {
		result := NormalizedExt("testdata/test.PNG")
		assert.Equal(t, "png", result)
	})

	t.Run("test.MOV", func(t *testing.T) {
		result := NormalizedExt("testdata/test.MOV")
		assert.Equal(t, "mov", result)
	})

	t.Run("test.xmp", func(t *testing.T) {
		result := NormalizedExt("testdata/test.xMp")
		assert.Equal(t, "xmp", result)
	})

	t.Run("test.MP", func(t *testing.T) {
		result := NormalizedExt("testdata/test.mp")
		assert.Equal(t, "mp", result)
	})
}

func TestTrimExt(t *testing.T) {
	t.Run("WithDot", func(t *testing.T) {
		assert.Equal(t, "raf", TrimExt(".raf"))
	})
	t.Run("Normalized", func(t *testing.T) {
		assert.Equal(t, "cr3", TrimExt("cr3"))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.Equal(t, "aaf", TrimExt("AAF"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", TrimExt(""))
	})
	t.Run("MixedCaseWithDot", func(t *testing.T) {
		assert.Equal(t, "raw", TrimExt(".Raw"))
	})
	t.Run("TypographicQuotes", func(t *testing.T) {
		assert.Equal(t, "jpeg", TrimExt(" “JPEG” "))
	})
}

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
	t.Run("eps.heic", func(t *testing.T) {
		result := StripKnownExt("eps.heic")
		assert.Equal(t, "eps", result)
	})
	t.Run("jpg.eps.heic", func(t *testing.T) {
		result := StripKnownExt("jpg.eps.heic")
		assert.Equal(t, "jpg", result)
	})
	t.Run("eps.jpg.heic", func(t *testing.T) {
		result := StripKnownExt("eps.jpg.heic")
		assert.Equal(t, "eps", result)
	})
	t.Run("/testdata/eps.heic", func(t *testing.T) {
		result := StripKnownExt("/testdata/eps.heic")
		assert.Equal(t, "/testdata/eps", result)
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
