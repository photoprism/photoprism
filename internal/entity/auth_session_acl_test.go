package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
)

// mixedPrincipalSession builds a restricted client session acting for a privileged account: an
// instance client (GrantSearchShared) owned by the admin user alice.
func mixedPrincipalSession() *Session {
	s := &Session{}
	s.SetClient(&Client{ClientRole: acl.RoleInstance.String(), AuthProvider: authn.ProviderClient.String()})
	s.SetUser(UserFixtures.Pointer("alice"))
	return s
}

// broadClientNarrowUserSession builds the mirror of mixedPrincipalSession: a full-access client
// role, which is what a credential is minted with by default, owned by a narrow account.
func broadClientNarrowUserSession() *Session {
	s := &Session{}
	s.SetClient(&Client{ClientRole: acl.RoleClient.String(), AuthProvider: authn.ProviderClient.String()})
	s.SetUser(UserFixtures.Pointer("guest"))

	return s
}

func TestSession_Grants(t *testing.T) {
	t.Run("NilUnrestricted", func(t *testing.T) {
		var s *Session
		assert.True(t, s.Grants(acl.ResourcePhotos, acl.AccessPrivate))
	})
	t.Run("AdminGranted", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		assert.True(t, s.Grants(acl.ResourcePhotos, acl.AccessPrivate))
		assert.True(t, s.Grants(acl.ResourcePlaces, acl.AccessLibrary))
	})
	t.Run("GuestDeniedPrivate", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("guest"))
		assert.False(t, s.Grants(acl.ResourcePhotos, acl.AccessPrivate))
	})
	t.Run("ClientRoleLimitsPrivilegedUser", func(t *testing.T) {
		s := mixedPrincipalSession()
		assert.True(t, s.IsClient())
		assert.Equal(t, acl.RoleInstance, s.GetClientRole())
		// The admin owner alone would grant these; the instance client role does not.
		assert.False(t, s.Grants(acl.ResourcePhotos, acl.AccessPrivate))
		assert.False(t, s.Grants(acl.ResourcePlaces, acl.AccessPrivate))
		assert.False(t, s.Grants(acl.ResourceAlbums, acl.AccessLibrary))
		assert.False(t, s.Grants(acl.ResourceFiles, acl.AccessAll))
	})
	t.Run("UserRoleLimitsBroadClient", func(t *testing.T) {
		// The mirror of the case above, and the shape a credential takes by default: the client role
		// carries full access and the owning account does not, so the account is what limits it.
		s := broadClientNarrowUserSession()
		assert.True(t, s.IsClient())
		assert.Equal(t, acl.RoleClient, s.GetClientRole())
		assert.True(t, acl.Rules.Allow(acl.ResourcePhotos, acl.RoleClient, acl.AccessPrivate))
		assert.False(t, s.Grants(acl.ResourcePhotos, acl.AccessPrivate))
		assert.False(t, s.Grants(acl.ResourceFiles, acl.AccessLibrary))
		assert.False(t, s.Grants(acl.ResourceAlbums, acl.AccessLibrary))
	})
	t.Run("ClientWithoutUserKeepsClientRole", func(t *testing.T) {
		s := &Session{}
		s.SetClient(&Client{ClientRole: acl.RoleInstance.String(), AuthProvider: authn.ProviderClient.String()})
		assert.True(t, s.NoUser())
		assert.True(t, s.Grants(acl.ResourcePhotos, acl.ActionSearch))
	})
}

func TestSession_GrantsAny(t *testing.T) {
	t.Run("NilTrue", func(t *testing.T) {
		var s *Session
		assert.True(t, s.GrantsAny(acl.ResourcePhotos, acl.Permissions{acl.AccessAll}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, mixedPrincipalSession().GrantsAny(acl.ResourcePhotos, acl.Permissions{}))
	})
	t.Run("MixedPrincipalNoLibrary", func(t *testing.T) {
		s := mixedPrincipalSession()
		assert.False(t, s.GrantsAny(acl.ResourcePhotos, acl.Permissions{acl.AccessAll, acl.AccessLibrary}))
		assert.True(t, s.GrantsAny(acl.ResourcePhotos, acl.Permissions{acl.AccessAll, acl.AccessShared}))
	})
}

func TestSession_Denies(t *testing.T) {
	t.Run("MixedPrincipal", func(t *testing.T) {
		s := mixedPrincipalSession()
		assert.True(t, s.Denies(acl.ResourcePhotos, acl.AccessPrivate))
		assert.False(t, s.Denies(acl.ResourcePhotos, acl.AccessShared))
	})
	t.Run("Nil", func(t *testing.T) {
		var s *Session
		assert.False(t, s.Denies(acl.ResourcePhotos, acl.AccessPrivate))
	})
}

func TestSession_DeniesAll(t *testing.T) {
	t.Run("MixedPrincipal", func(t *testing.T) {
		s := mixedPrincipalSession()
		assert.True(t, s.DeniesAll(acl.ResourcePhotos, acl.Permissions{acl.AccessAll, acl.AccessLibrary}))
		assert.False(t, s.DeniesAll(acl.ResourcePhotos, acl.Permissions{acl.AccessAll, acl.AccessShared}))
	})
	t.Run("RoleNoneHasNoLibraryReach", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("unauthorized"))
		assert.True(t, s.DeniesAll(acl.ResourcePhotos, acl.Permissions{acl.AccessAll, acl.AccessLibrary}))
	})
}

func TestSession_HasSharedAccessOnly(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var s *Session
		assert.False(t, s.HasSharedAccessOnly(acl.ResourcePhotos))
	})
	t.Run("Admin", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		assert.False(t, s.HasSharedAccessOnly(acl.ResourcePhotos))
		assert.False(t, s.HasSharedAccessOnly(acl.ResourceAlbums))
	})
	t.Run("MixedPrincipal", func(t *testing.T) {
		// The admin owner is not shared-only, but the instance client is, so the session is.
		s := mixedPrincipalSession()
		assert.False(t, s.GetUser().HasSharedAccessOnly(acl.ResourcePhotos))
		assert.True(t, s.HasSharedAccessOnly(acl.ResourcePhotos))
		assert.True(t, s.HasSharedAccessOnly(acl.ResourceAlbums))
	})
	t.Run("BroadClientNarrowUser", func(t *testing.T) {
		// A full-access client role does not lift the owning account out of shared-only access.
		s := broadClientNarrowUserSession()
		assert.False(t, acl.Rules.Deny(acl.ResourcePhotos, acl.RoleClient, acl.AccessLibrary))
		assert.True(t, s.HasSharedAccessOnly(acl.ResourcePhotos))
		assert.True(t, s.HasSharedAccessOnly(acl.ResourceAlbums))
	})
	t.Run("OwnAccessIsNotSharedAccess", func(t *testing.T) {
		// An instance client holds GrantUseOwn on places, which carries neither access_shared nor
		// search, so it is not shared-only: it has less reach than that, not more.
		s := mixedPrincipalSession()
		assert.False(t, s.HasSharedAccessOnly(acl.ResourcePlaces))
		assert.True(t, s.Denies(acl.ResourcePlaces, acl.AccessShared))
		assert.True(t, s.Denies(acl.ResourcePlaces, acl.ActionSearch))
	})
}

// fullAccessClientSession returns what a credential is minted with by default: the client role, no
// scope restriction, and no account.
func fullAccessClientSession() *Session {
	s := &Session{}
	s.SetClient(&Client{ClientRole: acl.RoleClient.String(), AuthProvider: authn.ProviderClient.String(),
		AuthScope: "*"})

	return s
}

func TestSession_SeesAnyDetail(t *testing.T) {
	t.Run("NilUnrestricted", func(t *testing.T) {
		var s *Session
		assert.True(t, s.SeesAnyDetail(acl.ResourceAlbums))
	})
	t.Run("Admin", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		assert.True(t, s.SeesAnyDetail(acl.ResourceAlbums))
	})
	// A share reaches a record, so a session holding one still views the resource - what it may not
	// do is reach a resource its credential was never issued for.
	t.Run("ShareLinkVisitor", func(t *testing.T) {
		assert.True(t, SessionFixtures.Pointer("visitor").SeesAnyDetail(acl.ResourceAlbums))
	})
	t.Run("Guest", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("guest"))
		assert.True(t, s.SeesAnyDetail(acl.ResourceAlbums))
		assert.False(t, s.SeesFullDetail(acl.ResourceAlbums), "without reaching every record")
	})
	t.Run("ScopeWithoutTheResource", func(t *testing.T) {
		s := fullAccessClientSession()
		s.SetScope("photos")
		assert.False(t, s.SeesAnyDetail(acl.ResourceAlbums))
		assert.True(t, s.SeesAnyDetail(acl.ResourcePhotos))
	})
	t.Run("WriteOnlyScope", func(t *testing.T) {
		s := fullAccessClientSession()
		s.SetScope("write photos")
		assert.False(t, s.SeesAnyDetail(acl.ResourcePhotos))
	})
	t.Run("RoleWithoutView", func(t *testing.T) {
		s := &Session{}
		s.SetClient(&Client{ClientRole: acl.RoleInstance.String(), AuthScope: "*",
			AuthProvider: authn.ProviderClient.String()})
		require.True(t, s.Grants(acl.ResourcePlaces, acl.AccessOwn), "it holds something on places")
		assert.False(t, s.SeesAnyDetail(acl.ResourcePlaces), "but not view")
	})
}

func TestSession_SeesFullDetail(t *testing.T) {
	t.Run("NilUnrestricted", func(t *testing.T) {
		var s *Session
		assert.True(t, s.SeesFullDetail(acl.ResourcePhotos))
	})
	t.Run("Admin", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		assert.True(t, s.SeesFullDetail(acl.ResourcePhotos))
		assert.True(t, s.SeesFullDetail(acl.ResourceAlbums))
	})
	t.Run("ShareLinkVisitor", func(t *testing.T) {
		assert.False(t, SessionFixtures.Pointer("visitor").SeesFullDetail(acl.ResourcePhotos))
		assert.False(t, SessionFixtures.Pointer("visitor_token_metrics").SeesFullDetail(acl.ResourcePhotos))
	})
	t.Run("Guest", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("guest"))
		assert.False(t, s.SeesFullDetail(acl.ResourcePhotos))
	})
	t.Run("FullAccessClient", func(t *testing.T) {
		s := fullAccessClientSession()
		assert.True(t, s.NotRegistered(), "the case exists because it has no account")
		assert.True(t, s.SeesFullDetail(acl.ResourcePhotos))
		assert.True(t, s.SeesFullDetail(acl.ResourceAlbums))
	})
	t.Run("NarrowlyScopedClient", func(t *testing.T) {
		assert.False(t, SessionFixtures.Pointer("client_metrics").SeesFullDetail(acl.ResourcePhotos))
		assert.False(t, SessionFixtures.Pointer("client_analytics").SeesFullDetail(acl.ResourcePhotos))
	})
	t.Run("ScopeNamingPhotos", func(t *testing.T) {
		s := fullAccessClientSession()
		s.SetScope("photos")
		assert.True(t, s.SeesFullDetail(acl.ResourcePhotos))
		assert.False(t, s.SeesFullDetail(acl.ResourceAlbums), "another resource is out of scope")
	})
	t.Run("WriteOnlyScope", func(t *testing.T) {
		// A credential admitted to change a picture is not thereby admitted to read its detail.
		s := fullAccessClientSession()
		s.SetScope("write photos")
		assert.False(t, s.SeesFullDetail(acl.ResourcePhotos))

		s.SetScope("read photos")
		assert.True(t, s.SeesFullDetail(acl.ResourcePhotos), "a read scope is the one this answers for")

		s.SetScope("read")
		assert.True(t, s.SeesFullDetail(acl.ResourcePhotos), "as is a global read scope")
	})
	t.Run("EmptyClientScope", func(t *testing.T) {
		// AuthAny refuses a credential with no scope, so it is not full detail here either.
		s := fullAccessClientSession()
		s.AuthScope = ""
		assert.False(t, s.SeesFullDetail(acl.ResourcePhotos))
	})
	t.Run("AccountSessionWithoutScope", func(t *testing.T) {
		assert.True(t, SessionFixtures.Pointer("alice").SeesFullDetail(acl.ResourcePhotos),
			"a browser session has no scope and is unrestricted by it")
		assert.True(t, SessionFixtures.Pointer("alice_app_password_full_access").SeesFullDetail(acl.ResourcePhotos))
		assert.False(t, SessionFixtures.Pointer("alice_app_password_webdav").SeesFullDetail(acl.ResourcePhotos),
			"an app password is scoped like any other credential")
	})
	t.Run("ClientRoleWithoutLibraryAccess", func(t *testing.T) {
		// An instance client is shared-only on photos, and holds less than that on places - neither
		// reaches the whole library, so the role decides before the scope is read.
		s := &Session{}
		s.SetClient(&Client{ClientRole: acl.RoleInstance.String(), AuthProvider: authn.ProviderClient.String(),
			AuthScope: "*"})
		assert.False(t, s.SeesFullDetail(acl.ResourcePhotos))
		assert.False(t, s.SeesFullDetail(acl.ResourcePlaces))
	})
	t.Run("ClientActingForAccount", func(t *testing.T) {
		assert.False(t, broadClientNarrowUserSession().SeesFullDetail(acl.ResourcePhotos),
			"the owning account is shared-only, and a client role does not lift it out")
		assert.False(t, mixedPrincipalSession().SeesFullDetail(acl.ResourcePhotos),
			"and a narrow client is not lifted by an admin owner either")
	})
	t.Run("ClientForAnAccountThatIsNotOne", func(t *testing.T) {
		// Grants reads the intersection, so a client naming a user the database does not have is
		// answered by that user's role, as it is on the handler.
		s := fullAccessClientSession()
		s.UserUID = "us6sg6bxpogaabcd"
		assert.False(t, s.SeesFullDetail(acl.ResourcePhotos))
	})
	t.Run("NeitherAccountNorCredential", func(t *testing.T) {
		assert.False(t, (&Session{}).SeesFullDetail(acl.ResourcePhotos))
	})
	t.Run("UserWithoutAValidAccount", func(t *testing.T) {
		// A session carrying a privileged user record without a valid account uid is neither, so
		// the role it would otherwise be read on never gets there.
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		s.UserUID = ""

		require.Equal(t, acl.RoleAdmin, s.GetUserRole())
		assert.False(t, s.SeesFullDetail(acl.ResourcePhotos))
	})
	t.Run("AccountSessionWithScope", func(t *testing.T) {
		// An account session that does carry a scope is limited by it, as it is on the handler.
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		s.SetScope("metrics")
		assert.False(t, s.SeesFullDetail(acl.ResourcePhotos))

		s.SetScope("photos")
		assert.True(t, s.SeesFullDetail(acl.ResourcePhotos))
	})
	// No registered grant carries whole-library reach without view, so no role can tell the two
	// requirements apart today. This pins that, so a grant that separates them fails here rather
	// than handing out detail the role may not view.
	t.Run("LibraryReachImpliesView", func(t *testing.T) {
		reach := acl.Permissions{acl.AccessAll, acl.AccessLibrary}

		checked := 0

		for resource, roles := range acl.Rules {
			for role := range roles {
				if acl.Rules.AllowAny(resource, role, reach) {
					checked++

					assert.True(t, acl.Rules.Allow(resource, role, acl.ActionView),
						"%s may reach the whole library of %s", role, resource)
				}
			}
		}

		require.NotZero(t, checked, "an empty rule table would pass this without asserting anything")
	})
}
