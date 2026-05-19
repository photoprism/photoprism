package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
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
		result := entity.UnscopedDb().
			Model(&entity.File{}).
			Where("photo_uid = ?", file.photoUID).
			Updates(entity.Values{
				"file_root":    entity.RootOriginals,
				"file_name":    file.fileName,
				"file_hash":    file.fileHash,
				"file_missing": false,
				"deleted_at":   nil,
			})

		if result.Error != nil {
			t.Fatalf("reset fixture %s failed: %v", file.photoUID, result.Error)
		}

		if result.RowsAffected < 1 {
			t.Fatalf("reset fixture %s failed: no rows updated", file.photoUID)
		}
	}
}
