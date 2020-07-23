package entity

import (
	"github.com/photoprism/photoprism/internal/form"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFolder(t *testing.T) {
	t.Run("2020/05", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "2020/05", nil)
		assert.Equal(t, RootOriginals, folder.Root)
		assert.Equal(t, "2020/05", folder.Path)
		assert.Equal(t, "May 2020", folder.FolderTitle)
		assert.Equal(t, "", folder.FolderDescription)
		assert.Equal(t, "", folder.FolderType)
		assert.Equal(t, SortOrderName, folder.FolderOrder)
		assert.IsType(t, "", folder.FolderUID)
		assert.Equal(t, false, folder.FolderFavorite)
		assert.Equal(t, false, folder.FolderIgnore)
		assert.Equal(t, false, folder.FolderWatch)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "zz", folder.FolderCountry)
	})

	t.Run("/2020/05/01/", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "/2020/05/01/", nil)
		assert.Equal(t, "2020/05/01", folder.Path)
		assert.Equal(t, "May 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "zz", folder.FolderCountry)
	})

	t.Run("/2020/05/23/", func(t *testing.T) {
		folder := NewFolder(RootImport, "/2020/05/23/", nil)
		assert.Equal(t, "2020/05/23", folder.Path)
		assert.Equal(t, "May 23, 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "zz", folder.FolderCountry)
	})

	t.Run("/2020/05/23/Iceland 2020", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "/2020/05/23/Iceland 2020", nil)
		assert.Equal(t, "2020/05/23/Iceland 2020", folder.Path)
		assert.Equal(t, "Iceland 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "is", folder.FolderCountry)
	})

	t.Run("/London/2020/05/23", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "/London/2020/05/23", nil)
		assert.Equal(t, "London/2020/05/23", folder.Path)
		assert.Equal(t, "May 23, 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "gb", folder.FolderCountry)
	})

	t.Run("empty", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "", nil)
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Originals", folder.FolderTitle)
		assert.Equal(t, 0, folder.FolderYear)
		assert.Equal(t, 0, folder.FolderMonth)
		assert.Equal(t, "zz", folder.FolderCountry)
	})

	t.Run("root", func(t *testing.T) {
		folder := NewFolder(RootOriginals, RootPath, nil)
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Originals", folder.FolderTitle)
		assert.Equal(t, 0, folder.FolderYear)
		assert.Equal(t, 0, folder.FolderMonth)
		assert.Equal(t, "zz", folder.FolderCountry)
	})

	t.Run("pathName equals root path", func(t *testing.T) {
		folder := NewFolder("", "", nil)
		assert.Equal(t, "", folder.Path)
	})
}

func TestFirstOrCreateFolder(t *testing.T) {
	folder := NewFolder(RootOriginals, RootPath, nil)
	result := FirstOrCreateFolder(&folder)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if folder.FolderTitle != "Originals" {
		t.Errorf("FolderTitle should be 'Originals'")
	}

	if folder.FolderCountry != "zz" {
		t.Errorf("FolderCountry should be 'zz'")
	}

	found := FindFolder(RootOriginals, RootPath)

	if found == nil {
		t.Fatal("found should not be nil")
	}

	if found.FolderTitle != "Originals" {
		t.Errorf("FolderTitle should be 'Originals'")
	}

	if found.FolderCountry != "zz" {
		t.Errorf("FolderCountry should be 'zz'")
	}
}

func TestFolder_SetValuesFromPath(t *testing.T) {
	t.Run("/", func(t *testing.T) {
		folder := NewFolder("new", "", nil)
		folder.SetValuesFromPath()
		assert.Equal(t, "New", folder.FolderTitle)
	})
}

func TestFolder_Slug(t *testing.T) {
	t.Run("/", func(t *testing.T) {
		folder := Folder{FolderTitle: "Beautiful beach"}
		assert.Equal(t, "beautiful-beach", folder.Slug())
	})
}

func TestFolder_Title(t *testing.T) {
	t.Run("/", func(t *testing.T) {
		folder := Folder{FolderTitle: "Beautiful beach"}
		assert.Equal(t, "Beautiful beach", folder.Title())
	})
}

func TestFindFolder(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, FindFolder("vvfgt", "jgfuyf"))
	})
}

func TestFolder_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		folder := NewFolder("oldRoot", "oldPath", nil)

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

		folder := NewFolder("oldRoot", "oldPath", nil)

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
