package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// AuthJWTIssueCommand issues portal-signed JWTs for cluster nodes.
var AuthJWTIssueCommand = &cli.Command{
	Name:  "issue",
	Usage: "Issues a portal-signed JWT for a node",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "node", Aliases: []string{"n"}, Usage: "target node uuid, client id, or DNS label", Required: true},
		&cli.StringSliceFlag{Name: "scope", Aliases: []string{"s"}, Usage: "token scope", Value: cli.NewStringSlice("cluster")},
		&cli.DurationFlag{Name: "ttl", Usage: "token lifetime", Value: clusterjwt.TokenTTL},
		&cli.StringFlag{Name: "subject", Usage: "token subject (default portal:<clusterUUID>)"},
		JsonFlag(),
	},
	Action: authJWTIssueAction,
}

// authJWTIssueAction handles CLI issuance of portal-signed JWTs for nodes.
func authJWTIssueAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if err := requirePortal(conf); err != nil {
			return err
		}

		node, err := resolveNode(conf, ctx.String("node"))
		if err != nil {
			return err
		}

		scopes, err := normalizeScopes(ctx.StringSlice("scope"), "cluster")
		if err != nil {
			return err
		}

		ttl := ctx.Duration("ttl")
		if ttl <= 0 {
			ttl = clusterjwt.TokenTTL
		}

		clusterUUID := strings.TrimSpace(conf.ClusterUUID())
		if clusterUUID == "" {
			return cli.Exit(fmt.Errorf("cluster uuid not configured"), 1)
		}

		subject := strings.TrimSpace(ctx.String("subject"))
		if subject == "" {
			subject = fmt.Sprintf("portal:%s", clusterUUID)
		}

		var token string
		if subject == fmt.Sprintf("portal:%s", clusterUUID) {
			token, err = get.IssuePortalJWTForNode(node.UUID, scopes, ttl)
		} else {
			spec := clusterjwt.ClaimsSpec{
				Issuer:   fmt.Sprintf("portal:%s", clusterUUID),
				Subject:  subject,
				Audience: fmt.Sprintf("node:%s", node.UUID),
				Scope:    scopes,
				TTL:      ttl,
			}
			token, err = get.IssuePortalJWT(spec)
		}
		if err != nil {
			return cli.Exit(err, 1)
		}

		header, claims, err := decodeJWTClaims(token)
		if err != nil {
			return cli.Exit(err, 1)
		}

		if ctx.Bool("json") {
			type nodePayload struct {
				UUID     string `json:"UUID"`
				ClientID string `json:"ClientID"`
				Name     string `json:"Name"`
				Role     string `json:"Role"`
			}
			response := struct {
				Token  string             `json:"Token"`
				Header map[string]any     `json:"Header"`
				Claims *clusterjwt.Claims `json:"Claims"`
				Node   nodePayload        `json:"Node"`
			}{
				Token:  token,
				Header: header,
				Claims: claims,
				Node: nodePayload{
					UUID:     node.UUID,
					ClientID: node.ClientID,
					Name:     node.Name,
					Role:     string(node.Role),
				},
			}
			return printJSON(response)
		}

		expires := "unknown"
		if claims.ExpiresAt != nil {
			expires = claims.ExpiresAt.Time.UTC().Format(time.RFC3339)
		}
		audience := strings.Join(claims.Audience, " ")
		if audience == "" {
			audience = "(none)"
		}

		fmt.Printf("\nIssued JWT for node %s (%s)\n", node.Name, node.UUID)
		fmt.Printf("Scopes: %s\n", strings.Join(scopes, " "))
		fmt.Printf("Expires: %s\n", expires)
		fmt.Printf("Audience: %s\n", audience)
		fmt.Printf("Subject: %s\n", claims.Subject)
		fmt.Printf("Key ID: %v\n", header["kid"])
		fmt.Printf("\n%s\n", token)

		return nil
	})
}
