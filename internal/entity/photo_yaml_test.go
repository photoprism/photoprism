package entity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestPhoto_Yaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()
		result, err := m.Yaml()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("YAML: %s", result)
	})
}

func TestPhoto_SaveAsYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		fileName := filepath.Join(os.TempDir(), ".photoprism_test.yml")

		if err := m.SaveAsYaml(fileName); err != nil {
			t.Fatal(err)
		}

		if err := m.LoadFromYaml(fileName); err != nil {
			t.Fatal(err)
		}

		if err := os.Remove(fileName); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("FilenameEmpty", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		err := m.SaveAsYaml("")

		assert.Error(t, err)
	})
	t.Run("NoPhotoUID", func(t *testing.T) {
		m := Photo{}
		m.PreloadFiles()

		fileName := filepath.Join(os.TempDir(), ".photoprism_test.yml")

		err := m.SaveAsYaml(fileName)

		assert.Error(t, err)
	})
}

func TestPhoto_YamlFileName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()
		fileName, relative, err := m.YamlFileName("xxx", "yyy")
		assert.NoError(t, err)
		assert.Equal(t, "xxx/2790/02/yyy/Photo01.yml", fileName)
		assert.Equal(t, "2790/02/Photo01.yml", relative)

		if err := os.RemoveAll("xxx"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("NotFileName", func(t *testing.T) {
		t.Cleanup(func() { _ = os.RemoveAll("xxx") })
		for _, photoPath := range []string{"2790/02", ""} {
			for _, name := range []string{"", ".", "..", "a/b"} {
				m := &Photo{PhotoPath: photoPath, PhotoName: name}
				fileName, relative, err := m.YamlFileName("xxx", "yyy")
				assert.ErrorIs(t, err, ErrPhotoYamlName, name)
				assert.Empty(t, fileName, name)
				assert.Empty(t, relative, name)
			}
		}
		assert.NoDirExists(t, "xxx")
	})
	t.Run("Nil", func(t *testing.T) {
		_, _, err := (*Photo)(nil).YamlFileName("xxx", "yyy")
		assert.Error(t, err)
	})
}

func TestPhoto_SaveSidecarYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		basePath := fs.Abs("testdata/yaml")
		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		if err := fs.MkdirAll(originalsPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := fs.MkdirAll(sidecarPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := m.SaveSidecarYaml(originalsPath, sidecarPath); err != nil {
			t.Error(err)
		}

		if err := os.RemoveAll(basePath); err != nil {
			t.Error(err)
		}
	})
	t.Run("PhotoNameEmpty", func(t *testing.T) {
		m := Photo{}
		m.PreloadFiles()

		basePath := fs.Abs("testdata/yaml")
		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		if err := fs.MkdirAll(originalsPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := fs.MkdirAll(sidecarPath); err != nil {
			t.Fatal(err)
			return
		}

		err := m.SaveSidecarYaml(originalsPath, sidecarPath)

		assert.Error(t, err)

		if err := os.RemoveAll(basePath); err != nil {
			t.Error(err)
		}
	})
	t.Run("PhotoUIDEmpty", func(t *testing.T) {
		m := Photo{PhotoName: "testphoto"}
		m.PreloadFiles()

		basePath := fs.Abs("testdata/yaml")
		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		if err := fs.MkdirAll(originalsPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := fs.MkdirAll(sidecarPath); err != nil {
			t.Fatal(err)
			return
		}

		err := m.SaveSidecarYaml(originalsPath, sidecarPath)

		assert.Error(t, err)

		if err := os.RemoveAll(basePath); err != nil {
			t.Error(err)
		}
	})
}

func TestPhoto_LoadFromYaml(t *testing.T) {
	t.Run("EmptyFilename", func(t *testing.T) {
		m := Photo{}

		err := m.LoadFromYaml("")

		assert.Error(t, err)
	})
	t.Run("ClipsValues", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "oversized.yml")
		data := "UID: ps6sg6be2lvl0yh0\n" +
			"Title: " + strings.Repeat("ä", 250) + "\n" +
			"TitleSrc: unknownsource\n" +
			"Caption: Short caption\n" +
			"CaptionSrc: manual\n" +
			"OriginalName: " + strings.Repeat("ö", 400) + "\n"
		require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeFile))

		m := Photo{}
		require.NoError(t, m.LoadFromYaml(fileName))
		assert.Equal(t, 200, utf8.RuneCountInString(m.PhotoTitle))
		assert.Equal(t, "", m.TitleSrc)
		assert.Equal(t, "Short caption", m.PhotoCaption)
		assert.Equal(t, SrcManual, m.CaptionSrc)
		assert.LessOrEqual(t, len(m.OriginalName), 755)
		assert.True(t, utf8.ValidString(m.OriginalName))
	})
}

func TestPhoto_ClipYamlValues(t *testing.T) {
	t.Run("WithinLimits", func(t *testing.T) {
		m := Photo{PhotoTitle: " Title ", TitleSrc: SrcManual, PhotoCaption: "Caption\n", PhotoType: MediaImage}
		m.clipYamlValues()
		assert.Equal(t, " Title ", m.PhotoTitle)
		assert.Equal(t, SrcManual, m.TitleSrc)
		assert.Equal(t, "Caption\n", m.PhotoCaption)
		assert.Equal(t, MediaImage, m.PhotoType)
	})
	t.Run("Oversized", func(t *testing.T) {
		m := Photo{
			PhotoType:     "panorama-image",
			TakenSrc:      "123456789",
			PhotoCaption:  strings.Repeat("x", 5000),
			TimeZone:      strings.Repeat("z", 100),
			PhotoExposure: strings.Repeat("ü", 40),
			CameraSerial:  strings.Repeat("s", 200),
			CreatedBy:     strings.Repeat("u", 50),
			UUID:          strings.Repeat("d", 300),
		}
		m.clipYamlValues()
		assert.Equal(t, MediaImage, m.PhotoType)
		assert.Equal(t, "", m.TakenSrc)
		assert.Equal(t, 4096, len(m.PhotoCaption))
		assert.Equal(t, 64, len(m.TimeZone))
		assert.Equal(t, strings.Repeat("ü", 32), m.PhotoExposure)
		assert.Equal(t, 160, len(m.CameraSerial))
		assert.Equal(t, 42, len(m.CreatedBy))
		assert.Equal(t, 255, len(m.UUID))
	})
}
