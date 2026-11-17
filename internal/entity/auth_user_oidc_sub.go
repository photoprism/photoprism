package entity

import (
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// ValidOIDCSub reports whether sub fits the auth_id column and is not changed by cleaning.
func ValidOIDCSub(sub string) bool {
	return len(sub) >= 4 && len(sub) <= 255 && clean.Auth(sub) == sub
}

// InitOIDCSub links the account to the OIDC subject sub if it has no OIDC subject yet and no other super
// admin holds it, detaching it from other accounts. An existing link, the password, and 2FA settings are
// kept, and no issuer is stored, so the first sign-in binds the issuer of the configured provider.
func (m *User) InitOIDCSub(sub string) (updated bool) {
	if sub == "" || m == nil || m.InvalidUID() {
		return false
	} else if !ValidOIDCSub(sub) {
		event.SystemWarn([]string{"config", "init", "admin account", "invalid oidc subject", status.Skipped})
		return false
	}

	user := FindUserByUID(m.UserUID)

	if user == nil {
		return false
	}

	name := clean.LogQuote(user.Username())
	provider := user.Provider()

	switch {
	case user.IsDeleted():
		event.SystemWarn([]string{"config", "init", "admin account %s", "deleted account cannot be linked to oidc subject", status.Skipped}, name)
		return false
	case !provider.IsDefault() && !provider.IsLocal():
		event.SystemWarn([]string{"config", "init", "admin account %s", "provider %s cannot be linked to oidc subject", status.Skipped}, name, clean.Log(provider.String()))
		return false
	case provider.Is(authn.ProviderOIDC) && user.AuthID == sub:
		return false
	case provider.Is(authn.ProviderOIDC) && user.AuthID != "":
		event.SystemWarn([]string{"config", "init", "admin account %s", "already linked to another oidc subject", status.Skipped}, name)
		return false
	}

	// Super admins keep their subject, since SetAuthID does not detach it from them.
	holders := Users{}

	if err := UnscopedDb().Where("user_uid <> ? AND auth_provider = ? AND auth_id = ?", user.UserUID, authn.ProviderOIDC.String(), sub).
		Find(&holders).Error; err != nil {
		event.SystemError([]string{"config", "init", "admin account %s", "failed to find oidc subject", status.Error(err)}, name)
		return false
	}

	for _, other := range holders {
		if other.SuperAdmin {
			event.SystemWarn([]string{"config", "init", "admin account %s", "oidc subject is already linked to %s", status.Skipped}, name, clean.LogQuote(other.Username()))
			return false
		}
	}

	user.AuthProvider = authn.ProviderOIDC.String()

	if err := user.Updates(Values{"auth_provider": user.AuthProvider, "auth_id": sub, "auth_issuer": ""}); err != nil {
		event.SystemError([]string{"config", "init", "admin account %s", "failed to link oidc subject", status.Error(err)}, name)
		return false
	}

	// Detach the subject from other accounts, as the users mod command does.
	user.SetAuthID(sub, "")

	for _, other := range holders {
		event.AuditWarn([]string{"config", "init", "admin account %s", "oidc subject detached from %s", status.Updated}, name, clean.LogQuote(other.Username()))
	}

	m.AuthProvider = user.AuthProvider
	m.AuthID = user.AuthID
	m.AuthIssuer = user.AuthIssuer

	event.AuditWarn([]string{"config", "init", "admin account %s", "linked to oidc subject", status.Updated}, name)

	return true
}
