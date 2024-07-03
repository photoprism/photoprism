package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
)

func TestNewFolder(t *testing.T) {
	t.Run("2020/05", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "2020/05", time.Now().UTC())
		assert.Equal(t, RootOriginals, folder.Root)
		assert.Equal(t, "2020/05", folder.Path)
		assert.Equal(t, "May 2020", folder.FolderTitle)
		assert.Equal(t, "", folder.FolderDescription)
		assert.Equal(t, "", folder.FolderType)
		assert.Equal(t, sortby.Name, folder.FolderOrder)
		assert.IsType(t, "", folder.FolderUID)
		assert.Equal(t, false, folder.FolderFavorite)
		assert.Equal(t, false, folder.FolderIgnore)
		assert.Equal(t, false, folder.FolderWatch)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})

	t.Run("/2020/05/01/", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "/2020/05/01/", time.Now().UTC())
		assert.Equal(t, "2020/05/01", folder.Path)
		assert.Equal(t, "May 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})

	t.Run("/2020/05/23/", func(t *testing.T) {
		folder := NewFolder(RootImport, "/2020/05/23/", time.Now().UTC())
		assert.Equal(t, "2020/05/23", folder.Path)
		assert.Equal(t, "May 23, 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})

	t.Run("/2020/05/23/Iceland 2020", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "/2020/05/23/Iceland 2020", time.Now().UTC())
		assert.Equal(t, "2020/05/23/Iceland 2020", folder.Path)
		assert.Equal(t, "Iceland 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "is", folder.FolderCountry)
	})

	t.Run("/London/2020/05/23", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "/London/2020/05/23", time.Now().UTC())
		assert.Equal(t, "London/2020/05/23", folder.Path)
		assert.Equal(t, "May 23, 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "gb", folder.FolderCountry)
	})

	t.Run("empty", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "", time.Time{})
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Originals", folder.FolderTitle)
		assert.Equal(t, 0, folder.FolderYear)
		assert.Equal(t, 0, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})

	t.Run("root", func(t *testing.T) {
		folder := NewFolder(RootOriginals, RootPath, time.Time{})
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Originals", folder.FolderTitle)
		assert.Equal(t, 0, folder.FolderYear)
		assert.Equal(t, 0, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})

	t.Run("pathName equals root path", func(t *testing.T) {
		folder := NewFolder("", RootPath, time.Now().UTC())
		assert.Equal(t, "", folder.Path)
	})
}

func TestFirstOrCreateFolder(t *testing.T) {
	folder := NewFolder(RootOriginals, RootPath, time.Now().UTC())
	result := FirstOrCreateFolder(&folder)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if folder.FolderTitle != "Originals" {
		t.Errorf("FolderTitle should be 'Originals'")
	}

	if folder.FolderCountry != UnknownID {
		t.Errorf("FolderCountry should be 'zz'")
	}

	found := FindFolder(RootOriginals, RootPath)

	if found == nil {
		t.Fatal("found should not be nil")
	}

	if found.FolderTitle != "Originals" {
		t.Errorf("FolderTitle should be 'Originals'")
	}

	if found.FolderCountry != UnknownID {
		t.Errorf("FolderCountry should be 'zz'")
	}
}

func TestFolder_SetValuesFromPath(t *testing.T) {
	t.Run("/", func(t *testing.T) {
		folder := NewFolder("new", "", time.Now().UTC())
		folder.SetValuesFromPath()
		assert.Equal(t, "New", folder.FolderTitle)
	})
}

func TestFolder_Slug(t *testing.T) {
	t.Run("/", func(t *testing.T) {
		folder := Folder{FolderTitle: "Beautiful beach", Root: "sidecar", Path: "ugly/beach"}
		assert.Equal(t, "ugly-beach", folder.Slug())
	})
}

func TestFolder_Title(t *testing.T) {
	t.Run("/", func(t *testing.T) {
		folder := Folder{FolderTitle: "Beautiful beach"}
		assert.Equal(t, "Beautiful beach", folder.Title())
	})
}

func TestFolder_RootPath(t *testing.T) {
	t.Run("/rainbow", func(t *testing.T) {
		folder := Folder{FolderTitle: "Beautiful beach", Root: "/", Path: "rainbow"}
		assert.Equal(t, "/rainbow", folder.RootPath())
	})
}
func TestFindFolder(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, FindFolder("vvfgt", "jgfuyf"))
	})
	t.Run("pathName === rootPath", func(t *testing.T) {
		assert.Nil(t, FindFolder("vvfgt", RootPath))
	})
}

func TestFolder_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		folder := NewFolder("oldRoot", "oldPath", time.Now().UTC())

		assert.Equal(t, "oldRoot", folder.Root)
		assert.Equal(t, "oldPath", folder.Path)

		err := folder.Updates(Folder{Root: "newRoot", Path: "newPath"})

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "newRoot", folder.Root)
		assert.Equal(t, "newPath", folder.Path)
	})
}

func TestFolder_SetForm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		formValues := Folder{FolderTitle: "Beautiful beach"}

		folderForm, err := form.NewFolder(formValues)

		folder := NewFolder("oldRoot", "oldPath", time.Now().UTC())

		assert.Equal(t, "oldRoot", folder.Root)
		assert.Equal(t, "oldPath", folder.Path)
		assert.Equal(t, "OldPath", folder.FolderTitle)

		err = folder.SetForm(folderForm)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", folder.Root)
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Beautiful beach", folder.FolderTitle)
	})
}

func TestFolder_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		folder := Folder{FolderTitle: "Holiday 2020", Root: RootOriginals, Path: "2020/Greece"}
		err := folder.Create()
		if err != nil {
			t.Fatal(err)
		}
		result := FindFolder(RootOriginals, "2020/Greece")
		assert.Equal(t, "2020-greece", result.Slug())
		assert.Equal(t, "Holiday 2020", result.Title())
	})
}
