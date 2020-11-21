package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangePassword(t *testing.T) {
	t.Run("not existing user", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ChangePassword(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/users/xxx/password", `{}`)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
