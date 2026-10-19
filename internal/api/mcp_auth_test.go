package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestMcpChallenge(t *testing.T) {
	conf := prepareMCPTest(t)
	metadataUrl := conf.OAuthResourceMetadataUrl()
	require.True(t, conf.OAuthAppsEnabled())

	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, `Bearer resource_metadata="`+metadataUrl+`"`, McpChallenge(conf, ""))
	})
	t.Run("InvalidToken", func(t *testing.T) {
		assert.Equal(t, `Bearer resource_metadata="`+metadataUrl+`", error="invalid_token"`, McpChallenge(conf, McpErrInvalidToken))
	})
	t.Run("InsufficientScope", func(t *testing.T) {
		assert.Equal(t, `Bearer resource_metadata="`+metadataUrl+`", error="insufficient_scope", scope="mcp photos"`, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos"))
	})
	t.Run("AppsDisabled", func(t *testing.T) {
		// Without the app flow, the challenge does not name the protected resource metadata.
		prev := conf.Options().OAuthApps
		t.Cleanup(func() { conf.Options().OAuthApps = prev })
		conf.Options().OAuthApps = config.OAuthAppsNone
		assert.Equal(t, `Bearer error="insufficient_scope", scope="mcp photos"`, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos"))
		assert.Empty(t, McpChallenge(conf, ""))
	})
	t.Run("PublicMode", func(t *testing.T) {
		conf.Options().Public = true
		t.Cleanup(func() { conf.Options().Public = false })
		assert.Equal(t, `Bearer error="insufficient_scope", scope="mcp photos"`, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos"))
	})
}

func TestMcpTokenInfo(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, mcpTokenInfo(nil))
	})
	t.Run("User", func(t *testing.T) {
		expires := time.Now().Add(time.Hour).Unix()
		s := &entity.Session{UserUID: "uqxetse3cy5eo9z2", ClientUID: "cs5gfen1bgxz7s9i", RefID: "sessxkkcxxxz", AuthScope: "mcp photos", SessExpires: expires}
		info := mcpTokenInfo(s)
		require.NotNil(t, info)
		assert.Equal(t, "uqxetse3cy5eo9z2", info.UserID)
		assert.Equal(t, []string{"mcp", "photos"}, info.Scopes)
		assert.True(t, info.Expiration.IsZero())
		assert.Same(t, s, info.Extra[mcp.SessionKey])
	})
	t.Run("Client", func(t *testing.T) {
		s := &entity.Session{ClientUID: "cs5gfen1bgxz7s9i", RefID: "sessxkkcxxxz", AuthScope: "*"}
		info := mcpTokenInfo(s)
		require.NotNil(t, info)
		assert.Equal(t, "cs5gfen1bgxz7s9i", info.UserID)
		assert.Equal(t, []string{"*"}, info.Scopes)
	})
	t.Run("JWT", func(t *testing.T) {
		a := &entity.Session{RefID: "jti-one", GrantType: authn.GrantJwtBearer.String(), AuthIssuer: "https://portal.example.com", ClientName: "portal:0198a3c4-3b26-7d4a-9f31-1d0c6c2a7f10"}
		b := &entity.Session{RefID: "jti-two", GrantType: authn.GrantJwtBearer.String(), AuthIssuer: "https://portal.example.com", ClientName: "portal:0198a3c4-3b26-7d4a-9f31-1d0c6c2a7f10"}
		c := &entity.Session{RefID: "jti-three", GrantType: authn.GrantJwtBearer.String(), AuthIssuer: "https://other.example.com", ClientName: "portal:0198a3c4-3b26-7d4a-9f31-1d0c6c2a7f10"}
		require.NotNil(t, mcpTokenInfo(a))
		assert.Equal(t, mcpTokenInfo(a).UserID, mcpTokenInfo(b).UserID)
		assert.NotEqual(t, mcpTokenInfo(a).UserID, mcpTokenInfo(c).UserID)
	})
	t.Run("RefID", func(t *testing.T) {
		info := mcpTokenInfo(&entity.Session{RefID: "sessxkkcxxxz"})
		require.NotNil(t, info)
		assert.Equal(t, "sessxkkcxxxz", info.UserID)
	})
	t.Run("NoOwner", func(t *testing.T) {
		assert.Nil(t, mcpTokenInfo(&entity.Session{}))
	})
}

func TestMcpBearerRequest(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, mcpBearerRequest(nil, "token"))
	})
	t.Run("NoToken", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/mcp", nil)
		assert.Same(t, r, mcpBearerRequest(r, ""))
	})
	t.Run("SameBearer", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/mcp", nil)
		header.SetAuthorization(r, "abc")
		assert.Same(t, r, mcpBearerRequest(r, "abc"))
	})
	t.Run("XAuthToken", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/mcp", nil)
		r.Header.Set(header.XAuthToken, "abc")
		result := mcpBearerRequest(r, "abc")
		assert.NotSame(t, r, result)
		assert.Equal(t, "Bearer abc", result.Header.Get(header.Auth))
		assert.Empty(t, r.Header.Get(header.Auth))
	})
	t.Run("OtherBearer", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/mcp", nil)
		r.Header.Set(header.XAuthToken, "abc")
		header.SetAuthorization(r, "def")
		result := mcpBearerRequest(r, "abc")
		assert.Equal(t, "Bearer abc", result.Header.Get(header.Auth))
		assert.Equal(t, "Bearer def", r.Header.Get(header.Auth))
	})
}

func TestMcpSessionHandler(t *testing.T) {
	var got *auth.TokenInfo

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = auth.TokenInfoFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	t.Run("NoTokenInfo", func(t *testing.T) {
		got = nil
		h := mcpSessionHandler(inner, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/mcp", nil))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Nil(t, got)
	})
	t.Run("TokenInfo", func(t *testing.T) {
		got = nil
		info := &auth.TokenInfo{UserID: "uqxetse3cy5eo9z2"}
		h := mcpSessionHandler(inner, info)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/mcp", nil)
		header.SetAuthorization(r, "abc")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Same(t, info, got)
	})
	t.Run("NoBearer", func(t *testing.T) {
		got = nil
		h := mcpSessionHandler(inner, &auth.TokenInfo{UserID: "uqxetse3cy5eo9z2"})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/mcp", nil))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Nil(t, got)
	})
}
