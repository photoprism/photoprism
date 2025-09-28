package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

type healthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// ClusterHealthCommand prints a minimal health response (Portal-only).
var ClusterHealthCommand = &cli.Command{
	Name:   "health",
	Usage:  "Shows cluster health status",
	Flags:  report.CliFlags,
	Hidden: true, // Required for cluster-management only.
	Action: clusterHealthAction,
}

func clusterHealthAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if !conf.IsPortal() {
			return fmt.Errorf("cluster health is only available on a Portal node")
		}

		resp := healthResponse{Status: "ok", Time: time.Now().UTC().Format(time.RFC3339)}

		if ctx.Bool("json") {
			b, _ := json.Marshal(resp)
			fmt.Println(string(b))
			return nil
		}

		cols := []string{"Status", "Time"}
		rows := [][]string{{resp.Status, resp.Time}}
		out, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))
		fmt.Printf("\n%s\n", out)
		return err
	})
}
