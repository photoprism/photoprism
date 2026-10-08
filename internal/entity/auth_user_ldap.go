package entity

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ldapNameProvider checks if a directory login may resolve an account with this provider by username.
func ldapNameProvider(provider string) bool {
	switch authn.ProviderType(provider) {
	case authn.ProviderUndefined, authn.ProviderDefault, authn.ProviderLDAP:
		return true
	default:
		return false
	}
}

// ldapNameMatch checks if a directory login may resolve the account by its username: the stored name
// must be identical to the login name, the provider must allow it, and a super admin account must
// already be a directory account.
func ldapNameMatch(m *User, name string) bool {
	if m == nil || name == "" || m.UserName != name || !ldapNameProvider(m.AuthProvider) {
		return false
	}

	return !m.SuperAdmin || m.AuthProvider == authn.ProviderLDAP.String()
}

// LdapAuthID returns the distinguished name as it is stored, or an empty string if it cannot be
// stored unchanged, so that only identical names are compared.
func LdapAuthID(dn string) string {
	if len(dn) > ldapAuthIDLength || clean.Auth(dn) != dn {
		return ""
	}

	return dn
}

// ldapAuthIDLength is the size of the auth_id column in bytes.
const ldapAuthIDLength = 255

// FindLdapUser returns the account of a directory user, including a deleted one: the LDAP account that
// stores its distinguished name, or else the account whose username is identical to the login name and
// whose provider allows it to be resolved by name. The username is compared byte for byte.
func FindLdapUser(username, dn string) *User {
	if authId := LdapAuthID(dn); authId != "" {
		m := &User{}

		if err := UnscopedDb().
			Where("auth_provider = ? AND auth_id = ?", authn.ProviderLDAP.String(), authId).
			Order("id").First(m).Error; err == nil {
			return m.LoadRelated()
		}
	}

	name := clean.Username(username)

	if name == "" {
		return nil
	}

	var found []User

	if err := UnscopedDb().Where("user_name = ?", name).Order("id").Find(&found).Error; err != nil {
		return nil
	}

	for i := range found {
		if ldapNameMatch(&found[i], name) {
			return found[i].LoadRelated()
		}
	}

	return nil
}
