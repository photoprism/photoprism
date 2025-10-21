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

		if !conf.Portal() {
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
		event.AuditDebug([]string{
			ClientIP(c),
			"session %s",
			string(acl.ResourceCluster),
			"list nodes",
			"count %d offset %d returned %d",
			event.Succeeded,
		}, s.RefID, count, offset, len(resp))

		c.JSON(http.StatusOK, resp)
	})
}

// ClusterGetNode returns a single node by uuid.
//
//	@Summary	get node by uuid
//	@Id			ClusterGetNode
//	@Tags		Cluster
//	@Produce	json
//	@Param		uuid			path		string	true	"node uuid"
//	@Success	200				{object}	cluster.Node
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/nodes/{uuid} [get]
func ClusterGetNode(router *gin.RouterGroup) {
	router.GET("/cluster/nodes/:uuid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCluster, acl.ActionView)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if !conf.Portal() {
			AbortFeatureDisabled(c)
			return
		}

		uuid := c.Param("uuid")

		// Validate id to avoid path traversal and unexpected file access.
		if !isSafeNodeID(uuid) {
			AbortEntityNotFound(c)
			return
		}

		regy, err := reg.NewClientRegistryWithConfig(conf)

		if err != nil {
			AbortUnexpectedError(c)
			return
		}

		// Prefer NodeUUID identifier for cluster nodes.
		n, err := regy.FindByNodeUUID(uuid)
		if err != nil || n == nil {
			AbortEntityNotFound(c)
			return
		}

		// Build response with session-based redaction.
		opts := reg.NodeOptsForSession(s)
		resp := reg.BuildClusterNode(*n, opts)

		// Audit get access.
		event.AuditInfo([]string{
			ClientIP(c),
			"session %s",
			string(acl.ResourceCluster),
			"get node %s",
			event.Succeeded,
		}, s.RefID, uuid)

		c.JSON(http.StatusOK, resp)
	})
}

// ClusterUpdateNode updates mutable fields: role, labels, AdvertiseUrl.
//
//	@Summary	update node fields
//	@Id			ClusterUpdateNode
//	@Tags		Cluster
//	@Accept		json
//	@Produce	json
//	@Param		uuid				path		string	true	"node uuid"
//	@Param		node				body		object	true	"properties to update (Role, Labels, AdvertiseUrl, SiteUrl)"
//	@Success	200					{object}	cluster.StatusResponse
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/nodes/{uuid} [patch]
func ClusterUpdateNode(router *gin.RouterGroup) {
	router.PATCH("/cluster/nodes/:uuid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCluster, acl.ActionManage)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if !conf.Portal() {
			AbortFeatureDisabled(c)
			return
		}

		uuid := c.Param("uuid")

		var req struct {
			Role         string            `json:"Role"`
			Labels       map[string]string `json:"Labels"`
			AdvertiseUrl string            `json:"AdvertiseUrl"`
			SiteUrl      string            `json:"SiteUrl"`
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

		// Resolve by NodeUUID first (preferred).
		n, err := regy.FindByNodeUUID(uuid)
		if err != nil || n == nil {
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

		if u := normalizeSiteURL(req.SiteUrl); u != "" {
			n.SiteUrl = u
		}

		n.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

		if err = regy.Put(n); err != nil {
			AbortUnexpectedError(c)
			return
		}

		event.AuditInfo([]string{
			ClientIP(c),
			"session %s",
			string(acl.ResourceCluster),
			"node %s",
			event.Updated,
		}, s.RefID, uuid)

		c.JSON(http.StatusOK, cluster.StatusResponse{Status: "ok"})
	})
}

// ClusterDeleteNode removes a node entry from the registry.
//
//	@Summary	delete node by uuid
//	@Id			ClusterDeleteNode
//	@Tags		Cluster
//	@Produce	json
//	@Param		uuid			path		string	true	"node uuid"
//	@Success	200				{object}	cluster.StatusResponse
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/nodes/{uuid} [delete]
func ClusterDeleteNode(router *gin.RouterGroup) {
	router.DELETE("/cluster/nodes/:uuid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCluster, acl.ActionManage)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if !conf.Portal() {
			AbortFeatureDisabled(c)
			return
		}

		uuid := c.Param("uuid")
		// Validate uuid format to avoid path traversal or unexpected input.
		if !isSafeNodeID(uuid) {
			AbortEntityNotFound(c)
			return
		}

		regy, err := reg.NewClientRegistryWithConfig(conf)

		if err != nil {
			AbortUnexpectedError(c)
			return
		}

		// Delete by NodeUUID
		if err = regy.Delete(uuid); err != nil {
			if err == reg.ErrNotFound {
				AbortEntityNotFound(c)
			} else {
				AbortUnexpectedError(c)
			}
			return
		}

		event.AuditWarn([]string{
			ClientIP(c),
			"session %s",
			string(acl.ResourceCluster),
			"node %s",
			event.Deleted,
		}, s.RefID, uuid)

		c.JSON(http.StatusOK, cluster.StatusResponse{Status: "ok"})
	})
}
