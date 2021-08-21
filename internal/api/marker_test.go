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
}
