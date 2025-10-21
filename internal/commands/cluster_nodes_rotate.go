package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/internal/service/cluster/theme"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

var (
	rotateDatabaseFlag = &cli.BoolFlag{Name: "database", Aliases: []string{"db"}, Usage: "rotate DB credentials"}
	rotateSecretFlag   = &cli.BoolFlag{Name: "secret", Usage: "rotate node secret"}
	rotatePortalURL    = &cli.StringFlag{Name: "portal-url", Usage: "Portal base `URL` (defaults to config)"}
	rotatePortalTok    = &cli.StringFlag{Name: "join-token", Usage: "Portal access `TOKEN` (defaults to config)"}
)

// ClusterNodesRotateCommand triggers rotation via the register endpoint.
var ClusterNodesRotateCommand = &cli.Command{
	Name:      "rotate",
	Usage:     "Rotates a node's DB and/or secret via Portal (HTTP)",
	ArgsUsage: "<id|name>",
	Flags: append([]cli.Flag{
		DryRunFlag("preview rotation without contacting the Portal"),
		rotateDatabaseFlag,
		rotateSecretFlag,
		rotatePortalURL,
		rotatePortalTok,
		YesFlag(),
	}, report.CliFlags...),
	Action: clusterNodesRotateAction,
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
			if portalURL == "" {
				log.Infof("dry-run: would rotate %s for %s (portal URL not set)", txt.JoinAnd(what), target)
			} else {
				log.Infof("dry-run: would rotate %s for %s via %s", txt.JoinAnd(what), target, clean.Log(portalURL))
			}
			return nil
		}

		if portalURL == "" {
			return cli.Exit(fmt.Errorf("portal URL is required (use --portal-url or set portal-url)"), 2)
		}
		if token == "" {
			return cli.Exit(fmt.Errorf("portal token is required (use --join-token or set join-token)"), 2)
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
				log.Infof("rotation cancelled for %s", clean.LogQuote(name))
				return nil
			}
		}

		payload := cluster.RegisterRequest{
			NodeName:       name,
			RotateDatabase: rotateDatabase,
			RotateSecret:   rotateSecret,
			AppName:        clean.TypeUnicode(conf.About()),
			AppVersion:     clean.TypeUnicode(conf.Version()),
		}

		if themeVersion, err := theme.DetectVersion(conf.ThemePath()); err == nil && themeVersion != "" {
			payload.Theme = themeVersion
		}

		b, _ := json.Marshal(payload)

		endpointUrl := stringsTrimRightSlash(portalURL) + "/api/v1/cluster/nodes/register"

		var resp cluster.RegisterResponse
		if err := postWithBackoff(endpointUrl, token, b, &resp); err != nil {
			// Map common HTTP errors similarly to register command
			var he *httpError
			if errors.As(err, &he) {
				switch he.Status {
				case 401, 403:
					return cli.Exit(fmt.Errorf("%s", he.Error()), 4)
				case 409:
					return cli.Exit(fmt.Errorf("%s", he.Error()), 5)
				case 400:
					return cli.Exit(fmt.Errorf("%s", he.Error()), 2)
				case 404:
					return cli.Exit(fmt.Errorf("%s", he.Error()), 3)
				case 429:
					return cli.Exit(fmt.Errorf("%s", he.Error()), 6)
				}
			}
			return cli.Exit(err, 1)
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
			"rotate node %s",
		}
		args := []interface{}{clean.Log(nodeID)}
		if detail != "" {
			segments = append(segments, "%s")
			args = append(args, clean.Log(detail))
		}
		segments = append(segments, event.Succeeded)

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
			if resp.Secrets != nil && resp.Secrets.ClientSecret != "" && resp.Database.Password != "" {
				fmt.Printf("\n%s\n", report.Credentials("Node Client Secret", resp.Secrets.ClientSecret, "DB Password", resp.Database.Password))
			} else if resp.Secrets != nil && resp.Secrets.ClientSecret != "" {
				fmt.Printf("\n%s\n", report.Credentials("Node Client Secret", resp.Secrets.ClientSecret, "", ""))
			} else if resp.Database.Password != "" {
				fmt.Printf("\n%s\n", report.Credentials("DB User", resp.Database.User, "DB Password", resp.Database.Password))
			}
			if resp.Database.DSN != "" {
				fmt.Printf("DSN: %s\n", resp.Database.DSN)
			}
		}

		return nil
	})
}
