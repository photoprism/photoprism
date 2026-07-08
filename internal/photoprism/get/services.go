package get

import (
	gc "github.com/patrickmn/go-cache"

	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/auth/oidc"
	"github.com/photoprism/photoprism/internal/auth/session"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
)

var conf *config.Config

var services struct {
	FolderCache *gc.Cache
	CoverCache  *gc.Cache
	ThumbCache  *gc.Cache
	Convert     *photoprism.Convert
	Files       *photoprism.Files
	Photos      *photoprism.Photos
	Import      *photoprism.Import
	Index       *photoprism.Index
	Moments     *photoprism.Moments
	Faces       *photoprism.Faces
	Places      *photoprism.Places
	Purge       *photoprism.Purge
	CleanUp     *photoprism.CleanUp
	Query       *query.Query
	Thumbs      *photoprism.Thumbs
	Session     *session.Session
	OIDC        *oidc.Client
	JWTManager  *clusterjwt.Manager
	JWTIssuer   *clusterjwt.Issuer
	JWTVerifier *clusterjwt.Verifier
}

// SetConfig stores the shared Config for service constructors.
func SetConfig(c *config.Config) {
	if c == nil {
		log.Panic("panic: argument is nil in get.SetConfig(c *config.Config)")
		return
	}

	resetJWT()
	resetOIDC()

	conf = c

	photoprism.SetConfig(c)
}

// Config returns the shared Config used by the service registry.
func Config() *config.Config {
	if conf == nil {
		log.Panic("panic: conf is nil in get.Config()")
		return nil
	}

	return conf
}
