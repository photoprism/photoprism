package api

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestGetVideoActions verifies real preparation and attachment delivery through the routed handler.
func TestGetVideoActions(t *testing.T) {
	app, router, conf := NewApiTest()
	if conf.FFmpegBin() == "" {
		t.Skip("ffmpeg unavailable")
	}
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options; conf.Propagate() })
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().CachePath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().FFmpegEncoder = "libx264"
	GetVideo(router)
	f := videoDownloadFixture(t, conf, "export-hevc.mp4", "libx265", 64)
	base := "/api/v1/videos/" + f.FileHash + "/" + conf.DownloadToken() + "/avc?action="
	t.Run("ReadOnlyProbe", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, base+"probe")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		var probe photoprism.ExportProbe
		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &probe))
		require.NotNil(t, probe.Avc)
		assert.True(t, probe.Avc.Preparable)
		assert.False(t, probe.Avc.Ready)
		assert.Equal(t, "no-store", r.Header().Get("Cache-Control"))
		entries, err := os.ReadDir(conf.CachePath())
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
	t.Run("DownloadDoesNotPrepare", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, base+"download")
		assert.Equal(t, http.StatusNotFound, r.Code)
		assert.Empty(t, r.Header().Get("Content-Disposition"))
		entries, err := os.ReadDir(conf.CachePath())
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
	t.Run("Prepare", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, base+"prepare")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		var output photoprism.ExportOutput
		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &output))
		assert.True(t, output.Ready)
		assert.Equal(t, "avc1", output.Codec)
		assert.Equal(t, "mp4", output.Format)
		assert.NotContains(t, r.Body.String(), conf.CachePath())
	})
	t.Run("Attachment", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, base+"download&filename="+url.QueryEscape("旅行.mov"))
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		_, params, err := mime.ParseMediaType(r.Header().Get("Content-Disposition"))
		require.NoError(t, err)
		assert.Equal(t, "旅行.mp4", params["filename"])
		assert.Equal(t, "video/mp4", r.Header().Get("Content-Type"))
		assert.Greater(t, r.Body.Len(), 0)
		assert.Equal(t, "no-store", r.Header().Get("Cache-Control"))
	})
	t.Run("SanitizedFallback", func(t *testing.T) {
		original := f.OriginalName
		f.OriginalName = "clip:name.mp4"
		require.NoError(t, entity.Db().Model(f).UpdateColumn("original_name", f.OriginalName).Error)
		defer func() { _ = entity.Db().Model(f).UpdateColumn("original_name", original).Error }()
		r := PerformRequest(app, http.MethodGet, base+"download&name=original")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		_, params, err := mime.ParseMediaType(r.Header().Get("Content-Disposition"))
		require.NoError(t, err)
		assert.Equal(t, "clipname_avc.mp4", params["filename"])
	})
	t.Run("InvalidOperation", func(t *testing.T) {
		for _, op := range []string{"invalid", "../prepare"} {
			r := PerformRequest(app, http.MethodGet, base+op)
			assert.Equal(t, http.StatusBadRequest, r.Code)
		}
	})
	t.Run("InvalidActionFormat", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.DownloadToken()+"/mp4?action=prepare")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RemovedExportFormat", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.PreviewToken()+"/export?export=probe")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("PairingKeyRedaction", func(t *testing.T) {
		dir, err := fs.CachePath(conf.CachePath(), f.FileHash, "json", true)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, f.FileHash+"_exiftool.json"), []byte(`[{"ContentIdentifier":"export-pair"}]`), fs.ModeFile))
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		visitor := entity.SessionFixtures.Get("visitor")
		admin := entity.SessionFixtures.Get("alice_app_password_full_access")
		relation := entity.NewPhotoAlbum(f.PhotoUID, "as6sg6bxpogaaba8")
		var existing entity.PhotoAlbum
		if err := entity.Db().Where("photo_uid = ? AND album_uid = ?", relation.PhotoUID, relation.AlbumUID).First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			require.NoError(t, relation.Create())
			defer entity.Db().Delete(relation)
		} else {
			require.NoError(t, err)
		}
		for _, tc := range []struct{ name, token, key string }{
			{"FullDetail", tokens.SignDownload(admin.ID), "export-pair"},
			{"SharedDetail", tokens.SignDownload(visitor.ID), ""},
		} {
			r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+tc.token+"/avc?action=probe")
			require.Equal(t, http.StatusOK, r.Code, tc.name+": "+r.Body.String())
			var probe photoprism.ExportProbe
			require.NoError(t, json.Unmarshal(r.Body.Bytes(), &probe))
			assert.Equal(t, tc.key, probe.AppleKey, tc.name)
		}
	})
}

// TestGetVideoActionsEmbedded verifies native attachment bytes and ranges without cached extraction.
func TestGetVideoActionsEmbedded(t *testing.T) {
	app, router, conf := NewApiTest()
	if conf.FFmpegBin() == "" {
		t.Skip("ffmpeg unavailable")
	}
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options; conf.Propagate() })
	conf.Options().OriginalsPath, conf.Options().SidecarPath, conf.Options().CachePath = t.TempDir(), t.TempDir(), t.TempDir()
	GetVideo(router)
	image, err := os.ReadFile("../photoprism/testdata/flash.jpg")
	require.NoError(t, err)
	for _, format := range []string{"mp4", "mov"} {
		t.Run(strings.ToUpper(format), func(t *testing.T) {
			clip := filepath.Join(t.TempDir(), "native."+format)
			data, err := exec.Command(conf.FFmpegBin(), "-v", "error", "-f", "lavfi", "-i", "testsrc2=size=64x48:rate=5", "-t", "0.4", "-threads", "1", "-c:v", "libx264", "-y", clip).CombinedOutput() //nolint:gosec // paths and executable are disposable test fixtures
			require.NoError(t, err, "%s", data)
			video, err := os.ReadFile(clip) //nolint:gosec // clip is inside t.TempDir
			require.NoError(t, err)
			f := videoDownloadFixture(t, conf, "motion-db-"+format+".mp4", "libx264", 64)
			f.FileName, f.FileType, f.FileVideo, f.MediaType = "motion-"+format+".jpg", "jpg", false, entity.MediaLive
			name := filepath.Join(conf.OriginalsPath(), f.FileName)
			require.NoError(t, os.WriteFile(name, append(append([]byte{}, image...), video...), fs.ModeFile)) //nolint:gosec // source path is confined to disposable originals
			f.FileHash = fs.Hash(name)
			require.NoError(t, f.Save())
			base := "/api/v1/videos/" + f.FileHash + "/" + conf.DownloadToken() + "/native?action="
			probeResponse := PerformRequest(app, http.MethodGet, base+"probe")
			require.Equal(t, http.StatusOK, probeResponse.Code, probeResponse.Body.String())
			var probe photoprism.ExportProbe
			require.NoError(t, json.Unmarshal(probeResponse.Body.Bytes(), &probe))
			require.NotNil(t, probe.VideoPart)
			assert.True(t, probe.VideoPart.Ready)
			assert.Equal(t, format, probe.VideoPart.Format)
			assert.Equal(t, int64(len(video)), probe.VideoPart.Bytes)
			prepare := PerformRequest(app, http.MethodGet, base+"prepare")
			require.Equal(t, http.StatusOK, prepare.Code, prepare.Body.String())
			url := base + "download&filename=" + url.QueryEscape("旅行.avi")
			attachment := PerformRequest(app, http.MethodGet, url)
			require.Equal(t, http.StatusOK, attachment.Code)
			assert.Equal(t, video, attachment.Body.Bytes())
			_, params, err := mime.ParseMediaType(attachment.Header().Get("Content-Disposition"))
			require.NoError(t, err)
			assert.Equal(t, "旅行."+format, params["filename"])
			request := httptest.NewRequest(http.MethodGet, url, nil)
			request.Header.Set("Range", "bytes=0-31")
			partial := httptest.NewRecorder()
			app.ServeHTTP(partial, request)
			require.Equal(t, http.StatusPartialContent, partial.Code)
			assert.Equal(t, video[:32], partial.Body.Bytes())
			assert.Equal(t, "no-store", partial.Header().Get("Cache-Control"))
			request = httptest.NewRequest(http.MethodGet, url, nil)
			request.Header.Set("Range", "bytes="+strconv.Itoa(len(video))+"-")
			unsatisfiable := httptest.NewRecorder()
			app.ServeHTTP(unsatisfiable, request)
			assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, unsatisfiable.Code)
			assert.Equal(t, "bytes */"+strconv.Itoa(len(video)), unsatisfiable.Header().Get("Content-Range"))
			assert.Equal(t, "no-store", unsatisfiable.Header().Get("Cache-Control"))
			request = httptest.NewRequest(http.MethodGet, url, nil)
			request.Header.Set("Range", "bytes=-16")
			suffix := httptest.NewRecorder()
			app.ServeHTTP(suffix, request)
			require.Equal(t, http.StatusPartialContent, suffix.Code)
			assert.Equal(t, video[len(video)-16:], suffix.Body.Bytes())
			for _, path := range []string{conf.SidecarPath(), conf.CachePath()} {
				entries, err := os.ReadDir(path)
				require.NoError(t, err)
				assert.Empty(t, entries)
			}
		})
	}
}

// TestGetVideoActionsAuthorization checks path credentials, scope, and uniform hidden-source errors.
func TestGetVideoActionsAuthorization(t *testing.T) {
	app, router, conf := NewApiTest()
	router.Use(func(c *gin.Context) { c.Query("t"); DownloadSession(c); c.Next() })
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
	GetVideo(router)
	admin := entity.SessionFixtures.Get("alice_app_password_full_access")
	refused := entity.SessionFixtures.Get("alice_app_password_shares")
	const hash = "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"
	t.Run("PathCredentialIsAuthoritative", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+hash+"/invalid/avc?action=probe&t="+tokens.SignDownload(admin.ID))
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("PreviewTokenDoesNotAuthorize", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+hash+"/"+conf.PreviewToken()+"/avc?action=probe")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("RefusedScope", func(t *testing.T) {
		for _, mode := range []string{"probe", "prepare", "download"} {
			r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+hash+"/"+tokens.SignDownload(refused.ID)+"/avc?action="+mode)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
	t.Run("MissingAndUnknownAreIdentical", func(t *testing.T) {
		token := tokens.SignDownload(admin.ID)
		missing := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+hash+"/"+token+"/avc?action=probe")
		unknown := PerformRequest(app, http.MethodGet, "/api/v1/videos/1234567890123456789012345678901234567890/"+token+"/avc?action=probe")
		assert.Equal(t, http.StatusNotFound, missing.Code)
		assert.Equal(t, missing.Code, unknown.Code)
		assert.Equal(t, missing.Body.String(), unknown.Body.String())
	})
	t.Run("HiddenAndUnknownAreIdentical", func(t *testing.T) {
		visitor := entity.SessionFixtures.Get("visitor")
		file := entity.FileFixtures.Get("Photo18.jpg")
		require.NotEmpty(t, file.FileHash)
		visible, err := search.FileDownloadable(file.FileHash, &visitor)
		require.NoError(t, err)
		require.False(t, visible)
		token := tokens.SignDownload(visitor.ID)
		for _, mode := range []string{"probe", "prepare", "download"} {
			hidden := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+file.FileHash+"/"+token+"/avc?action="+mode)
			unknown := PerformRequest(app, http.MethodGet, "/api/v1/videos/1234567890123456789012345678901234567890/"+token+"/avc?action="+mode)
			assert.Equal(t, http.StatusNotFound, hidden.Code)
			assert.Equal(t, hidden.Code, unknown.Code)
			assert.Equal(t, hidden.Body.String(), unknown.Body.String())
		}
	})
}
