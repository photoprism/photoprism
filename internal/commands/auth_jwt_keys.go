package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// AuthJWTKeysCommand groups JWT key management helpers.
var AuthJWTKeysCommand = &cli.Command{
	Name:  "keys",
	Usage: "JWT signing key helpers",
	Subcommands: []*cli.Command{
		AuthJWTKeysListCommand,
	},
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
			status := ""
			if row.Active {
				status = " (active)"
			}
			parts := []string{fmt.Sprintf("KID: %s%s", row.Kid, status)}
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
