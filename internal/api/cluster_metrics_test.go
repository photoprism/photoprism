package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestClusterMetrics_EmptyCounts(t *testing.T) {
	// Remove the fixture record
	require.NoError(t, entity.UnscopedDb().Delete(entity.Client{}, "client_uid = ?", entity.ClientFixtures.Get("node").ClientUID).Error)
	defer func() {
		require.NoError(t, entity.Db().Create(entity.ClientFixtures.Pointer("node")).Error)
	}()

	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)
	conf.Options().ClusterCIDR = "192.0.2.0/24"

	ClusterMetrics(router)
	token := AuthenticateAdmin(app, router)

	resp := AuthenticatedRequest(app, http.MethodGet, "/api/v1/cluster/metrics", token)
	assert.Equal(t, http.StatusOK, resp.Code)

	body := resp.Body.String()
	assert.Equal(t, "192.0.2.0/24", gjson.Get(body, "ClusterCIDR").String())
	assert.Equal(t, int64(0), gjson.Get(body, "Nodes.total").Int())
}
