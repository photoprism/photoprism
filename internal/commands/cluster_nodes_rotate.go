package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/internal/service/cluster/provisioner"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/internal/service/cluster/theme"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// provisionerTimeout bounds a local database credential rotation, and exceeds what
// provisioner.EnsureCredentials budgets across its statements.
const provisionerTimeout = 90 * time.Second

var (
	rotateDatabaseFlag = &cli.BoolFlag{Name: "database", Aliases: []string{"db"}, Usage: "rotate DB credentials"}
	rotateSecretFlag   = &cli.BoolFlag{Name: "secret", Usage: "rotate node secret"}
	rotatePortalURL    = &cli.StringFlag{Name: "portal-url", Usage: "Portal base `URL` (defaults to config)"}
	rotatePortalTok    = &cli.StringFlag{Name: "join-token", Usage: "Portal access `TOKEN` (defaults to config)"}
)

// ClusterNodesRotateCommand rotates a node's credentials, locally on a Portal and through
// the Portal on an instance.
var ClusterNodesRotateCommand = &cli.Command{
	Name:      "rotate",
	Usage:     "Rotates the database credentials and/or secret of a node",
	ArgsUsage: "<id|name>",
	Flags: append([]cli.Flag{
		DryRunFlag("preview rotation without replacing any credentials"),
		rotateDatabaseFlag,
		rotateSecretFlag,
		rotatePortalURL,
		rotatePortalTok,
		YesFlag(),
	}, report.CliFlags...),
	Action: clusterNodesRotateAction,
}

// rotateNodeInRegistry rotates a node's secret and database credentials through the local
// registry, and reports the result with the operator exposure policy. The Portal-only fields
// of a register response (JWKS and login URLs, the theme hint) do not apply here.
func rotateNodeInRegistry(conf *config.Config, name string, rotateDatabase, rotateSecret bool) (cluster.RegisterResponse, error) {
	resp := cluster.RegisterResponse{UUID: conf.ClusterUUID(), ClusterCIDR: conf.ClusterCIDR()}

	regy, err := reg.NewClientRegistryWithConfig(conf)

	if err != nil {
		return resp, cli.Exit(err, 1)
	}

	n, err := regy.FindByName(name)

	if err != nil || n == nil {
		return resp, cli.Exit(fmt.Errorf("node %s is not registered", clean.LogQuote(name)), 3)
	}

	// The registry shares its name space with ordinary OAuth clients, so only a node
	// registration is eligible.
	if role := cluster.NormalizeNodeRole(n.Role); n.UUID == "" || (role != cluster.RoleInstance && role != cluster.RoleService) {
		return resp, cli.Exit(fmt.Errorf("%s is not a cluster node", clean.LogQuote(name)), 3)
	}

	if rotateSecret {
		rotated, rotateErr := regy.RotateSecretByClientID(n.ClientID)

		if rotateErr != nil {
			return resp, cli.Exit(fmt.Errorf("failed to rotate the secret of node %s: %w", clean.LogQuote(name), rotateErr), 1)
		}

		resp.Secrets = &cluster.RegisterSecrets{ClientSecret: rotated.ClientSecret, RotatedAt: rotated.RotatedAt}
		n = rotated
	}

	if rotateDatabase {
		ctx, cancel := context.WithTimeout(context.Background(), provisionerTimeout)
		defer cancel()

		creds, _, credsErr := provisioner.EnsureCredentials(ctx, conf, n.UUID, n.Name, true)

		if credsErr != nil {
			return resp, cli.Exit(credsErr, 5)
		}

		if n.Database == nil {
			n.Database = &cluster.NodeDatabase{}
		}

		n.Database.Name, n.Database.User, n.Database.Driver, n.Database.RotatedAt = creds.Name, creds.User, creds.Driver, creds.RotatedAt

		if err = regy.Put(n); err != nil {
			return resp, cli.Exit(err, 1)
		}

		resp.Database.Password, resp.Database.DSN, resp.Database.RotatedAt = creds.Password, creds.DSN, creds.RotatedAt
	}

	if n.Database != nil {
		driver := n.Database.Driver

		if driver == "" {
			driver = provisioner.DatabaseDriver
		}

		resp.Database.Driver, resp.Database.Host, resp.Database.Port = driver, conf.DatabaseHost(), conf.DatabasePort()
		resp.Database.Name, resp.Database.User = n.Database.Name, n.Database.User

		if resp.Database.RotatedAt == "" {
			resp.Database.RotatedAt = n.Database.RotatedAt
		}
	}

	resp.Node = reg.BuildClusterNode(*n, reg.NodeOptsForOperator())
	resp.AlreadyRegistered = true
	resp.AlreadyProvisioned = n.Database != nil && n.Database.Name != ""

	return resp, nil
}

// rotateNodeViaPortal asks the Portal to rotate the local node's credentials, authorized by
// an access token minted from the node's own client credentials.
func rotateNodeViaPortal(conf *config.Config, portalURL, joinToken, name string, rotateDatabase, rotateSecret bool) (cluster.RegisterResponse, error) {
	resp := cluster.RegisterResponse{}

	if portalURL == "" {
		return resp, cli.Exit(fmt.Errorf("portal URL is required (use --portal-url or set portal-url)"), 2)
	}

	token, err := clusterRegisterToken(conf, portalURL, joinToken, name)

	if err != nil {
		return resp, cli.Exit(err, clusterTokenExitCode(err))
	}

	payload := cluster.RegisterRequest{
		NodeName:       name,
		RotateDatabase: rotateDatabase,
		RotateSecret:   rotateSecret,
		AppName:        clean.TypeUnicode(conf.About()),
		AppVersion:     clean.TypeUnicode(conf.Version()),
	}

	if themeVersion, themeErr := theme.DetectVersion(conf.ThemePath()); themeErr == nil && themeVersion != "" {
		payload.Theme = themeVersion
	}

	endpointUrl := stringsTrimRightSlash(portalURL) + "/api/v1/cluster/nodes/register"

	if err = postWithBackoff(endpointUrl, token, marshalRegisterRequest(payload), &resp); err != nil {
		// Map the Portal's response status to a CLI exit code.
		var he *httpError

		if errors.As(err, &he) {
			switch he.Status {
			case 401, 403:
				return resp, cli.Exit(fmt.Errorf("%s", he.Error()), 4)
			case 409:
				return resp, cli.Exit(fmt.Errorf("%s", he.Error()), 5)
			case 400:
				return resp, cli.Exit(fmt.Errorf("%s", he.Error()), 2)
			case 404:
				return resp, cli.Exit(fmt.Errorf("%s", he.Error()), 3)
			case 429:
				return resp, cli.Exit(fmt.Errorf("%s", he.Error()), 6)
			}
		}

		return resp, cli.Exit(err, 1)
	}

	return resp, nil
}

func clusterNodesRotateAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		key := ctx.Args().First()
		if key == "" {
			return cli.Exit(fmt.Errorf("node id or name is required"), 2)
		}

		// Determine node name. On portal, resolve id->name via registry; otherwise treat key as name.
		name := clean.DNSLabel(key)
		if conf.Portal() {
			if r, err := reg.NewClientRegistryWithConfig(conf); err == nil {
				if n, err := r.FindByNodeUUID(key); err == nil && n != nil {
					name = n.Name
				} else if n, err := r.FindByClientID(key); err == nil && n != nil {
					name = n.Name
				} else if n, err := r.FindByName(clean.DNSLabel(key)); err == nil && n != nil {
					name = n.Name
				}
			}
		}
		if name == "" {
			return cli.Exit(fmt.Errorf("invalid node identifier"), 2)
		}

		// Portal URL and token
		portalURL := ctx.String("portal-url")
		if portalURL == "" {
			portalURL = conf.PortalUrl()
		}
		if portalURL == "" {
			portalURL = os.Getenv(config.EnvVar("portal-url"))
		}
		token := ctx.String("join-token")
		if token == "" {
			token = os.Getenv(config.EnvVar("join-token"))
		}
		if token == "" {
			token = conf.JoinToken()
		}

		// Default: rotate DB only if no flag given (safer default)
		rotateDatabase := ctx.Bool("database") || (!ctx.IsSet("database") && !ctx.IsSet("secret"))
		rotateSecret := ctx.Bool("secret")

		if ctx.Bool("dry-run") {
			target := clean.LogQuote(name)
			if target == "" {
				target = "(unnamed node)"
			}
			var what []string
			if rotateDatabase {
				what = append(what, "database credentials")
			}
			if rotateSecret {
				what = append(what, "node secret")
			}
			if len(what) == 0 {
				what = append(what, "no resources (no rotation flags set)")
			}
			switch {
			case conf.Portal():
				log.Infof("dry-run: would rotate %s for %s in the local registry", txt.JoinAnd(what), target)
			case portalURL == "":
				log.Infof("dry-run: would rotate %s for %s (portal URL not set)", txt.JoinAnd(what), target)
			default:
				log.Infof("dry-run: would rotate %s for %s via %s", txt.JoinAnd(what), target, clean.Log(portalURL))
			}
			return nil
		}

		if !conf.Portal() && portalURL == "" {
			return cli.Exit(fmt.Errorf("portal URL is required (use --portal-url or set portal-url)"), 2)
		}

		confirmed := RunNonInteractively(ctx.Bool("yes"))
		if !confirmed {
			var what string
			switch {
			case rotateDatabase && rotateSecret:
				what = "DB credentials and node secret"
			case rotateDatabase:
				what = "DB credentials"
			case rotateSecret:
				what = "node secret"
			}
			prompt := promptui.Prompt{Label: fmt.Sprintf("Rotate %s for %s?", what, clean.LogQuote(name)), IsConfirm: true}
			if _, err := prompt.Run(); err != nil {
				log.Infof("rotation canceled for %s", clean.LogQuote(name))
				return nil
			}
		}

		var resp cluster.RegisterResponse
		var err error

		// The register endpoint accepts a node's own access token, which a Portal cannot hold
		// for another node, so a Portal rotates through its local registry instead.
		if conf.Portal() {
			resp, err = rotateNodeInRegistry(conf, name, rotateDatabase, rotateSecret)
		} else {
			resp, err = rotateNodeViaPortal(conf, portalURL, token, name, rotateDatabase, rotateSecret)
		}

		if err != nil {
			return err
		}

		nodeID := resp.Node.UUID
		if nodeID == "" {
			nodeID = resp.Node.Name
		}

		rotatedParts := make([]string, 0, 2)
		if rotateDatabase {
			rotatedParts = append(rotatedParts, "database")
		}
		if rotateSecret {
			rotatedParts = append(rotatedParts, "secret")
		}

		detail := strings.Join(rotatedParts, ", ")

		who := clusterAuditWho(ctx, conf)
		segments := []string{
			string(acl.ResourceCluster),
			"rotate node", "%s",
		}
		args := []any{clean.Log(nodeID)}
		if detail != "" {
			segments = append(segments, "%s")
			args = append(args, clean.Log(detail))
		}
		segments = append(segments, status.Succeeded)

		event.AuditInfo(append(who, segments...), args...)

		if ctx.Bool("json") {
			jb, _ := json.Marshal(resp)
			fmt.Println(string(jb))
			return nil
		}

		cols := []string{"UUID", "ClientID", "Name", "Role", "DB Driver", "DB Name", "DB User", "Host", "Port"}
		rows := [][]string{{resp.Node.UUID, resp.Node.ClientID, resp.Node.Name, resp.Node.Role, resp.Database.Driver, resp.Database.Name, resp.Database.User, resp.Database.Host, fmt.Sprintf("%d", resp.Database.Port)}}
		out, _ := report.RenderFormat(rows, cols, report.CliFormat(ctx))
		fmt.Printf("\n%s\n", out)

		if (resp.Secrets != nil && resp.Secrets.ClientSecret != "") || resp.Database.Password != "" {
			fmt.Println("PLEASE WRITE DOWN THE FOLLOWING CREDENTIALS; THEY WILL NOT BE SHOWN AGAIN:")
			switch {
			case resp.Secrets != nil && resp.Secrets.ClientSecret != "" && resp.Database.Password != "":
				fmt.Printf("\n%s\n", report.Credentials("Node Client Secret", resp.Secrets.ClientSecret, "DB Password", resp.Database.Password))
			case resp.Secrets != nil && resp.Secrets.ClientSecret != "":
				fmt.Printf("\n%s\n", report.Credentials("Node Client Secret", resp.Secrets.ClientSecret, "", ""))
			case resp.Database.Password != "":
				fmt.Printf("\n%s\n", report.Credentials("DB User", resp.Database.User, "DB Password", resp.Database.Password))
			}
			if resp.Database.DSN != "" {
				fmt.Printf("DSN: %s\n", resp.Database.DSN)
			}
		}

		return nil
	})
}
