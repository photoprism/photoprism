package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
)

// ClusterMetrics returns lightweight metrics about the cluster.
//
//	@Summary	temporary cluster metrics (counts only)
//	@Id			ClusterMetrics
//	@Tags		Cluster
//	@Produce	json
//	@Success	200			{object}	cluster.MetricsResponse
//	@Failure	401,403,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/metrics [get]
func ClusterMetrics(router *gin.RouterGroup) {
	router.GET("/cluster/metrics", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCluster, acl.ActionView)
		if s.Abort(c) {
			return
		}

		conf := get.Config()
		if !conf.Portal() {
			AbortFeatureDisabled(c)
			return
		}

		regy, err := reg.NewClientRegistryWithConfig(conf)
		if err != nil {
			AbortUnexpectedError(c)
			return
		}

		nodes, _ := regy.List()
		counts := map[string]int{"total": len(nodes)}
		for _, node := range nodes {
			role := node.Role
			if role == "" {
				role = "unknown"
			}
			counts[role]++
		}

		c.JSON(http.StatusOK, cluster.MetricsResponse{
			UUID:        conf.ClusterUUID(),
			ClusterCIDR: conf.ClusterCIDR(),
			Nodes:       counts,
			Time:        time.Now().UTC().Format(time.RFC3339),
		})
	})
}
