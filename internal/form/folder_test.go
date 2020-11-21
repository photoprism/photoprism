package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFolder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var folder = struct {
			Path              string
			Root              string
			FolderType        string
			FolderTitle       string
			FolderCategory    string
			FolderDescription string
			FolderOrder       string
			FolderCountry     string
			FolderYear        int
			FolderMonth       int
			FolderFavorite    bool
			FolderPrivate     bool
			FolderIgnore      bool
			FolderWatch       bool
		}{Path: "HD/2011/11-WG-Party",
			Root:              "",
			FolderType:        "",
			FolderTitle:       "testTitle",
			FolderCategory:    "family",
			FolderDescription: "",
			FolderOrder:       "name",
			FolderCountry:     "de",
			FolderYear:        2020,
			FolderMonth:       07,
			FolderFavorite:    false,
			FolderPrivate:     false,
			FolderIgnore:      false,
			FolderWatch:       false,
		}

		r, err := NewFolder(folder)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "HD/2011/11-WG-Party", r.Path)
		assert.Equal(t, "", r.Root)
		assert.Equal(t, "", r.FolderType)
		assert.Equal(t, "testTitle", r.FolderTitle)
		assert.Equal(t, "family", r.FolderCategory)
		assert.Equal(t, "", r.FolderDescription)
		assert.Equal(t, "name", r.FolderOrder)
		assert.Equal(t, "de", r.FolderCountry)
		assert.Equal(t, 2020, r.FolderYear)
		assert.Equal(t, false, r.FolderPrivate)
		assert.Equal(t, 07, r.FolderMonth)
		assert.Equal(t, false, r.FolderFavorite)
		assert.Equal(t, false, r.FolderIgnore)
		assert.Equal(t, false, r.FolderWatch)
	})
}
