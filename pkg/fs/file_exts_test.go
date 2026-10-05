package fs

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExtensions_Known(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Extensions.Known(""))
	})
	t.Run("Jpg", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test.jpg"))
	})
	t.Run("Jpeg", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test.jpeg"))
	})
	t.Run("Cr2", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/.xxx/test (jpg).cr2"))
	})
	t.Run("CR2", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test (jpg).CR2"))
	})
	t.Run("CR5", func(t *testing.T) {
		assert.False(t, Extensions.Known("testdata/test (jpg).CR5"))
	})
	t.Run("Mp4", func(t *testing.T) {
		assert.True(t, Extensions.Known("file.mp4"))
	})
	t.Run("Mxf", func(t *testing.T) {
		assert.True(t, Extensions.Known("file.mxf"))
	})
}

func TestIsSecondaryRaw(t *testing.T) {
	t.Run("Ori", func(t *testing.T) {
		assert.True(t, IsSecondaryRaw("P1010101.ori"))
		assert.True(t, IsSecondaryRaw("/photos/P1010101.ORI"))
	})
	t.Run("Orf", func(t *testing.T) {
		// The High Res Shot composite is the camera's result and stays the main file.
		assert.False(t, IsSecondaryRaw("P1010101.orf"))
	})
	t.Run("OtherRaw", func(t *testing.T) {
		assert.False(t, IsSecondaryRaw("IMG_2567.CR2"))
		assert.False(t, IsSecondaryRaw("canon_eos_6d.dng"))
	})
	t.Run("NotRaw", func(t *testing.T) {
		assert.False(t, IsSecondaryRaw("cat.jpg"))
		assert.False(t, IsSecondaryRaw(""))
		assert.False(t, IsSecondaryRaw("noextension"))
	})
}

func TestExtensionList(t *testing.T) {
	t.Run("Unique", func(t *testing.T) {
		seen := make(map[string]bool, len(ExtensionList))
		for _, e := range ExtensionList {
			assert.False(t, seen[e.Ext], e.Ext)
			assert.Equal(t, strings.ToLower(e.Ext), e.Ext)
			assert.NotEmpty(t, e.Type, e.Ext)
			seen[e.Ext] = true
		}
	})
	t.Run("Extensions", func(t *testing.T) {
		assert.Len(t, Extensions, len(ExtensionList))
		for _, e := range ExtensionList {
			assert.Equal(t, e.Type, Extensions[e.Ext], e.Ext)
		}
	})
}

func TestFileExtensionList_Map(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := FileExtensionList{{ExtJpeg, ImageJpeg}, {ExtPng, ImagePng}}.Map()
		assert.Equal(t, FileExtensions{ExtJpeg: ImageJpeg, ExtPng: ImagePng}, m)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, FileExtensionList{}.Map())
	})
}

func TestFileExtensionList_Types(t *testing.T) {
	t.Run("Order", func(t *testing.T) {
		l := FileExtensionList{{ExtJpeg, ImageJpeg}, {ExtPng, ImagePng}, {".jpeg", ImageJpeg}}
		assert.Equal(t, TypesExt{ImageJpeg: {".jpg", ".JPG", ".jpeg", ".JPEG"}, ImagePng: {".png", ".PNG"}}, l.Types(false))
		assert.Equal(t, TypesExt{ImageJpeg: {".jpg", ".jpeg"}, ImagePng: {".png"}}, l.Types(true))
	})
	t.Run("DefaultFirst", func(t *testing.T) {
		for fileType, exts := range ExtensionList.Types(true) {
			if slices.Contains(exts, fileType.DefaultExt()) {
				assert.Equal(t, fileType.DefaultExt(), exts[0], fileType)
			}
		}
		assert.Equal(t, []string{".yml", ".yaml"}, ExtensionList.Types(true)[SidecarYaml])
	})
	t.Run("PreferredFirst", func(t *testing.T) {
		// Repeated, as a list built by iterating a map would only sometimes start with the same extensions.
		for i := 0; i < 10; i++ {
			assert.Equal(t, []string{".jpg", ".JPG", ".jpeg", ".JPEG"}, ExtensionList.Types(false)[ImageJpeg][:4])
			assert.Equal(t, []string{".png", ".PNG"}, ExtensionList.Types(false)[ImagePng][:2])
		}
	})
}
