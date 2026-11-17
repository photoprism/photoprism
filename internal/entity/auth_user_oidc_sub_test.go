package entity

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// createOIDCSubTestUser creates an account with the given provider and auth ID, and deletes it when the test ends.
func createOIDCSubTestUser(t *testing.T, provider authn.ProviderType, authId string, superAdmin bool) *User {
	t.Helper()

	user := NewUser()
	user.UserName = "oidcsub-" + rnd.Base36(6)
	user.UserRole = acl.RoleAdmin.String()
	user.SuperAdmin = superAdmin
	user.CanLogin = true
	user.AuthProvider = provider.String()
	user.AuthID = authId

	require.NoError(t, user.Create())

	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(&Password{}, "uid = ?", user.UserUID).Error)
		assert.NoError(t, UnscopedDb().Delete(&UserDetails{}, "user_uid = ?", user.UserUID).Error)
		assert.NoError(t, UnscopedDb().Delete(&UserSettings{}, "user_uid = ?", user.UserUID).Error)
		assert.NoError(t, UnscopedDb().Delete(&User{}, "user_uid = ?", user.UserUID).Error)
	})

	return user
}

// oidcSubTestValue returns a random OIDC subject, so that no fixture or other test holds it.
func oidcSubTestValue() string {
	return "sub-" + rnd.Base36(12)
}

// oidcSubAuditWarnings returns the number of audit warnings that contain the message.
func oidcSubAuditWarnings(hook *test.Hook, message string) (n int) {
	for _, e := range hook.AllEntries() {
		if e.Level == logrus.WarnLevel && strings.Contains(e.Message, message) {
			n++
		}
	}

	return n
}

func TestValidOIDCSub(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, ValidOIDCSub("usg73p55zwgr1gbq"))
		assert.True(t, ValidOIDCSub("abcd"))
		assert.True(t, ValidOIDCSub("4f2c1a7e-0d3b-4c8e-9a51-6e2f8b7d9c10"))
		assert.True(t, ValidOIDCSub(strings.Repeat("a", 255)))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, ValidOIDCSub(""))
		assert.False(t, ValidOIDCSub("abc"))
		assert.False(t, ValidOIDCSub("abc\x01def"))
		assert.False(t, ValidOIDCSub("<abcdef>"))
		assert.False(t, ValidOIDCSub(" abcdef"))
		assert.False(t, ValidOIDCSub(strings.Repeat("a", 256)))
		assert.False(t, ValidOIDCSub(strings.Repeat("ä", 200)))
	})
}

func TestUser_InitOIDCSub(t *testing.T) {
	t.Run("LinkLocal", func(t *testing.T) {
		user := createOIDCSubTestUser(t, authn.ProviderLocal, "", true)
		require.NoError(t, user.SetPassword("insecure-oidc"))
		require.NoError(t, user.Updates(Values{"auth_method": authn.Method2FA.String(), "auth_issuer": "https://old.example.com/"}))
		user.AuthIssuer = "https://old.example.com/"
		sub := oidcSubTestValue()
		hook := captureAuditLog(t)

		assert.True(t, user.InitOIDCSub(sub))
		assert.Equal(t, 1, oidcSubAuditWarnings(hook, "linked to oidc subject"))
		assert.Equal(t, authn.ProviderOIDC.String(), user.AuthProvider)
		assert.Equal(t, sub, user.AuthID)
		assert.Equal(t, "", user.AuthIssuer)

		found := FindUserByUID(user.UserUID)
		require.NotNil(t, found)
		assert.Equal(t, authn.ProviderOIDC.String(), found.AuthProvider)
		assert.Equal(t, sub, found.AuthID)
		assert.Equal(t, "", found.AuthIssuer)
		assert.Equal(t, authn.Method2FA.String(), found.AuthMethod)
		assert.True(t, found.SuperAdmin)
		assert.Equal(t, acl.RoleAdmin.String(), found.UserRole)
		assert.False(t, found.InvalidPassword("insecure-oidc"))
		assert.Equal(t, found.UserUID, FindUser(User{AuthProvider: authn.ProviderOIDC.String(), AuthID: sub, AuthIssuer: "https://idp.example.com/"}).UserUID)
	})
	t.Run("LinkDefault", func(t *testing.T) {
		user := createOIDCSubTestUser(t, authn.ProviderDefault, "", false)
		sub := oidcSubTestValue()

		assert.True(t, user.InitOIDCSub(sub))
		assert.Equal(t, sub, FindUserByUID(user.UserUID).AuthID)
	})
	t.Run("LinkUnlinkedOIDC", func(t *testing.T) {
		user := createOIDCSubTestUser(t, authn.ProviderOIDC, "", true)
		sub := oidcSubTestValue()

		assert.True(t, user.InitOIDCSub(sub))
		assert.Equal(t, sub, FindUserByUID(user.UserUID).AuthID)
	})
	t.Run("ReplaceLocalAuthID", func(t *testing.T) {
		user := createOIDCSubTestUser(t, authn.ProviderLocal, "local-id", true)
		sub := oidcSubTestValue()

		assert.True(t, user.InitOIDCSub(sub))
		found := FindUserByUID(user.UserUID)
		assert.Equal(t, authn.ProviderOIDC.String(), found.AuthProvider)
		assert.Equal(t, sub, found.AuthID)
	})
	t.Run("RelinkAfterLocal", func(t *testing.T) {
		user := createOIDCSubTestUser(t, authn.ProviderLocal, "", true)
		sub := oidcSubTestValue()

		require.True(t, user.InitOIDCSub(sub))
		require.NoError(t, user.Updates(Values{"auth_provider": authn.ProviderLocal.String()}))

		assert.True(t, user.InitOIDCSub(sub))
		assert.Equal(t, authn.ProviderOIDC.String(), FindUserByUID(user.UserUID).AuthProvider)
	})
	t.Run("SameSub", func(t *testing.T) {
		sub := oidcSubTestValue()
		user := createOIDCSubTestUser(t, authn.ProviderOIDC, sub, true)
		require.NoError(t, user.Updates(Values{"auth_issuer": "https://idp.example.com/"}))
		hook := captureAuditLog(t)

		assert.False(t, user.InitOIDCSub(sub))
		assert.Equal(t, 0, oidcSubAuditWarnings(hook, "linked to oidc subject"))
		found := FindUserByUID(user.UserUID)
		assert.Equal(t, sub, found.AuthID)
		assert.Equal(t, "https://idp.example.com/", found.AuthIssuer)
	})
	t.Run("OtherSub", func(t *testing.T) {
		linked := oidcSubTestValue()
		user := createOIDCSubTestUser(t, authn.ProviderOIDC, linked, true)

		assert.False(t, user.InitOIDCSub(oidcSubTestValue()))
		assert.Equal(t, linked, user.AuthID)
		assert.Equal(t, linked, FindUserByUID(user.UserUID).AuthID)
	})
	t.Run("DetachOtherAccount", func(t *testing.T) {
		sub := oidcSubTestValue()
		other := createOIDCSubTestUser(t, authn.ProviderOIDC, sub, false)
		user := createOIDCSubTestUser(t, authn.ProviderLocal, "", true)
		hook := captureAuditLog(t)

		assert.True(t, user.InitOIDCSub(sub))
		assert.Equal(t, 1, oidcSubAuditWarnings(hook, "oidc subject detached from '"+other.UserName+"'"))
		assert.Equal(t, sub, FindUserByUID(user.UserUID).AuthID)
		found := FindUserByUID(other.UserUID)
		assert.Equal(t, authn.ProviderNone.String(), found.AuthProvider)
		assert.Equal(t, "", found.AuthID)
	})
	t.Run("SuperAdminHolder", func(t *testing.T) {
		sub := oidcSubTestValue()
		other := createOIDCSubTestUser(t, authn.ProviderOIDC, sub, true)
		user := createOIDCSubTestUser(t, authn.ProviderLocal, "", true)

		assert.False(t, user.InitOIDCSub(sub))
		assert.Equal(t, authn.ProviderLocal.String(), FindUserByUID(user.UserUID).AuthProvider)
		assert.Equal(t, "", FindUserByUID(user.UserUID).AuthID)
		found := FindUserByUID(other.UserUID)
		assert.Equal(t, authn.ProviderOIDC.String(), found.AuthProvider)
		assert.Equal(t, sub, found.AuthID)
	})
	t.Run("ProviderNone", func(t *testing.T) {
		user := createOIDCSubTestUser(t, authn.ProviderNone, "", true)

		assert.False(t, user.InitOIDCSub(oidcSubTestValue()))
		assert.Equal(t, authn.ProviderNone.String(), FindUserByUID(user.UserUID).AuthProvider)
	})
	t.Run("ProviderLDAP", func(t *testing.T) {
		user := createOIDCSubTestUser(t, authn.ProviderLDAP, "uid=admin,ou=people,dc=example,dc=com", true)

		assert.False(t, user.InitOIDCSub(oidcSubTestValue()))
		found := FindUserByUID(user.UserUID)
		assert.Equal(t, authn.ProviderLDAP.String(), found.AuthProvider)
		assert.Equal(t, "uid=admin,ou=people,dc=example,dc=com", found.AuthID)
	})
	t.Run("Deleted", func(t *testing.T) {
		user := createOIDCSubTestUser(t, authn.ProviderLocal, "", true)
		require.NoError(t, UnscopedDb().Model(user).UpdateColumn("deleted_at", Now()).Error)

		assert.False(t, user.InitOIDCSub(oidcSubTestValue()))
		assert.Equal(t, authn.ProviderLocal.String(), FindUserByUID(user.UserUID).AuthProvider)
	})
	t.Run("InvalidSub", func(t *testing.T) {
		user := createOIDCSubTestUser(t, authn.ProviderLocal, "", true)

		assert.False(t, user.InitOIDCSub("abc"))
		assert.False(t, user.InitOIDCSub("abc\x01def"))
		assert.Equal(t, authn.ProviderLocal.String(), FindUserByUID(user.UserUID).AuthProvider)
	})
	t.Run("EmptySub", func(t *testing.T) {
		sub := oidcSubTestValue()
		user := createOIDCSubTestUser(t, authn.ProviderOIDC, sub, true)

		assert.False(t, user.InitOIDCSub(""))
		assert.Equal(t, sub, FindUserByUID(user.UserUID).AuthID)
	})
	t.Run("InvalidUser", func(t *testing.T) {
		var user *User
		assert.False(t, user.InitOIDCSub(oidcSubTestValue()))
		assert.False(t, (&User{}).InitOIDCSub(oidcSubTestValue()))
		assert.False(t, (&User{UserUID: rnd.GenerateUID(UserUID)}).InitOIDCSub(oidcSubTestValue()))
	})
}
