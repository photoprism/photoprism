package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

// ClusterSummary returns a minimal overview of the cluster/portal.
//
//	@Summary	cluster summary
//	@Id			ClusterSummary
//	@Tags		Cluster
//	@Produce	json
//	@Success	200			{object}	cluster.SummaryResponse
//	@Failure	401,403,429	{object}	i18n.Response
//	@Router		/api/v1/cluster [get]
func ClusterSummary(router *gin.RouterGroup) {
	router.GET("/cluster", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCluster, acl.ActionView)
		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if !conf.IsPortal() {
			AbortFeatureDisabled(c)
			return
		}

		regy, err := reg.NewClientRegistryWithConfig(conf)

		if err != nil {
			AbortUnexpectedError(c)
			return
		}

		nodes, _ := regy.List()

		c.JSON(http.StatusOK, cluster.SummaryResponse{
			UUID:        conf.ClusterUUID(),
			ClusterCIDR: conf.ClusterCIDR(),
			Nodes:       len(nodes),
			Database:    cluster.DatabaseInfo{Driver: conf.DatabaseDriverName(), Host: conf.DatabaseHost(), Port: conf.DatabasePort()},
			Time:        time.Now().UTC().Format(time.RFC3339),
		})
	})
}

// ClusterHealth returns minimal health information.
//
//	@Summary	cluster health
//	@Id			ClusterHealth
//	@Tags		Cluster
//	@Produce	json
//	@Success	200			{object}	HealthResponse
//	@Failure	401,403,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/health [get]
func ClusterHealth(router *gin.RouterGroup) {
	router.GET("/cluster/health", func(c *gin.Context) {
		conf := get.Config()

		// Align headers with server-level health endpoints.
		c.Header(header.CacheControl, header.CacheControlNoStore)
		c.Header(header.AccessControlAllowOrigin, header.Any)

		// Return error if not a portal node.
		if !conf.IsPortal() {
			AbortFeatureDisabled(c)
			return
		}

		c.JSON(http.StatusOK, NewHealthResponse("ok"))
	})
}
