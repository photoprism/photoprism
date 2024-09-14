package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestCancelImport(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
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
