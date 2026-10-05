package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestDeletePhoto(t *testing.T) {
	c := Config()

	// newPhoto creates a photo with the given path and name in the index.
	newPhoto := func(t *testing.T, photoPath, photoName string) *entity.Photo {
		t.Helper()
		p := &entity.Photo{PhotoUID: rnd.GenerateUID(entity.PhotoUID), PhotoPath: photoPath, PhotoName: photoName}
		require.NoError(t, p.Create())
		return p
	}

	// writeYaml writes a YAML backup with the given relative name to the sidecar folder.
	writeYaml := func(t *testing.T, relName string) string {
		t.Helper()
		fileName := filepath.Join(c.SidecarPath(), relName)
		require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
		require.NoError(t, os.WriteFile(fileName, []byte("UID: test\n"), fs.ModeFile))
		t.Cleanup(func() { _ = os.Remove(fileName) })
		return fileName
	}

	t.Run("Success", func(t *testing.T) {
		folder := "delete-photo-" + rnd.Base36(8)
		p := newPhoto(t, folder, "IMG_1")
		yamlName := writeYaml(t, filepath.Join(folder, "IMG_1.yml"))
		numFiles, err := DeletePhoto(p, false, false)
		require.NoError(t, err)
		assert.Equal(t, 1, numFiles)
		assert.NoFileExists(t, yamlName)
	})
	t.Run("NotFileName", func(t *testing.T) {
		// The backup named after the folder belongs to another photo.
		for _, name := range []string{"", "."} {
			folder := "delete-photo-" + rnd.Base36(8)
			p := newPhoto(t, folder, name)
			yamlName := writeYaml(t, folder+".yml")
			numFiles, err := DeletePhoto(p, false, false)
			require.NoError(t, err)
			assert.Equal(t, 0, numFiles)
			assert.FileExists(t, yamlName)
		}
	})
	t.Run("RootDot", func(t *testing.T) {
		// A root-level photo named "." would name the backup of the photo named after the originals folder.
		p := newPhoto(t, "", ".")
		yamlName := writeYaml(t, filepath.Join(filepath.Dir(c.OriginalsPath()), filepath.Base(c.OriginalsPath())+".yml"))
		numFiles, err := DeletePhoto(p, false, false)
		require.NoError(t, err)
		assert.Equal(t, 0, numFiles)
		assert.FileExists(t, yamlName)
	})
	t.Run("Nil", func(t *testing.T) {
		_, err := DeletePhoto(nil, false, false)
		assert.Error(t, err)
	})
}
