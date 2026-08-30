package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// AuthJWTKeysCommand groups JWT key management helpers.
var AuthJWTKeysCommand = &cli.Command{
	Name:  "keys",
	Usage: "JWT signing key helpers",
	Subcommands: []*cli.Command{
		AuthJWTKeysListCommand,
		AuthJWTKeysRotateCommand,
	},
}

// AuthJWTKeysRotateCommand replaces the active JWT signing key.
var AuthJWTKeysRotateCommand = &cli.Command{
	Name:      "rotate",
	Usage:     "Replaces the active JWT signing key",
	ArgsUsage: "",
	Flags: []cli.Flag{
		JsonFlag(),
	},
	Action: authJWTKeysRotateAction,
}

// authJWTKeysRotateAction issues a new portal signing key and reports both key IDs.
// The replaced key stays in the JWKS for jwt.RotationOverlap.
func authJWTKeysRotateAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if err := requirePortal(conf); err != nil {
			return err
		}

		manager := get.JWTManager()
		if manager == nil {
			return cli.Exit(errors.New("jwt manager not available"), 1)
		}

		prev, _ := manager.ActiveKey()

		// A new key signs as soon as it exists, so an error alongside one means only the
		// retirement is outstanding. A plain failure would invite a key-minting retry.
		key, err := manager.RotateKey()
		if err != nil && key == nil {
			return cli.Exit(err, 1)
		} else if err != nil {
			fmt.Printf("Rotated to %s, but retiring the previous key did not complete: %s\n", key.Kid, err)
		}

		event.AuditInfo([]string{"cli", "jwt", "rotate signing key", status.Succeeded})

		prevKid := ""
		retired := ""
		if prev != nil && prev.Kid != key.Kid {
			prevKid = prev.Kid
			for _, k := range manager.AllKeys() {
				if k.Kid == prevKid && k.NotAfter > 0 {
					retired = time.Unix(k.NotAfter, 0).UTC().Format(time.RFC3339)
				}
			}
		}

		if ctx.Bool("json") {
			return printJSON(map[string]any{
				"kid":      key.Kid,
				"replaced": prevKid,
				"notAfter": retired,
			})
		}

		fmt.Println()
		fmt.Printf("New signing key: %s\n", key.Kid)
		switch {
		case prevKid == "":
			fmt.Println("No previous key to replace.")
		case retired == "":
			fmt.Printf("Replaced key %s.\n", prevKid)
		default:
			fmt.Printf("Replaced key %s, which verifies until %s.\n", prevKid, retired)
		}
		fmt.Println("A running portal loads its keys at startup, so restart it to use the new key.")
		fmt.Println()

		return nil
	})
}

// AuthJWTKeysListCommand lists JWT signing keys.
var AuthJWTKeysListCommand = &cli.Command{
	Name:      "ls",
	Usage:     "Lists JWT signing keys",
	Aliases:   []string{"list"},
	ArgsUsage: "",
	Flags: []cli.Flag{
		JsonFlag(),
	},
	Action: authJWTKeysListAction,
}

// authJWTKeysListAction lists portal signing keys with metadata.
func authJWTKeysListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if err := requirePortal(conf); err != nil {
			return err
		}

		manager := get.JWTManager()
		if manager == nil {
			return cli.Exit(errors.New("jwt manager not available"), 1)
		}

		keys := manager.AllKeys()
		active, _ := manager.ActiveKey()
		activeKid := ""
		if active != nil {
			activeKid = active.Kid
		}

		type keyInfo struct {
			Kid       string `json:"kid"`
			CreatedAt string `json:"createdAt"`
			NotAfter  string `json:"notAfter,omitempty"`
			Active    bool   `json:"active"`
		}

		rows := make([]keyInfo, 0, len(keys))
		for _, k := range keys {
			info := keyInfo{Kid: k.Kid, Active: k.Kid == activeKid}
			if k.CreatedAt > 0 {
				info.CreatedAt = time.Unix(k.CreatedAt, 0).UTC().Format(time.RFC3339)
			}
			if k.NotAfter > 0 {
				info.NotAfter = time.Unix(k.NotAfter, 0).UTC().Format(time.RFC3339)
			}
			rows = append(rows, info)
		}

		if ctx.Bool("json") {
			payload := map[string]any{
				"keys": rows,
			}
			return printJSON(payload)
		}

		if len(rows) == 0 {
			fmt.Println()
			fmt.Println("No signing keys found.")
			fmt.Println()
			return nil
		}

		fmt.Println()
		fmt.Println("JWT signing keys:")
		for _, row := range rows {
			stat := ""
			if row.Active {
				stat = " (active)"
			}
			parts := []string{fmt.Sprintf("KID: %s%s", row.Kid, stat)}
			if row.CreatedAt != "" {
				parts = append(parts, fmt.Sprintf("created %s", row.CreatedAt))
			}
			if row.NotAfter != "" {
				parts = append(parts, fmt.Sprintf("expires %s", row.NotAfter))
			}
			fmt.Printf("- %s\n", strings.Join(parts, ", "))
		}
		fmt.Println()
		return nil
	})
}
