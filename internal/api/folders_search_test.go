package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/pkg/dirs"
)

func TestGetFoldersOriginals(t *testing.T) {
	t.Run("Flat", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, counts, err := dirs.Dirs(conf.OriginalsPath(), false, true)

		if err != nil {
			t.Fatal(err)
		}

		SearchFoldersOriginals(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/originals?includeroot=true")

		// t.Logf("RESPONSE: %s", r.Body.Bytes())

		var resp FoldersResponse
		err = json.Unmarshal(r.Body.Bytes(), &resp)

		if err != nil {
			t.Fatal(err)
		}

		folders := resp.Folders

		assert.Len(t, folders, len(expected), "folder length incorrect")

		if len(folders) == 0 {
			// There are no existing folders, that's ok.
			return
		}

		for _, folder := range folders {
			assert.Equal(t, "", folder.FolderDescription)
			assert.Equal(t, entity.MediaUnknown, folder.FolderType)
			assert.Equal(t, sortby.Name, folder.FolderOrder)
			assert.Equal(t, entity.RootOriginals, folder.Root)
			assert.IsType(t, "", folder.FolderUID)
			assert.Equal(t, false, folder.FolderFavorite)
			assert.Equal(t, false, folder.FolderIgnore)
			assert.Equal(t, false, folder.FolderWatch)
			assert.Equal(t, counts["/"+folder.Path], folder.FileCount)
		}
	})
	t.Run("Recursive", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, counts, err := dirs.Dirs(conf.OriginalsPath(), true, true)

		if err != nil {
			t.Fatal(err)
		}
		SearchFoldersOriginals(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/originals?recursive=true&includeroot=true")

		// t.Logf("RESPONSE: %s", r.Body.Bytes())

		var resp FoldersResponse
		err = json.Unmarshal(r.Body.Bytes(), &resp)

		if err != nil {
			t.Fatal(err)
		}

		folders := resp.Folders

		assert.Len(t, folders, len(expected), "folder length incorrect")

		for _, folder := range folders {
			assert.Equal(t, "", folder.FolderDescription)
			assert.Equal(t, entity.MediaUnknown, folder.FolderType)
			assert.Equal(t, sortby.Name, folder.FolderOrder)
			assert.Equal(t, entity.RootOriginals, folder.Root)
			assert.IsType(t, "", folder.FolderUID)
			assert.Equal(t, false, folder.FolderFavorite)
			assert.Equal(t, false, folder.FolderIgnore)
			assert.Equal(t, false, folder.FolderWatch)
			assert.Equal(t, counts["/"+folder.Path], folder.FileCount)
		}
	})
}

func TestGetFoldersImport(t *testing.T) {
	t.Run("Flat", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, counts, err := dirs.Dirs(conf.ImportPath(), false, true)

		if err != nil {
			t.Fatal(err)
		}

		SearchFoldersImport(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/import?includeroot=true")

		// t.Logf("RESPONSE: %s", r.Body.Bytes())

		var resp FoldersResponse
		err = json.Unmarshal(r.Body.Bytes(), &resp)

		if err != nil {
			t.Fatal(err)
		}

		folders := resp.Folders

		assert.Len(t, folders, len(expected), "folder length incorrect")

		if len(folders) == 0 {
			// There are no existing folders, that's ok.
			return
		}

		for _, folder := range folders {
			assert.Equal(t, "", folder.FolderDescription)
			assert.Equal(t, entity.MediaUnknown, folder.FolderType)
			assert.Equal(t, sortby.Name, folder.FolderOrder)
			assert.Equal(t, entity.RootImport, folder.Root)
			assert.IsType(t, "", folder.FolderUID)
			assert.Equal(t, false, folder.FolderFavorite)
			assert.Equal(t, false, folder.FolderIgnore)
			assert.Equal(t, false, folder.FolderWatch)
			assert.Equal(t, counts["/"+folder.Path], folder.FileCount)
		}

	})
	t.Run("FlatNoRoot", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, counts, err := dirs.Dirs(conf.ImportPath(), false, true)

		if err != nil {
			t.Fatal(err)
		}

		SearchFoldersImport(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/import?uncached=true")

		// t.Logf("RESPONSE: %s", r.Body.Bytes())

		var resp FoldersResponse
		err = json.Unmarshal(r.Body.Bytes(), &resp)

		if err != nil {
			t.Fatal(err)
		}

		folders := resp.Folders

		assert.Len(t, folders, len(expected)-1, "folder length incorrect")

		if len(folders) == 0 {
			// There are no existing folders, that's ok.
			return
		}

		for _, folder := range folders {
			assert.Equal(t, "", folder.FolderDescription)
			assert.Equal(t, entity.MediaUnknown, folder.FolderType)
			assert.Equal(t, sortby.Name, folder.FolderOrder)
			assert.Equal(t, entity.RootImport, folder.Root)
			assert.IsType(t, "", folder.FolderUID)
			assert.Equal(t, false, folder.FolderFavorite)
			assert.Equal(t, false, folder.FolderIgnore)
			assert.Equal(t, false, folder.FolderWatch)
			assert.Equal(t, counts["/"+folder.Path], folder.FileCount)
		}

	})
	t.Run("Recursive", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, counts, err := dirs.Dirs(conf.ImportPath(), true, true)

		if err != nil {
			t.Fatal(err)
		}

		SearchFoldersImport(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/import?recursive=true&includeroot=true")

		var resp FoldersResponse
		err = json.Unmarshal(r.Body.Bytes(), &resp)

		if err != nil {
			t.Fatal(err)
		}

		folders := resp.Folders

		assert.Len(t, folders, len(expected), "folder length incorrect")

		for _, folder := range folders {
			assert.Equal(t, "", folder.FolderDescription)
			assert.Equal(t, entity.MediaUnknown, folder.FolderType)
			assert.Equal(t, sortby.Name, folder.FolderOrder)
			assert.Equal(t, entity.RootImport, folder.Root)
			assert.IsType(t, "", folder.FolderUID)
			assert.Equal(t, false, folder.FolderFavorite)
			assert.Equal(t, false, folder.FolderIgnore)
			assert.Equal(t, false, folder.FolderWatch)
			assert.Equal(t, counts["/"+folder.Path], folder.FileCount)
		}
	})
}
