package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetLabels(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetLabels(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/labels?count=15")
		t.Log(result.Body)
		assert.Equal(t, http.StatusOK, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetLabels(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/labels?xxx=10")
		t.Log(result.Body)

		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		t.Log(router)
		t.Log(ctx)
		result := PerformRequest(app, "GET", "/api/v1/labels?xxx=10")
		t.Log(result.Body)

		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}

func TestLikeLabel(t *testing.T) {
	t.Run("like not existing album", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikeLabel(router, ctx)

		result := PerformRequest(app, "POST", "/api/v1/label/8775789/like")
		t.Log(result.Body)
		//assert.Equal(t, http.StatusNotFound, result.Code)
	})

}

func TestDislikeLabel(t *testing.T) {
	t.Run("dislike not existing album", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikeLabel(router, ctx)

		result := PerformRequest(app, "DELETE", "/api/v1/labels/5678/like")
		t.Log(result.Body)
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
