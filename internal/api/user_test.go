package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestChangePassword(t *testing.T) {
	t.Run("not existing person", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ChangePassword(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/users/xxx/password", `{}`)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
