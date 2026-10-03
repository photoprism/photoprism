package api

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
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
		var found int

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

// videoStillTestPhoto creates a video photo whose originals, generated still and sidecar exist on disk.
// The download settings are restored and the file roots point to temporary folders.
func videoStillTestPhoto(t *testing.T, conf *config.Config) *entity.Photo {
	t.Helper()

	options, settings := *conf.Options(), *conf.Settings()
	t.Cleanup(func() {
		*conf.Options() = options
		*conf.Settings() = settings
		conf.Propagate()
	})
	conf.Options().TempPath = t.TempDir()
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Propagate()

	photo := entity.NewPhoto(false)
	photo.PhotoType = media.Video.String()
	require.NoError(t, photo.Save())
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&photo)
	})

	for _, f := range []entity.File{
		{FileName: "still/clip.mp4", FileRoot: entity.RootOriginals, FileType: "mp4", MediaType: "video", FileVideo: true},
		{FileName: "still/clip.jpg", FileRoot: entity.RootOriginals, FileType: "jpg", MediaType: "image"},
		{FileName: "still/clip.xmp", FileRoot: entity.RootOriginals, FileType: "xmp", MediaType: "sidecar", FileSidecar: true},
		{FileName: "still/clip.mp4.jpg", FileRoot: entity.RootSidecar, FileType: "jpg", MediaType: "image"},
	} {
		f.PhotoID, f.PhotoUID, f.FileHash = photo.ID, photo.PhotoUID, rnd.Base36(40)
		require.NoError(t, f.Create())
		CreateTestOriginal(t, &f)
	}

	return &photo
}

// zipEntryNames returns the names of the files in a zip archive.
func zipEntryNames(t *testing.T, body []byte) (names []string) {
	t.Helper()

	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	require.NoError(t, err)

	for _, f := range archive.File {
		names = append(names, f.Name)
	}

	return names
}

// TestZipCreate_VideoStills checks that selection archives skip the generated still images of videos.
func TestZipCreate_VideoStills(t *testing.T) {
	app, router, conf := NewApiTest()
	ZipCreate(router)
	ZipDownload(router)

	photo := videoStillTestPhoto(t, conf)

	archive := func(t *testing.T) []string {
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/zip", fmt.Sprintf(`{"photos": [%q]}`, photo.PhotoUID))
		require.Equal(t, http.StatusOK, r.Code)
		filename := gjson.Get(r.Body.String(), "filename").String()
		response := PerformRequest(app, http.MethodGet, "/api/v1/zip/"+filename+"?t="+conf.DownloadToken())
		require.Equal(t, http.StatusOK, response.Code)
		return zipEntryNames(t, response.Body.Bytes())
	}

	conf.Settings().Download.Name = "file"
	conf.Settings().Download.Originals = false
	conf.Settings().Download.MediaRaw = true

	t.Run("SidecarsOn", func(t *testing.T) {
		conf.Settings().Download.MediaSidecar = true
		assert.ElementsMatch(t, []string{"clip.mp4", "clip.jpg", "clip.xmp"}, archive(t))
	})
	t.Run("SidecarsOff", func(t *testing.T) {
		conf.Settings().Download.MediaSidecar = false
		assert.ElementsMatch(t, []string{"clip.mp4", "clip.jpg"}, archive(t))
	})
}
