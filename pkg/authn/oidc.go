package authn

// OpenID Connect (OIDC) scope and claim identifiers:
// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
const (
	OidcClaimPreferredUsername = "preferred_username"
	OidcClaimEmail             = "email"
	OidcClaimName              = "name"
	OidcClaimNickname          = "nickname"
	OidcRequiredScopes         = "openid email profile"
	OidcDefaultScopes          = "openid email profile address"
)
