package api

import (
	"image/color"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/frame"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// sharedPreviewAlbum creates a manual album holding the given photos, shares it with a link and
// returns the album UID and the link token.
func sharedPreviewAlbum(t *testing.T, title string, photoUIDs ...string) (albumUID, token string) {
	t.Helper()

	album := entity.NewAlbum(title, entity.AlbumManual)
	require.NoError(t, album.Create())

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Delete(album).Error
		entity.FlushAlbumCache()
	})

	for _, uid := range photoUIDs {
		entry := entity.NewPhotoAlbum(uid, album.AlbumUID)
		require.NoError(t, entry.Save())

		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(entry).Error })
	}

	link := entity.NewUserLink(album.AlbumUID, entity.Admin.UserUID)
	require.NoError(t, link.Save())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(link).Error })

	require.Len(t, entity.FindValidLinksByToken(link.LinkToken, album.AlbumUID), 1)

	return album.AlbumUID, link.LinkToken
}

// previewOriginal writes an original for a file fixture and drops the thumbnail its hash resolves
// to, so the request under test renders from the bytes this call wrote. A file written as
// unreadable is asserted to produce no thumbnail, which is the premise of the cases that rely on it.
func previewOriginal(t *testing.T, conf *config.Config, fixture string, readable bool) {
	t.Helper()

	f := entity.FileFixtures.Pointer(fixture)
	CreateTestOriginal(t, f)

	origName := photoprism.FileName(f.FileRoot, f.FileName)

	if !readable {
		require.NoError(t, os.WriteFile(origName, []byte("not an image"), fs.ModeFile))
	}

	size := thumb.Sizes[thumb.Tile500]

	cached, err := thumb.FileName(f.FileHash, conf.ThumbCachePath(), size.Width, size.Height, size.Options...)
	require.NoError(t, err)

	if err = os.Remove(cached); err != nil {
		require.ErrorIs(t, err, os.ErrNotExist)
	}

	t.Cleanup(func() { _ = os.Remove(cached) })

	if !readable {
		_, err = thumb.FromFile(origName, f.FileHash, conf.ThumbCachePath(), size.Width, size.Height, f.FileOrientation, size.Options...)
		require.Error(t, err, "the original must not render a thumbnail")
		require.NoFileExists(t, cached)
	}
}

// renderedPixels counts the pixels of a rendered preview that differ from the collage background,
// so a test can tell a composed card from an empty one. It allows for the resampling and the JPEG
// encoding the handler applies.
func renderedPixels(t *testing.T, fileName string) int {
	t.Helper()

	img, _, err := fs.DecodeImageFile(fileName)
	require.NoError(t, err)

	bg := frame.CollageBackground
	b := img.Bounds()
	n := 0

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)

			if channelDiff(c.R, bg.R) > 12 || channelDiff(c.G, bg.G) > 12 || channelDiff(c.B, bg.B) > 12 {
				n++
			}
		}
	}

	return n
}

// channelDiff returns the absolute difference between two color channels.
func channelDiff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}

	return int(b - a)
}

// TestSharePreview_Collage covers what a share link renders when the album leaves the collage a
// single image, when it leaves none, and when it leaves several.
func TestSharePreview_Collage(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	SharePreview(router)

	// request performs the share preview request and returns the response with the file it wrote.
	request := func(t *testing.T, albumUID, token string) (int, string) {
		t.Helper()

		previewFile := filepath.Join(path.Join(conf.ThumbCachePath(), "share"), albumUID+fs.ExtJpeg)
		_ = os.Remove(previewFile)
		t.Cleanup(func() { _ = os.Remove(previewFile) })

		r := PerformRequest(app, "GET", "/api/v1/"+token+"/"+albumUID+"/preview")

		return r.Code, previewFile
	}

	t.Run("OnePhoto", func(t *testing.T) {
		previewOriginal(t, conf, "exampleFileName.jpg", true)

		albumUID, token := sharedPreviewAlbum(t, "Share Preview One Photo", "ps6sg6be2lvl0yh7")
		code, previewFile := request(t, albumUID, token)

		assert.Equal(t, http.StatusOK, code)
		require.FileExists(t, previewFile)
		mimeType, _ := fs.DetectMimeType(previewFile)
		assert.Equal(t, header.ContentTypeJpeg, mimeType)
		assert.NotZero(t, renderedPixels(t, previewFile), "the preview must hold the album image")
	})
	t.Run("OneReadablePhoto", func(t *testing.T) {
		previewOriginal(t, conf, "exampleFileName.jpg", true)
		previewOriginal(t, conf, "bridge1.jpg", false)

		albumUID, token := sharedPreviewAlbum(t, "Share Preview One Readable Photo", "ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh9")
		code, previewFile := request(t, albumUID, token)

		assert.Equal(t, http.StatusOK, code)
		require.FileExists(t, previewFile)
		mimeType, _ := fs.DetectMimeType(previewFile)
		assert.Equal(t, header.ContentTypeJpeg, mimeType)
		assert.NotZero(t, renderedPixels(t, previewFile), "the preview must hold the readable image")
	})
	t.Run("NoReadablePhoto", func(t *testing.T) {
		previewOriginal(t, conf, "bridge1.jpg", false)

		albumUID, token := sharedPreviewAlbum(t, "Share Preview No Readable Photo", "ps6sg6be2lvl0yh9")
		code, previewFile := request(t, albumUID, token)

		// Nothing renders, so the request serves the site preview and marks the album empty.
		assert.Equal(t, http.StatusTemporaryRedirect, code)
		require.FileExists(t, previewFile)

		info, err := os.Stat(previewFile)
		require.NoError(t, err)
		assert.Zero(t, info.Size())

		// A repeat answers from that marker rather than composing again.
		repeat := PerformRequest(app, "GET", "/api/v1/"+token+"/"+albumUID+"/preview")
		assert.Equal(t, http.StatusTemporaryRedirect, repeat.Code)

		info, err = os.Stat(previewFile)
		require.NoError(t, err)
		assert.Zero(t, info.Size())
	})
	t.Run("TwoPhotos", func(t *testing.T) {
		previewOriginal(t, conf, "exampleFileName.jpg", true)
		previewOriginal(t, conf, "bridge1.jpg", true)

		albumUID, token := sharedPreviewAlbum(t, "Share Preview Two Photos", "ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh9")
		code, previewFile := request(t, albumUID, token)

		assert.Equal(t, http.StatusOK, code)
		require.FileExists(t, previewFile)
		assert.NotZero(t, renderedPixels(t, previewFile))
	})
}

func TestMarkEmptyPreview(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "as6sg6bxpogaaba7"+fs.ExtJpeg)
		markEmptyPreview(fileName)

		info, err := os.Stat(fileName)
		require.NoError(t, err)
		assert.Zero(t, info.Size())
	})
	t.Run("KeepsExistingFile", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "as6sg6bxpogaaba7"+fs.ExtJpeg)
		require.NoError(t, os.WriteFile(fileName, []byte("card"), fs.ModeFile))

		markEmptyPreview(fileName)

		data, err := os.ReadFile(fileName) // #nosec G304 -- the name comes from t.TempDir.
		require.NoError(t, err)
		assert.Equal(t, "card", string(data))
	})
	t.Run("MissingFolder", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "missing", "as6sg6bxpogaaba7"+fs.ExtJpeg)

		markEmptyPreview(fileName)

		assert.NoFileExists(t, fileName)
	})
}
