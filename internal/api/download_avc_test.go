package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/media/video"
)

func TestAvcDownloadable(t *testing.T) {
	original := func(values entity.File) *entity.File {
		values.FileRoot = entity.RootOriginals
		values.FileHash = "3cad9168fa6acc5c5c2965ddf6ec465ca42fd818"
		return &values
	}
	t.Run("Accepted", func(t *testing.T) {
		for name, f := range map[string]*entity.File{
			"Hevc":          original(entity.File{FileType: "mov", MediaType: "video", FileVideo: true, FileCodec: "hvc1"}),
			"NoCodec":       original(entity.File{FileType: "mp4", MediaType: "video", FileVideo: true}),
			"Live":          original(entity.File{FileType: "heic", MediaType: "live", FileVideo: true}),
			"Animated":      original(entity.File{FileType: "gif", MediaType: "animated"}),
			"Frames":        original(entity.File{FileType: "png", MediaType: "image", FileFrames: 2}),
			"Duration":      original(entity.File{FileType: "webp", MediaType: "image", FileDuration: 1}),
			"Insv":          original(entity.File{FileType: "insv", MediaType: "video", FileVideo: true, FileCodec: "avc1"}),
			"Lrv":           original(entity.File{FileType: "lrv", MediaType: "video", FileVideo: true, FileCodec: "avc1"}),
			"NoMediaType":   original(entity.File{FileType: "mkv", FileCodec: "vp09"}),
			"UnknownFormat": original(entity.File{FileType: "xyz", FileVideo: true}),
		} {
			assert.True(t, avcDownloadable(f), name)
		}
	})
	t.Run("Refused", func(t *testing.T) {
		for name, f := range map[string]*entity.File{
			"Nil":          nil,
			"Avc":          original(entity.File{FileType: "mp4", MediaType: "video", FileVideo: true, FileCodec: "avc1"}),
			"H264":         original(entity.File{FileType: "mov", MediaType: "video", FileVideo: true, FileCodec: "H.264"}),
			"Still":        original(entity.File{FileType: "jpg", MediaType: "image"}),
			"Raw":          original(entity.File{FileType: "raw", MediaType: "raw"}),
			"Audio":        original(entity.File{FileType: "mp3", MediaType: "audio"}),
			"MotionNoType": original(entity.File{FileType: "jpg", FileVideo: true}),
			"Sidecar":      {FileRoot: entity.RootSidecar, FileHash: "3cad9168fa6acc5c5c2965ddf6ec465ca42fd818", FileType: "mov", MediaType: "video", FileVideo: true},
			"SidecarFlag":  original(entity.File{FileType: "mov", MediaType: "video", FileVideo: true, FileSidecar: true}),
			"Missing":      original(entity.File{FileType: "mov", MediaType: "video", FileVideo: true, FileMissing: true}),
			"Error":        original(entity.File{FileType: "mov", MediaType: "video", FileVideo: true, FileError: "failed"}),
			"NoHash":       {FileRoot: entity.RootOriginals, FileType: "mov", MediaType: "video", FileVideo: true},
		} {
			assert.False(t, avcDownloadable(f), name)
		}
	})
}

func TestAvcDownloadName(t *testing.T) {
	t.Run("Original", func(t *testing.T) {
		f := &entity.File{FileType: "mov", FileName: "2024/clip.mov", OriginalName: "camera/Holiday.MOV"}
		assert.Equal(t, "Holiday.mp4", avcDownloadName(f, customize.DownloadNameOriginal))
		assert.Equal(t, "clip.mp4", avcDownloadName(f, customize.DownloadNameFile))
	})
	t.Run("NoExtension", func(t *testing.T) {
		f := &entity.File{FileType: "mov", FileName: "2024/clip"}
		assert.Equal(t, "clip.mp4", avcDownloadName(f, customize.DownloadNameFile))
	})
	t.Run("Unsafe", func(t *testing.T) {
		f := &entity.File{FileType: "gif", FileName: "2024/.gif"}
		assert.Equal(t, "download.mp4", avcDownloadName(f, customize.DownloadNameFile))
	})
}

func TestGetDownload_FormatAvc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	app, router, conf := NewApiTest()
	GetDownload(router)
	GetVideo(router)

	if !conf.FFmpegEnabled() {
		t.Skip("ffmpeg is not available")
	}

	options := *conf.Options()
	settings := *conf.Settings()
	t.Cleanup(func() {
		*conf.Options() = options
		*conf.Settings() = settings
		conf.Propagate()
	})
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().TempPath = t.TempDir()
	conf.Options().CachePath = t.TempDir()
	conf.Propagate()
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(&entity.Folder{}, "path LIKE ?", "avc-export%") })

	ind := photoprism.NewIndex(conf, photoprism.NewConvert(conf), photoprism.NewFiles(), photoprism.NewPhotos())

	// index copies a sample to the originals folder, indexes it, and returns its file row.
	index := func(t *testing.T, sample, name string) *entity.File {
		data, err := os.ReadFile(sample) //nolint:gosec // Test reads a controlled fixture.
		require.NoError(t, err)
		fileName := filepath.Join(conf.OriginalsPath(), name)
		require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
		require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile)) //nolint:gosec // Test writes to its own temporary folder.
		m, err := photoprism.NewMediaFile(fileName)
		require.NoError(t, err)
		related, err := m.RelatedFiles(false)
		require.NoError(t, err)
		result := photoprism.IndexRelated(related, ind, photoprism.IndexOptionsNone(conf))
		require.True(t, result.Success(), "%+v", result)
		f := &entity.File{}
		require.NoError(t, entity.Db().Where("file_root = ? AND file_name = ?", entity.RootOriginals, name).Preload("Photo").First(f).Error)
		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "photo_uid = ?", f.PhotoUID)
			entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", f.PhotoID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", f.PhotoID)
			entity.UnscopedDb().Unscoped().Delete(&entity.PhotoKeyword{}, "photo_id = ?", f.PhotoID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", f.PhotoID)
		})
		return f
	}

	// get requests a download with the coarse instance token and the given query.
	get := func(hash, query string) *http.Response {
		w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+hash+"?t="+conf.DownloadToken()+query)
		return w.Result()
	}

	// play requests a video stream with the preview token.
	play := func(hash string) *httptest.ResponseRecorder {
		return PerformRequest(app, http.MethodGet, "/api/v1/videos/"+hash+"/"+conf.PreviewToken()+"/avc")
	}

	// probeBody writes a response body to a file and returns its video probe.
	probeBody := func(t *testing.T, body []byte) video.Info {
		fileName := filepath.Join(t.TempDir(), "body.mp4")
		require.NoError(t, os.WriteFile(fileName, body, fs.ModeFile))
		info, err := video.ProbeFile(fileName)
		require.NoError(t, err)
		return info
	}

	unknown := PerformRequest(app, http.MethodGet, "/api/v1/dl/3cad9168fa6acc5c5c2965ddf6ec465ca42fd000?t="+conf.DownloadToken()+"&format=avc")
	require.Equal(t, http.StatusNotFound, unknown.Code)

	clip := index(t, "../../pkg/media/video/testdata/quicktime-hvc1.mov", "avc-export/clip.mov")
	clipData, err := os.ReadFile(filepath.Join(conf.OriginalsPath(), clip.FileName))
	require.NoError(t, err)

	t.Run("WithoutFormat", func(t *testing.T) {
		w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+clip.FileHash+"?t="+conf.DownloadToken()+"&name=original")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, clipData, w.Body.Bytes())
		assert.Equal(t, `attachment; filename="clip.mov"`, w.Header().Get("Content-Disposition"))
	})
	t.Run("UnknownFormat", func(t *testing.T) {
		for _, format := range []string{"xyz", "AVC", "mp4"} {
			w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+clip.FileHash+"?t="+conf.DownloadToken()+"&format="+format)
			assert.Equal(t, http.StatusNotFound, w.Code, format)
			assert.Equal(t, unknown.Body.Bytes(), w.Body.Bytes(), format)
			assert.Empty(t, w.Header().Get("Content-Disposition"), format)
		}
	})
	t.Run("InsufficientStorage", func(t *testing.T) {
		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(conf.StoragePath(), 1*disk.MB, 1000*disk.MB)
		storageCheck := config.DisableStorageCheck.Load()
		t.Cleanup(func() { config.DisableStorageCheck.Store(storageCheck) })
		config.DisableStorageCheck.Store(false)
		w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+clip.FileHash+"?t="+conf.DownloadToken()+"&format=avc")
		assert.Equal(t, http.StatusInsufficientStorage, w.Code)
		assert.Empty(t, w.Header().Get("Content-Disposition"))
		assert.Equal(t, http.StatusInsufficientStorage, play(clip.FileHash).Code)
	})
	t.Run("Video", func(t *testing.T) {
		w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+clip.FileHash+"?t="+conf.DownloadToken()+"&name=original&format=avc")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, `attachment; filename="clip.mp4"`, w.Header().Get("Content-Disposition"))
		assert.Contains(t, w.Header().Get("Content-Type"), "video/mp4")
		assert.Empty(t, w.Header().Get("Cache-Control"))
		assert.Equal(t, video.CodecAvc1, probeBody(t, w.Body.Bytes()).VideoCodec)
	})
	t.Run("Playback", func(t *testing.T) {
		w := play(clip.FileHash)
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("Cache-Control"))
		assert.Equal(t, header.ContentTypeMp4AvcMain, w.Header().Get("Content-Type"))
		assert.Equal(t, video.CodecAvc1, probeBody(t, w.Body.Bytes()).VideoCodec)
	})
	t.Run("PlaybackMissing", func(t *testing.T) {
		f := index(t, "../../pkg/media/video/testdata/mp42-hvc1.mp4", "avc-export/missing.mp4")
		require.NoError(t, os.Remove(filepath.Join(conf.OriginalsPath(), f.FileName)))
		w := play(f.FileHash)
		assert.Equal(t, brokenVideo, w.Body.Bytes())
		missing := &entity.File{}
		require.NoError(t, entity.UnscopedDb().First(missing, f.ID).Error)
		assert.True(t, missing.FileMissing)
	})
	t.Run("Avc", func(t *testing.T) {
		f := index(t, "../../pkg/media/video/testdata/isom-avc1.mp4", "avc-export/avc.mp4")
		w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+f.FileHash+"?t="+conf.DownloadToken()+"&format=avc")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, unknown.Body.Bytes(), w.Body.Bytes())
		assert.Equal(t, http.StatusOK, get(f.FileHash, "").StatusCode)
		original, err := os.ReadFile(filepath.Join(conf.OriginalsPath(), f.FileName))
		require.NoError(t, err)
		w = play(f.FileHash)
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, original, w.Body.Bytes())
		bitrate := conf.Options().FFmpegBitrate
		t.Cleanup(func() { conf.Options().FFmpegBitrate = bitrate })
		conf.Options().FFmpegBitrate = 1
		require.NoError(t, f.Update("FileDuration", time.Millisecond))
		w = play(f.FileHash)
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotEqual(t, original, w.Body.Bytes())
		assert.Equal(t, header.ContentTypeMp4AvcMain, w.Header().Get("Content-Type"))
	})
	t.Run("Live", func(t *testing.T) {
		f := index(t, "../../assets/samples/samsung-motion-photo.jpg", "avc-export/live.jpg")
		require.Equal(t, entity.MediaLive, f.MediaType)
		info, err := video.ProbeFile(filepath.Join(conf.OriginalsPath(), f.FileName))
		require.NoError(t, err)
		w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+f.FileHash+"?t="+conf.DownloadToken()+"&name=file&format=avc")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="live.mp4"`, w.Header().Get("Content-Disposition"))
		assert.Equal(t, info.VideoSize(), int64(w.Body.Len()))
		assert.Equal(t, video.CodecAvc1, probeBody(t, w.Body.Bytes()).VideoCodec)
		stream := play(f.FileHash)
		require.Equal(t, http.StatusOK, stream.Code)
		assert.Equal(t, info.VideoContentType(), stream.Header().Get("Content-Type"))
		assert.NotEmpty(t, stream.Header().Get("Cache-Control"))
		assert.Equal(t, w.Body.Bytes(), stream.Body.Bytes())
	})
	t.Run("Animated", func(t *testing.T) {
		f := index(t, "../../assets/samples/pythagoras.gif", "avc-export/animated.gif")
		w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+f.FileHash+"?t="+conf.DownloadToken()+"&name=file&format=avc")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="animated.mp4"`, w.Header().Get("Content-Disposition"))
		assert.Equal(t, video.CodecAvc1, probeBody(t, w.Body.Bytes()).VideoCodec)
	})
	t.Run("Insta360", func(t *testing.T) {
		f := index(t, "../../internal/photoprism/testdata/insta360.insv", "avc-export/VID_20220625_140410_00_008.insv")
		w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+f.FileHash+"?t="+conf.DownloadToken()+"&name=file&format=avc")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, unknown.Body.Bytes(), w.Body.Bytes())
		stitched, err := os.ReadFile("../../pkg/media/video/testdata/isom-avc1.mp4")
		require.NoError(t, err)
		avcName := filepath.Join(conf.SidecarPath(), "avc-export", "VID_20220625_140410_00_008.insv.avc")
		require.NoError(t, fs.MkdirAll(filepath.Dir(avcName)))
		require.NoError(t, os.WriteFile(avcName, stitched, fs.ModeFile)) //nolint:gosec // Test writes to its own temporary folder.
		w = PerformRequest(app, http.MethodGet, "/api/v1/dl/"+f.FileHash+"?t="+conf.DownloadToken()+"&name=file&format=avc")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="VID_20220625_140410_00_008.mp4"`, w.Header().Get("Content-Disposition"))
		assert.Equal(t, stitched, w.Body.Bytes())
	})
	t.Run("SharedHash", func(t *testing.T) {
		// The visitor's album holds the picture of the second file with this hash.
		const albumUID = "as6sg6bxpogaaba8"
		mode := conf.AuthMode()
		conf.SetAuthMode(config.AuthModePasswd)
		conf.Propagate()
		t.Cleanup(func() { conf.SetAuthMode(mode); conf.Propagate() })
		hidden := entity.NewPhoto(false)
		require.NoError(t, hidden.Create())
		t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", hidden.ID) })
		visible := *clip
		visible.ID = 0
		visible.FileUID = ""
		visible.FileName = "avc-export/clip-copy.mov"
		visible.Photo = nil
		require.NoError(t, entity.Db().Model(clip).UpdateColumns(entity.Values{"photo_id": hidden.ID, "photo_uid": hidden.PhotoUID}).Error)
		require.NoError(t, visible.Create())
		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "id = ?", visible.ID)
			entity.Db().Model(clip).UpdateColumns(entity.Values{"photo_id": visible.PhotoID, "photo_uid": visible.PhotoUID})
		})
		require.NoError(t, entity.NewPhotoAlbum(visible.PhotoUID, albumUID).Create())
		visitorToken := tokens.DownloadToken(entity.SessionFixtures.Get("visitor").ID)
		require.Equal(t, http.StatusOK, PerformRequest(app, http.MethodGet, "/api/v1/dl/"+clip.FileHash+"?t="+visitorToken).Code)
		w := PerformRequest(app, http.MethodGet, "/api/v1/dl/"+clip.FileHash+"?t="+visitorToken+"&format=avc")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, unknown.Body.Bytes(), w.Body.Bytes())
	})
}
