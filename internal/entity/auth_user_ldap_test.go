package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/authn"
)

// newLdapTestUser stores an account with the given name, provider and identifier, bypassing validation.
func newLdapTestUser(t *testing.T, name, provider, authId string) *User {
	t.Helper()

	m := NewUser()
	m.UserName = name
	m.AuthProvider = provider
	m.AuthID = authId
	m.UserRole = "guest"

	require.NoError(t, UnscopedDb().Create(m).Error)
	t.Cleanup(func() { UnscopedDb().Delete(&User{}, "user_uid = ?", m.UserUID) })

	return m
}

func TestLdapAuthID(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", LdapAuthID(""))
	})
	t.Run("Unchanged", func(t *testing.T) {
		assert.Equal(t, "cn=jane,ou=people,dc=example,dc=com", LdapAuthID("cn=jane,ou=people,dc=example,dc=com"))
	})
	t.Run("MaxLength", func(t *testing.T) {
		dn := "cn=jane," + strings.Repeat("o", 255-len("cn=jane,"))
		require.Len(t, dn, 255)
		assert.Equal(t, dn, LdapAuthID(dn))
	})
	t.Run("NotStoredUnchanged", func(t *testing.T) {
		for _, dn := range []string{
			"cn=jane,\n<dc=example>,dc=com",
			"cn=x\\<y,ou=people,dc=example,dc=com",
			"cn=jane," + strings.Repeat("ou=unit,", 100) + "dc=example,dc=com",
			"cn=" + strings.Repeat("ü", 200) + ",dc=example,dc=com",
		} {
			assert.Equal(t, "", LdapAuthID(dn), dn)
		}
	})
}

func TestFindLdapUser(t *testing.T) {
	t.Run("ByDistinguishedName", func(t *testing.T) {
		m := newLdapTestUser(t, "ldap-find-dn", string(authn.ProviderLDAP), "cn=ldap-find-dn,dc=example,dc=com")

		if found := FindLdapUser("ldap-find-renamed", "cn=ldap-find-dn,dc=example,dc=com"); assert.NotNil(t, found) {
			assert.Equal(t, m.UserUID, found.UserUID)
		}
	})
	t.Run("DistinguishedNameFirst", func(t *testing.T) {
		a := newLdapTestUser(t, "ldap-find-first-a", string(authn.ProviderLDAP), "cn=ldap-find-first-a,dc=example,dc=com")
		newLdapTestUser(t, "ldap-find-first-b", string(authn.ProviderLDAP), "cn=ldap-find-first-b,dc=example,dc=com")

		if found := FindLdapUser("ldap-find-first-b", "cn=ldap-find-first-a,dc=example,dc=com"); assert.NotNil(t, found) {
			assert.Equal(t, a.UserUID, found.UserUID)
		}
	})
	t.Run("MovedEntry", func(t *testing.T) {
		m := newLdapTestUser(t, "ldap-find-moved", string(authn.ProviderLDAP), "cn=ldap-find-moved,ou=old,dc=example,dc=com")

		if found := FindLdapUser("LDAP-Find-Moved", "cn=ldap-find-moved,ou=new,dc=example,dc=com"); assert.NotNil(t, found) {
			assert.Equal(t, m.UserUID, found.UserUID)
		}
	})
	t.Run("LongDistinguishedName", func(t *testing.T) {
		prefix := "cn=ldap-find-long," + strings.Repeat("ou=unit,", 40)
		dnA, dnB := prefix+"dc=example,dc=com", prefix+"dc=example,dc=org"
		require.Equal(t, dnA[:255], dnB[:255])
		m := newLdapTestUser(t, "ldap-find-long-a", string(authn.ProviderLDAP), dnA[:255])

		assert.Nil(t, FindLdapUser("ldap-find-long-b", dnB))

		if found := FindLdapUser("ldap-find-long-a", dnA); assert.NotNil(t, found) {
			assert.Equal(t, m.UserUID, found.UserUID)
		}
	})
	t.Run("ReservedCharacters", func(t *testing.T) {
		newLdapTestUser(t, "ldap-find-reserved", string(authn.ProviderLDAP), "cn=xy,ou=people,dc=example,dc=com")

		assert.Nil(t, FindLdapUser("ldap-find-reserved-other", "cn=x<y,ou=people,dc=example,dc=com"))
		assert.Nil(t, FindLdapUser("ldap-find-reserved-other", "cn=x>y,ou=people,dc=example,dc=com"))
	})
	t.Run("SuperAdmin", func(t *testing.T) {
		local := newLdapTestUser(t, "ldap-find-super-default", string(authn.ProviderDefault), "")
		require.NoError(t, UnscopedDb().Model(local).Update("SuperAdmin", true).Error)
		ldap := newLdapTestUser(t, "ldap-find-super-ldap", string(authn.ProviderLDAP), "cn=ldap-find-super-ldap,ou=old,dc=example,dc=com")
		require.NoError(t, UnscopedDb().Model(ldap).Update("SuperAdmin", true).Error)

		assert.Nil(t, FindLdapUser("ldap-find-super-default", "cn=ldap-find-super-default,dc=example,dc=com"))

		if found := FindLdapUser("ldap-find-super-ldap", "cn=ldap-find-super-ldap,ou=new,dc=example,dc=com"); assert.NotNil(t, found) {
			assert.Equal(t, ldap.UserUID, found.UserUID)
		}
	})
	t.Run("LegacyProviders", func(t *testing.T) {
		for _, provider := range []string{"", string(authn.ProviderDefault)} {
			name := "ldap-find-legacy-" + strings.TrimSpace(provider+"x")
			m := newLdapTestUser(t, name, provider, "")

			if found := FindLdapUser(name, "cn="+name+",dc=example,dc=com"); assert.NotNil(t, found, "provider %q", provider) {
				assert.Equal(t, m.UserUID, found.UserUID)
				assert.Equal(t, provider, found.AuthProvider)
			}
		}
	})
	t.Run("OtherProvidersNotByName", func(t *testing.T) {
		for _, provider := range []authn.ProviderType{authn.ProviderNone, authn.ProviderLocal, authn.ProviderOIDC, authn.ProviderLink} {
			name := "ldap-find-other-" + provider.String()
			newLdapTestUser(t, name, string(provider), "")

			assert.Nil(t, FindLdapUser(name, "cn="+name+",dc=example,dc=com"), "provider %s", provider)
		}
	})
	t.Run("IdenticalNameOnly", func(t *testing.T) {
		newLdapTestUser(t, "ldap-find-jose", string(authn.ProviderLDAP), "cn=ldap-find-jose,dc=example,dc=com")
		newLdapTestUser(t, "ldap-find-strasse", string(authn.ProviderLDAP), "cn=ldap-find-strasse,dc=example,dc=com")

		assert.Nil(t, FindLdapUser("ldap-find-josé", "cn=ldap-find-josé,dc=example,dc=com"))
		assert.Nil(t, FindLdapUser("ldap-find-straße", "cn=ldap-find-straße,dc=example,dc=com"))
	})
	t.Run("Deleted", func(t *testing.T) {
		m := newLdapTestUser(t, "ldap-find-deleted", string(authn.ProviderLDAP), "cn=ldap-find-deleted,dc=example,dc=com")
		require.NoError(t, UnscopedDb().Model(m).Update("DeletedAt", Now()).Error)

		// Found by name.
		if found := FindLdapUser("ldap-find-deleted", "cn=ldap-find-deleted,ou=new,dc=example,dc=com"); assert.NotNil(t, found) {
			assert.Equal(t, m.UserUID, found.UserUID)
			assert.True(t, found.IsDeleted())
		}

		// Found by distinguished name only.
		if found := FindLdapUser("ldap-find-deleted-renamed", "cn=ldap-find-deleted,dc=example,dc=com"); assert.NotNil(t, found) {
			assert.Equal(t, m.UserUID, found.UserUID)
			assert.True(t, found.IsDeleted())
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.Nil(t, FindLdapUser("ldap-find-missing", "cn=ldap-find-missing,dc=example,dc=com"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, FindLdapUser("", ""))
	})
}

func TestLdapNameProvider(t *testing.T) {
	t.Run("Allowed", func(t *testing.T) {
		for _, provider := range []string{"", "default", "ldap"} {
			assert.True(t, ldapNameProvider(provider), "provider %q", provider)
		}
	})
	t.Run("Refused", func(t *testing.T) {
		for _, provider := range []string{"none", "local", "oidc", "link", "client", "application", "*", "LDAP"} {
			assert.False(t, ldapNameProvider(provider), "provider %q", provider)
		}
	})
}

func TestLdapNameMatch(t *testing.T) {
	t.Run("Identical", func(t *testing.T) {
		assert.True(t, ldapNameMatch(&User{UserName: "jose", AuthProvider: "ldap"}, "jose"))
		assert.True(t, ldapNameMatch(&User{UserName: "jose", AuthProvider: ""}, "jose"))
	})
	t.Run("DifferentBytes", func(t *testing.T) {
		assert.False(t, ldapNameMatch(&User{UserName: "jose", AuthProvider: "ldap"}, "josé"))
		assert.False(t, ldapNameMatch(&User{UserName: "strasse", AuthProvider: "ldap"}, "straße"))
		assert.False(t, ldapNameMatch(&User{UserName: "jose", AuthProvider: "ldap"}, "Jose"))
	})
	t.Run("SuperAdmin", func(t *testing.T) {
		assert.False(t, ldapNameMatch(&User{UserName: "admin", AuthProvider: "", SuperAdmin: true}, "admin"))
		assert.False(t, ldapNameMatch(&User{UserName: "admin", AuthProvider: "default", SuperAdmin: true}, "admin"))
		assert.True(t, ldapNameMatch(&User{UserName: "admin", AuthProvider: "ldap", SuperAdmin: true}, "admin"))
	})
	t.Run("Provider", func(t *testing.T) {
		assert.False(t, ldapNameMatch(&User{UserName: "jose", AuthProvider: "none"}, "jose"))
		assert.False(t, ldapNameMatch(&User{UserName: "jose", AuthProvider: "oidc"}, "jose"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, ldapNameMatch(nil, "jose"))
		assert.False(t, ldapNameMatch(&User{UserName: "", AuthProvider: "ldap"}, ""))
	})
}

func TestUser_LdapNameMatch(t *testing.T) {
	t.Run("Provider", func(t *testing.T) {
		for _, provider := range []string{"", "default", "ldap"} {
			assert.True(t, (&User{UserName: "jose", AuthProvider: provider}).LdapNameMatch(), "provider %q", provider)
		}
		for _, provider := range []string{"none", "local", "oidc"} {
			assert.False(t, (&User{UserName: "jose", AuthProvider: provider}).LdapNameMatch(), "provider %q", provider)
		}
	})
	t.Run("SuperAdmin", func(t *testing.T) {
		assert.False(t, (&User{UserName: "admin", AuthProvider: "default", SuperAdmin: true}).LdapNameMatch())
		assert.True(t, (&User{UserName: "admin", AuthProvider: "ldap", SuperAdmin: true}).LdapNameMatch())
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, (*User)(nil).LdapNameMatch())
	})
}
