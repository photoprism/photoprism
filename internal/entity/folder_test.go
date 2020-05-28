package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFolder(t *testing.T) {
	t.Run("2020/05", func(t *testing.T) {
		folder := NewFolder(RootDefault, "2020/05", nil)
		assert.Equal(t, RootDefault, folder.Root)
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
		folder := NewFolder(RootDefault, "/2020/05/01/", nil)
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
		folder := NewFolder(RootDefault, "/2020/05/23/Iceland 2020", nil)
		assert.Equal(t, "2020/05/23/Iceland 2020", folder.Path)
		assert.Equal(t, "Iceland 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "is", folder.FolderCountry)
	})

	t.Run("/London/2020/05/23", func(t *testing.T) {
		folder := NewFolder(RootDefault, "/London/2020/05/23", nil)
		assert.Equal(t, "London/2020/05/23", folder.Path)
		assert.Equal(t, "May 23, 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "gb", folder.FolderCountry)
	})

	t.Run("empty", func(t *testing.T) {
		folder := NewFolder(RootDefault, "", nil)
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Originals", folder.FolderTitle)
		assert.Equal(t, 0, folder.FolderYear)
		assert.Equal(t, 0, folder.FolderMonth)
		assert.Equal(t, "zz", folder.FolderCountry)
	})

	t.Run("root", func(t *testing.T) {
		folder := NewFolder(RootDefault, RootPath, nil)
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Originals", folder.FolderTitle)
		assert.Equal(t, 0, folder.FolderYear)
		assert.Equal(t, 0, folder.FolderMonth)
		assert.Equal(t, "zz", folder.FolderCountry)
	})
}

func TestFirstOrCreateFolder(t *testing.T) {
	folder := NewFolder(RootDefault, RootPath, nil)
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

	found := FindFolder(RootDefault, RootPath)

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
