package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/form"
)

func TestCreateMarker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		CreateMarker(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		assert.Equal(t, http.StatusOK, r.Code)

		photoUid := gjson.Get(r.Body.String(), "UID").String()
		fileUid := gjson.Get(r.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(r.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := "/api/v1/markers"

		frm := form.Marker{
			FileUID:       fileUid,
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "",
			MarkerName:    "",
			MarkerReview:  false,
			MarkerInvalid: false,
		}

		if b, err := json.Marshal(frm); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("POST %s", u)
			r = PerformRequestWithBody(app, "POST", u, string(b))
		}

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("SuccessWithName", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		CreateMarker(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		assert.Equal(t, http.StatusOK, r.Code)

		photoUid := gjson.Get(r.Body.String(), "UID").String()
		fileUid := gjson.Get(r.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(r.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := "/api/v1/markers"

		frm := form.Marker{
			FileUID:       fileUid,
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Jens Mander",
			MarkerReview:  false,
			MarkerInvalid: false,
		}

		if b, err := json.Marshal(frm); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("POST %s", u)
			r = PerformRequestWithBody(app, "POST", u, string(b))
		}

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidArea", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		CreateMarker(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		assert.Equal(t, http.StatusOK, r.Code)

		photoUid := gjson.Get(r.Body.String(), "UID").String()
		fileUid := gjson.Get(r.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(r.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := "/api/v1/markers"

		frm := form.Marker{
			FileUID:       fileUid,
			MarkerType:    "face",
			X:             0.5,
			Y:             0.5,
			W:             0,
			H:             0,
			SubjSrc:       "",
			MarkerName:    "",
			MarkerReview:  false,
			MarkerInvalid: false,
		}

		if b, err := json.Marshal(frm); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("POST %s", u)
			r = PerformRequestWithBody(app, "POST", u, string(b))
		}

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestUpdateMarker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		UpdateMarker(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		assert.Equal(t, http.StatusOK, r.Code)

		photoUid := gjson.Get(r.Body.String(), "UID").String()
		fileUid := gjson.Get(r.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(r.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := fmt.Sprintf("/api/v1/markers/%s", markerUid)

		var m = form.Marker{
			SubjSrc:       "manual",
			MarkerInvalid: true,
			MarkerName:    "Foo",
		}

		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("PUT %s", u)
			r = PerformRequestWithBody(app, "PUT", u, string(b))
		}

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NonPrimaryFile", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowu1000", "test")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("bad request file and photouid not matching", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowu1000", "test")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("file not existing", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/1112", "test")

		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("marker not existing", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/1112", "test")

		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("empty photouid", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowu1000", "test")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("update cluster with existing subject", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		var m = form.Marker{
			SubjSrc:       "manual",
			MarkerInvalid: false,
			MarkerName:    "Actress A",
		}

		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowuy666", string(b))

			assert.Equal(t, http.StatusOK, r.Code)

			ClearMarkerSubject(router)

			r = PerformRequestWithBody(app, "DELETE", "/api/v1/markers/ms6sg6b1wowuy666/subject", "")

			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
	t.Run("update cluster with existing subject 2", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		var m = form.Marker{
			SubjSrc:       "manual",
			MarkerInvalid: false,
			MarkerName:    "Actress A",
		}

		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowuy666", string(b))

			assert.Equal(t, http.StatusOK, r.Code)

			ClearMarkerSubject(router)

			r = PerformRequestWithBody(app, "DELETE", "/api/v1/markers/ms6sg6b1wowuy666/subject", "")

			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
	t.Run("invalid body", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		var m = struct {
			ID      int
			Type    string
			Src     int
			Name    int
			SubjUID string
			SubjSrc string
			FaceID  string
		}{ID: 8,
			Type:    "face",
			Src:     123,
			Name:    456,
			SubjUID: "js6sg6b1h1njaaac",
			SubjSrc: "manual",
			FaceID:  "GMH5NISEEULNJL6RATITOA3TMZXMTMCI"}
		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			r := PerformRequestWithBody(app, "PUT", "/api/v1/markers/ms6sg6b1wowuy666", string(b))

			assert.Equal(t, http.StatusBadRequest, r.Code)
		}
	})
}

func TestClearMarkerSubject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		ClearMarkerSubject(router)

		photoResp := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")

		if photoResp == nil {
			t.Fatal("response is nil")
		}

		assert.Equal(t, http.StatusOK, photoResp.Code)

		if photoResp.Body.String() == "" {
			t.Fatal("body is empty")
		}

		photoUid := gjson.Get(photoResp.Body.String(), "UID").String()
		fileUid := gjson.Get(photoResp.Body.String(), "Files.0.UID").String()
		markerUid := gjson.Get(photoResp.Body.String(), "Files.0.Markers.0.UID").String()

		assert.NotEmpty(t, photoUid)
		assert.NotEmpty(t, fileUid)
		assert.NotEmpty(t, markerUid)

		u := fmt.Sprintf("/api/v1/markers/%s/subject", markerUid)

		// t.Logf("DELETE %s", u)

		resp := PerformRequestWithBody(app, "DELETE", u, "")

		assert.Equal(t, http.StatusOK, resp.Code)
	})
	t.Run("non-primary file", func(t *testing.T) {
		app, router, _ := NewApiTest()

		ClearMarkerSubject(router)

		r := PerformRequestWithBody(app, "DELETE", "/api/v1/markers/ms6sg6b1wowu1000/subject", "")

		assert.Equal(t, http.StatusOK, r.Code)
	})
}
