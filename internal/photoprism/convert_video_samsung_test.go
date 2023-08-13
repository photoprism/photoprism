package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestConvert_ToSamsungVideo(t *testing.T) {
	conf := config.TestConfig()
	convert := NewConvert(conf)

	t.Run(
		"samsung-motion-photo.jpg", func(t *testing.T) {
			fileName := filepath.Join(
				conf.ExamplesPath(),
				"samsung-motion-photo.jpg",
			)
			sidecarFileName := filepath.Join(
				conf.ExamplesPath(),
				"samsung-motion-photo.json",
			)
			outputName := filepath.Join(
				conf.SidecarPath(), "samsung-motion-photo.mp4",
			)

			_ = os.Remove(outputName)

			assert.Truef(
				t, fs.FileExists(fileName), "input file does not exist: %s",
				fileName,
			)

			mf, err := NewMediaFile(fileName)

			if err != nil {
				t.Fatal(err)
			}

			videoFile, err := convert.ToSamsungVideo(mf, sidecarFileName, false)

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, videoFile.FileName(), outputName)
			assert.Truef(
				t, fs.FileExists(videoFile.FileName()),
				"output file does not exist: %s", videoFile.FileName(),
			)

			_ = os.Remove(outputName)
		},
	)

	t.Run(
		"beach_sand.jpg", func(t *testing.T) {
			fileName := filepath.Join(
				conf.ExamplesPath(),
				"beach_sand.jpg",
			)
			sidecarFileName := filepath.Join(
				conf.ExamplesPath(),
				"beach_sand.json",
			)

			assert.Truef(
				t, fs.FileExists(fileName), "input file does not exist: %s",
				fileName,
			)

			mf, err := NewMediaFile(fileName)

			if err != nil {
				t.Fatal(err)
			}

			videoFile, err := convert.ToSamsungVideo(mf, sidecarFileName, false)

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, (*MediaFile)(nil), videoFile)
		},
	)
}
