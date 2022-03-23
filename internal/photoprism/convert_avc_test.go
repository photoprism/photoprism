package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestConvert_ToAvc(t *testing.T) {
	t.Run("gopher-video.mp4", func(t *testing.T) {
		conf := config.TestConfig()
		convert := NewConvert(conf)

		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")
		outputName := filepath.Join(conf.SidecarPath(), conf.ExamplesPath(), "gopher-video.mp4.avc")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		avcFile, err := convert.ToAvc(mf, "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, avcFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(avcFile.FileName()), "output file does not exist: %s", avcFile.FileName())

		t.Logf("video metadata: %+v", avcFile.MetaData())

		_ = os.Remove(outputName)
	})

	t.Run("jpg", func(t *testing.T) {
		conf := config.TestConfig()
		convert := NewConvert(conf)

		fileName := filepath.Join(conf.ExamplesPath(), "cat_black.jpg")
		outputName := filepath.Join(conf.SidecarPath(), conf.ExamplesPath(), "cat_black.jpg.avc")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		avcFile, err := convert.ToAvc(mf, "")
		assert.Error(t, err)
		assert.Nil(t, avcFile)
	})
}
