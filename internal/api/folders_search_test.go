package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestGetFoldersOriginals(t *testing.T) {
	t.Run("flat", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, err := fs.Dirs(conf.OriginalsPath(), false, true)

		if err != nil {
			t.Fatal(err)
		}

		SearchFoldersOriginals(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/originals")

		// t.Logf("RESPONSE: %s", r.Body.Bytes())

		var resp FoldersResponse
		err = json.Unmarshal(r.Body.Bytes(), &resp)

		if err != nil {
			t.Fatal(err)
		}

		folders := resp.Folders

		if len(folders) != len(expected) {
			t.Fatalf("response contains %d folders", len(folders))
		}

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
		}
	})
	t.Run("recursive", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, err := fs.Dirs(conf.OriginalsPath(), true, true)

		if err != nil {
			t.Fatal(err)
		}
		SearchFoldersOriginals(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/originals?recursive=true")

		// t.Logf("RESPONSE: %s", r.Body.Bytes())

		var resp FoldersResponse
		err = json.Unmarshal(r.Body.Bytes(), &resp)

		if err != nil {
			t.Fatal(err)
		}

		folders := resp.Folders

		if len(folders) != len(expected) {
			t.Fatalf("response contains %d folders", len(folders))
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
		}
	})
}

func TestGetFoldersImport(t *testing.T) {
	t.Run("flat", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, err := fs.Dirs(conf.ImportPath(), false, true)

		if err != nil {
			t.Fatal(err)
		}

		SearchFoldersImport(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/import")

		// t.Logf("RESPONSE: %s", r.Body.Bytes())

		var resp FoldersResponse
		err = json.Unmarshal(r.Body.Bytes(), &resp)

		if err != nil {
			t.Fatal(err)
		}

		folders := resp.Folders

		if len(folders) != len(expected) {
			t.Fatalf("response contains %d folders", len(folders))
		}

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
		}

	})
	t.Run("recursive", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, err := fs.Dirs(conf.ImportPath(), true, true)

		if err != nil {
			t.Fatal(err)
		}

		SearchFoldersImport(router)
		r := PerformRequest(app, "GET", "/api/v1/folders/import?recursive=true")

		var resp FoldersResponse
		err = json.Unmarshal(r.Body.Bytes(), &resp)

		if err != nil {
			t.Fatal(err)
		}

		folders := resp.Folders

		if len(folders) != len(expected) {
			t.Fatalf("response contains %d folders", len(folders))
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
		}
	})
}
