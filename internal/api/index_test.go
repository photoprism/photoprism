package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestCancelIndex(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CancelIndexing(router)
		r := PerformRequest(app, "DELETE", "/api/v1/index")

		var resp i18n.Response

		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}

		assert.True(t, resp.Success())
		assert.Equal(t, i18n.Msg(i18n.MsgIndexingCanceled), resp.Msg)
		assert.Equal(t, i18n.Msg(i18n.MsgIndexingCanceled), resp.String())
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestStartIndexing(t *testing.T) {
	t.Run("InsufficientStorage", func(t *testing.T) {
		app, router, conf := NewApiTest()

		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(conf.StoragePath(), 1*disk.MB, 1000*disk.MB)
		// Reset the storage-check latch, which an earlier test may have tripped by probing a
		// temp path that duf cannot resolve to a mount point.
		config.DisableStorageCheck.Store(false)

		StartIndexing(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/index", "{}")

		assert.Equal(t, http.StatusInsufficientStorage, r.Code)
	})
}
