package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
)

// AuthJWTInspectCommand inspects and verifies portal-issued JWTs.
var AuthJWTInspectCommand = &cli.Command{
	Name:      "inspect",
	Usage:     "Decodes and verifies a portal JWT",
	ArgsUsage: "<token>",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "read token from file"},
		&cli.StringFlag{Name: "expect-audience", Usage: "expected audience (e.g., node:<uuid>)"},
		&cli.StringSliceFlag{Name: "require-scope", Usage: "require specific scope(s)"},
		&cli.BoolFlag{Name: "skip-verify", Usage: "decode without signature verification"},
		JsonFlag(),
	},
	Action: authJWTInspectAction,
}

// authJWTInspectAction decodes and optionally verifies a portal-issued JWT.
func authJWTInspectAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if err := requirePortal(conf); err != nil {
			return err
		}

		token, err := readTokenInput(ctx)
		if err != nil {
			return err
		}

		header, claims, err := decodeJWTClaims(token)
		if err != nil {
			return cli.Exit(err, 1)
		}

		var verified bool
		tokenScopes := clean.Scopes(claims.Scope)

		if !ctx.Bool("skip-verify") {
			expected := clusterjwt.ExpectedClaims{}
			if clusterUUID := strings.TrimSpace(conf.ClusterUUID()); clusterUUID != "" {
				expected.Issuer = fmt.Sprintf("portal:%s", clusterUUID)
			} else if portal := strings.TrimSpace(conf.PortalUrl()); portal != "" {
				expected.Issuer = strings.TrimRight(portal, "/")
			}

			if expectAud := strings.TrimSpace(ctx.String("expect-audience")); expectAud != "" {
				expected.Audience = expectAud
			} else if len(claims.Audience) > 0 {
				expected.Audience = claims.Audience[0]
			}

			if required := ctx.StringSlice("require-scope"); len(required) > 0 {
				scopes, scopeErr := normalizeScopes(required)
				if scopeErr != nil {
					return scopeErr
				}
				expected.Scope = scopes
			} else {
				expected.Scope = tokenScopes
			}

			if _, err := verifyPortalToken(conf, token, expected); err != nil {
				return cli.Exit(err, 1)
			}
			verified = true
		}

		if ctx.Bool("json") {
			payload := map[string]any{
				"token":    token,
				"verified": verified,
				"header":   header,
				"claims":   claims,
			}
			return printJSON(payload)
		}

		fmt.Println()
		fmt.Println("JWT header:")
		for k, v := range header {
			fmt.Printf("  %s: %v\n", k, v)
		}

		fmt.Println("\nJWT claims:")
		fmt.Printf("  issuer: %s\n", claims.Issuer)
		fmt.Printf("  subject: %s\n", claims.Subject)
		fmt.Printf("  audience: %s\n", strings.Join(claims.Audience, " "))
		fmt.Printf("  scope: %s\n", strings.Join(tokenScopes, " "))
		if claims.IssuedAt != nil {
			fmt.Printf("  issuedAt: %s\n", claims.IssuedAt.Time.UTC().Format(time.RFC3339))
		}
		if claims.ExpiresAt != nil {
			fmt.Printf("  expiresAt: %s\n", claims.ExpiresAt.Time.UTC().Format(time.RFC3339))
		}
		if claims.NotBefore != nil {
			fmt.Printf("  notBefore: %s\n", claims.NotBefore.Time.UTC().Format(time.RFC3339))
		}
		if claims.ID != "" {
			fmt.Printf("  jti: %s\n", claims.ID)
		}

		if verified {
			fmt.Println("\nSignature: verified")
		} else {
			fmt.Println("\nSignature: not verified (skipped)")
		}

		fmt.Printf("\nToken:\n%s\n\n", token)
		return nil
	})
}

// readTokenInput loads the token from CLI args, file, or STDIN.
func readTokenInput(ctx *cli.Context) (string, error) {
	if file := strings.TrimSpace(ctx.String("file")); file != "" {
		data, err := os.ReadFile(file) //nolint:gosec // user-supplied path is intended
		if err != nil {
			return "", cli.Exit(err, 1)
		}
		return strings.TrimSpace(string(data)), nil
	}

	if ctx.Args().Len() == 0 {
		return "", cli.Exit(errors.New("token argument required"), 2)
	}

	token := strings.TrimSpace(ctx.Args().First())
	if token == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", cli.Exit(err, 1)
		}
		token = strings.TrimSpace(string(data))
	}
	if token == "" {
		return "", cli.Exit(errors.New("token argument required"), 2)
	}
	return token, nil
}
