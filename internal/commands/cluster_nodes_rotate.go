package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
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
	Flags:     append([]cli.Flag{rotateDatabaseFlag, rotateSecretFlag, &cli.BoolFlag{Name: "yes", Aliases: []string{"y"}, Usage: "runs the command non-interactively"}, rotatePortalURL, rotatePortalTok}, report.CliFlags...),
	Action:    clusterNodesRotateAction,
}

func clusterNodesRotateAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		key := ctx.Args().First()
		if key == "" {
			return cli.Exit(fmt.Errorf("node id or name is required"), 2)
		}

		// Determine node name. On portal, resolve id->name via registry; otherwise treat key as name.
		name := clean.DNSLabel(key)
		if conf.IsPortal() {
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
		if portalURL == "" {
			return cli.Exit(fmt.Errorf("portal URL is required (use --portal-url or set portal-url)"), 2)
		}
		token := ctx.String("join-token")
		if token == "" {
			token = conf.JoinToken()
		}
		if token == "" {
			token = os.Getenv(config.EnvVar("join-token"))
		}
		if token == "" {
			return cli.Exit(fmt.Errorf("portal token is required (use --join-token or set join-token)"), 2)
		}

		// Default: rotate DB only if no flag given (safer default)
		rotateDatabase := ctx.Bool("database") || (!ctx.IsSet("database") && !ctx.IsSet("secret"))
		rotateSecret := ctx.Bool("secret")

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

		body := map[string]interface{}{
			"nodeName":     name,
			"rotate":       rotateDatabase,
			"rotateSecret": rotateSecret,
		}
		b, _ := json.Marshal(body)

		url := stringsTrimRightSlash(portalURL) + "/api/v1/cluster/nodes/register"
		var resp cluster.RegisterResponse
		if err := postWithBackoff(url, token, b, &resp); err != nil {
			// Map common HTTP errors similarly to register command
			if he, ok := err.(*httpError); ok {
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
