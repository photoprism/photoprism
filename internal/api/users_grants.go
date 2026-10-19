package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// UserGrant represents an app connected to a user account.
type UserGrant struct {
	UID        string    `json:"UID"`
	ClientID   string    `json:"ClientID"`
	ClientName string    `json:"ClientName"`
	Origin     string    `json:"Origin"`
	Scope      string    `json:"Scope"`
	CreatedAt  time.Time `json:"CreatedAt"`
	UsedAt     time.Time `json:"UsedAt"`
}

// NewUserGrant returns the connected app of a grant.
func NewUserGrant(m entity.Grant) UserGrant {
	return UserGrant{
		UID:        m.GrantUID,
		ClientID:   m.ClientID,
		ClientName: m.ClientName,
		Origin:     oauth.ClientOrigin(m.ClientID, m.RedirectURI),
		Scope:      m.Scope,
		CreatedAt:  m.CreatedAt,
		UsedAt:     m.UsedAt,
	}
}

// FindUserGrants returns the apps connected to a user account.
//
//	@Summary	list the apps connected to a user account
//	@Id			FindUserGrants
//	@Tags		Users, Authentication
//	@Produce	json
//	@Param		uid					path		string	true	"user uid"
//	@Success	200					{object}	[]UserGrant
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Router		/api/v1/users/{uid}/grants [get]
func FindUserGrants(router *gin.RouterGroup) {
	router.GET("/users/:uid/grants", func(c *gin.Context) {
		s, uid, ok := userGrantsAuth(c)

		if !ok {
			return
		}

		grants, err := entity.FindUserGrants(uid)

		if err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "find grants of user %s", status.Error(err)}, s.RefID, clean.Log(uid))
			AbortUnexpectedError(c)
			return
		}

		result := make([]UserGrant, 0, len(grants))

		for i := range grants {
			result = append(result, NewUserGrant(grants[i]))
		}

		c.JSON(http.StatusOK, result)
	})
}

// DeleteUserGrant disconnects an app from a user account by revoking its grant, tokens and MCP sessions.
//
//	@Summary	disconnect an app from a user account
//	@Id			DeleteUserGrant
//	@Tags		Users, Authentication
//	@Produce	json
//	@Param		uid					path		string	true	"user uid"
//	@Param		grant				path		string	true	"grant uid"
//	@Success	200					{object}	UserGrant
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Router		/api/v1/users/{uid}/grants/{grant} [delete]
func DeleteUserGrant(router *gin.RouterGroup) {
	router.DELETE("/users/:uid/grants/:grant", func(c *gin.Context) {
		s, uid, ok := userGrantsAuth(c)

		if !ok {
			return
		}

		clientIp := ClientIP(c)
		grant := entity.FindGrant(clean.UID(c.Param("grant")))

		if grant == nil || grant.UserUID != uid {
			Abort(c, http.StatusNotFound, i18n.ErrNotFound)
			return
		}

		if err := grant.Delete(); err != nil {
			event.AuditErr([]string{clientIp, "session %s", "revoke grant %s", status.Error(err)}, s.RefID, clean.Log(grant.GrantUID))
			AbortUnexpectedError(c)
			return
		}

		event.AuditInfo([]string{clientIp, "session %s", "client %s", "user %s", "revoke grant %s", status.Deleted}, s.RefID, clean.Log(grant.ClientID), clean.Log(grant.UserName), clean.Log(grant.GrantUID))

		c.JSON(http.StatusOK, NewUserGrant(*grant))
	})
}

// userGrantsAuth authorizes access to the connected apps of the user in the request path: a user may manage
// their own apps, and an admin who may manage all sessions those of other users.
func userGrantsAuth(c *gin.Context) (s *entity.Session, uid string, ok bool) {
	s = AuthAny(c, acl.ResourceSessions, acl.Permissions{acl.ActionManage, acl.ActionManageOwn, acl.ActionView, acl.ActionDelete})

	if s.Abort(c) {
		return s, "", false
	}

	conf := get.Config()
	uid = clean.UID(c.Param("uid"))

	switch {
	case conf.Public() || conf.Demo() || !rnd.IsUID(uid, entity.UserUID):
		Abort(c, http.StatusNotFound, i18n.ErrNotFound)
		return s, uid, false
	case s.UserUID == uid && s.HasRegisteredUser():
		return s, uid, true
	case !s.IsClient() && s.HasRegisteredUser() && acl.Rules.AllowAll(acl.ResourceSessions, s.GetUserRole(), acl.Permissions{acl.AccessAll, acl.ActionManage}):
		return s, uid, true
	}

	event.AuditWarn([]string{ClientIP(c), "session %s", "access grants of user %s", status.Denied}, s.RefID, clean.Log(uid))
	Abort(c, http.StatusForbidden, i18n.ErrForbidden)

	return s, uid, false
}
