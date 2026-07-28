package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/log/status"
)

func TestZip(t *testing.T) {
	app, router, conf := NewApiTest()
	ZipCreate(router)
	ZipDownload(router)

	originalOptions := *conf.Options()

	t.Cleanup(func() {
		*conf.Options() = originalOptions
	})

	// Isolate ZIP output from shared singleton config mutations in other tests.
	conf.Options().TempPath = t.TempDir()

	t.Run("Download", func(t *testing.T) {
		resetZipDownloadFixtures(t)

		r := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": ["ps6sg6be2lvl0y12", "ps6sg6be2lvl0y11"]}`)
		message := gjson.Get(r.Body.String(), "message")
		assert.Contains(t, message.String(), "Zip created")
		assert.Equal(t, http.StatusOK, r.Code)
		filename := gjson.Get(r.Body.String(), "filename")
		response := PerformRequest(app, "GET", "/api/v1/zip/"+filename.String()+"?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("ErrNoItemsSelected", func(t *testing.T) {
		response := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": []}`)
		val := gjson.Get(response.Body.String(), "error")
		assert.Equal(t, "No items selected", val.String())
		assert.Equal(t, http.StatusBadRequest, response.Code)
	})
	t.Run("ErrBadRequest", func(t *testing.T) {
		response := PerformRequestWithBody(app, "POST", "/api/v1/zip", `{"photos": [123, "ps6sg6be2lvl0yxx"]}`)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	})
	t.Run("ErrNotFound", func(t *testing.T) {
		response := PerformRequest(app, "GET", "/api/v1/zip/xxx?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, response.Code)
	})
}

// resetZipDownloadFixtures restores file rows used by TestZip/Download, making
// test independent of any previous tests that may have marked them as missing.
func resetZipDownloadFixtures(t *testing.T) {
	t.Helper()

	reset := []struct {
		photoUID string
		fileName string
		fileHash string
	}{
		{
			photoUID: "ps6sg6be2lvl0y11",
			fileName: "Germany/bridge.jpg",
			fileHash: "pcad9168fa6acc5c5c2965ddf6ec465ca42fd818",
		},
		{
			photoUID: "ps6sg6be2lvl0y12",
			fileName: "2015/11/20151101_000000_51C501B5.jpg",
			fileHash: "acad9168fa6acc5c5c2965ddf6ec465ca42fd818",
		},
	}

	for _, file := range reset {
		if err := entity.UnscopedDb().
			Model(&entity.File{}).
			Where("photo_uid = ?", file.photoUID).
			Updates(entity.Values{
				"file_root":    entity.RootOriginals,
				"file_name":    file.fileName,
				"file_hash":    file.fileHash,
				"file_missing": false,
				"deleted_at":   nil,
			}).Error; err != nil {
			t.Fatalf("reset fixture %s failed: %v", file.photoUID, err)
		}

		// The row count is verified separately, as MySQL reports the number of
		// rows an UPDATE changed while SQLite reports the number it matched.
		var found int64

		if err := entity.UnscopedDb().
			Model(&entity.File{}).
			Where("photo_uid = ? AND file_name = ? AND file_missing = 0", file.photoUID, file.fileName).
			Count(&found).Error; err != nil {
			t.Fatalf("reset fixture %s failed: %v", file.photoUID, err)
		} else if found < 1 {
			t.Fatalf("reset fixture %s failed: no rows updated", file.photoUID)
		}
	}
}

func TestAuditArchiveAccess(t *testing.T) {
	orig := event.AuditLog
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.AuditLog = logger

	t.Cleanup(func() {
		event.AuditLog = orig
	})

	// newTestContext returns a gin context backed by a request, as ClientIP requires one.
	newTestContext := func() *gin.Context {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/zip/photoprism-download-20260727-094439-zihqtuw4.zip", nil)
		return c
	}
	t.Run("WithSession", func(t *testing.T) {
		hook.Reset()
		auditArchiveAccess(newTestContext(), &entity.Session{RefID: "sessxkkcabcd"}, "download %s", status.Succeeded, "photoprism-download-20260727-094439-zihqtuw4.zip")
		entries := hook.AllEntries()

		if len(entries) != 1 {
			t.Fatalf("expected 1 audit entry, got %d", len(entries))
		}

		msg := entries[0].Message
		assert.Contains(t, msg, "sessxkkcabcd")
		assert.Contains(t, msg, "photoprism-download-20260727-094439-***.zip")
		assert.Contains(t, msg, status.Succeeded)
		assert.NotContains(t, msg, "zihqtuw4")
	})
	t.Run("WithoutSession", func(t *testing.T) {
		hook.Reset()
		auditArchiveAccess(newTestContext(), nil, "download %s", status.NotFound, "photoprism-download-20260727-094439-zihqtuw4.zip")
		entries := hook.AllEntries()

		if len(entries) != 1 {
			t.Fatalf("expected 1 audit entry, got %d", len(entries))
		}

		msg := entries[0].Message
		assert.Contains(t, msg, "photoprism-download-20260727-094439-***.zip")
		assert.Contains(t, msg, status.NotFound)
		assert.NotContains(t, msg, "session")
		assert.NotContains(t, msg, "zihqtuw4")
	})
}
