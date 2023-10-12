package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConvert_FixJpeg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cnf := config.TestConfig()
	cnf.InitializeTestData()
	convert := NewConvert(cnf)

	t.Run("elephants.jpg", func(t *testing.T) {
		fileName := filepath.Join(cnf.ExamplesPath(), "elephants.jpg")
		outputName := filepath.Join(cnf.MediaFileCachePath("b10447b54c3330eb13566735322e971cc1dcbc41"),
			"b10447b54c3330eb13566735322e971cc1dcbc41.jpg")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jpegFile, err := convert.FixJpeg(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, jpegFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(jpegFile.FileName()), "output file does not exist: %s", jpegFile.FileName())

		t.Logf("old jpeg filename: %s", mf.FileName())
		t.Logf("old jpeg metadata: %#v", mf.MetaData())
		t.Logf("new jpeg filename: %s", jpegFile.FileName())
		t.Logf("new jpeg metadata: %#v", jpegFile.MetaData())

		_ = os.Remove(outputName)
	})
}
