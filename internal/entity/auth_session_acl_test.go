package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
