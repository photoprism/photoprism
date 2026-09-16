package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestOIDCSessionEligible(t *testing.T) {
	conf := get.Config()
	user := entity.FindLocalUser("alice")
	require.NotNil(t, user)

	interactive := func(scope string) *entity.Session {
		sess := entity.NewSession(conf.SessionMaxAge(), 0)
		sess.SetProvider(authn.ProviderLocal)
		sess.SetMethod(authn.MethodDefault)
		sess.SetScope(scope)
		sess.SetUser(user)
		return sess
	}

	t.Run("Nil", func(t *testing.T) {
		assert.False(t, OIDCSessionEligible(nil))
	})
	t.Run("InteractiveUnscoped", func(t *testing.T) {
		assert.True(t, OIDCSessionEligible(interactive("*")))
	})
	t.Run("InteractiveWithoutAScope", func(t *testing.T) {
		// An empty scope is not a restriction, so the scope check must not apply.
		sess := interactive("*")
		sess.AuthScope = ""
		assert.True(t, OIDCSessionEligible(sess))
	})
	t.Run("InteractiveWithClusterScope", func(t *testing.T) {
		assert.True(t, OIDCSessionEligible(interactive("cluster")))
	})
	t.Run("InteractiveWithoutClusterScope", func(t *testing.T) {
		assert.False(t, OIDCSessionEligible(interactive("photos")))
	})
	t.Run("SessionWithoutAnAccount", func(t *testing.T) {
		sess := entity.NewSession(conf.SessionMaxAge(), 0)
		sess.SetProvider(authn.ProviderLocal)
		assert.False(t, OIDCSessionEligible(sess))
	})
	t.Run("Expired", func(t *testing.T) {
		sess := interactive("*")
		sess.Expires(time.Now().Add(-time.Hour))
		assert.False(t, OIDCSessionEligible(sess))
	})
	t.Run("AppPasswordUnrestricted", func(t *testing.T) {
		sess, err := entity.AddClientSession("admission-app-any", conf.SessionMaxAge(), "*", authn.GrantPassword, user)
		require.NoError(t, err)
		require.True(t, sess.IsApplication())
		assert.False(t, OIDCSessionEligible(sess))
	})
	t.Run("AppPasswordRestricted", func(t *testing.T) {
		sess, err := entity.AddClientSession("admission-app-photos", conf.SessionMaxAge(), "photos:read", authn.GrantPassword, user)
		require.NoError(t, err)
		assert.False(t, OIDCSessionEligible(sess))
	})
	t.Run("ClientCredentials", func(t *testing.T) {
		sess, err := entity.AddClientSession("admission-client", conf.SessionMaxAge(), "cluster", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		assert.False(t, OIDCSessionEligible(sess))
	})
	t.Run("ClientCredentialsWithUser", func(t *testing.T) {
		sess, err := entity.AddClientSession("admission-client-user", conf.SessionMaxAge(), "*", authn.GrantClientCredentials, user)
		require.NoError(t, err)
		assert.False(t, OIDCSessionEligible(sess))
	})
	// Copies so the shared fixture users other tests read stay untouched.
	withLogin := func(name string, canLogin bool) *entity.Session {
		found := entity.FindLocalUser(name)
		require.NotNil(t, found)

		u := *found
		u.CanLogin = canLogin

		sess := entity.NewSession(conf.SessionMaxAge(), 0)
		sess.SetProvider(authn.ProviderLocal)
		sess.SetMethod(authn.MethodDefault)
		sess.SetUser(&u)

		return sess
	}
	t.Run("UserCannotLogIn", func(t *testing.T) {
		assert.False(t, OIDCSessionEligible(withLogin("bob", false)))
	})
	t.Run("SuperAdminKeepsLogin", func(t *testing.T) {
		require.True(t, entity.FindLocalUser("alice").SuperAdmin)
		assert.True(t, OIDCSessionEligible(withLogin("alice", false)))
	})
	t.Run("DeletedAccount", func(t *testing.T) {
		deleted := entity.FindLocalUser("deleted")
		require.NotNil(t, deleted)

		sess := entity.NewSession(conf.SessionMaxAge(), 0)
		sess.SetProvider(authn.ProviderLink)
		sess.SetUser(deleted)

		assert.False(t, OIDCSessionEligible(sess))
	})
}
