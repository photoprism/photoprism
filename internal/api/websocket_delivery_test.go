package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// deliveryUser returns a user value with the given role and UID, as the connection map holds one.
func deliveryUser(role acl.Role, uid string) entity.User {
	return entity.User{UserUID: uid, UserName: "delivery-" + role.String(), UserRole: role.String()}
}

func TestWsRelease(t *testing.T) {
	const connId = "connection-under-release"

	t.Run("EveryMap", func(t *testing.T) {
		wsAuth.mutex.Lock()
		wsAuth.sid[connId] = "sid"
		wsAuth.rid[connId] = "rid"
		wsAuth.user[connId] = deliveryUser(acl.RoleAdmin, "us6sg6bxpogaaba1")
		wsAuth.client[connId] = wsClientRole{Role: acl.RoleInstance, Present: true}
		wsAuth.mutex.Unlock()

		wsRelease(connId)

		wsAuth.mutex.RLock()
		defer wsAuth.mutex.RUnlock()

		assert.NotContains(t, wsAuth.sid, connId)
		assert.NotContains(t, wsAuth.rid, connId)
		assert.NotContains(t, wsAuth.user, connId)
		assert.NotContains(t, wsAuth.client, connId)
	})
	t.Run("UnknownConnection", func(t *testing.T) {
		assert.NotPanics(t, func() { wsRelease("connection-never-opened") })
	})
}

func TestWsSessionClient(t *testing.T) {
	t.Run("NilSession", func(t *testing.T) {
		assert.Equal(t, wsClientRole{}, wsSessionClient(nil))
	})
	t.Run("UserSession", func(t *testing.T) {
		sess := (&entity.Session{}).SetUser(entity.UserFixtures.Pointer("alice"))
		assert.Equal(t, wsClientRole{}, wsSessionClient(sess))
	})
	t.Run("AppPassword", func(t *testing.T) {
		// An app password belongs to a client or an agent rather than to a person at a browser. It
		// records no client of its own, so it takes the default client role.
		sess := entity.NewClientSession("app", 3600, "*", authn.GrantPassword, entity.UserFixtures.Pointer("alice"))

		require.False(t, sess.HasClient())
		assert.Equal(t, wsClientRole{Role: acl.RoleClient, Present: true}, wsSessionClient(sess))
	})
	t.Run("AccessToken", func(t *testing.T) {
		sess := entity.NewClientSession("token", 3600, "*", authn.GrantClientCredentials, nil)

		assert.Equal(t, wsClientRole{Role: acl.RoleClient, Present: true}, wsSessionClient(sess))
	})
	t.Run("RegisteredClientWithTheDefaultRole", func(t *testing.T) {
		client := &entity.Client{
			ClientUID:    rnd.GenerateUID(entity.ClientUID),
			ClientName:   "default-role",
			ClientRole:   acl.RoleClient.String(),
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   authn.MethodOAuth2.String(),
		}

		assert.Equal(t, wsClientRole{Role: acl.RoleClient, Present: true}, wsSessionClient((&entity.Session{}).SetClient(client)))
	})
	t.Run("ClientWithAnotherProvider", func(t *testing.T) {
		// The provider decides, and reading the role rewrites it from the client record, so the two are
		// read in that order. Here the record names a provider that is not a client one, and the REST
		// handlers read the session the same way.
		client := &entity.Client{
			ClientUID:    rnd.GenerateUID(entity.ClientUID),
			ClientName:   "provider-test",
			ClientRole:   acl.RoleInstance.String(),
			AuthProvider: authn.ProviderLocal.String(),
			AuthMethod:   authn.MethodOAuth2.String(),
		}
		sess := (&entity.Session{}).SetClient(client)

		require.False(t, sess.IsClient())
		assert.Equal(t, wsClientRole{}, wsSessionClient(sess))
	})
}

func TestWsDelivery(t *testing.T) {
	const (
		userUID   = "us6sg6bxpogaaba1"
		sessionID = "sessionidfordelivery"
	)

	admin := deliveryUser(acl.RoleAdmin, userUID)
	guest := deliveryUser(acl.RoleGuest, userUID)

	t.Run("UserSessionReceivesItsChannels", func(t *testing.T) {
		// A session with no client of its own is decided by the account role alone, which is what an
		// ordinary browser login is.
		ev, ok := wsDelivery("log.info", sessionID, admin, wsClientRole{})
		assert.Equal(t, "log.info", ev)
		assert.True(t, ok)
	})
	t.Run("RegisteredClientReceivesNothing", func(t *testing.T) {
		// Clients and agents read through the REST API, so a client session receives no events at all -
		// including under the admin role, which an operator assigns to lift a client's other limits.
		for _, role := range []acl.Role{acl.RoleClient, acl.RoleAdmin, acl.RoleInstance, acl.RoleService, acl.RolePortal} {
			_, ok := wsDelivery("log.info", sessionID, admin, wsClientRole{Role: role, Present: true})
			assert.False(t, ok, "role %s", role)
		}
	})
	t.Run("InstanceClientForAdminAccount", func(t *testing.T) {
		for _, topic := range []string{"log.info", "log.error", "audit.log.info", "notify.info", "index.updated"} {
			_, ok := wsDelivery(topic, sessionID, admin, wsClientRole{Role: acl.RoleInstance, Present: true})
			assert.False(t, ok, "topic %s", topic)
		}
	})
	t.Run("ServiceClientForAdminAccount", func(t *testing.T) {
		for _, topic := range []string{"log.info", "audit.log.info", "photos.updated"} {
			_, ok := wsDelivery(topic, sessionID, admin, wsClientRole{Role: acl.RoleService, Present: true})
			assert.False(t, ok, "topic %s", topic)
		}
	})
	t.Run("ClientWithAnUnlistedRole", func(t *testing.T) {
		// A client role this edition does not list resolves to no role at all. The session still
		// authenticates a client, which is what decides.
		_, ok := wsDelivery("log.info", sessionID, admin, wsClientRole{Present: true})
		assert.False(t, ok)
	})
	t.Run("NarrowAccountLimitsItsOwnSession", func(t *testing.T) {
		// A narrow account is held to its own role, with no client involved.
		_, ok := wsDelivery("log.info", sessionID, guest, wsClientRole{})
		assert.False(t, ok)
	})
	t.Run("AddressedUserTopic", func(t *testing.T) {
		ev, ok := wsDelivery("user."+userUID+".albums.updated", sessionID, guest, wsClientRole{})
		assert.Equal(t, "albums.updated", ev)
		assert.True(t, ok)
	})
	t.Run("AddressedUserTopicRefusedForInstanceClient", func(t *testing.T) {
		ev, ok := wsDelivery("user."+userUID+".albums.updated", sessionID, guest, wsClientRole{Role: acl.RoleInstance, Present: true})
		assert.Equal(t, "albums.updated", ev)
		assert.False(t, ok)
	})
	t.Run("AddressedUserTopicOfAnotherAccount", func(t *testing.T) {
		_, ok := wsDelivery("user.us6sg6bxpogaaba2.albums.updated", sessionID, guest, wsClientRole{})
		assert.False(t, ok)
	})
	t.Run("AddressedSessionTopic", func(t *testing.T) {
		ev, ok := wsDelivery("session."+sessionID+".albums.updated", sessionID, guest, wsClientRole{})
		assert.Equal(t, "albums.updated", ev)
		assert.True(t, ok)
	})
	t.Run("AddressedSessionTopicRefusedForInstanceClient", func(t *testing.T) {
		_, ok := wsDelivery("session."+sessionID+".albums.updated", sessionID, guest, wsClientRole{Role: acl.RoleInstance, Present: true})
		assert.False(t, ok)
	})
	t.Run("AddressedSessionTopicOfAnotherSession", func(t *testing.T) {
		_, ok := wsDelivery("session.othersessionid.albums.updated", sessionID, guest, wsClientRole{})
		assert.False(t, ok)
	})
	t.Run("FourPartTopicFromTheChannelTable", func(t *testing.T) {
		// An admin reaches a four-part topic through the channel in its third segment, without the
		// address matching.
		ev, ok := wsDelivery("user.us6sg6bxpogaaba2.albums.updated", sessionID, admin, wsClientRole{})
		assert.Equal(t, "albums.updated", ev)
		assert.True(t, ok)
	})
	t.Run("FourPartTopicRefusedForInstanceClientOfAdminAccount", func(t *testing.T) {
		// The account reaches the channel named in the third segment, which is the arm a client session
		// would otherwise be delivered through.
		_, ok := wsDelivery("user.us6sg6bxpogaaba2.albums.updated", sessionID, admin, wsClientRole{Role: acl.RoleInstance, Present: true})
		assert.False(t, ok)
	})
	t.Run("ThreePartTopicDelivered", func(t *testing.T) {
		// The level-tagged audit topics Pro and Portal forward take the same arm as a two-part topic.
		ev, ok := wsDelivery("audit.log.info", sessionID, admin, wsClientRole{})
		assert.Equal(t, "audit.log.info", ev)
		assert.True(t, ok)
	})
	t.Run("UnknownTopicShape", func(t *testing.T) {
		for _, topic := range []string{"", "log", "user.uid.albums.updated.extra"} {
			_, ok := wsDelivery(topic, sessionID, admin, wsClientRole{})
			assert.False(t, ok, "topic %q", topic)
		}
	})
	t.Run("UnknownUserReceivesNothing", func(t *testing.T) {
		for _, topic := range []string{"log.info", "photos.updated", "user." + userUID + ".albums.updated"} {
			_, ok := wsDelivery(topic, sessionID, entity.UnknownUser, wsClientRole{})
			assert.False(t, ok, "topic %s", topic)
		}
	})
}
