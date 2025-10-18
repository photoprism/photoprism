package commands

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
)

// flags for nodes mod
var (
	nodesModRoleFlag = &cli.StringFlag{Name: "role", Aliases: []string{"t"}, Usage: "node `ROLE` (portal, instance, service)"}
	nodesModInternal = &cli.StringFlag{Name: "advertise-url", Aliases: []string{"i"}, Usage: "internal service `URL`"}
	nodesModLabel    = &cli.StringSliceFlag{Name: "label", Aliases: []string{"l"}, Usage: "`k=v` label (repeatable)"}
)

// ClusterNodesModCommand updates node fields.
var ClusterNodesModCommand = &cli.Command{
	Name:      "mod",
	Usage:     "Updates node properties",
	ArgsUsage: "<id|name>",
	Flags: []cli.Flag{
		DryRunFlag("preview updates without modifying the registry"),
		nodesModRoleFlag,
		nodesModInternal,
		nodesModLabel,
		YesFlag(),
	},
	Hidden: true, // Required for cluster-management only.
	Action: clusterNodesModAction,
}

func clusterNodesModAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if !conf.Portal() {
			return cli.Exit(fmt.Errorf("node update is only available on a Portal node"), 2)
		}

		key := ctx.Args().First()
		if key == "" {
			return cli.Exit(fmt.Errorf("node id or name is required"), 2)
		}

		r, err := reg.NewClientRegistryWithConfig(conf)
		if err != nil {
			return cli.Exit(err, 1)
		}

		// Resolve by NodeUUID first, then by client UID, then by normalized name.
		var n *reg.Node
		var getErr error
		if n, getErr = r.FindByNodeUUID(key); getErr != nil || n == nil {
			n, getErr = r.FindByClientID(key)
		}
		if getErr != nil || n == nil {
			name := clean.DNSLabel(key)
			if name == "" {
				return cli.Exit(fmt.Errorf("invalid node identifier"), 2)
			}
			n, getErr = r.FindByName(name)
		}
		if getErr != nil || n == nil {
			return cli.Exit(fmt.Errorf("node not found"), 3)
		}

		changes := make([]string, 0, 4)

		if v := ctx.String("role"); v != "" {
			n.Role = clean.TypeLowerDash(v)
			changes = append(changes, fmt.Sprintf("role=%s", clean.Log(n.Role)))
		}
		if v := ctx.String("advertise-url"); v != "" {
			n.AdvertiseUrl = v
			changes = append(changes, fmt.Sprintf("advertise-url=%s", clean.Log(n.AdvertiseUrl)))
		}
		if labels := ctx.StringSlice("label"); len(labels) > 0 {
			if n.Labels == nil {
				n.Labels = map[string]string{}
			}
			for _, kv := range labels {
				if k, v, ok := splitKV(kv); ok {
					n.Labels[k] = v
				}
			}
			changes = append(changes, fmt.Sprintf("labels+=%s", clean.Log(strings.Join(labels, ","))))
		}

		if ctx.Bool("dry-run") {
			if len(changes) == 0 {
				log.Infof("dry-run: no updates to apply for node %s", clean.LogQuote(n.Name))
			} else {
				log.Infof("dry-run: would update node %s (%s)", clean.LogQuote(n.Name), strings.Join(changes, ", "))
			}
			return nil
		}

		confirmed := RunNonInteractively(ctx.Bool("yes"))
		if !confirmed {
			prompt := promptui.Prompt{Label: fmt.Sprintf("Update node %s?", clean.LogQuote(n.Name)), IsConfirm: true}
			if _, err := prompt.Run(); err != nil {
				log.Infof("update cancelled for %s", clean.LogQuote(n.Name))
				return nil
			}
		}

		if err := r.Put(n); err != nil {
			return cli.Exit(err, 1)
		}

		log.Infof("node %s has been updated", clean.LogQuote(n.Name))
		return nil
	})
}

func splitKV(s string) (string, string, bool) {
	if s == "" {
		return "", "", false
	}
	i := strings.IndexByte(s, '=')
	if i <= 0 || i >= len(s)-1 {
		return "", "", false
	}
	return s[:i], s[i+1:], true
}
