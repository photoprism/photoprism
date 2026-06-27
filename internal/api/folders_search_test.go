package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestGetFoldersOriginals(t *testing.T) {
	t.Run("Flat", func(t *testing.T) {
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
	t.Run("Recursive", func(t *testing.T) {
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

func TestSearchFolders_Headers(t *testing.T) {
	app, router, conf := NewApiTest()
	_ = conf.CreateDirectories()
	SearchFoldersOriginals(router)

	// Request uncached results so the count headers are always written
	// (a cache hit returns early before the headers are added).
	r := PerformRequest(app, "GET", "/api/v1/folders/originals?uncached=true")
	assert.Equal(t, http.StatusOK, r.Code)
	result := r.Result()

	xFiles, err := strconv.Atoi(result.Header.Get("X-Files"))
	assert.NoError(t, err, "X-Files header should be an integer")
	xFolders, err := strconv.Atoi(result.Header.Get("X-Folders"))
	assert.NoError(t, err, "X-Folders header should be an integer")
	xCount, err := strconv.Atoi(result.Header.Get("X-Count"))
	assert.NoError(t, err, "X-Count header should be an integer")

	// X-Count is the combined number of files and folders returned.
	assert.Equal(t, xFiles+xFolders, xCount)

	// An unpaginated request reports the form's zero-value count and offset.
	assert.Equal(t, "0", result.Header.Get("X-Limit"))
	assert.Equal(t, "0", result.Header.Get("X-Offset"))
}

func TestGetFoldersImport(t *testing.T) {
	t.Run("Flat", func(t *testing.T) {
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
	t.Run("Recursive", func(t *testing.T) {
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
