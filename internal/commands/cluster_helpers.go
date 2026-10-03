package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/internal/service/cluster/node"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// obtainClientCredentialsViaRegister calls the portal register endpoint using a join token
// to (re)register the node, rotating the secret when necessary, and returns client id/secret.
func obtainClientCredentialsViaRegister(portalURL, joinToken, nodeName string) (id, secret string, err error) {
	u, err := url.Parse(strings.TrimRight(portalURL, "/"))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", "", fmt.Errorf("invalid portal-url: %s", portalURL)
	}
	endpoint := *u
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api/v1/cluster/nodes/register"

	payload := cluster.RegisterRequest{
		NodeName:     nodeName,
		NodeRole:     cluster.RoleInstance,
		RotateSecret: true,
	}
	b := marshalRegisterRequest(payload)
	// endpoint is derived from a parsed portal URL with explicit scheme/host validation above.
	req, _ := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(b)) //nolint:gosec
	req.Header.Set("Content-Type", "application/json")
	header.SetAuthorization(req, joinToken)

	resp, err := (&http.Client{}).Do(req) //nolint:gosec
	if err != nil {
		return "", "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Debugf("cluster: %s (close register response body)", clean.ErrorFull(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return "", "", fmt.Errorf("%s", resp.Status)
	}
	var regResp cluster.RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return "", "", err
	}
	id = regResp.Node.ClientID
	if regResp.Secrets != nil {
		secret = regResp.Secrets.ClientSecret
	}
	if id == "" || secret == "" {
		return "", "", fmt.Errorf("missing client credentials in response")
	}
	return id, secret, nil
}

// ErrMissingPortalToken reports that a register request has no credential to send.
var ErrMissingPortalToken = errors.New("portal token is required (use --join-token or set join-token)")

// clusterTokenExitCode maps a clusterRegisterToken failure to a CLI exit code, so a missing
// flag reads as a usage error and a rejected credential reads as an authentication error.
func clusterTokenExitCode(err error) int {
	if errors.Is(err, ErrMissingPortalToken) {
		return 2
	}

	return 4
}

// clusterRegisterToken returns the bearer token for a register request against the Portal.
// A node mutating its own registration uses an access token minted from its client credentials.
// A first join, or credentials the Portal no longer honors, uses the join token; the Portal
// still refuses that for a name it holds, so the fallback widens nothing.
func clusterRegisterToken(conf *config.Config, portalURL, joinToken, nodeName string) (string, error) {
	id, secret := strings.TrimSpace(conf.NodeClientID()), strings.TrimSpace(conf.NodeClientSecret())

	if id == "" || secret == "" || !strings.EqualFold(conf.NodeName(), nodeName) {
		if joinToken == "" {
			return "", ErrMissingPortalToken
		}

		return joinToken, nil
	}

	u, err := url.Parse(strings.TrimRight(portalURL, "/"))

	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid portal-url: %s", clean.Log(portalURL))
	}

	token, err := node.OAuthAccessToken(u, id, secret, node.OAuthScope)

	if err == nil {
		return token, nil
	}

	if joinToken == "" {
		return "", fmt.Errorf("portal access token request failed: %w", err)
	}

	log.Warnf("cluster: %s, retrying with the join token", clean.ErrorFull(err))

	return joinToken, nil
}

// marshalRegisterRequest JSON-encodes a cluster register payload for portal requests.
func marshalRegisterRequest(payload cluster.RegisterRequest) []byte {
	b, _ := json.Marshal(payload) //nolint:gosec

	return b
}

// clusterAuditWho builds the leading audit log segments for CLI commands.
func clusterAuditWho(ctx *cli.Context, conf *config.Config) []string {
	actor := clean.Log(conf.NodeName())
	if actor == "" {
		actor = clean.Log(conf.SiteUrl())
	}
	if actor == "" {
		actor = "cli"
	}

	context := "cli"
	if ctx != nil && ctx.Command != nil {
		if full := strings.TrimSpace(ctx.Command.FullName()); full != "" {
			context = "cli " + full
		}
	}

	return []string{actor, context}
}
