package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMarker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		UpdateMarker(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0y11")

		assert.Equal(t, http.StatusOK, r.Code)

		photoUID := gjson.Get(r.Body.String(), "UID").String()
		fileUID := gjson.Get(r.Body.String(), "Files.0.UID").String()
		markerID := gjson.Get(r.Body.String(), "Files.0.Markers.0.ID").String()

		assert.NotEmpty(t, photoUID)
		assert.NotEmpty(t, fileUID)
		assert.NotEmpty(t, markerID)

		u := fmt.Sprintf("/api/v1/photos/%s/files/%s/markers/%s", photoUID, fileUID, markerID)

		var m = struct {
			RefUID        string
			RefSrc        string
			MarkerSrc     string
			MarkerType    string
			MarkerScore   int
			MarkerInvalid bool
			MarkerLabel   string
		}{
			RefUID:        "3h59wvth837b5vyiub35",
			RefSrc:        "meta",
			MarkerSrc:     "image",
			MarkerType:    "Face",
			MarkerScore:   100,
			MarkerInvalid: true,
			MarkerLabel:   "Foo",
		}

		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("PUT %s", u)
			r = PerformRequestWithBody(app, "PUT", u, string(b))
		}

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("bad request non- primary file", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/pt9jtdre2lvl0y17/files/ft72s39w45bnlqdw/markers/12", "test")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("bad request file and photouid not matching", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/pt9jtdre2lvl0y16/files/fikjs39w45bnlqdw/markers/12", "test")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("file not existing", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/pt9jtdre2lvl0y17/files/fxxxx39w45bnlqdw/markers/1112", "test")

		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("marker not existing", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/pt9jtdre2lvl0y17/files/fikjs39w45bnlqdw/markers/1112", "test")

		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("empty photouid", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos//files/fikjs39w45bnlqdw/markers/1112", "test")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("update cluster with existing subject", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		var m = struct {
			ID         int
			Type       string
			Src        string
			Name       string
			SubjectUID string
			SubjectSrc string
			FaceID     string
		}{ID: 8,
			Type:       "face",
			Src:        "image",
			Name:       "Actress A",
			SubjectUID: "jqy1y111h1njaaac",
			SubjectSrc: "manual",
			FaceID:     "GMH5NISEEULNJL6RATITOA3TMZXMTMCI"}
		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/pt9jtdre2lvl0y12/files/ft3es39w45bnlqdw/markers/8", string(b))

			assert.Equal(t, http.StatusOK, r.Code)

			ClearMarkerSubject(router)

			r = PerformRequestWithBody(app, "DELETE", "/api/v1/photos/pt9jtdre2lvl0y12/files/ft3es39w45bnlqdw/markers/8/subject", "")

			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
	t.Run("update cluster with existing subject", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		var m = struct {
			ID         int
			Type       string
			Src        string
			Name       string
			SubjectUID string
			SubjectSrc string
			FaceID     string
		}{ID: 8,
			Type:       "face",
			Src:        "image",
			Name:       "Actress A",
			SubjectUID: "jqy1y111h1njaaac",
			SubjectSrc: "manual",
			FaceID:     "GMH5NISEEULNJL6RATITOA3TMZXMTMCI"}
		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/pt9jtdre2lvl0y12/files/ft3es39w45bnlqdw/markers/8", string(b))

			assert.Equal(t, http.StatusOK, r.Code)

			ClearMarkerSubject(router)

			r = PerformRequestWithBody(app, "DELETE", "/api/v1/photos/pt9jtdre2lvl0y12/files/ft3es39w45bnlqdw/markers/8/subject", "")

			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
	t.Run("invalid body", func(t *testing.T) {
		app, router, _ := NewApiTest()

		UpdateMarker(router)

		var m = struct {
			ID         int
			Type       string
			Src        int
			Name       int
			SubjectUID string
			SubjectSrc string
			FaceID     string
		}{ID: 8,
			Type:       "face",
			Src:        123,
			Name:       456,
			SubjectUID: "jqy1y111h1njaaac",
			SubjectSrc: "manual",
			FaceID:     "GMH5NISEEULNJL6RATITOA3TMZXMTMCI"}
		if b, err := json.Marshal(m); err != nil {
			t.Fatal(err)
		} else {
			r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/pt9jtdre2lvl0y12/files/ft3es39w45bnlqdw/markers/8", string(b))

			assert.Equal(t, http.StatusBadRequest, r.Code)
		}
	})
}

func TestClearMarkerSubject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPhoto(router)
		ClearMarkerSubject(router)

		photoResp := PerformRequest(app, "GET", "/api/v1/photos/pt9jtdre2lvl0y11")

		assert.Equal(t, http.StatusOK, photoResp.Code)

		photoUID := gjson.Get(photoResp.Body.String(), "UID").String()
		fileUID := gjson.Get(photoResp.Body.String(), "Files.0.UID").String()
		markerID := gjson.Get(photoResp.Body.String(), "Files.0.Markers.0.ID").String()

		assert.NotEmpty(t, photoUID)
		assert.NotEmpty(t, fileUID)
		assert.NotEmpty(t, markerID)

		u := fmt.Sprintf("/api/v1/photos/%s/files/%s/markers/%s/subject", photoUID, fileUID, markerID)

		t.Logf("DELETE %s", u)

		resp := PerformRequestWithBody(app, "DELETE", u, "")

		assert.Equal(t, http.StatusOK, resp.Code)
	})
	t.Run("bad request non- primary file", func(t *testing.T) {
		app, router, _ := NewApiTest()

		ClearMarkerSubject(router)

		r := PerformRequestWithBody(app, "DELETE", "/api/v1/photos/pt9jtdre2lvl0y17/files/ft72s39w45bnlqdw/markers/12/subject", "")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
