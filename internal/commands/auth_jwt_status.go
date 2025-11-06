package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// AuthJWTStatusCommand reports verifier cache diagnostics.
var AuthJWTStatusCommand = &cli.Command{
	Name:  "status",
	Usage: "Shows JWT verifier cache status",
	Flags: []cli.Flag{
		JsonFlag(),
	},
	Action: authJWTStatusAction,
}

// authJWTStatusAction prints JWKS cache diagnostics for the current node.
func authJWTStatusAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		verifier := get.JWTVerifier()
		if verifier == nil {
			return cli.Exit(errors.New("jwt verifier not available"), 1)
		}

		ttl := time.Duration(conf.JWKSCacheTTL()) * time.Second
		s := verifier.Status(ttl)
		s.JWKSURL = strings.TrimSpace(conf.JWKSUrl())

		if ctx.Bool("json") {
			return printJSON(s)
		}

		fmt.Println()
		fmt.Printf("JWKS URL: %s\n", s.JWKSURL)
		fmt.Printf("Cache Path: %s\n", s.CachePath)
		fmt.Printf("Cache URL: %s\n", s.CacheURL)
		fmt.Printf("Cache ETag: %s\n", s.CacheETag)
		fmt.Printf("Cached Keys: %d\n", s.KeyCount)
		if len(s.KeyIDs) > 0 {
			fmt.Printf("Key IDs: %s\n", strings.Join(s.KeyIDs, ", "))
		}
		if !s.CacheFetchedAt.IsZero() {
			fmt.Printf("Last Fetch: %s\n", s.CacheFetchedAt.Format(time.RFC3339))
		} else {
			fmt.Println("Last Fetch: never")
		}
		fmt.Printf("Cache Age: %ds\n", s.CacheAgeSeconds)
		if s.CacheTTLSeconds > 0 {
			fmt.Printf("Cache TTL: %ds\n", s.CacheTTLSeconds)
		}
		if s.CacheStale {
			fmt.Println("Cache Status: STALE")
		} else {
			fmt.Println("Cache Status: fresh")
		}
		fmt.Println()
		return nil
	})
}
