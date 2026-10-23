package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// videoSourceTestConfig points the test config to temporary originals and sidecar folders.
func videoSourceTestConfig(t *testing.T) *config.Config {
	_, _, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() {
		*conf.Options() = options
		conf.Propagate()
	})
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().CachePath = t.TempDir()
	conf.Propagate()
	return conf
}

// copyVideoSample copies a test sample to the originals folder and returns its file name.
func copyVideoSample(t *testing.T, conf *config.Config, sample, name string) string {
	data, err := os.ReadFile(sample) //nolint:gosec // Test reads a controlled fixture.
	require.NoError(t, err)
	fileName := filepath.Join(conf.OriginalsPath(), name)
	require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile)) //nolint:gosec // Test writes to its own temporary folder.
	return fileName
}

func TestNewVideoSource(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		f := &entity.File{FileRoot: entity.RootOriginals, FileName: "2024/clip.mov", FileType: "mov", FileVideo: true, FileCodec: "hvc1", FileMime: "video/quicktime"}
		src := newVideoSource(f)
		assert.True(t, filepath.IsAbs(src.FileName))
		assert.Equal(t, "clip.mov", filepath.Base(src.FileName))
		assert.Equal(t, fs.VideoMov, src.FileType)
		assert.Equal(t, f.ContentType(), src.ContentType)
	})
	t.Run("NoVideo", func(t *testing.T) {
		f := &entity.File{FileRoot: entity.RootOriginals, FileName: "still.jpg", FileType: "jpg", FileMime: "image/jpeg"}
		assert.Equal(t, "image/jpeg", newVideoSource(f).ContentType)
	})
}

func TestOpenEmbeddedVideo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		f := &entity.File{FileName: "image-isom-avc1.jpg"}
		info, reader, err := openEmbeddedVideo(f, "../../pkg/media/video/testdata/image-isom-avc1.jpg")
		require.NoError(t, err)
		require.NotNil(t, reader)
		defer reader.Close()
		assert.Equal(t, video.CodecAvc1, info.VideoCodec)
		assert.Positive(t, info.VideoOffset)
	})
	t.Run("NoVideo", func(t *testing.T) {
		f := &entity.File{FileName: "image.jpg"}
		_, reader, err := openEmbeddedVideo(f, "../../pkg/media/video/testdata/image.jpg")
		assert.ErrorIs(t, err, errVideoUnavailable)
		assert.Nil(t, reader)
	})
}

func TestCacheEmbeddedVideo(t *testing.T) {
	conf := videoSourceTestConfig(t)
	sample := "../../pkg/media/video/testdata/image-isom-avc1.jpg"
	t.Run("Success", func(t *testing.T) {
		f := &entity.File{FileName: "image-isom-avc1.jpg", FileType: "jpg", FileHash: "3cad9168fa6acc5c5c2965ddf6ec465ca42fd901"}
		info, reader, err := openEmbeddedVideo(f, sample)
		require.NoError(t, err)
		defer reader.Close()
		src, err := cacheEmbeddedVideo(conf, f, info, reader)
		require.NoError(t, err)
		assert.Equal(t, info.VideoFileType(), src.FileType)
		assert.Equal(t, info.VideoContentType(), src.ContentType)
		data, err := os.ReadFile(src.FileName)
		require.NoError(t, err)
		assert.Equal(t, info.VideoSize(), int64(len(data)))
	})
	t.Run("Unwritable", func(t *testing.T) {
		f := &entity.File{FileName: "image-isom-avc1.jpg", FileType: "jpg", FileHash: "3cad9168fa6acc5c5c2965ddf6ec465ca42fd902"}
		info, reader, err := openEmbeddedVideo(f, sample)
		require.NoError(t, err)
		defer reader.Close()
		require.NoError(t, fs.MkdirAll(filepath.Join(conf.MediaFileCachePath(f.FileHash), f.FileHash+info.VideoFileExt())))
		_, err = cacheEmbeddedVideo(conf, f, info, reader)
		assert.ErrorIs(t, err, errVideoUnavailable)
	})
}

func TestResolveVideo(t *testing.T) {
	conf := videoSourceTestConfig(t)
	avcName := copyVideoSample(t, conf, "../../pkg/media/video/testdata/isom-avc1.mp4", "avc.mp4")
	avc := &entity.File{FileRoot: entity.RootOriginals, FileName: "avc.mp4", FileType: "mp4", FileVideo: true, FileCodec: "avc1", FileMime: header.ContentTypeMp4}
	t.Run("Compatible", func(t *testing.T) {
		fileName, contentType, err := resolveVideo(conf, avc, newVideoSource(avc), video.Avc, true)
		require.NoError(t, err)
		assert.Equal(t, avcName, fileName)
		assert.Equal(t, avc.ContentType(), contentType)
	})
	t.Run("DownloadIgnoresBitrate", func(t *testing.T) {
		if !conf.FFmpegEnabled() {
			t.Skip("ffmpeg is not available")
		}
		src := newVideoSource(avc)
		src.Bitrate = 100000
		fileName, _, err := resolveVideo(conf, avc, src, video.Avc, true)
		require.NoError(t, err)
		assert.Equal(t, avcName, fileName)
		fileName, contentType, err := resolveVideo(conf, avc, src, video.Avc, false)
		require.NoError(t, err)
		assert.NotEqual(t, avcName, fileName)
		assert.Equal(t, header.ContentTypeMp4AvcMain, contentType)
	})
	t.Run("Missing", func(t *testing.T) {
		f := &entity.File{PhotoID: 1000000, FileRoot: entity.RootOriginals, FileName: "resolve-video/missing.mp4", FileType: "mp4", FileVideo: true}
		require.NoError(t, f.Create())
		t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(f) })
		_, _, err := resolveVideo(conf, f, newVideoSource(f), video.Avc, true)
		assert.ErrorIs(t, err, errVideoUnavailable)
		missing := &entity.File{}
		require.NoError(t, entity.UnscopedDb().First(missing, f.ID).Error)
		assert.True(t, missing.FileMissing)
	})
	t.Run("Insta360WithoutDerivative", func(t *testing.T) {
		copyVideoSample(t, conf, "../../pkg/media/video/testdata/isom-avc1.mp4", "LRV_20240415_213145_01_035.lrv")
		f := &entity.File{FileRoot: entity.RootOriginals, FileName: "LRV_20240415_213145_01_035.lrv", FileType: "lrv", FileVideo: true, FileCodec: "avc1"}
		_, _, err := resolveVideo(conf, f, newVideoSource(f), video.Avc, true)
		assert.ErrorIs(t, err, errVideoUnavailable)
		if conf.FFmpegEnabled() {
			fileName, _, err := resolveVideo(conf, f, newVideoSource(f), video.Avc, false)
			require.NoError(t, err)
			assert.NotEmpty(t, fileName)
		}
	})
	t.Run("Insta360Derivative", func(t *testing.T) {
		copyVideoSample(t, conf, "../../pkg/media/video/testdata/mp4v-avc1.mp4", "LRV_20240416_213145_01_036.lrv")
		f := &entity.File{FileRoot: entity.RootOriginals, FileName: "LRV_20240416_213145_01_036.lrv", FileType: "lrv", FileVideo: true, FileCodec: "avc1"}
		avcName := filepath.Join(conf.SidecarPath(), "LRV_20240416_213145_01_036.lrv.avc")
		data, err := os.ReadFile("../../pkg/media/video/testdata/isom-avc1.mp4")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(avcName, data, fs.ModeFile)) //nolint:gosec // Test writes to its own temporary folder.
		derivative, err := photoprism.NewMediaFile(avcName)
		require.NoError(t, err)
		fileName, contentType, err := resolveVideo(conf, f, newVideoSource(f), video.Avc, true)
		require.NoError(t, err)
		assert.Equal(t, avcName, fileName)
		assert.Equal(t, derivative.ContentType(), contentType)
	})
}

func TestInsta360Capture(t *testing.T) {
	t.Run("Capture", func(t *testing.T) {
		assert.True(t, insta360Capture(&entity.File{FileType: "insv"}))
		assert.True(t, insta360Capture(&entity.File{FileType: "lrv"}))
	})
	t.Run("Other", func(t *testing.T) {
		assert.False(t, insta360Capture(&entity.File{FileType: "mp4"}))
		assert.False(t, insta360Capture(&entity.File{FileType: "insp"}))
	})
}

func TestSendAvcDownload(t *testing.T) {
	conf := videoSourceTestConfig(t)
	copyVideoSample(t, conf, "../../pkg/media/video/testdata/image-isom-avc1.jpg", "live.jpg")
	live := func(photoUID string) *entity.File {
		return &entity.File{PhotoUID: photoUID, FileRoot: entity.RootOriginals, FileName: "live.jpg", FileType: "jpg", MediaType: entity.MediaLive, FileVideo: true, FileHash: "3cad9168fa6acc5c5c2965ddf6ec465ca42fd903"}
	}
	// send calls sendAvcDownload for f with a nil session, as for a coarse download token.
	send := func(f *entity.File) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/dl/"+f.FileHash+"?format=avc&name=file", nil)
		sendAvcDownload(c, f, nil)
		return w
	}
	t.Run("Success", func(t *testing.T) {
		photo := entity.NewPhoto(false)
		require.NoError(t, photo.Create())
		t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", photo.ID) })
		w := send(live(photo.PhotoUID))
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="live.mp4"`, w.Header().Get("Content-Disposition"))
		assert.Contains(t, w.Header().Get("Content-Type"), "video/mp4")
	})
	t.Run("NotVisible", func(t *testing.T) {
		w := send(live("ps6sg6be2lvl0y12"))
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, brokenIconSvg, w.Body.Bytes())
	})
	t.Run("Refused", func(t *testing.T) {
		f := live("ps6sg6be2lvl0yh8")
		f.MediaType = entity.MediaImage
		f.FileVideo = false
		w := send(f)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, brokenIconSvg, w.Body.Bytes())
	})
}
