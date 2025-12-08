package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeExt(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		result := NormalizedExt("testdata/test")
		assert.Equal(t, "", result)
	})
	t.Run("Dot", func(t *testing.T) {
		result := NormalizedExt("testdata/test.")
		assert.Equal(t, "", result)
	})
	t.Run("TestZ", func(t *testing.T) {
		result := NormalizedExt("testdata/test.z")
		assert.Equal(t, "z", result)
	})
	t.Run("TestJpg", func(t *testing.T) {
		result := NormalizedExt("testdata/test.jpg")
		assert.Equal(t, "jpg", result)
	})
	t.Run("TestPng", func(t *testing.T) {
		result := NormalizedExt("testdata/test.PNG")
		assert.Equal(t, "png", result)
	})
	t.Run("TestMov", func(t *testing.T) {
		result := NormalizedExt("testdata/test.MOV")
		assert.Equal(t, "mov", result)
	})
	t.Run("TestXmp", func(t *testing.T) {
		result := NormalizedExt("testdata/test.xMp")
		assert.Equal(t, "xmp", result)
	})
	t.Run("TestMp", func(t *testing.T) {
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
	t.Run("TestJpg", func(t *testing.T) {
		result := StripExt("/testdata/Test.jpg")
		assert.Equal(t, "/testdata/Test", result)
	})
	t.Run("TestJpgJson", func(t *testing.T) {
		result := StripExt("/testdata/Test.jpg.json")
		assert.Equal(t, "/testdata/Test.jpg", result)
	})
	t.Run("TestCopyThreeFoo", func(t *testing.T) {
		result := StripExt("/testdata/Test copy 3.foo")
		assert.Equal(t, "/testdata/Test copy 3", result)
	})
	t.Run("MpJpg", func(t *testing.T) {
		assert.Equal(t, "name.MP", StripExt("name.MP.jpg"))
	})
}

func TestStripKnownExt(t *testing.T) {
	t.Run("TestJpg", func(t *testing.T) {
		result := StripKnownExt("/testdata/Test.jpg")
		assert.Equal(t, "/testdata/Test", result)
	})
	t.Run("TestJpgJson", func(t *testing.T) {
		result := StripKnownExt("/testdata/Test.jpg.json")
		assert.Equal(t, "/testdata/Test", result)
	})
	t.Run("TestCopyThreeFoo", func(t *testing.T) {
		result := StripKnownExt("/testdata/Test copy 3.foo")
		assert.Equal(t, "/testdata/Test copy 3.foo", result)
	})
	t.Run("MyFileJpgJsonXmp", func(t *testing.T) {
		result := StripKnownExt("my/file.jpg.json.xmp")
		assert.Equal(t, "my/file", result)
	})
	t.Run("MyJpgAviFooBarBaz", func(t *testing.T) {
		result := StripKnownExt("my/jpg/avi.foo.bar.baz")
		assert.Equal(t, "my/jpg/avi.foo.bar.baz", result)
	})
	t.Run("EpsHeic", func(t *testing.T) {
		result := StripKnownExt("eps.heic")
		assert.Equal(t, "eps", result)
	})
	t.Run("JpgEpsHeic", func(t *testing.T) {
		result := StripKnownExt("jpg.eps.heic")
		assert.Equal(t, "jpg", result)
	})
	t.Run("EpsJpgHeic", func(t *testing.T) {
		result := StripKnownExt("eps.jpg.heic")
		assert.Equal(t, "eps", result)
	})
	t.Run("TestdataEpsHeic", func(t *testing.T) {
		result := StripKnownExt("/testdata/eps.heic")
		assert.Equal(t, "/testdata/eps", result)
	})
}

func TestExt(t *testing.T) {
	t.Run("TestJpg", func(t *testing.T) {
		result := Ext("/testdata/Test.jpg")
		assert.Equal(t, ".jpg", result)
	})
	t.Run("TestJpgJson", func(t *testing.T) {
		result := Ext("/testdata/Test.jpg.json")
		assert.Equal(t, ".jpg.json", result)
	})
	t.Run("TestCopyThreeFoo", func(t *testing.T) {
		result := Ext("/testdata/Test copy 3.foo")
		assert.Equal(t, ".foo", result)
	})
	t.Run("Test", func(t *testing.T) {
		result := Ext("/testdata/Test")
		assert.Equal(t, "", result)
	})
}

func TestArchiveExt(t *testing.T) {
	t.Run("Zip", func(t *testing.T) {
		assert.Equal(t, ExtZip, ArchiveExt("/testdata/archive.ZIP"))
	})
	t.Run("NotArchive", func(t *testing.T) {
		assert.Equal(t, "", ArchiveExt("/testdata/file.jpg"))
	})
}
