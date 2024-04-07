package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrantType_String(t *testing.T) {
	assert.Equal(t, "", GrantUndefined.String())
	assert.Equal(t, "client_credentials", GrantClientCredentials.String())
	assert.Equal(t, "session", GrantSession.String())
	assert.Equal(t, "password", GrantPassword.String())
	assert.Equal(t, "refresh_token", GrantRefreshToken.String())
	assert.Equal(t, "authorization_code", GrantAuthorizationCode.String())
	assert.Equal(t, "authorization_code", GrantType("Authorization Code ").String())
	assert.Equal(t, GrantAuthorizationCode.String(), GrantType("Authorization Code ").String())
	assert.Equal(t, "urn:ietf:params:oauth:grant-type:jwt-bearer", GrantJwtBearer.String())
}

func TestGrantType_Is(t *testing.T) {
	assert.Equal(t, true, GrantUndefined.Is(GrantUndefined))
	assert.Equal(t, true, GrantClientCredentials.Is(GrantClientCredentials))
	assert.Equal(t, true, GrantSession.Is(GrantSession))
	assert.Equal(t, true, GrantPassword.Is(GrantPassword))
	assert.Equal(t, false, GrantClientCredentials.Is(GrantPassword))
	assert.Equal(t, false, GrantClientCredentials.Is(GrantRefreshToken))
	assert.Equal(t, false, GrantClientCredentials.Is(GrantAuthorizationCode))
	assert.Equal(t, false, GrantClientCredentials.Is(GrantJwtBearer))
	assert.Equal(t, false, GrantClientCredentials.Is(GrantSamlBearer))
	assert.Equal(t, false, GrantClientCredentials.Is(GrantTokenExchange))
	assert.Equal(t, false, GrantClientCredentials.Is(GrantUndefined))
}

func TestGrantType_IsNot(t *testing.T) {
	assert.Equal(t, false, GrantUndefined.IsNot(GrantUndefined))
	assert.Equal(t, false, GrantClientCredentials.IsNot(GrantClientCredentials))
	assert.Equal(t, false, GrantPassword.IsNot(GrantPassword))
	assert.Equal(t, true, GrantClientCredentials.IsNot(GrantPassword))
	assert.Equal(t, true, GrantClientCredentials.IsNot(GrantRefreshToken))
	assert.Equal(t, true, GrantClientCredentials.IsNot(GrantAuthorizationCode))
	assert.Equal(t, true, GrantClientCredentials.IsNot(GrantJwtBearer))
	assert.Equal(t, true, GrantClientCredentials.IsNot(GrantSamlBearer))
	assert.Equal(t, true, GrantClientCredentials.IsNot(GrantTokenExchange))
	assert.Equal(t, true, GrantClientCredentials.IsNot(GrantUndefined))
}

func TestGrantType_IsUndefined(t *testing.T) {
	assert.Equal(t, true, GrantUndefined.IsUndefined())
	assert.Equal(t, false, GrantClientCredentials.IsUndefined())
	assert.Equal(t, false, GrantSession.IsUndefined())
	assert.Equal(t, false, GrantPassword.IsUndefined())
}

func TestGrantType_Pretty(t *testing.T) {
	assert.Equal(t, "", GrantUndefined.Pretty())
	assert.Equal(t, "CLI", GrantCLI.Pretty())
	assert.Equal(t, "Client Credentials", GrantClientCredentials.Pretty())
	assert.Equal(t, "Session", GrantSession.Pretty())
	assert.Equal(t, "Password", GrantPassword.Pretty())
	assert.Equal(t, "Refresh Token", GrantRefreshToken.Pretty())
	assert.Equal(t, "Authorization Code", GrantAuthorizationCode.Pretty())
	assert.Equal(t, "JWT Bearer Assertion", GrantJwtBearer.Pretty())
	assert.Equal(t, "SAML2 Bearer Assertion", GrantSamlBearer.Pretty())
}

func TestGrantType_Equal(t *testing.T) {
	assert.True(t, GrantClientCredentials.Equal("Client_Credentials"))
	assert.True(t, GrantClientCredentials.Equal("Client Credentials"))
	assert.True(t, GrantClientCredentials.Equal("client_credentials"))
	assert.True(t, GrantClientCredentials.Equal("client"))
	assert.True(t, GrantUndefined.Equal(""))
	assert.True(t, GrantSession.Equal("session"))
	assert.True(t, GrantPassword.Equal("Password"))
	assert.True(t, GrantPassword.Equal("password"))
	assert.True(t, GrantPassword.Equal("pass"))
}

func TestGrantType_NotEqual(t *testing.T) {
	assert.False(t, GrantClientCredentials.NotEqual("Client_Credentials"))
	assert.False(t, GrantClientCredentials.NotEqual("Client Credentials"))
	assert.False(t, GrantClientCredentials.NotEqual("client_credentials"))
	assert.False(t, GrantClientCredentials.NotEqual("client"))
	assert.True(t, GrantClientCredentials.NotEqual("access_token"))
	assert.True(t, GrantClientCredentials.NotEqual(""))
	assert.False(t, GrantUndefined.NotEqual(""))
	assert.False(t, GrantPassword.NotEqual("Password"))
	assert.False(t, GrantPassword.NotEqual("password"))
	assert.False(t, GrantPassword.NotEqual("pass"))
	assert.True(t, GrantPassword.NotEqual("passw"))
}

func TestGrant(t *testing.T) {
	assert.Equal(t, GrantUndefined, Grant(""))
	assert.Equal(t, GrantCLI, Grant("cli"))
	assert.Equal(t, GrantImplicit, Grant("implicit"))
	assert.Equal(t, GrantSession, Grant("session"))
	assert.Equal(t, GrantPassword, Grant("pass"))
	assert.Equal(t, GrantPassword, Grant("password"))
	assert.Equal(t, GrantClientCredentials, Grant("client credentials"))
	assert.Equal(t, GrantClientCredentials, Grant("client_credentials"))
	assert.Equal(t, GrantShareToken, Grant("share_token"))
	assert.Equal(t, GrantRefreshToken, Grant("refresh_token"))
	assert.Equal(t, GrantAuthorizationCode, Grant("auth_code"))
	assert.Equal(t, GrantAuthorizationCode, Grant("authorization_code"))
	assert.Equal(t, GrantAuthorizationCode, Grant("authorization code"))
	assert.Equal(t, GrantJwtBearer, Grant("jwt-bearer"))
	assert.Equal(t, GrantJwtBearer, Grant("jwt_bearer"))
	assert.Equal(t, GrantJwtBearer, Grant("jwt bearer"))
	assert.Equal(t, GrantSamlBearer, Grant("saml"))
	assert.Equal(t, GrantSamlBearer, Grant("saml2"))
	assert.Equal(t, GrantSamlBearer, Grant("saml2-bearer"))
	assert.Equal(t, GrantTokenExchange, Grant("token-exchange"))
	assert.Equal(t, GrantTokenExchange, Grant("token_exchange"))
	assert.Equal(t, GrantTokenExchange, Grant("token exchange"))
}
