package api

import (
	"encoding/json"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestGetFoldersOriginals(t *testing.T) {
	t.Run("flat", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, err := fs.Dirs(conf.OriginalsPath(), false)

		if err != nil {
			t.Fatal(err)
		}

		GetFoldersOriginals(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/folders/originals")

		var folders []entity.Folder
		err = json.Unmarshal(r.Body.Bytes(), &folders)

		if err != nil {
			t.Fatal(err)
		}

		if len(folders) != len(expected){
			t.Fatalf("response contains %d folders", len(folders))
		}

		if len(folders) == 0 {
			// There are no existing folders, that's ok.
			return
		}

		for _, folder := range folders {
			assert.Equal(t, "", folder.FolderDescription)
			assert.Equal(t, entity.TypeDefault, folder.FolderType)
			assert.Equal(t, entity.SortOrderName, folder.FolderOrder)
			assert.Equal(t, entity.FolderRootOriginals, folder.Root)
			assert.Equal(t, "", folder.FolderUUID)
			assert.Equal(t, false, folder.FolderFavorite)
			assert.Equal(t, false, folder.FolderHidden)
			assert.Equal(t, false, folder.FolderIgnore)
			assert.Equal(t, false, folder.FolderWatch)
		}

		// t.Logf("ORIGINALS: %+v", folders)
	})
	t.Run("recursive", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, err := fs.Dirs(conf.OriginalsPath(), true)

		if err != nil {
			t.Fatal(err)
		}
		GetFoldersOriginals(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/folders/originals?recursive=true")
		var folders []entity.Folder
		err = json.Unmarshal(r.Body.Bytes(), &folders)

		if err != nil {
			t.Fatal(err)
		}

		if len(folders) != len(expected){
			t.Fatalf("response contains %d folders", len(folders))
		}

		for _, folder := range folders {
			assert.Equal(t, "", folder.FolderDescription)
			assert.Equal(t, entity.TypeDefault, folder.FolderType)
			assert.Equal(t, entity.SortOrderName, folder.FolderOrder)
			assert.Equal(t, entity.FolderRootOriginals, folder.Root)
			assert.Equal(t, "", folder.FolderUUID)
			assert.Equal(t, false, folder.FolderFavorite)
			assert.Equal(t, false, folder.FolderHidden)
			assert.Equal(t, false, folder.FolderIgnore)
			assert.Equal(t, false, folder.FolderWatch)
		}

		// t.Logf("ORIGINALS RECURSIVE: %+v", folders)
	})
}

func TestGetFoldersImport(t *testing.T) {
	t.Run("flat", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, err := fs.Dirs(conf.ImportPath(), false)

		if err != nil {
			t.Fatal(err)
		}

		GetFoldersImport(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/folders/import")
		var folders []entity.Folder
		err = json.Unmarshal(r.Body.Bytes(), &folders)

		if err != nil {
			t.Fatal(err)
		}

		if len(folders) != len(expected){
			t.Fatalf("response contains %d folders", len(folders))
		}

		// t.Logf("IMPORT FOLDERS: %+v", folders)

		if len(folders) == 0 {
			// There are no existing folders, that's ok.
			return
		}

		for _, folder := range folders {
			assert.Equal(t, "", folder.FolderDescription)
			assert.Equal(t, entity.TypeDefault, folder.FolderType)
			assert.Equal(t, entity.SortOrderName, folder.FolderOrder)
			assert.Equal(t, entity.FolderRootImport, folder.Root)
			assert.Equal(t, "", folder.FolderUUID)
			assert.Equal(t, false, folder.FolderFavorite)
			assert.Equal(t, false, folder.FolderHidden)
			assert.Equal(t, false, folder.FolderIgnore)
			assert.Equal(t, false, folder.FolderWatch)
		}

	})
	t.Run("recursive", func(t *testing.T) {
		app, router, conf := NewApiTest()
		_ = conf.CreateDirectories()
		expected, err := fs.Dirs(conf.ImportPath(), true)

		if err != nil {
			t.Fatal(err)
		}
		GetFoldersImport(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/folders/import?recursive=true")
		var folders []entity.Folder
		err = json.Unmarshal(r.Body.Bytes(), &folders)

		if err != nil {
			t.Fatal(err)
		}

		if len(folders) != len(expected){
			t.Fatalf("response contains %d folders", len(folders))
		}

		for _, folder := range folders {
			assert.Equal(t, "", folder.FolderDescription)
			assert.Equal(t, entity.TypeDefault, folder.FolderType)
			assert.Equal(t, entity.SortOrderName, folder.FolderOrder)
			assert.Equal(t, entity.FolderRootImport, folder.Root)
			assert.Equal(t, "", folder.FolderUUID)
			assert.Equal(t, false, folder.FolderFavorite)
			assert.Equal(t, false, folder.FolderHidden)
			assert.Equal(t, false, folder.FolderIgnore)
			assert.Equal(t, false, folder.FolderWatch)
		}

		// t.Logf("IMPORT FOLDERS RECURSIVE: %+v", folders)
	})
}
