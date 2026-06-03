package api

import "github.com/gin-gonic/gin"

// Cluster OAuth/OIDC delegation hooks.
//
// The shared /api/v1/oauth/{authorize,token,userinfo} handlers serve general
// OAuth2 clients on every edition. On Portal builds the same endpoints must also
// drive the cluster OIDC OP flow (which issues EdDSA id_tokens to cluster
// instances). Because internal/api must not import portal/*, the Portal injects
// that behavior by assigning these function variables during extension
// registration (mirroring the server.WebDAVHandler pattern).
//
// Each hook returns handled=true only when it has fully written the response;
// returning false (or being nil) lets the default CE flow proceed unchanged, so
// non-cluster requests are byte-for-byte identical to a non-Portal build.
var (
	// OAuthClusterAuthorize handles a cluster-instance authorize request
	// (client with the instance role) with a direct OP redirect instead of the
	// general consent page.
	OAuthClusterAuthorize func(c *gin.Context) (handled bool)

	// OAuthClusterToken redeems a cluster authorization code (cluster_oidc_codes)
	// for an EdDSA id_token. It defers every other grant — including instance
	// client_credentials and general authorization codes — to the opaque path.
	OAuthClusterToken func(c *gin.Context) (handled bool)

	// OAuthClusterUserinfo serves userinfo for a cluster OP access token. It
	// defers regular session tokens to the CE userinfo handler.
	OAuthClusterUserinfo func(c *gin.Context) (handled bool)
)
