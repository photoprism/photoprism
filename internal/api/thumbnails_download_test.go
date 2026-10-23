package api

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestThumbDownloadName(t *testing.T) {
	photo := &entity.Photo{TakenAtLocal: time.Date(2019, 1, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Morning"}
	original := func(name, originalName string) *entity.File {
		return &entity.File{Photo: photo, FileRoot: entity.RootOriginals, FileType: "jpg", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc", FileName: name, OriginalName: originalName}
	}
	t.Run("Original", func(t *testing.T) {
		f := original("2024/IMG_1234.jpg", "camera/Holiday.JPG")
		assert.Equal(t, "Holiday_fit_720.jpg", thumbDownloadName(f, customize.DownloadNameOriginal, thumb.Fit720, "x.jpg"))
		assert.Equal(t, "IMG_1234_fit_720.jpg", thumbDownloadName(f, customize.DownloadNameFile, thumb.Fit720, "x.jpg"))
	})
	t.Run("LastExtension", func(t *testing.T) {
		f := original("2024/IMG_1234.HEIC.jpg", "")
		assert.Equal(t, "IMG_1234.HEIC_fit_720.jpg", thumbDownloadName(f, customize.DownloadNameOriginal, thumb.Fit720, "x.jpg"))
	})
	t.Run("Png", func(t *testing.T) {
		f := original("2024/graphic.png", "")
		f.FileType = "png"
		assert.Equal(t, "graphic_fit_1920.jpg", thumbDownloadName(f, customize.DownloadNameFile, thumb.Fit1920, "x.jpg"))
	})
	t.Run("ThumbExtension", func(t *testing.T) {
		f := original("2024/IMG_1234.jpg", "")
		assert.Equal(t, "IMG_1234_colors.png", thumbDownloadName(f, customize.DownloadNameFile, thumb.Colors, "x.png"))
	})
	t.Run("Share", func(t *testing.T) {
		f := original("2024/IMG_1234.jpg", "")
		assert.Equal(t, "20190115-000000-Morning_tile_500.jpg", thumbDownloadName(f, customize.DownloadNameShare, thumb.Tile500, "x.jpg"))
	})
	t.Run("Unsafe", func(t *testing.T) {
		f := original("2024/x.jpg", "camera/a:b.jpg")
		assert.Equal(t, "a_b_fit_720.jpg", thumbDownloadName(f, customize.DownloadNameOriginal, thumb.Fit720, "x.jpg"))
	})
	t.Run("Header", func(t *testing.T) {
		thumbName := filepath.Join(t.TempDir(), "thumb.jpg")
		require.NoError(t, os.WriteFile(thumbName, NewTestJpeg(t, 8, 8), fs.ModeFile)) //nolint:gosec // Test writes to its own temporary folder.
		for _, originalName := range []string{"a\r\nSet-Cookie: x=1.jpg", `a"b\c.jpg`, "a\u202Eb.jpg", "Café Paris.jpg"} {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.FileAttachment(thumbName, thumbDownloadName(original("2024/x.jpg", originalName), customize.DownloadNameOriginal, thumb.Fit720, thumbName))
			disposition := w.Header().Values("Content-Disposition")
			require.Len(t, disposition, 1, originalName)
			assert.NotContains(t, disposition[0], "\r", originalName)
			assert.NotContains(t, disposition[0], "\n", originalName)
			assert.Empty(t, w.Header().Get("Set-Cookie"), originalName)
			assert.Contains(t, disposition[0], "_fit_720.jpg", originalName)
		}
	})
	t.Run("UnsavedSidecar", func(t *testing.T) {
		f := &entity.File{Photo: photo, FileRoot: entity.RootSidecar, FileType: "jpg", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc", FileName: "2024/IMG_1234.HEIC.jpg"}
		assert.Equal(t, "IMG_1234.HEIC_fit_720.jpg", thumbDownloadName(f, customize.DownloadNameFile, thumb.Fit720, "x.jpg"))
	})
}

func TestThumbOwner(t *testing.T) {
	// addPhoto creates a picture with file rows of the given roots and names.
	addPhoto := func(t *testing.T, files ...[3]string) []*entity.File {
		photo := entity.NewPhoto(false)
		require.NoError(t, photo.Create())
		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", photo.ID)
		})
		result := make([]*entity.File, 0, len(files))
		for _, rn := range files {
			f := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: rn[0], FileName: rn[1], FileType: "jpg", FileSidecar: rn[2] == "sidecar"}
			require.NoError(t, f.Create())
			result = append(result, f)
		}
		return result
	}
	t.Run("FullName", func(t *testing.T) {
		files := addPhoto(t, [3]string{entity.RootOriginals, "thumb-owner/a/IMG_1.HEIC", ""}, [3]string{entity.RootOriginals, "thumb-owner/a/IMG_1.HEIC.mov", ""}, [3]string{entity.RootSidecar, "thumb-owner/a/IMG_1.HEIC.jpg", ""})
		assert.Equal(t, files[0].ID, thumbOwner(files[2]).ID)
	})
	t.Run("Stem", func(t *testing.T) {
		files := addPhoto(t, [3]string{entity.RootOriginals, "thumb-owner/b/raw.cr2", ""}, [3]string{entity.RootOriginals, "thumb-owner/b/raw.xmp", "sidecar"}, [3]string{entity.RootSidecar, "thumb-owner/b/raw.jpg", ""})
		assert.Equal(t, files[0].ID, thumbOwner(files[2]).ID)
	})
	t.Run("DottedFolder", func(t *testing.T) {
		files := addPhoto(t, [3]string{entity.RootOriginals, "thumb-owner/2024.06/IMG_4", ""}, [3]string{entity.RootSidecar, "thumb-owner/2024.jpg", ""})
		assert.Equal(t, files[1].ID, thumbOwner(files[1]).ID)
	})
	t.Run("FullNameBeforeStems", func(t *testing.T) {
		files := addPhoto(t, [3]string{entity.RootOriginals, "thumb-owner/f/IMG_3.cr3", ""}, [3]string{entity.RootOriginals, "thumb-owner/f/IMG_3.hif", ""}, [3]string{entity.RootOriginals, "thumb-owner/f/IMG_3", ""}, [3]string{entity.RootSidecar, "thumb-owner/f/IMG_3.jpg", ""})
		assert.Equal(t, files[2].ID, thumbOwner(files[3]).ID)
	})
	t.Run("Ambiguous", func(t *testing.T) {
		files := addPhoto(t, [3]string{entity.RootOriginals, "thumb-owner/c/dup.cr3", ""}, [3]string{entity.RootOriginals, "thumb-owner/c/dup.hif", ""}, [3]string{entity.RootSidecar, "thumb-owner/c/dup.jpg", ""})
		assert.Equal(t, files[2].ID, thumbOwner(files[2]).ID)
	})
	t.Run("OtherPicture", func(t *testing.T) {
		other := addPhoto(t, [3]string{entity.RootOriginals, "thumb-owner/d/IMG_2.HEIC", ""})
		files := addPhoto(t, [3]string{entity.RootSidecar, "thumb-owner/d/IMG_2.HEIC.jpg", ""})
		assert.NotEqual(t, other[0].ID, thumbOwner(files[0]).ID)
		assert.Equal(t, files[0].ID, thumbOwner(files[0]).ID)
	})
	t.Run("Original", func(t *testing.T) {
		files := addPhoto(t, [3]string{entity.RootOriginals, "thumb-owner/e/photo.jpg", ""})
		assert.Same(t, files[0], thumbOwner(files[0]))
	})
}

func TestAbortThumb(t *testing.T) {
	t.Run("Download", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		abortThumb(c, true, http.StatusNotFound, brokenIconSvg)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, brokenIconSvg, w.Body.Bytes())
	})
	t.Run("Inline", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		abortThumb(c, false, http.StatusNotFound, brokenIconSvg)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "image/svg+xml", w.Header().Get("Content-Type"))
	})
}

func TestGetThumb_Download(t *testing.T) {
	app, router, conf := NewApiTest()
	GetThumb(router)

	options := *conf.Options()
	t.Cleanup(func() {
		*conf.Options() = options
		conf.Propagate()
	})
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().CachePath = t.TempDir()
	conf.Options().ThumbUncached = true
	conf.Propagate()

	photo := entity.NewPhoto(false)
	require.NoError(t, photo.Create())
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", photo.ID)
	})

	// addFile writes data below root and adds it to the test photo.
	addFile := func(t *testing.T, root, name, fileType string, data []byte) *entity.File {
		dir := conf.OriginalsPath()
		if root == entity.RootSidecar {
			dir = conf.SidecarPath()
		}
		fileName := filepath.Join(dir, name)
		require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
		require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile)) //nolint:gosec // Test writes to its own temporary folder.
		f := &entity.File{
			PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: root, FileName: name, FileType: fileType,
			FileHash: fs.Hash(fileName), FileSize: int64(len(data)), FileWidth: 1600, FileHeight: 1200, MediaType: entity.MediaImage,
		}
		require.NoError(t, f.Create())
		t.Cleanup(func() {
			get.ThumbCache().Flush()
		})
		return f
	}

	// request fetches a thumbnail with the preview token.
	request := func(hash, size, query string) *httptest.ResponseRecorder {
		return PerformRequest(app, http.MethodGet, "/api/v1/t/"+hash+"/"+conf.PreviewToken()+"/"+size+query)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1600, 1200))
	for y := 0; y < 1200; y++ {
		for x := 0; x < 1600; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var pngData bytes.Buffer
	require.NoError(t, png.Encode(&pngData, img))

	jpg := addFile(t, entity.RootOriginals, "thumb-download/IMG_1234.jpg", "jpg", NewTestJpeg(t, 1600, 1200))
	heic := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: "thumb-download/internal.heic", OriginalName: "camera/Holiday.heic", FileType: "heic"}
	require.NoError(t, heic.Create())
	preview := addFile(t, entity.RootSidecar, "thumb-download/internal.heic.png", "png", pngData.Bytes())

	t.Run("Miss", func(t *testing.T) {
		get.ThumbCache().Flush()
		w := request(jpg.FileHash, "fit_720", "?download=1&name=file")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="IMG_1234_fit_720.jpg"`, w.Header().Get("Content-Disposition"))
		assert.Equal(t, "image/jpeg", w.Header().Get("Content-Type"))
	})
	t.Run("Hit", func(t *testing.T) {
		inline := request(jpg.FileHash, "fit_720", "")
		require.Equal(t, http.StatusOK, inline.Code)
		assert.Empty(t, inline.Header().Get("Content-Disposition"))
		_, cached := get.ThumbCache().Get(CacheKey("thumbs", jpg.FileHash, "fit_720"))
		require.True(t, cached)
		w := request(jpg.FileHash, "fit_720", "?download=1&name=file")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="IMG_1234_fit_720.jpg"`, w.Header().Get("Content-Disposition"))
		assert.Equal(t, inline.Body.Bytes(), w.Body.Bytes())
	})
	t.Run("StaleCacheEntry", func(t *testing.T) {
		key := CacheKey("thumbs", jpg.FileHash, "tile_500")
		get.ThumbCache().SetDefault(key, ThumbCache{FileName: filepath.Join(t.TempDir(), "missing.jpg")})
		w := request(jpg.FileHash, "tile_500", "?download=1&name=file")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="IMG_1234_tile_500.jpg"`, w.Header().Get("Content-Disposition"))
	})
	t.Run("PngSource", func(t *testing.T) {
		w := request(preview.FileHash, "fit_720", "?download=1&name=original")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="Holiday_fit_720.jpg"`, w.Header().Get("Content-Disposition"))
		assert.Equal(t, "image/jpeg", w.Header().Get("Content-Type"))
		w = request(preview.FileHash, "fit_720", "?download=1&name=file")
		assert.Equal(t, `attachment; filename="internal_fit_720.jpg"`, w.Header().Get("Content-Disposition"))
	})
	t.Run("UnknownFile", func(t *testing.T) {
		const unknown = "3cad9168fa6acc5c5c2965ddf6ec465ca42fd000"
		assert.Equal(t, http.StatusNotFound, request(unknown, "fit_720", "?download=1").Code)
		assert.Equal(t, http.StatusOK, request(unknown, "fit_720", "").Code)
	})
	t.Run("InvalidSize", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, request(jpg.FileHash, "xxx", "?download=1").Code)
		assert.Equal(t, http.StatusOK, request(jpg.FileHash, "xxx", "").Code)
	})
	t.Run("RenderFailure", func(t *testing.T) {
		broken := addFile(t, entity.RootOriginals, "thumb-download/broken.jpg", "jpg", []byte("not a jpeg"))
		w := request(broken.FileHash, "fit_720", "?download=1")
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "image/svg+xml", w.Header().Get("Content-Type"))
		assert.Empty(t, w.Header().Get("Content-Disposition"))
		w = request(broken.FileHash, "fit_720", "")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, brokenIconSvg, w.Body.Bytes())
	})
	t.Run("NotCached", func(t *testing.T) {
		conf.Options().ThumbUncached = false
		t.Cleanup(func() { conf.Options().ThumbUncached = true })
		fresh := addFile(t, entity.RootOriginals, "thumb-download/fresh.jpg", "jpg", NewTestJpeg(t, 1600, 1199))
		w := request(fresh.FileHash, "fit_720", "?download=1")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Empty(t, w.Header().Get("Content-Disposition"))
		assert.Equal(t, http.StatusOK, request(fresh.FileHash, "fit_720", "").Code)
	})
	t.Run("FileError", func(t *testing.T) {
		require.NoError(t, jpg.Update("FileError", "failed"))
		t.Cleanup(func() { _ = jpg.Update("FileError", "") })
		assert.Equal(t, http.StatusNotFound, request(jpg.FileHash, "fit_720", "?download=1").Code)
		get.ThumbCache().Flush()
		w := request(jpg.FileHash, "tile_224", "")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, brokenIconSvg, w.Body.Bytes())
	})
	t.Run("Missing", func(t *testing.T) {
		missing := addFile(t, entity.RootOriginals, "thumb-download/missing.jpg", "jpg", NewTestJpeg(t, 64, 48))
		require.NoError(t, os.Remove(photoprism.FileName(missing.FileRoot, missing.FileName)))
		assert.Equal(t, http.StatusNotFound, request(missing.FileHash, "fit_720", "?download=1").Code)
		found := &entity.File{}
		require.NoError(t, entity.UnscopedDb().First(found, missing.ID).Error)
		assert.True(t, found.FileMissing)
	})
	t.Run("Crop", func(t *testing.T) {
		w := request(jpg.FileHash+"-016014058037", "tile_160", "?download=1")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="tile_160.jpg"`, w.Header().Get("Content-Disposition"))
	})
	t.Run("CropFailure", func(t *testing.T) {
		const unknown = "3cad9168fa6acc5c5c2965ddf6ec465ca42fd000-016014058037"
		assert.Equal(t, http.StatusNotFound, request(unknown, "tile_160", "?download=1").Code)
		assert.Equal(t, http.StatusOK, request(unknown, "tile_160", "").Code)
		assert.Equal(t, http.StatusBadRequest, request(unknown, "xxx", "?download=1").Code)
	})
}
