package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestBatchPhotosArchive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhoto(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "DeletedAt")
		assert.Empty(t, val.String())

		BatchPhotosArchive(router)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": ["ps6sg6be2lvl0yh7", "ps6sg6be2lvl0ycc"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "Selection archived")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7")
		assert.Equal(t, http.StatusOK, r3.Code)
		val3 := gjson.Get(r3.Body.String(), "DeletedAt")
		assert.NotEmpty(t, val3.String())
	})
	t.Run("MissingSelection", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosArchive(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosArchive(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosArchive(router)

		body := `{"photos":["` + strings.Repeat("p", int(MaxSelectionRequestBytes)) + `"]}`
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/batch/photos/archive", body)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
}

func TestBatchPhotosRestore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		BatchPhotosArchive(router)
		GetPhoto(router)
		BatchPhotosRestore(router)

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/archive", `{"photos": ["ps6sg6be2lvl0yh8", "ps6sg6be2lvl0ycc"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "Selection archived")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh8")
		assert.Equal(t, http.StatusOK, r3.Code)
		val3 := gjson.Get(r3.Body.String(), "DeletedAt")
		assert.NotEmpty(t, val3.String())

		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/restore", `{"photos": ["ps6sg6be2lvl0yh8", "ps6sg6be2lvl0ycc"]}`)
		val := gjson.Get(r.Body.String(), "message")
		assert.Contains(t, val.String(), "Selection restored")
		assert.Equal(t, http.StatusOK, r.Code)

		r4 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh8")
		assert.Equal(t, http.StatusOK, r4.Code)
		val4 := gjson.Get(r4.Body.String(), "DeletedAt")
		assert.Empty(t, val4.String())
	})
	t.Run("MissingSelection", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosRestore(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/restore", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosRestore(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/restore", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchAlbumsDelete(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateAlbum(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/albums", `{"Title": "BatchDelete", "Description": "To be deleted", "Notes": "", "Favorite": true}`)
	assert.Equal(t, http.StatusCreated, r.Code)
	uid := gjson.Get(r.Body.String(), "UID").String()

	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		GetAlbum(router)
		BatchAlbumsDelete(router)

		r := PerformRequest(app, "GET", "/api/v1/albums/"+uid)
		val := gjson.Get(r.Body.String(), "Slug")
		assert.Equal(t, "batchdelete", val.String())

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/albums/delete", fmt.Sprintf(`{"albums": ["%s", "ps6sg6be2lvl0ycc"]}`, uid))
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), i18n.Msg(i18n.MsgAlbumsDeleted))
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/albums/"+uid)
		val3 := gjson.Get(r3.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAlbumNotFound), val3.String())
		assert.Equal(t, http.StatusNotFound, r3.Code)
	})
	t.Run("NoAlbumsSelected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchAlbumsDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/albums/delete", `{"albums": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoAlbumsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchAlbumsDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/albums/delete", `{"albums": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchAlbumsDelete(router)

		body := `{"albums":["` + strings.Repeat("a", int(MaxSelectionRequestBytes)) + `"]}`
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/batch/albums/delete", body)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
}

func TestBatchPhotosPrivate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		GetPhoto(router)
		BatchPhotosPrivate(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh8")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Private")
		assert.Equal(t, "false", val.String())

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/private", `{"photos": ["ps6sg6be2lvl0yh8", "ps6sg6be2lvl0ycc"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "Selection marked as private")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh8")
		assert.Equal(t, http.StatusOK, r3.Code)
		val3 := gjson.Get(r3.Body.String(), "Private")
		assert.Equal(t, "true", val3.String())
	})
	t.Run("MissingSelection", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosPrivate(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/private", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosPrivate(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/private", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestBatchLabelsDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		SearchLabels(router)
		BatchLabelsDelete(router)

		r := PerformRequest(app, "GET", "/api/v1/labels?count=15")
		val := gjson.Get(r.Body.String(), `#(Name=="Batch Delete").Slug`)
		assert.Equal(t, val.String(), "batch-delete")

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/labels/delete", `{"labels": ["ls6sg6b1wowuy3c6", "ps6sg6be2lvl0ycc"]}`)

		var resp i18n.Response

		if err := json.Unmarshal(r2.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}

		assert.True(t, resp.Success())
		assert.Equal(t, i18n.Msg(i18n.MsgLabelsDeleted), resp.Message)
		assert.Equal(t, i18n.Msg(i18n.MsgLabelsDeleted), resp.String())
		assert.Equal(t, http.StatusOK, r2.Code)
		assert.Equal(t, http.StatusOK, resp.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/labels?count=15")
		val3 := gjson.Get(r3.Body.String(), `#(Name=="BatchDelete").Slug`)
		assert.Equal(t, val3.String(), "")
	})
	t.Run("NoLabelsSelected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchLabelsDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/labels/delete", `{"labels": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoLabelsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchLabelsDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/labels/delete", `{"labels": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchLabelsDelete(router)

		body := `{"labels":["` + strings.Repeat("l", int(MaxSelectionRequestBytes)) + `"]}`
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/batch/labels/delete", body)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
}

func TestBatchPhotosApprove(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Register routes.
		GetPhoto(router)
		BatchPhotosApprove(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y50")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Quality")
		assert.Equal(t, "1", val.String())
		val4 := gjson.Get(r.Body.String(), "EditedAt")
		assert.Empty(t, val4.String())

		r2 := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/approve", `{"photos": ["ps6sg6be2lvl0y50", "ps6sg6be2lvl0y90"]}`)
		val2 := gjson.Get(r2.Body.String(), "message")
		assert.Contains(t, val2.String(), "Selection approved")
		assert.Equal(t, http.StatusOK, r2.Code)

		r3 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y50")
		assert.Equal(t, http.StatusOK, r3.Code)
		val5 := gjson.Get(r3.Body.String(), "Quality")
		assert.Equal(t, "7", val5.String())
		val6 := gjson.Get(r3.Body.String(), "EditedAt")
		assert.NotEmpty(t, val6.String())
	})
	t.Run("MissingSelection", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosApprove(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/approve", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosApprove(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/approve", `{"photos": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

// batchDeleteTestFolder returns a new originals folder for batch delete tests and removes it,
// along with the rows of the photos created in it, when the test ends.
func batchDeleteTestFolder(t *testing.T, conf *config.Config) string {
	t.Helper()

	folder := "zz-batch-delete-" + rnd.Base36(8)
	dir := filepath.Join(conf.OriginalsPath(), folder)
	require.NoError(t, fs.MkdirAll(dir))

	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
		db := entity.UnscopedDb()
		ids := db.Table(entity.Photo{}.TableName()).Select("id").Where("photo_path = ?", folder).QueryExpr()
		_ = db.Where("photo_id IN (?)", ids).Delete(&entity.PhotoLabel{}).Error
		_ = db.Where("photo_id IN (?)", ids).Delete(&entity.Details{}).Error
		_ = db.Where("file_name LIKE ?", folder+"/%").Delete(&entity.File{}).Error
		_ = db.Where("photo_path = ?", folder).Delete(&entity.Photo{}).Error
	})

	return folder
}

// batchDeleteTestPhoto creates a photo with an original file in folder and returns it with the file's
// absolute path, so batch delete tests do not permanently remove shared fixtures.
func batchDeleteTestPhoto(t *testing.T, conf *config.Config, folder, name string) (*entity.Photo, string) {
	t.Helper()

	photo := &entity.Photo{
		PhotoPath:    folder,
		PhotoName:    name,
		PhotoType:    entity.MediaImage,
		PhotoQuality: 3,
	}

	require.NoError(t, photo.Create())

	fileName := filepath.Join(conf.OriginalsPath(), folder, name+".jpg")
	require.NoError(t, fs.Copy("./testdata/london_160x160.jpg", fileName, true))

	file := &entity.File{
		PhotoID:     photo.ID,
		PhotoUID:    photo.PhotoUID,
		FileName:    folder + "/" + name + ".jpg",
		FileRoot:    entity.RootOriginals,
		FileHash:    rnd.GenerateUID(entity.FileUID),
		FileType:    fs.ImageJpeg.String(),
		FileMime:    "image/jpeg",
		FilePrimary: true,
	}

	require.NoError(t, file.Create())

	return photo, fileName
}

// batchDeleteTestPhotoExists reports whether the photo is still indexed.
func batchDeleteTestPhotoExists(t *testing.T, photo *entity.Photo) bool {
	t.Helper()

	var count int
	require.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_uid = ?", photo.PhotoUID).Count(&count).Error)

	return count > 0
}

// assertBatchDeleteTestPhotoKept asserts that the photo, its file, and its original were left unchanged.
func assertBatchDeleteTestPhotoKept(t *testing.T, photo *entity.Photo, fileName string) {
	t.Helper()

	var result entity.Photo

	if !assert.NoError(t, entity.UnscopedDb().First(&result, "photo_uid = ?", photo.PhotoUID).Error, photo.PhotoName) {
		return
	}

	assert.Equal(t, photo.DeletedAt == nil, result.DeletedAt == nil, photo.PhotoName)
	assert.Equal(t, photo.PhotoQuality, result.PhotoQuality, photo.PhotoName)

	var files int
	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("photo_id = ? AND deleted_at IS NULL", result.ID).Count(&files).Error)
	assert.Equal(t, 1, files, photo.PhotoName)
	assert.FileExists(t, fileName)
}

func TestBatchPhotosDelete(t *testing.T) {
	t.Run("ErrNoItemsSelected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/delete", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NoneArchived", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosDelete(router)
		folder := batchDeleteTestFolder(t, conf)
		photo1, file1 := batchDeleteTestPhoto(t, conf, folder, "visible1")
		photo2, file2 := batchDeleteTestPhoto(t, conf, folder, "visible2")

		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/delete", fmt.Sprintf(`{"photos": [%q, %q]}`, photo1.PhotoUID, photo2.PhotoUID))
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), gjson.Get(r.Body.String(), "error").String())

		assertBatchDeleteTestPhotoKept(t, photo1, file1)
		assertBatchDeleteTestPhotoKept(t, photo2, file2)
	})
	t.Run("Mixed", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosDelete(router)
		folder := batchDeleteTestFolder(t, conf)
		archived, archivedFile := batchDeleteTestPhoto(t, conf, folder, "archived")
		visible, visibleFile := batchDeleteTestPhoto(t, conf, folder, "visible")
		require.NoError(t, archived.Archive())

		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/delete", fmt.Sprintf(`{"photos": [%q, %q]}`, visible.PhotoUID, archived.PhotoUID))
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, i18n.Msg(i18n.MsgPermanentlyDeleted), gjson.Get(r.Body.String(), "message").String())

		assert.False(t, batchDeleteTestPhotoExists(t, archived))
		assert.NoFileExists(t, archivedFile)
		assertBatchDeleteTestPhotoKept(t, visible, visibleFile)
	})
	t.Run("AllArchived", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosDelete(router)
		folder := batchDeleteTestFolder(t, conf)
		photo1, file1 := batchDeleteTestPhoto(t, conf, folder, "archived1")
		photo2, file2 := batchDeleteTestPhoto(t, conf, folder, "archived2")
		require.NoError(t, photo1.Archive())
		require.NoError(t, photo2.Archive())

		// The archive also lists pictures with the lowest quality score.
		require.NoError(t, photo2.Update("photo_quality", 0))

		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/delete", fmt.Sprintf(`{"photos": [%q, %q]}`, photo1.PhotoUID, photo2.PhotoUID))
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, i18n.Msg(i18n.MsgPermanentlyDeleted), gjson.Get(r.Body.String(), "message").String())

		assert.False(t, batchDeleteTestPhotoExists(t, photo1))
		assert.False(t, batchDeleteTestPhotoExists(t, photo2))
		assert.NoFileExists(t, file1)
		assert.NoFileExists(t, file2)
	})
	t.Run("OtherSelectionFields", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosDelete(router)
		folder := batchDeleteTestFolder(t, conf)
		selected, selectedFile := batchDeleteTestPhoto(t, conf, folder, "selected")
		labelArchived, labelArchivedFile := batchDeleteTestPhoto(t, conf, folder, "label-archived")
		labelVisible, labelVisibleFile := batchDeleteTestPhoto(t, conf, folder, "label-visible")
		albumVisible, albumVisibleFile := batchDeleteTestPhoto(t, conf, folder, "album-visible")

		label := entity.NewLabel("Batch Delete "+rnd.Base36(8), 0)
		require.NoError(t, label.Create())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(label).Error })

		for _, p := range []*entity.Photo{labelArchived, labelVisible} {
			require.NoError(t, entity.NewPhotoLabel(p.ID, label.ID, 0, entity.SrcManual).Create())
		}

		album := entity.NewAlbum("Batch Delete "+rnd.Base36(8), entity.AlbumManual)
		require.NoError(t, album.Create())
		t.Cleanup(func() {
			_ = entity.UnscopedDb().Where("album_uid = ?", album.AlbumUID).Delete(&entity.PhotoAlbum{}).Error
			_ = entity.UnscopedDb().Delete(album).Error
		})
		require.NoError(t, entity.NewPhotoAlbum(albumVisible.PhotoUID, album.AlbumUID).Create())

		require.NoError(t, labelArchived.Archive())

		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/delete",
			fmt.Sprintf(`{"photos": [%q], "labels": [%q], "albums": [%q]}`, selected.PhotoUID, label.LabelUID, album.AlbumUID))
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, i18n.Msg(i18n.MsgPermanentlyDeleted), gjson.Get(r.Body.String(), "message").String())

		assert.False(t, batchDeleteTestPhotoExists(t, labelArchived))
		assert.NoFileExists(t, labelArchivedFile)

		for p, fileName := range map[*entity.Photo]string{selected: selectedFile, labelVisible: labelVisibleFile, albumVisible: albumVisibleFile} {
			assertBatchDeleteTestPhotoKept(t, p, fileName)
		}

		var labels, albums int
		require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoLabel{}).Where("photo_id = ? AND label_id = ?", labelVisible.ID, label.ID).Count(&labels).Error)
		require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).Where("photo_uid = ? AND album_uid = ? AND hidden = 0", albumVisible.PhotoUID, album.AlbumUID).Count(&albums).Error)
		assert.Equal(t, 1, labels)
		assert.Equal(t, 1, albums)
	})
	t.Run("Removed", func(t *testing.T) {
		app, router, conf := NewApiTest()
		BatchPhotosDelete(router)
		folder := batchDeleteTestFolder(t, conf)
		removed, removedFile := batchDeleteTestPhoto(t, conf, folder, "removed")
		require.NoError(t, removed.Archive())
		require.NoError(t, removed.Update("photo_quality", -1))

		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/delete", fmt.Sprintf(`{"photos": [%q]}`, removed.PhotoUID))
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), gjson.Get(r.Body.String(), "error").String())

		assertBatchDeleteTestPhotoKept(t, removed, removedFile)
	})
}
