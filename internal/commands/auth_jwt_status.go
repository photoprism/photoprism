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
		status := verifier.Status(ttl)
		status.JWKSURL = strings.TrimSpace(conf.JWKSUrl())

		if ctx.Bool("json") {
			return printJSON(status)
		}

		fmt.Println()
		fmt.Printf("JWKS URL: %s\n", status.JWKSURL)
		fmt.Printf("Cache Path: %s\n", status.CachePath)
		fmt.Printf("Cache URL: %s\n", status.CacheURL)
		fmt.Printf("Cache ETag: %s\n", status.CacheETag)
		fmt.Printf("Cached Keys: %d\n", status.KeyCount)
		if len(status.KeyIDs) > 0 {
			fmt.Printf("Key IDs: %s\n", strings.Join(status.KeyIDs, ", "))
		}
		if !status.CacheFetchedAt.IsZero() {
			fmt.Printf("Last Fetch: %s\n", status.CacheFetchedAt.Format(time.RFC3339))
		} else {
			fmt.Println("Last Fetch: never")
		}
		fmt.Printf("Cache Age: %ds\n", status.CacheAgeSeconds)
		if status.CacheTTLSeconds > 0 {
			fmt.Printf("Cache TTL: %ds\n", status.CacheTTLSeconds)
		}
		if status.CacheStale {
			fmt.Println("Cache Status: STALE")
		} else {
			fmt.Println("Cache Status: fresh")
		}
		fmt.Println()
		return nil
	})
}
