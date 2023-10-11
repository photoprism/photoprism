package commands

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/report"
)

// ShowConfigYamlCommand configures the command name, flags, and action.
var ShowConfigYamlCommand = cli.Command{
	Name:   "config-yaml",
	Usage:  "Displays supported YAML config options and CLI flags",
	Flags:  report.CliFlags,
	Action: showConfigYamlAction,
}

// showConfigYamlAction displays supported YAML config options and CLI flag.
func showConfigYamlAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.TraceLevel)

	rows, cols := conf.Options().Report()

	// CSV Export?
	if ctx.Bool("csv") || ctx.Bool("tsv") {
		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Println(result)

		return err
	}

	type Section struct {
		Start string
		Title string
		Info  string
	}

	s := []Section{
		{Start: "AuthMode", Title: "Authentication"},
		{Start: "LogLevel", Title: "Logging"},
		{Start: "ConfigPath", Title: "Storage"},
		{Start: "Workers", Title: "Index Workers"},
		{Start: "ReadOnly", Title: "Feature Flags"},
		{Start: "DefaultTheme", Title: "Customization"},
		{Start: "CdnUrl", Title: "Site Information"},
		{Start: "HttpsProxy", Title: "Web Server"},
		{Start: "DatabaseDriver", Title: "Database Connection"},
		{Start: "SipsBin", Title: "File Converters"},
		{Start: "DownloadToken", Title: "Security Tokens"},
		{Start: "ThumbColor", Title: "Image Quality"},
		{Start: "PIDFilename", Title: "Daemon Mode",
			Info: "If you start the server as a *daemon* in the background, you can additionally specify a filename for the log and the process ID:"},
	}

	j := 0

	for i, sec := range s {
		fmt.Printf("### %s ###\n\n", sec.Title)
		if sec.Info != "" && ctx.Bool("md") {
			fmt.Printf("%s\n\n", sec.Info)
		}

		secRows := make([][]string, 0, len(rows))

		for {
			row := rows[j]

			if len(row) < 1 {
				continue
			}

			if i < len(s)-1 {
				if s[i+1].Start == row[0] {
					break
				}
			}

			secRows = append(secRows, row)
			j++

			if j >= len(rows) {
				break
			}
		}

		result, err := report.RenderFormat(secRows, cols, report.CliFormat(ctx))

		if err != nil {
			return err
		}

		fmt.Println(result)

		if j >= len(rows) {
			break
		}
	}

	return nil
}
