package fs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasePrefix(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", BasePrefix("", true))
		assert.Equal(t, "", BasePrefix("", false))
	})
	t.Run("ScreenshotNum2019Num05Num21AtTenNum45Num52Png", func(t *testing.T) {
		regular := BasePrefix("Screenshot 2019-05-21 at 10.45.52.png", false)
		assert.Equal(t, "Screenshot 2019-05-21 at 10.45.52", regular)
		stripped := BasePrefix("Screenshot 2019-05-21 at 10.45.52.png", true)
		assert.Equal(t, "Screenshot 2019-05-21 at 10.45.52", stripped)
	})
	t.Run("TestJpg", func(t *testing.T) {
		result := BasePrefix("/testdata/Test.jpg", true)
		assert.Equal(t, "Test", result)
	})
	t.Run("TestJpgJson", func(t *testing.T) {
		result := BasePrefix("/testdata/Test.jpg.json", true)
		assert.Equal(t, "Test", result)
	})
	t.Run("TestCopyThreeJpg", func(t *testing.T) {
		result := BasePrefix("/testdata/Test copy 3.jpg", true)
		assert.Equal(t, "Test", result)
	})
	t.Run("TestThreeJpg", func(t *testing.T) {
		result := BasePrefix("/testdata/Test (3).jpg", true)
		assert.Equal(t, "Test", result)
	})
	t.Run("DotName", func(t *testing.T) {
		// A sequence form that reduces to a dot name keeps the unstripped base.
		assert.Equal(t, " .", BasePrefix("/testdata/2024/ ..jpg", true))
		assert.Equal(t, " .(1)", BasePrefix("/testdata/2024/ .(1).jpg", true))
		assert.Equal(t, " ..", BasePrefix("/testdata/2024/ ...jpg", true))
		assert.Equal(t, filepath.Join("/testdata/2024", " ."), AbsPrefix("/testdata/2024/ ..jpg", true))
		// A base that is a dot name keeps its extensions.
		for _, strip := range []bool{false, true} {
			assert.Equal(t, ".jpg", BasePrefix("/testdata/2024/.jpg", strip))
			assert.Equal(t, "..jpg", BasePrefix("/testdata/2024/..jpg", strip))
			assert.Equal(t, "...jpg", BasePrefix("/testdata/2024/...jpg", strip))
			assert.Equal(t, "/testdata/2024/..jpg", AbsPrefix("/testdata/2024/..jpg", strip))
			assert.Equal(t, "2024/..jpg", RelPrefix("/testdata/2024/..jpg", "/testdata", strip))
			assert.Equal(t, "", BasePrefix("/testdata/2024/..", strip))
		}
	})
	t.Run("SequenceOnly", func(t *testing.T) {
		assert.Equal(t, "(1)", BasePrefix("/testdata/(1).heic", true))
		assert.Equal(t, " copy 2", BasePrefix("/testdata/ copy 2.jpg", true))
		assert.Equal(t, "(1)", BasePrefix("/testdata/(1).heic", false))
	})
	t.Run("TestJpg", func(t *testing.T) {
		result := BasePrefix("/testdata/Test.jpg", false)
		assert.Equal(t, "Test", result)
	})
	t.Run("TestNum3453453Jpg", func(t *testing.T) {
		regular := BasePrefix("/testdata/Test.3453453.jpg", false)
		assert.Equal(t, "Test.3453453", regular)

		stripped := BasePrefix("/testdata/Test.3453453.jpg", true)
		assert.Equal(t, "Test", stripped)
	})
	t.Run("FooBarNum0000Zip", func(t *testing.T) {
		regular := BasePrefix("/foo/bar.0000.ZIP", false)
		assert.Equal(t, "bar.0000", regular)

		stripped := BasePrefix("/foo/bar.0000.ZIP", true)
		assert.Equal(t, "bar.0000", stripped)
	})
	t.Run("FooBarNum00001Zip", func(t *testing.T) {
		regular := BasePrefix("/foo/bar.00001.ZIP", false)
		assert.Equal(t, "bar.00001", regular)

		stripped := BasePrefix("/foo/bar.00001.ZIP", true)
		assert.Equal(t, "bar", stripped)
	})
	t.Run("TestCopyThreeJpg", func(t *testing.T) {
		result := BasePrefix("/testdata/Test copy 3.jpg", false)
		assert.Equal(t, "Test copy 3", result)
	})
	t.Run("TestThreeJpg", func(t *testing.T) {
		result := BasePrefix("/testdata/Test (3).jpg", false)
		assert.Equal(t, "Test (3)", result)
	})
	t.Run("Num20180506Num091537DscNum02122Jpg", func(t *testing.T) {
		result := BasePrefix("20180506_091537_DSC02122.JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
	t.Run("Num20180506Num091537DscNum02122ThreeThreeJpg", func(t *testing.T) {
		result := BasePrefix("20180506_091537_DSC02122 (+3.3).JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
	t.Run("Num20180506Num091537DscNum02122TwoSevenJpg", func(t *testing.T) {
		result := BasePrefix("20180506_091537_DSC02122 (-2.7).JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
	t.Run("Num20180506Num091537DscNum02122ThreeThreeJpg", func(t *testing.T) {
		result := BasePrefix("20180506_091537_DSC02122(+3.3).JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
	t.Run("Num20180506Num091537DscNum02122TwoSevenJpg", func(t *testing.T) {
		result := BasePrefix("20180506_091537_DSC02122(-2.7).JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
	t.Run("Num1996Num001Jpg", func(t *testing.T) {
		result := BasePrefix("1996 001.jpg", true)
		assert.Equal(t, "1996 001", result)
	})
}

func TestRelPrefix(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", RelPrefix("", "", true))
		assert.Equal(t, "", RelPrefix("", "", false))
	})
	t.Run("FooBarNum0000Zip", func(t *testing.T) {
		regular := RelPrefix("/foo/bar.0000.ZIP", "/bar", false)
		assert.Equal(t, "/foo/bar.0000", regular)

		stripped := RelPrefix("/foo/bar.0000.ZIP", "/bar", true)
		assert.Equal(t, "/foo/bar.0000", stripped)
	})
	t.Run("FooBarNum00001Zip", func(t *testing.T) {
		regular := RelPrefix("/foo/bar.00001.ZIP", "/bar", false)
		assert.Equal(t, "/foo/bar.00001", regular)

		stripped := RelPrefix("/foo/bar.00001.ZIP", "/bar", true)
		assert.Equal(t, "/foo/bar", stripped)
	})
	t.Run("TestCopyThreeJpg", func(t *testing.T) {
		result := RelPrefix("/testdata/foo/Test copy 3.jpg", "/testdata", false)
		assert.Equal(t, "foo/Test copy 3", result)
	})
	t.Run("TestThreeJpg", func(t *testing.T) {
		result := RelPrefix("/testdata/foo/Test (3).jpg", "/testdata", false)
		assert.Equal(t, "foo/Test (3)", result)
	})
	t.Run("TestThreeJpg", func(t *testing.T) {
		result := RelPrefix("/testdata/foo/Test (3).jpg", "/testdata/foo/Test (3).jpg", false)
		assert.Equal(t, "Test (3)", result)
	})
}

func TestAbsPrefix(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", AbsPrefix("", true))
		assert.Equal(t, "", AbsPrefix("", false))
	})
	t.Run("ImgNum4120", func(t *testing.T) {
		assert.Equal(t, "/foo/bar/IMG_4120", AbsPrefix("/foo/bar/IMG_4120.JPG", false))
		assert.Equal(t, "/foo/bar/IMG_E4120", AbsPrefix("/foo/bar/IMG_E4120.JPG", false))
	})
	t.Run("TestCopyThreeJpg", func(t *testing.T) {
		result := AbsPrefix("/testdata/Test (4).jpg", true)

		assert.Equal(t, "/testdata/Test", result)
	})
	t.Run("TestThreeJpg", func(t *testing.T) {
		result := AbsPrefix("/testdata/Test (4).jpg", false)

		assert.Equal(t, "/testdata/Test (4)", result)
	})
	t.Run("Sequence", func(t *testing.T) {
		assert.Equal(t, "/foo/bar/Test", AbsPrefix("/foo/bar/Test (4).jpg", true))
		assert.Equal(t, "/foo/bar/Test (4)", AbsPrefix("/foo/bar/Test (4).jpg", false))
	})
	t.Run("LowerCase", func(t *testing.T) {
		assert.Equal(t, "/foo/bar/IMG_E4120", AbsPrefix("/foo/bar/IMG_E4120.JPG", false))
	})
}

func TestNoBaseName(t *testing.T) {
	t.Run("NoName", func(t *testing.T) {
		assert.True(t, NoBaseName(""))
		assert.True(t, NoBaseName("."))
		assert.True(t, NoBaseName(".."))
	})
	t.Run("Name", func(t *testing.T) {
		assert.False(t, NoBaseName(" ."))
		assert.False(t, NoBaseName("..."))
		assert.False(t, NoBaseName("IMG_1"))
	})
}
