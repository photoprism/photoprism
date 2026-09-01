package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/media"
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

func TestBatchPhotosStack(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		primary := entity.NewPhoto(true)
		primary.PhotoUID = rnd.GenerateUID(entity.PhotoUID)
		primary.PhotoPath = "api-stack"
		primary.PhotoName = "Primary"
		primary.PhotoType = entity.MediaImage
		primary.TypeSrc = entity.SrcAuto

		if err := primary.Create(); err != nil {
			t.Fatal(err)
		}

		secondary := entity.NewPhoto(true)
		secondary.PhotoUID = rnd.GenerateUID(entity.PhotoUID)
		secondary.PhotoPath = "api-stack"
		secondary.PhotoName = "Secondary"
		secondary.PhotoType = entity.MediaImage
		secondary.TypeSrc = entity.SrcAuto

		if err := secondary.Create(); err != nil {
			t.Fatal(err)
		}

		primaryFile := entity.File{
			PhotoID:     primary.ID,
			PhotoUID:    primary.PhotoUID,
			FileUID:     rnd.GenerateUID(entity.FileUID),
			FileName:    "api-stack/" + primary.PhotoUID + ".jpg",
			FileRoot:    entity.RootOriginals,
			FileHash:    "api-stack-primary-" + primary.PhotoUID,
			FileType:    "jpg",
			MediaType:   media.Image.String(),
			FilePrimary: true,
		}

		if err := primaryFile.Create(); err != nil {
			t.Fatal(err)
		}

		secondaryFile := entity.File{
			PhotoID:     secondary.ID,
			PhotoUID:    secondary.PhotoUID,
			FileUID:     rnd.GenerateUID(entity.FileUID),
			FileName:    "api-stack/" + secondary.PhotoUID + ".jpg",
			FileRoot:    entity.RootOriginals,
			FileHash:    "api-stack-secondary-" + secondary.PhotoUID,
			FileType:    "jpg",
			MediaType:   media.Image.String(),
			FilePrimary: true,
		}

		if err := secondaryFile.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			fileUIDs := []string{primaryFile.FileUID, secondaryFile.FileUID}
			photoIDs := []uint{primary.ID, secondary.ID}

			_ = entity.UnscopedDb().Where("file_uid IN (?)", fileUIDs).Delete(entity.File{}).Error
			_ = entity.UnscopedDb().Where("photo_id IN (?)", photoIDs).Delete(entity.Details{}).Error
			_ = entity.UnscopedDb().Where("id IN (?)", photoIDs).Delete(entity.Photo{}).Error
		})

		app, router, _ := NewApiTest()
		BatchPhotosStack(router)

		body := fmt.Sprintf(`{"photos":["%s","%s"]}`, primary.PhotoUID, secondary.PhotoUID)
		response := PerformRequestWithBody(app, http.MethodPost, "/api/v1/batch/photos/stack", body)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, primary.PhotoUID, gjson.Get(response.Body.String(), "photo").String())
		assert.Equal(t, secondary.PhotoUID, gjson.Get(response.Body.String(), "stacked.0").String())

		var files entity.Files
		assert.NoError(t, entity.UnscopedDb().Where("file_uid IN (?)", []string{primaryFile.FileUID, secondaryFile.FileUID}).Find(&files).Error)
		assert.Len(t, files, 2)
		for _, file := range files {
			assert.Equal(t, primary.ID, file.PhotoID)
		}
	})

	t.Run("RequiresTwoPictures", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosStack(router)

		response := PerformRequestWithBody(app, http.MethodPost, "/api/v1/batch/photos/stack", `{"photos":["ps6sg6be2lvl0yh8"]}`)

		assert.Equal(t, http.StatusBadRequest, response.Code)
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

func TestBatchPhotosDelete(t *testing.T) {
	t.Run("ErrNoItemsSelected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosDelete(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/delete", `{"photos": []}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
