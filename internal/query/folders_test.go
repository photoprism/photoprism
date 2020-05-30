package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestFoldersByPath(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		folders, err := FoldersByPath(entity.RootDefault, "testdata", "", false)

		t.Logf("folders: %+v", folders)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, folders, 1)
	})

	t.Run("subdirectory", func(t *testing.T) {
		folders, err := FoldersByPath(entity.RootDefault, "testdata", "directory", false)

		t.Logf("folders: %+v", folders)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, folders, 2)
	})
}

func TestAlbumFolders(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		folders, err := AlbumFolders(1)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, folders, 1)

		t.Logf("folders: %+v", folders)
	})
}
