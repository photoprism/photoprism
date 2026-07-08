package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/oidc"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
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

		end := min(offset+count, len(items))

		page := items[offset:end]

		// Build response with session-based redaction.
		opts := reg.NodeOptsForSession(s)

		resp := reg.BuildClusterNodes(page, opts)

		// Audit list access.
		event.AuditDebug(
			[]string{ClientIP(c), "session %s", string(acl.ResourceCluster), "list nodes", "count %d offset %d returned %d", status.Succeeded},
			s.RefID,
			count,
			offset,
			len(resp),
		)

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
		event.AuditInfo(
			[]string{ClientIP(c), "session %s", string(acl.ResourceCluster), "get node", "%s", status.Succeeded},
			s.RefID,
			uuid,
		)

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
//	@Param		node				body		object	true	"properties to update (Role, DisplayName, Labels, AdvertiseUrl, SiteUrl, RedirectURIs, AllowGroups, AllowGroupRoles, GroupsFullView)"
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

		// Validate id to avoid path traversal and unexpected file access.
		if !isSafeNodeID(uuid) {
			AbortEntityNotFound(c)
			return
		}

		var req struct {
			Role            *string            `json:"Role"`
			DisplayName     *string            `json:"DisplayName"`
			Labels          map[string]string  `json:"Labels"`
			AdvertiseUrl    *string            `json:"AdvertiseUrl"`
			SiteUrl         *string            `json:"SiteUrl"`
			RedirectURIs    *[]string          `json:"RedirectURIs"`
			AllowGroups     *[]string          `json:"AllowGroups"`
			AllowGroupRoles *map[string]string `json:"AllowGroupRoles"`
			GroupsFullView  *bool              `json:"GroupsFullView"`
		}

		LimitRequestBodyBytes(c, MaxClusterRegisterBytes)

		if err := c.ShouldBindJSON(&req); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

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

		if req.Role != nil {
			role := cluster.NormalizeNodeRole(*req.Role)

			if role != cluster.RoleInstance && role != cluster.RoleService {
				AbortBadRequest(c, fmt.Errorf("invalid role"))
				return
			}

			n.Role = role
		}

		// An admin-set DisplayName (SrcManual) pins the value so it survives
		// later instance registrations; an empty value un-pins and falls back to
		// the instance-reported name.
		if req.DisplayName != nil {
			n.DisplayName = clean.TypeUnicode(strings.TrimSpace(*req.DisplayName))
			n.NameSrc = entity.SrcManual
		}

		if req.Labels != nil {
			n.Labels = req.Labels
		}

		if req.AdvertiseUrl != nil {
			advertise := strings.TrimSpace(*req.AdvertiseUrl)

			if advertise == "" {
				n.AdvertiseUrl = ""
			} else {
				if !validateAdvertiseURL(advertise) {
					AbortBadRequest(c, fmt.Errorf("invalid advertise url"))
					return
				}

				n.AdvertiseUrl = normalizeSiteURL(advertise)
			}
		}

		if req.SiteUrl != nil {
			siteUrl := strings.TrimSpace(*req.SiteUrl)

			if siteUrl == "" {
				n.SiteUrl = ""
			} else {
				if !validateSiteURL(siteUrl) {
					AbortBadRequest(c, fmt.Errorf("invalid site url"))
					return
				}

				n.SiteUrl = normalizeSiteURL(siteUrl)
			}
		}

		if req.RedirectURIs != nil {
			normalized, err := normalizeRedirectURIs(*req.RedirectURIs)
			if err != nil {
				AbortBadRequest(c, err)
				return
			}
			// Non-nil slice (even empty) replaces the persisted set; nil is "no change".
			n.RedirectURIs = normalized
		}

		if req.AllowGroups != nil {
			// Non-nil slice (even empty) replaces the persisted set; nil is "no change".
			normalized := oidc.MergeGroups(*req.AllowGroups)
			if normalized == nil {
				normalized = []string{}
			}
			n.AllowGroups = normalized
		}

		if req.AllowGroupRoles != nil {
			normalized, err := normalizeAllowGroupRoles(*req.AllowGroupRoles)
			if err != nil {
				AbortBadRequest(c, err)
				return
			}
			// Non-nil map (even empty) replaces the persisted mapping; nil is "no change".
			n.AllowGroupRoles = normalized
		}

		if req.GroupsFullView != nil {
			n.GroupsFullView = req.GroupsFullView
		}

		// An admin edit pins the group config so instance registrations can't
		// revert it; clearing all of it un-pins (see registry.applyGroupConfig).
		if req.AllowGroups != nil || req.AllowGroupRoles != nil || req.GroupsFullView != nil {
			n.GroupsSrc = entity.ClientGroupsSrcManual
		}

		n.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

		if err = regy.Put(n); err != nil {
			AbortUnexpectedError(c)
			return
		}

		event.AuditInfo(
			[]string{ClientIP(c), "session %s", string(acl.ResourceCluster), "node", "%s", status.Updated},
			s.RefID,
			uuid,
		)

		c.JSON(http.StatusOK, cluster.StatusResponse{Status: "ok"})
	})
}

// normalizeAllowGroupRoles validates a group → role mapping for an admin node
// PATCH: keys normalize via oidc.NormalizeGroupID (empty dropped) and roles must
// be cluster instance roles, so cluster_admin/visitor/unknown are rejected with an error.
func normalizeAllowGroupRoles(in map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(in))

	for group, roleName := range in {
		g := oidc.NormalizeGroupID(group)

		if g == "" {
			continue
		}

		role, ok := acl.ClusterInstanceRole(roleName)

		if !ok {
			return nil, fmt.Errorf("invalid role %s for group %s", clean.LogQuote(roleName), clean.LogQuote(group))
		}

		out[g] = role.String()
	}

	return out, nil
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

		event.AuditWarn(
			[]string{ClientIP(c), "session %s", string(acl.ResourceCluster), "node", "%s", status.Deleted},
			s.RefID,
			uuid,
		)

		c.JSON(http.StatusOK, cluster.StatusResponse{Status: "ok"})
	})
}
