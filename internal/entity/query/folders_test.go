package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestFolderCoverByUID(t *testing.T) {
	t.Run("1990/04", func(t *testing.T) {
		if result, err := FolderCoverByUID("dqo63pn2f87f02xj"); err != nil {
			t.Fatal(err)
		} else if result.FileUID == "" {
			t.Fatal("result must not be empty")
		} else if result.FileUID != "fs6sg6bw15bnlqdw" {
			t.Errorf("wrong result: %#v", result)
		}
	})
	t.Run("2007/12", func(t *testing.T) {
		if result, err := FolderCoverByUID("dqo63pn2f87f02oi"); err != nil {
			t.Fatal(err)
		} else if result.FileUID == "" {
			t.Fatal("result must not be empty")
		} else if result.FileUID != "fs6sg6bqhhinlplk" {
			t.Errorf("wrong result: %#v", result)
		}
	})
}

func TestFoldersByPath(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		folders, err := FoldersByPath(entity.RootOriginals, "testdata", "", false)

		t.Logf("folders: %+v", folders)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, folders, 1)
	})

	t.Run("subdirectory", func(t *testing.T) {
		folders, err := FoldersByPath(entity.RootOriginals, "testdata", "directory", false)

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
		assert.GreaterOrEqual(t, len(folders), 1)

		t.Logf("folders: %+v", folders)
	})
}

func TestUpdateFolderDates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if err := UpdateFolderDates(); err != nil {
			t.Fatal(err)
		}
	})
}
