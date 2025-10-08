package commands

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
)

var allowedJWTScope = func() map[string]struct{} {
	out := make(map[string]struct{}, len(acl.ResourceNames))
	for _, res := range acl.ResourceNames {
		out[res.String()] = struct{}{}
	}
	return out
}()

// requirePortal returns a CLI error when the active configuration is not a portal node.
func requirePortal(conf *config.Config) error {
	if conf == nil || !conf.IsPortal() {
		return cli.Exit(errors.New("command requires a Portal node"), 2)
	}
	return nil
}

// resolveNode finds a node by UUID, client ID, or DNS label using the portal registry.
func resolveNode(conf *config.Config, identifier string) (*reg.Node, error) {
	if err := requirePortal(conf); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(identifier)
	if key == "" {
		return nil, cli.Exit(errors.New("node identifier required"), 2)
	}

	registry, err := reg.NewClientRegistryWithConfig(conf)
	if err != nil {
		return nil, cli.Exit(err, 1)
	}

	if node, err := registry.FindByNodeUUID(key); err == nil && node != nil {
		return node, nil
	}
	if node, err := registry.FindByClientID(key); err == nil && node != nil {
		return node, nil
	}

	name := clean.DNSLabel(key)
	if name == "" {
		return nil, cli.Exit(errors.New("invalid node identifier"), 2)
	}

	node, err := registry.FindByName(name)
	if err != nil {
		if errors.Is(err, reg.ErrNotFound) {
			return nil, cli.Exit(fmt.Errorf("node %q not found", identifier), 3)
		}
		return nil, cli.Exit(err, 1)
	}
	return node, nil
}

// decodeJWTClaims decodes the compact JWT and returns header and claims without verifying the signature.
func decodeJWTClaims(token string) (map[string]any, *clusterjwt.Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, errors.New("jwt: token must contain three segments")
	}

	decode := func(segment string) ([]byte, error) {
		return base64.RawURLEncoding.DecodeString(segment)
	}

	headerBytes, err := decode(parts[0])
	if err != nil {
		return nil, nil, err
	}
	payloadBytes, err := decode(parts[1])
	if err != nil {
		return nil, nil, err
	}

	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, nil, err
	}

	claims := &clusterjwt.Claims{}
	if err := json.Unmarshal(payloadBytes, claims); err != nil {
		return nil, nil, err
	}

	return header, claims, nil
}

// verifyPortalToken verifies a JWT using the portal's in-memory key manager.
func verifyPortalToken(conf *config.Config, token string, expected clusterjwt.ExpectedClaims) (*clusterjwt.Claims, error) {
	if err := requirePortal(conf); err != nil {
		return nil, err
	}

	manager := get.JWTManager()
	if manager == nil {
		return nil, cli.Exit(errors.New("jwt issuer not available"), 1)
	}

	jwks := manager.JWKS()
	if jwks == nil || len(jwks.Keys) == 0 {
		return nil, cli.Exit(errors.New("jwks key set is empty"), 1)
	}

	leeway := time.Duration(conf.JWTLeeway()) * time.Second
	if leeway <= 0 {
		leeway = 60 * time.Second
	}

	claims, err := clusterjwt.VerifyTokenWithKeys(token, expected, jwks.Keys, leeway)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// normalizeScopes trims and de-duplicates scope values, falling back to defaults when necessary.
func normalizeScopes(values []string, defaults ...string) ([]string, error) {
	src := values
	if len(src) == 0 {
		src = defaults
	}
	out := make([]string, 0, len(src))
	seen := make(map[string]struct{}, len(src))
	for _, raw := range src {
		for _, parsed := range clean.Scopes(raw) {
			scope := clean.Scope(parsed)
			if scope == "" {
				continue
			}
			if _, exists := seen[scope]; exists {
				continue
			}
			if _, ok := allowedJWTScope[scope]; !ok {
				return nil, cli.Exit(fmt.Errorf("unsupported scope %q", scope), 2)
			}
			seen[scope] = struct{}{}
			out = append(out, scope)
		}
	}
	if len(out) == 0 {
		return nil, cli.Exit(errors.New("at least one scope is required"), 2)
	}
	return out, nil
}

// printJSON pretty-prints the payload as JSON.
func printJSON(payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return cli.Exit(err, 1)
	}
	fmt.Printf("%s\n", data)
	return nil
}
