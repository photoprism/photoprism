package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
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
		NodeRole:     cluster.RoleApp,
		RotateSecret: true,
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	header.SetAuthorization(req, joinToken)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
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
