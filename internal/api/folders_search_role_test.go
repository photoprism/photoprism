package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// mixedPrincipalToken registers a client with the given role owned by the named user, opens a
// session for it and returns the token, so a request can carry the same principal a handler sees in
// production.
func mixedPrincipalToken(t *testing.T, role acl.Role, userName string) string {
	t.Helper()

	user := entity.FindUserByName(userName)
	require.NotNil(t, user)

	client := &entity.Client{
		ClientUID:    rnd.GenerateUID(entity.ClientUID),
		UserUID:      user.UserUID,
		UserName:     user.UserName,
		ClientName:   "role-test-" + role.String(),
		ClientRole:   role.String(),
		ClientType:   authn.ClientConfidential,
		AuthProvider: authn.ProviderClient.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "*",
		AuthExpires:  3600,
		AuthTokens:   5,
		AuthEnabled:  true,
	}

	require.NoError(t, client.Create())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(client).Error })

	sess := entity.NewSession(client.AuthExpires, 0).SetClient(client).SetGrantType(authn.GrantClientCredentials)
	sess.SetUser(user)
	require.NoError(t, sess.Create())

	// Deleted through the entity, since a database-only delete leaves the credential resolvable from
	// the session cache with its preview token still registered.
	t.Cleanup(func() { _ = sess.Delete() })

	require.True(t, sess.IsClient())
	require.Equal(t, role, sess.GetClientRole())

	return sess.AuthToken()
}

// TestSearchFolders_EffectiveRole covers what the folder listing answers a client session acting for
// a privileged account.
func TestSearchFolders_EffectiveRole(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	SearchFoldersOriginals(router)

	t.Run("MixedPrincipalRefused", func(t *testing.T) {
		// The client role carries no library access to files, so the listing refuses the session even
		// though the owning account is an administrator. AuthAny answers this, before the handler.
		token := mixedPrincipalToken(t, acl.RoleInstance, "alice")
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/folders/originals?files=true", token)

		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("ClientRolePermitted", func(t *testing.T) {
		// A client role that does carry library access reaches the listing, so the refusal above
		// follows from the role rather than from the session being a client session.
		token := mixedPrincipalToken(t, acl.RoleClient, "alice")
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/folders/originals?files=true", token)

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NoRoleInThisTableReachesTheListingWithoutPrivateAccess", func(t *testing.T) {
		// A property of this rule table rather than of the handler: every role with library access to
		// files also has private access to pictures, and an intersection reaches the listing only if
		// both of its roles do. The editions have roles where the two come apart, so the behavior case
		// lives in their packages rather than here.
		reached := 0

		for role := range acl.Rules[acl.ResourceFiles] {
			if !acl.Rules.Allow(acl.ResourceFiles, role, acl.AccessLibrary) {
				continue
			}

			reached++

			assert.True(t, acl.Rules.Allow(acl.ResourcePhotos, role, acl.AccessPrivate),
				"role %s reaches the folder listing without private access, so its exclusion is now reachable and wants a case", role)
		}

		assert.NotZero(t, reached, "no role reaches the folder listing, so this case is checking nothing")
	})
}
