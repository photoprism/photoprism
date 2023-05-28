package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestConvert_ToJson(t *testing.T) {
	conf := config.TestConfig()
	convert := NewConvert(conf)

	t.Run("gopher-video.mp4", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})

	t.Run("IMG_4120.JPG", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "IMG_4120.JPG")
		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})

	t.Run("iphone_7.heic", func(t *testing.T) {
		fileName := conf.ExamplesPath() + "/iphone_7.heic"

		assert.True(t, fs.FileExists(fileName))

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})
}
