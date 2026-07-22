package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestSearchSubjects(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchSubjects(router)
		r := PerformRequest(app, "GET", "/api/v1/subjects?count=10")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(3), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchSubjects(router)
		r := PerformRequest(app, "GET", "/api/v1/subjects?xxx=10")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("Forbidden", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		SearchSubjects(router)

		// A session scoped to photos has no ResourcePeople grant. Subject search is now the
		// sole gate on people-name exposure (#5666), so it must be denied rather than leak
		// names to a caller that cannot act on them.
		sess, err := entity.AddClientSession("subjects-no-people", conf.SessionMaxAge(), "photos", authn.GrantClientCredentials, nil)
		require.NoError(t, err)

		r := AuthenticatedRequest(app, "GET", "/api/v1/subjects?count=10", sess.AuthToken())
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
