package api

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// videoDownloadFixture creates and indexes a small clip in a disposable library.
func videoDownloadFixture(t *testing.T, conf *config.Config, name, codec string, width int) *entity.File {
	t.Helper()
	args := []string{"-v", "error", "-f", "lavfi", "-i", "testsrc2=size=" + strconv.Itoa(width) + "x48:rate=5", "-t", "0.4"}
	args = append(args, "-threads", "1", "-c:v", codec)
	if codec == "libx265" {
		args = append(args, "-x265-params", "pools=1:frame-threads=1:log-level=error")
	}
	path := filepath.Join(conf.OriginalsPath(), name)
	args = append(args, path)
	output, err := exec.Command(conf.FFmpegBin(), args...).CombinedOutput() //nolint:gosec // executable and output belong to disposable test fixtures
	require.NoError(t, err, "%s", output)
	photo := entity.NewPhoto(false)
	photo.PhotoQuality = 3
	require.NoError(t, photo.Save())
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(photo)
	})
	f := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: name, FileHash: fs.Hash(path), FileType: filepath.Ext(name)[1:], FileVideo: true, MediaType: entity.MediaVideo, FileWidth: width, FileHeight: 48, FileDuration: 400 * time.Millisecond, FileCodec: "hvc1"}
	if codec == "libx264" {
		f.FileCodec = "avc1"
	}
	require.NoError(t, f.Create())
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(f) })
	return f
}

// TestVideoDownloadContracts compares playback downloads with exact-source exports using real media.
func TestVideoDownloadContracts(t *testing.T) {
	app, router, conf := NewApiTest()
	if conf.FFmpegBin() == "" {
		t.Skip("ffmpeg unavailable")
	}
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options; conf.Propagate() })
	originals := t.TempDir()
	conf.Options().OriginalsPath, conf.Options().SidecarPath, conf.Options().CachePath = originals, t.TempDir(), t.TempDir()
	conf.Options().FFmpegEncoder = "libx264"
	conf.Options().DisableExifTool = true
	GetVideo(router)
	t.Run("AvcOriginal", func(t *testing.T) {
		f := videoDownloadFixture(t, conf, "original.mov", "libx264", 64)
		r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.PreviewToken()+"/avc?download=1&name=original")
		require.Equal(t, http.StatusOK, r.Code)
		data, err := os.ReadFile(filepath.Join(conf.OriginalsPath(), f.FileName))
		require.NoError(t, err)
		assert.Equal(t, data, r.Body.Bytes())
		_, params, err := mime.ParseMediaType(r.Header().Get("Content-Disposition"))
		require.NoError(t, err)
		assert.Equal(t, f.FileName, params["filename"])
		playback := videoDownloadFixture(t, conf, "playback.mp4", "libx264", 64)
		data, err = os.ReadFile(filepath.Join(originals, playback.FileName)) // #nosec G304 -- Fixed fixture name under t.TempDir.
		require.NoError(t, err)
		for _, format := range []string{"mp4", "avc"} {
			t.Run(strings.ToUpper(format), func(t *testing.T) {
				url := "/api/v1/videos/" + playback.FileHash + "/" + conf.PreviewToken() + "/" + format
				full := PerformRequest(app, http.MethodGet, url)
				require.Equal(t, http.StatusOK, full.Code)
				if format == "avc" {
					assert.Equal(t, data, full.Body.Bytes())
				} else {
					info, err := video.Probe(bytes.NewReader(full.Body.Bytes()))
					require.NoError(t, err)
					assert.Equal(t, video.CodecAvc1, info.VideoCodec)
					assert.Equal(t, 64, info.VideoWidth)
				}
				served := full.Body.Bytes()
				assert.Equal(t, playback.ContentType(), full.Header().Get("Content-Type"))
				assert.Empty(t, full.Header().Get("Content-Disposition"))
				assert.Contains(t, full.Header().Get("Cache-Control"), "max-age=")
				modified := full.Header().Get("Last-Modified")
				require.NotEmpty(t, modified)
				for _, tc := range []struct {
					name, header, value, ranged string
					code                        int
					body                        []byte
				}{
					{"Range", "", "", "bytes=0-31", http.StatusPartialContent, served[:32]},
					{"SuffixRange", "", "", "bytes=-16", http.StatusPartialContent, served[len(served)-16:]},
					{"NotModified", "If-Modified-Since", modified, "", http.StatusNotModified, nil},
					{"IfRangeCurrent", "If-Range", modified, "bytes=0-31", http.StatusPartialContent, served[:32]},
					{"IfRangeStale", "If-Range", "Thu, 01 Jan 1970 00:00:00 GMT", "bytes=0-31", http.StatusOK, served},
				} {
					t.Run(tc.name, func(t *testing.T) {
						request := httptest.NewRequest(http.MethodGet, url, nil)
						if tc.header != "" {
							request.Header.Set(tc.header, tc.value)
						}
						if tc.ranged != "" {
							request.Header.Set("Range", tc.ranged)
						}
						response := httptest.NewRecorder()
						app.ServeHTTP(response, request)
						require.Equal(t, tc.code, response.Code)
						assert.Equal(t, tc.body, response.Body.Bytes())
						assert.Empty(t, response.Header().Get("Content-Disposition"))
					})
				}
			})
		}
	})
	t.Run("LegacyPlaybackFailures", func(t *testing.T) {
		for _, failure := range []string{"MissingFile", "TranscodeDisabled"} {
			t.Run(failure, func(t *testing.T) {
				f := videoDownloadFixture(t, conf, failure+".mov", "libx265", 64)
				if failure == "MissingFile" {
					require.NoError(t, os.Remove(filepath.Join(originals, f.FileName)))
				} else {
					disabled := conf.Options().DisableFFmpeg
					conf.Options().DisableFFmpeg = true
					t.Cleanup(func() { conf.Options().DisableFFmpeg = disabled })
				}
				response := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.PreviewToken()+"/avc?download=1")
				require.Equal(t, http.StatusOK, response.Code)
				assert.Equal(t, brokenVideo, response.Body.Bytes())
				assert.Empty(t, response.Header().Get("Content-Disposition"))
			})
		}
	})
	t.Run("HevcUsesSharedSidecar", func(t *testing.T) {
		f := videoDownloadFixture(t, conf, "converted.mov", "libx265", 80)
		legacy := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.PreviewToken()+"/avc?download=1")
		require.Equal(t, http.StatusOK, legacy.Code)
		info, err := video.Probe(bytes.NewReader(legacy.Body.Bytes()))
		require.NoError(t, err)
		assert.Equal(t, video.CodecAvc1, info.VideoCodec)
		assert.Equal(t, 80, info.VideoWidth)
		base := "/api/v1/videos/" + f.FileHash + "/" + conf.DownloadToken() + "/avc?action="
		probe := PerformRequest(app, http.MethodGet, base+"probe")
		require.Equal(t, http.StatusOK, probe.Code)
		var result photoprism.ExportProbe
		require.NoError(t, json.Unmarshal(probe.Body.Bytes(), &result))
		require.NotNil(t, result.Avc)
		assert.True(t, result.Avc.Ready)
		exported := PerformRequest(app, http.MethodGet, base+"download")
		require.Equal(t, http.StatusOK, exported.Code)
		assert.Equal(t, legacy.Body.Bytes(), exported.Body.Bytes())
		_, params, err := mime.ParseMediaType(exported.Header().Get("Content-Disposition"))
		require.NoError(t, err)
		assert.Equal(t, ".mp4", filepath.Ext(params["filename"]))
	})
	t.Run("AnimationUsesIndexedMetadata", func(t *testing.T) {
		for _, fixture := range []struct{ name, file, format string }{
			{"GIF", "2018-04-12 19_24_49.gif", "gif"},
			{"APNG", "animated.png", "png"},
			{"WebP", "windows95.webp", "webp"},
		} {
			t.Run(fixture.name, func(t *testing.T) {
				f := videoDownloadFixture(t, conf, "animation-source-"+fixture.format+".mp4", "libx264", 64)
				data, err := os.ReadFile(filepath.Join("../photoprism/testdata", fixture.file))
				require.NoError(t, err)
				f.FileName, f.FileType, f.FileVideo, f.MediaType, f.FileFrames, f.FileCodec = "animation."+fixture.format, fixture.format, false, entity.MediaAnimated, 2, ""
				name := filepath.Join(originals, "animation."+fixture.format)
				require.NoError(t, os.WriteFile(name, data, fs.ModeFile)) // #nosec G703 -- Fixed fixture name under t.TempDir.
				f.FileHash = fs.Hash(name)
				require.NoError(t, f.Save())
				r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.PreviewToken()+"/avc?download=1")
				require.Equal(t, http.StatusOK, r.Code)
				assert.False(t, bytes.Equal(brokenVideo, r.Body.Bytes()))
				assert.NotEmpty(t, r.Header().Get("Content-Disposition"))
				base := "/api/v1/videos/" + f.FileHash + "/" + conf.DownloadToken() + "/avc?action="
				prepared := PerformRequest(app, http.MethodGet, base+"prepare")
				require.Equal(t, http.StatusOK, prepared.Code, prepared.Body.String())
				exported := PerformRequest(app, http.MethodGet, base+"download")
				require.Equal(t, http.StatusOK, exported.Code)
				info, err := video.Probe(bytes.NewReader(exported.Body.Bytes()))
				require.NoError(t, err)
				assert.Equal(t, video.CodecAvc1, info.VideoCodec)
				assert.NotEmpty(t, exported.Header().Get("Content-Disposition"))
			})
		}
	})
	t.Run("StackSelection", func(t *testing.T) {
		first := videoDownloadFixture(t, conf, "stack-first.mp4", "libx265", 64)
		selected := videoDownloadFixture(t, conf, "stack-selected.mp4", "libx265", 96)
		selected.PhotoID, selected.PhotoUID = first.PhotoID, first.PhotoUID
		require.NoError(t, selected.Save())
		data, err := os.ReadFile("../photoprism/testdata/flash.jpg")
		require.NoError(t, err)
		name := filepath.Join(originals, "stack-cover.jpg")
		require.NoError(t, os.WriteFile(name, data, fs.ModeFile)) // #nosec G703 -- Fixed fixture name under t.TempDir.
		cover := &entity.File{PhotoID: first.PhotoID, PhotoUID: first.PhotoUID, FileRoot: entity.RootOriginals, FileName: "stack-cover.jpg", FileHash: fs.Hash(name), FileType: "jpg", MediaType: entity.MediaImage}
		require.NoError(t, cover.Create())
		t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(cover) })
		for _, tc := range []struct {
			name  string
			file  *entity.File
			width int
		}{{"SelectedVideo", selected, 96}, {"CoverFallback", cover, 64}} {
			t.Run(tc.name, func(t *testing.T) {
				r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+tc.file.FileHash+"/"+conf.PreviewToken()+"/avc?download=1")
				require.Equal(t, http.StatusOK, r.Code)
				info, err := video.Probe(bytes.NewReader(r.Body.Bytes()))
				require.NoError(t, err)
				assert.Equal(t, tc.width, info.VideoWidth)
				base := "/api/v1/videos/" + tc.file.FileHash + "/" + conf.DownloadToken() + "/avc?action="
				if tc.file.FileVideo {
					exported := PerformRequest(app, http.MethodGet, base+"download")
					require.Equal(t, http.StatusOK, exported.Code)
					assert.Equal(t, r.Body.Bytes(), exported.Body.Bytes())
				} else {
					prepared := PerformRequest(app, http.MethodGet, base+"prepare")
					assert.Equal(t, http.StatusUnprocessableEntity, prepared.Code)
					require.NoError(t, cover.Update("FileDuration", time.Second))
					withDuration := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+cover.FileHash+"/"+conf.PreviewToken()+"/avc?download=1")
					require.Equal(t, http.StatusOK, withDuration.Code)
					assert.Equal(t, r.Body.Bytes(), withDuration.Body.Bytes())
				}
			})
		}
	})
	t.Run("MotionVideoAttachment", func(t *testing.T) {
		f := videoDownloadFixture(t, conf, "motion-source.mp4", "libx264", 64)
		clip, err := os.ReadFile(filepath.Join(conf.OriginalsPath(), f.FileName))
		require.NoError(t, err)
		image, err := os.ReadFile("../photoprism/testdata/flash.jpg")
		require.NoError(t, err)
		f.FileName, f.FileType, f.FileVideo, f.MediaType = "motion.jpg", "jpg", false, entity.MediaLive
		name := filepath.Join(originals, "motion.jpg")
		require.NoError(t, os.WriteFile(name, append(image, clip...), fs.ModeFile)) // #nosec G703 -- Fixed fixture name under t.TempDir.
		f.FileHash = fs.Hash(name)
		require.NoError(t, f.Save())
		legacy := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.PreviewToken()+"/avc?download=1")
		require.Equal(t, http.StatusOK, legacy.Code)
		assert.Equal(t, clip, legacy.Body.Bytes())
		assert.NotEmpty(t, legacy.Header().Get("Content-Disposition"))
		request := httptest.NewRequest(http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.PreviewToken()+"/avc?download=1", nil)
		request.Header.Set("Range", "bytes=0-31")
		partial := httptest.NewRecorder()
		app.ServeHTTP(partial, request)
		require.Equal(t, http.StatusPartialContent, partial.Code)
		assert.Equal(t, clip[:32], partial.Body.Bytes())
		assert.NotEmpty(t, partial.Header().Get("Content-Disposition"))
		assert.NoFileExists(t, filepath.Join(conf.MediaFileCachePath(f.FileHash), f.FileHash+".mp4"))
		exported := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.DownloadToken()+"/native?action=download")
		require.Equal(t, http.StatusOK, exported.Code)
		assert.Equal(t, clip, exported.Body.Bytes())
		assert.NotEmpty(t, exported.Header().Get("Content-Disposition"))
	})
	t.Run("MotionVideoBitrateRange", func(t *testing.T) {
		f := videoDownloadFixture(t, conf, "motion-bitrate.mp4", "libx264", 640)
		clipName := filepath.Join(originals, "motion-bitrate.mp4")
		output, err := exec.Command(conf.FFmpegBin(), "-v", "error", "-f", "lavfi", "-i", "testsrc2=size=640x480:rate=30", "-t", "0.4", "-threads", "1", "-c:v", "libx264", "-preset", "ultrafast", "-qp", "0", "-y", clipName).CombinedOutput() //nolint:gosec // executable and output belong to disposable test fixtures
		require.NoError(t, err, "%s", output)
		clip, err := os.ReadFile(clipName) // #nosec G304 -- Fixed fixture name under t.TempDir.
		require.NoError(t, err)
		image, err := os.ReadFile("../photoprism/testdata/flash.jpg")
		require.NoError(t, err)
		f.FileName, f.FileType, f.FileVideo, f.MediaType = "motion-bitrate.jpg", "jpg", false, entity.MediaLive
		name := filepath.Join(originals, "motion-bitrate.jpg")
		require.NoError(t, os.WriteFile(name, append(image, clip...), fs.ModeFile)) // #nosec G703 -- Fixed fixture name under t.TempDir.
		f.FileHash = fs.Hash(name)
		require.NoError(t, f.Save())
		limit := conf.Options().FFmpegBitrate
		conf.Options().FFmpegBitrate = 1
		t.Cleanup(func() { conf.Options().FFmpegBitrate = limit })
		info, err := video.ProbeFile(name)
		require.NoError(t, err)
		require.True(t, conf.FFmpegBitrateExceeded(info.VideoBitrate()), "fixture must exceed the configured limit")
		url := "/api/v1/videos/" + f.FileHash + "/" + conf.PreviewToken() + "/avc?download=1"
		direct := PerformRequest(app, http.MethodGet, url)
		require.Equal(t, http.StatusOK, direct.Code)
		assert.Equal(t, clip, direct.Body.Bytes())
		request := httptest.NewRequest(http.MethodGet, url, nil)
		request.Header.Set("Range", "bytes=0-")
		stream := httptest.NewRecorder()
		app.ServeHTTP(stream, request)
		require.Equal(t, http.StatusPartialContent, stream.Code)
		assert.False(t, bytes.Equal(clip, stream.Body.Bytes()), "streaming must honor the configured bitrate limit")
		assert.NotEmpty(t, stream.Header().Get("Content-Disposition"))
		assert.FileExists(t, filepath.Join(conf.MediaFileCachePath(f.FileHash), f.FileHash+".mp4"))
		converted, err := video.Probe(bytes.NewReader(stream.Body.Bytes()))
		require.NoError(t, err)
		assert.Equal(t, video.CodecAvc1, converted.VideoCodec)
	})
	t.Run("MotionVideoTranscodeClosesSource", func(t *testing.T) {
		f := videoDownloadFixture(t, conf, "motion-hevc.mp4", "libx265", 64)
		clip, err := os.ReadFile(filepath.Join(originals, f.FileName)) // #nosec G304 -- Fixture name is fixed by videoDownloadFixture under t.TempDir.
		require.NoError(t, err)
		image, err := os.ReadFile("../photoprism/testdata/flash.jpg")
		require.NoError(t, err)
		f.FileName, f.FileType, f.FileVideo, f.MediaType = "motion-hevc.jpg", "jpg", false, entity.MediaLive
		name := filepath.Join(originals, "motion-hevc.jpg")
		require.NoError(t, os.WriteFile(name, append(image, clip...), fs.ModeFile)) // #nosec G703 -- Fixed fixture name under t.TempDir.
		f.FileHash = fs.Hash(name)
		require.NoError(t, f.Save())
		for range 3 {
			r := PerformRequest(app, http.MethodGet, "/api/v1/videos/"+f.FileHash+"/"+conf.PreviewToken()+"/avc?download=1")
			require.Equal(t, http.StatusOK, r.Code)
			require.NotEmpty(t, r.Header().Get("Content-Disposition"))
			info, probeErr := video.Probe(bytes.NewReader(r.Body.Bytes()))
			require.NoError(t, probeErr)
			assert.Equal(t, video.CodecAvc1, info.VideoCodec)
			entries, readErr := os.ReadDir("/proc/self/fd")
			if os.IsNotExist(readErr) {
				t.Log("source-descriptor assertions require procfs")
				continue
			}
			require.NoError(t, readErr)
			for _, entry := range entries {
				target, _ := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
				assert.NotEqual(t, name, target, "motion source descriptor remained open")
			}
		}
	})
}
