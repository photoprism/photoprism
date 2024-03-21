package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestAddPhotoLabel(t *testing.T) {
	t.Run("add new label", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotoLabel(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh8/label", `{"Name": "testAddLabel", "Uncertainty": 95, "Priority": 2}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), "TestAddLabel")
	})
	t.Run("add existing label", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotoLabel(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh8/label", `{"Name": "Flower", "Uncertainty": 10, "Priority": 2}`)
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Labels.#(LabelID==1000001).Uncertainty")
		assert.Equal(t, "10", val.String())
	})
	t.Run("not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotoLabel(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/xxx/label", `{"Name": "Flower", "Uncertainty": 10, "Priority": 2}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotoLabel(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh8/label", `{"Name": 123, "Uncertainty": 10, "Priority": 2}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

}

func TestRemovePhotoLabel(t *testing.T) {
	t.Run("photo with label", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/label/1000001")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Labels.#(LabelID==1000001).Uncertainty")
		assert.Equal(t, "100", val.String())
		assert.Contains(t, r.Body.String(), "cake")

	})
	t.Run("remove manually added label", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/label/1000002")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Labels")
		assert.NotContains(t, val.String(), "cake")
	})
	t.Run("photo not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/xxx/label/10000001")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("label not existing", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/label/xxx")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("try to remove wrong label", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/label/1000000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Record not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("not existing photo", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/xx/label/")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestUpdatePhotoLabel(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh0/label/1000006", `{"Label": {"Name": "NewLabelName"}}`)
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Title")
		assert.Contains(t, val.String(), "NewLabelName")
	})
	t.Run("photo not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/xxx/label/1000006", `{"Label": {"Name": "NewLabelName"}}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
	})
	t.Run("label not existing", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh0/label/9000006", `{"Label": {"Name": "NewLabelName"}}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("label not linked to photo", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh0/label/1000005", `{"Label": {"Name": "NewLabelName"}}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("bad request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh0/label/1000006", `{"Label": {"Name": 123}}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
