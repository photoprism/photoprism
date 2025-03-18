package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestCancelImport(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CancelImport(router)
		r := PerformRequest(app, "DELETE", "/api/v1/import")

		var resp i18n.Response

		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}

		assert.True(t, resp.Success())
		assert.Equal(t, i18n.Msg(i18n.MsgImportCanceled), resp.Msg)
		assert.Equal(t, i18n.Msg(i18n.MsgImportCanceled), resp.String())
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestStartImport(t *testing.T) {
	t.Run("ReadOnlyMode", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().ReadOnly = true
		StartImport(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/import/test", "{foo:123}")

		assert.Equal(t, http.StatusForbidden, r.Code)
		config.Options().ReadOnly = false
	})
	t.Run("QuotaExceeded", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().FilesQuota = 1
		StartImport(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/import/test", "{foo:123}")

		assert.Equal(t, http.StatusInsufficientStorage, r.Code)
		config.Options().FilesQuota = 0
	})
}
