package api

import (
	"archive/zip"
	"bytes"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/media"
)

// testJpeg returns a JPEG with the given dimensions.
func testJpeg(t *testing.T, width, height int) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, image.NewGray(image.Rect(0, 0, width, height)), &jpeg.Options{Quality: 10}))

	return buf.Bytes()
}

// TestUploadCheckFile_ResolutionLimit verifies that images above the resolution limit are rejected.
func TestUploadCheckFile_ResolutionLimit(t *testing.T) {
	data := testJpeg(t, 3000, 2000)

	t.Run("Exceeds", func(t *testing.T) {
		dir := t.TempDir()
		fileName := filepath.Join(dir, "large.jpg")
		require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile))
		remaining, err := UploadCheckFile(fileName, false, 5, 1<<20)
		require.Error(t, err)
		assert.Equal(t, int64(1<<20), remaining)
		assert.Contains(t, err.Error(), "large.jpg")
		assert.Contains(t, err.Error(), "resolution limit (6 / 5 MP)")
		assert.NotContains(t, clean.Error(err), dir)
		assert.NoFileExists(t, fileName)
	})
	t.Run("WithinLimit", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "large.jpg")
		require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile))
		remaining, err := UploadCheckFile(fileName, false, 6, 1<<20)
		require.NoError(t, err)
		assert.Equal(t, int64(1<<20-len(data)), remaining)
		assert.FileExists(t, fileName)
	})
	t.Run("Disabled", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "large.jpg")
		require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile))
		_, err := UploadCheckFile(fileName, false, -1, 1<<20)
		require.NoError(t, err)
		assert.FileExists(t, fileName)
	})
	t.Run("ContentNotExtension", func(t *testing.T) {
		for _, name := range []string{"large.mpo", "large.insp"} {
			t.Run(name, func(t *testing.T) {
				fileName := filepath.Join(t.TempDir(), name)
				require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile))
				_, err := UploadCheckFile(fileName, false, 5, 1<<20)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "resolution limit (6 / 5 MP)")
				assert.NoFileExists(t, fileName)
			})
		}
	})
	t.Run("ContentMismatch", func(t *testing.T) {
		for _, name := range []string{"large.bmp", "large.mp4"} {
			t.Run(name, func(t *testing.T) {
				fileName := filepath.Join(t.TempDir(), name)
				require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile))
				_, err := UploadCheckFile(fileName, false, 150, 1<<20)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid extension")
				assert.NoFileExists(t, fileName)
			})
		}
	})
	t.Run("Video", func(t *testing.T) {
		src, err := os.ReadFile(filepath.Join("..", "..", "assets", "samples", "blue-go-video.mp4"))
		if err != nil {
			t.Skip("video sample is not available")
		}
		fileName := filepath.Join(t.TempDir(), "video.mp4")
		require.NoError(t, os.WriteFile(fileName, src, fs.ModeFile)) //nolint:gosec // test writes to a temp path
		_, err = UploadCheckFile(fileName, false, 1, 1<<30)
		require.NoError(t, err)
		assert.FileExists(t, fileName)
	})
	t.Run("NotAnImage", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "notes.txt")
		require.NoError(t, os.WriteFile(fileName, []byte("text"), fs.ModeFile))
		_, err := UploadCheckFile(fileName, false, 1, 1<<20)
		require.NoError(t, err)
		assert.FileExists(t, fileName)
	})
}

// TestUploadUserFilesResolutionLimit verifies that the upload handler rejects images above the
// configured resolution limit before they are screened.
func TestUploadUserFilesResolutionLimit(t *testing.T) {
	app, router, conf := NewApiTest()
	previousStorage, previousAllow, previousNSFW, previousLimit := conf.Options().StoragePath, conf.Options().UploadAllow, conf.Options().UploadNSFW, conf.Options().ResolutionLimit
	t.Cleanup(func() {
		conf.Options().StoragePath, conf.Options().UploadAllow, conf.Options().UploadNSFW, conf.Options().ResolutionLimit = previousStorage, previousAllow, previousNSFW, previousLimit
	})
	conf.Options().StoragePath = t.TempDir()
	conf.Options().UploadAllow = "jpg"
	conf.Options().UploadNSFW = false
	UploadUserFiles(router)
	authToken := AuthenticateAdmin(app, router)
	stubNSFW(t, []nsfw.Result{nsfw.NewResult(0.01, nsfw.DefaultThreshold)}, nil)

	screened := 0
	vision.SetNSFWUploadFunc(func(files vision.Files, _ media.Src) ([]nsfw.Result, error) {
		screened += len(files)
		return []nsfw.Result{nsfw.NewResult(0.01, nsfw.DefaultThreshold)}, nil
	})

	data := testJpeg(t, 3000, 2000)

	for _, tc := range []struct {
		name     string
		limit    int
		token    string
		accepted bool
	}{
		{"Exceeds", 5, "resolution-exceeds", false},
		{"WithinLimit", 6, "resolution-within", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conf.Options().ResolutionLimit = tc.limit
			screened = 0
			body, contentType, err := buildMultipart(map[string][]byte{"large.jpg": data})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+entity.Admin.UserUID+"/upload/"+tc.token, body)
			req.Header.Set("Content-Type", contentType)
			header.SetAuthorization(req, authToken)
			out := httptest.NewRecorder()
			app.ServeHTTP(out, req)

			uploadBase := filepath.Join(conf.UserStoragePath(entity.Admin.UserUID), "upload")
			if tc.accepted {
				require.Equal(t, http.StatusOK, out.Code, out.Body.String())
				assert.NotEmpty(t, findUploadedFilesForToken(t, uploadBase, tc.token))
				assert.Equal(t, 1, screened)
			} else {
				assert.Empty(t, findUploadedFilesForToken(t, uploadBase, tc.token))
				assert.Zero(t, screened, "rejected files must not be screened")
			}
		})
	}
}

// TestUploadCheckFile_JpegScans verifies that JPEG files with more scans than supported are rejected.
func TestUploadCheckFile_JpegScans(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "pkg", "fs", "testdata", "progressive.jpg"))
	require.NoError(t, err)

	t.Run("Progressive", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "progressive.jpg")
		require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile)) //nolint:gosec // test writes to a temp path
		_, err := UploadCheckFile(fileName, false, 150, 1<<20)
		require.NoError(t, err)
		assert.FileExists(t, fileName)
	})
	t.Run("AboveLimit", func(t *testing.T) {
		previous := fs.MaxJpegScans
		fs.MaxJpegScans = 9
		t.Cleanup(func() { fs.MaxJpegScans = previous })
		dir := t.TempDir()
		fileName := filepath.Join(dir, "progressive.jpg")
		require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile)) //nolint:gosec // test writes to a temp path
		remaining, err := UploadCheckFile(fileName, false, 150, 1<<20)
		require.Error(t, err)
		assert.Equal(t, int64(1<<20), remaining)
		assert.Contains(t, err.Error(), "progressive.jpg")
		assert.NotContains(t, clean.Error(err), dir)
		assert.NoFileExists(t, fileName)
	})
}

// TestUploadMegapixels verifies that the resolution is read from the type and the content.
func TestUploadMegapixels(t *testing.T) {
	data := testJpeg(t, 3000, 2000)

	t.Run("Jpeg", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "image.jpg")
		require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile))
		mediaFile, err := photoprism.NewMediaFile(fileName)
		require.NoError(t, err)
		assert.Equal(t, 6, uploadMegapixels(mediaFile))
	})
	t.Run("JpegContent", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "image.mpo")
		require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile))
		mediaFile, err := photoprism.NewMediaFile(fileName)
		require.NoError(t, err)
		assert.Equal(t, 6, uploadMegapixels(mediaFile))
	})
	t.Run("Text", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "notes.txt")
		require.NoError(t, os.WriteFile(fileName, []byte("text"), fs.ModeFile))
		mediaFile, err := photoprism.NewMediaFile(fileName)
		require.NoError(t, err)
		assert.Zero(t, uploadMegapixels(mediaFile))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Zero(t, uploadMegapixels(nil))
	})
}

// TestUploadUserFilesResolutionLimitZip verifies that archive entries above the limit are rejected.
func TestUploadUserFilesResolutionLimitZip(t *testing.T) {
	app, router, conf := NewApiTest()
	previousStorage, previousAllow, previousNSFW, previousLimit, previousArchives := conf.Options().StoragePath, conf.Options().UploadAllow, conf.Options().UploadNSFW, conf.Options().ResolutionLimit, conf.Options().UploadArchives
	t.Cleanup(func() {
		conf.Options().StoragePath, conf.Options().UploadAllow, conf.Options().UploadNSFW, conf.Options().ResolutionLimit, conf.Options().UploadArchives = previousStorage, previousAllow, previousNSFW, previousLimit, previousArchives
	})
	conf.Options().StoragePath = t.TempDir()
	conf.Options().UploadAllow = "jpg,zip"
	conf.Options().UploadArchives = true
	conf.Options().UploadNSFW = true
	conf.Options().ResolutionLimit = 5
	UploadUserFiles(router)
	authToken := AuthenticateAdmin(app, router)

	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	for name, data := range map[string][]byte{"large.jpg": testJpeg(t, 3000, 2000), "small.jpg": testJpeg(t, 300, 200)} {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())

	const uploadToken = "resolution-zip"
	body, contentType, err := buildMultipart(map[string][]byte{"upload.zip": zbuf.Bytes()})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+entity.Admin.UserUID+"/upload/"+uploadToken, body)
	req.Header.Set("Content-Type", contentType)
	header.SetAuthorization(req, authToken)
	out := httptest.NewRecorder()
	app.ServeHTTP(out, req)
	require.Equal(t, http.StatusOK, out.Code, out.Body.String())

	var names []string
	for _, fileName := range findUploadedFilesForToken(t, filepath.Join(conf.UserStoragePath(entity.Admin.UserUID), "upload"), uploadToken) {
		names = append(names, filepath.Base(fileName))
	}
	assert.Contains(t, names, "small.jpg")
	assert.NotContains(t, names, "large.jpg")
}
