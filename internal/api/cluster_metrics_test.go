package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
)

// TestClusterMetrics_EmptyCounts checks the response for an empty node registry.
func TestClusterMetrics_EmptyCounts(t *testing.T) {
	var nodes []entity.Client
	require.NoError(t, entity.Db().Where("node_uuid <> ''").Find(&nodes).Error)
	t.Cleanup(func() {
		for _, node := range nodes {
			require.NoError(t, entity.Db().Model(&entity.Client{}).Where("client_uid = ?", node.ClientUID).UpdateColumn("node_uuid", node.NodeUUID).Error)
		}
	})
	require.NoError(t, entity.Db().Model(&entity.Client{}).Where("node_uuid <> ''").UpdateColumn("node_uuid", "").Error)

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
