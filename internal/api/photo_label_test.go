package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestAddPhotoLabel(t *testing.T) {
	t.Run("AddNewLabel", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotoLabel(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh8/label", `{"Name": "testAddLabel", "Uncertainty": 95, "Priority": 2}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), "TestAddLabel")
	})
	t.Run("AddExistingLabel", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotoLabel(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh8/label", `{"Name": "Flower", "Uncertainty": 10, "Priority": 2}`)
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Labels.#(LabelID==1000001).Uncertainty")
		assert.Equal(t, "10", val.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotoLabel(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/xxx/label", `{"Name": "Flower", "Uncertainty": 10, "Priority": 2}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		AddPhotoLabel(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh8/label", `{"Name": 123, "Uncertainty": 10, "Priority": 2}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

}

func TestRemovePhotoLabel(t *testing.T) {
	t.Run("PhotoWithLabel", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/label/1000001")
		assert.Equal(t, http.StatusOK, r.Code)
		uncertainty := gjson.Get(r.Body.String(), "Labels.#(LabelID==1000001).Uncertainty")
		src := gjson.Get(r.Body.String(), "Labels.#(LabelID==1000001).LabelSrc")
		name := gjson.Get(r.Body.String(), "Labels.#(LabelID==1000001).Label.Name")
		assert.Equal(t, "100", uncertainty.String())
		assert.Equal(t, "manual", src.String())
		assert.Equal(t, "Flower", name.String())
		assert.Contains(t, r.Body.String(), "cake")
	})
	t.Run("RemoveManuallyAddedLabel", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/label/1000002")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Labels")
		assert.NotContains(t, val.String(), "cake")
	})
	t.Run("PhotoNotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/xxx/label/10000001")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("LabelNotExisting", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/label/xxx")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("TryToRemoveWrongLabel", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/label/1000000")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Record not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("NotExistingPhoto", func(t *testing.T) {
		app, router, _ := NewApiTest()
		RemovePhotoLabel(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/xx/label/")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestUpdatePhotoLabel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh0/label/1000006", `{"Label": {"Name": "NewLabelName"}}`)
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Title")
		assert.Contains(t, val.String(), "NewLabelName")
	})
	t.Run("ReactivateRemovedLabel", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh9/label/1000003", `{"Uncertainty": 0}`)
		uncertainty := gjson.Get(r.Body.String(), "Labels.#(LabelID==1000003).Uncertainty")
		src := gjson.Get(r.Body.String(), "Labels.#(LabelID==1000003).LabelSrc")
		name := gjson.Get(r.Body.String(), "Labels.#(LabelID==1000003).Label.Name")
		assert.Equal(t, "0", uncertainty.String())
		assert.Equal(t, "manual", src.String())
		assert.Equal(t, "COW", name.String())
	})
	t.Run("photo not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/xxx/label/1000006", `{"Label": {"Name": "NewLabelName"}}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
	})
	t.Run("LabelNotExisting", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh0/label/9000006", `{"Label": {"Name": "NewLabelName"}}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("LabelNotLinkedToPhoto", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh0/label/1000005", `{"Label": {"Name": "NewLabelName"}}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhotoLabel(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0yh0/label/1000006", `{"Label": {"Name": 123}}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
