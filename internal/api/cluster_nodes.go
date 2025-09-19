package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// isSafeNodeID validates that an ID contains only allowed characters to avoid path traversal.
// Allows: lowercase letters, digits, and dashes; length 1..64.
func isSafeNodeID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, r := range id {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '-' {
			continue
		}
		return false
	}
	return true
}

// ClusterListNodes lists registered nodes from the file-backed registry.
//
//	@Summary	lists registered nodes
//	@Id			ClusterListNodes
//	@Tags		Cluster
//	@Produce	json
//	@Param		count		query		int	false	"maximum number of results (default 100, max 1000)"	minimum(1)	maximum(1000)
//	@Param		offset		query		int	false	"result offset"										minimum(0)
//	@Success	200			{array}		cluster.Node
//	@Failure	401,403,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/nodes [get]
func ClusterListNodes(router *gin.RouterGroup) {
	router.GET("/cluster/nodes", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCluster, acl.ActionSearch)

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

		items, err := regy.List()

		if err != nil {
			AbortUnexpectedError(c)
			return
		}

		// Pagination: count (1..1000), offset (>=0)
		count, offset := 100, 0
		if v := c.Query("count"); v != "" {
			if n := txt.Int(v); n > 0 && n <= 1000 {
				count = n
			}
		}

		if v := c.Query("offset"); v != "" {
			if n := txt.Int(v); n >= 0 {
				offset = n
			}
		}

		if offset > len(items) {
			offset = len(items)
		}

		end := offset + count

		if end > len(items) {
			end = len(items)
		}

		page := items[offset:end]

		// Build response with session-based redaction.
		opts := reg.NodeOptsForSession(s)

		resp := reg.BuildClusterNodes(page, opts)

		// Audit list access.
		event.AuditInfo([]string{ClientIP(c), "session %s", string(acl.ResourceCluster), "nodes", "list", event.Succeeded, "count=%d", "offset=%d", "returned=%d"}, s.RefID, count, offset, len(resp))

		c.JSON(http.StatusOK, resp)
	})
}

// ClusterGetNode returns a single node by id.
//
//	@Summary	get node by id
//	@Id			ClusterGetNode
//	@Tags		Cluster
//	@Produce	json
//	@Param		id				path		string	true	"node id"
//	@Success	200				{object}	cluster.Node
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/nodes/{id} [get]
func ClusterGetNode(router *gin.RouterGroup) {
	router.GET("/cluster/nodes/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCluster, acl.ActionView)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if !conf.IsPortal() {
			AbortFeatureDisabled(c)
			return
		}

		id := c.Param("id")

		// Validate id to avoid path traversal and unexpected file access.
		if !isSafeNodeID(id) {
			AbortEntityNotFound(c)
			return
		}

		regy, err := reg.NewClientRegistryWithConfig(conf)

		if err != nil {
			AbortUnexpectedError(c)
			return
		}

		n, err := regy.Get(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		// Build response with session-based redaction.
		opts := reg.NodeOptsForSession(s)
		resp := reg.BuildClusterNode(*n, opts)

		// Audit get access.
		event.AuditInfo([]string{ClientIP(c), "session %s", string(acl.ResourceCluster), "nodes", "get", n.ID, event.Succeeded}, s.RefID)

		c.JSON(http.StatusOK, resp)
	})
}

// ClusterUpdateNode updates mutable fields: role, labels, advertiseUrl.
//
//	@Summary	update node fields
//	@Id			ClusterUpdateNode
//	@Tags		Cluster
//	@Accept		json
//	@Produce	json
//	@Param		id					path		string	true	"node id"
//	@Param		node				body		object	true	"properties to update (role, labels, advertiseUrl, siteUrl)"
//	@Success	200					{object}	cluster.StatusResponse
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/nodes/{id} [patch]
func ClusterUpdateNode(router *gin.RouterGroup) {
	router.PATCH("/cluster/nodes/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCluster, acl.ActionManage)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if !conf.IsPortal() {
			AbortFeatureDisabled(c)
			return
		}

		id := c.Param("id")

		var req struct {
			Role         string            `json:"role"`
			Labels       map[string]string `json:"labels"`
			AdvertiseUrl string            `json:"advertiseUrl"`
			SiteUrl      string            `json:"siteUrl"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			AbortBadRequest(c, err)
			return
		}

		regy, err := reg.NewClientRegistryWithConfig(conf)

		if err != nil {
			AbortUnexpectedError(c)
			return
		}

		n, err := regy.Get(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if req.Role != "" {
			n.Role = clean.TypeLowerDash(req.Role)
		}

		if req.Labels != nil {
			n.Labels = req.Labels
		}

		if req.AdvertiseUrl != "" {
			n.AdvertiseUrl = req.AdvertiseUrl
		}
		if s := normalizeSiteURL(req.SiteUrl); s != "" {
			n.SiteUrl = s
		}

		n.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

		if err = regy.Put(n); err != nil {
			AbortUnexpectedError(c)
			return
		}

		event.AuditInfo([]string{ClientIP(c), string(acl.ResourceCluster), "nodes", "update", n.ID, event.Succeeded})
		c.JSON(http.StatusOK, cluster.StatusResponse{Status: "ok"})
	})
}

// ClusterDeleteNode removes a node entry from the registry.
//
//	@Summary	delete node by id
//	@Id			ClusterDeleteNode
//	@Tags		Cluster
//	@Produce	json
//	@Param		id				path		string	true	"node id"
//	@Success	200				{object}	cluster.StatusResponse
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/nodes/{id} [delete]
func ClusterDeleteNode(router *gin.RouterGroup) {
	router.DELETE("/cluster/nodes/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCluster, acl.ActionManage)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if !conf.IsPortal() {
			AbortFeatureDisabled(c)
			return
		}

		id := c.Param("id")

		regy, err := reg.NewClientRegistryWithConfig(conf)

		if err != nil {
			AbortUnexpectedError(c)
			return
		}

		if _, err = regy.Get(id); err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err = regy.Delete(id); err != nil {
			AbortUnexpectedError(c)
			return
		}

		event.AuditInfo([]string{ClientIP(c), string(acl.ResourceCluster), "nodes", "delete", id, event.Succeeded})
		c.JSON(http.StatusOK, cluster.StatusResponse{Status: "ok"})
	})
}
