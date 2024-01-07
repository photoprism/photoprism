package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestGetFace(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFace(router)
		// Example:
		// {"ID":"PN6QO5INYTUSAATOFL43LL2ABAV5ACZK","Src":"","SubjUID":"js6sg6b1qekk9jx8","Samples":5,"SampleRadius":0.8,"Collisions":1,"CollisionRadius":0,"MatchedAt":null,"CreatedAt":"2021-09-18T12:06:39Z","UpdatedAt":"2021-09-18T12:06:39Z"}
		r := PerformRequest(app, "GET", "/api/v1/faces/TOSCDXCS4VI3PGIUTCNIQCNI6HSFXQVZ")
		t.Logf("GET /api/v1/faces/TOSCDXCS4VI3PGIUTCNIQCNI6HSFXQVZ: %s", r.Body.String())
		val := gjson.Get(r.Body.String(), "ID")
		assert.Equal(t, "TOSCDXCS4VI3PGIUTCNIQCNI6HSFXQVZ", val.String())
		val2 := gjson.Get(r.Body.String(), "Samples")
		assert.LessOrEqual(t, int64(4), val2.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("Lowercase", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFace(router)
		r := PerformRequest(app, "GET", "/api/v1/faces/PN6QO5INYTUSAATOFL43LL2ABAV5ACzk")
		val := gjson.Get(r.Body.String(), "ID")
		assert.Equal(t, "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", val.String())
		val2 := gjson.Get(r.Body.String(), "SubjUID")
		assert.Equal(t, "js6sg6b1qekk9jx8", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFace(router)
		r := PerformRequestWithBody(app, "GET", "/api/v1/faces/xxx", `{"Name": "Updated01", "Priority": 4, "Uncertainty": 80}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Face not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestUpdateFace(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateFace(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/faces/PN6QO5INYTUSAATOFL43LL2ABAV5ACzk", `{"SubjUID": "js6sg6b1qekk9jx8"}`)
		t.Logf("PUT /api/v1/faces/PN6QO5INYTUSAATOFL43LL2ABAV5ACzk: %s", r.Body.String())
		val := gjson.Get(r.Body.String(), "ID")
		assert.Equal(t, "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", val.String())
		val2 := gjson.Get(r.Body.String(), "SubjUID")
		assert.Equal(t, "js6sg6b1qekk9jx8", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateFace(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/faces/xxx", `{"Name": "Updated01", "Priority": 4, "Uncertainty": 80}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Face not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
