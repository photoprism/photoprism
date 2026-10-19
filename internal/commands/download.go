package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
)

var downloadExamples = `
Usage examples:

photoprism dl --cookies cookies.txt \
 --header 'Authorization: Bearer <token>' \
 --method file --remux auto -- \
 https://example.com/a.mp4 https://example.com/b.jpg

photoprism dl -a 'Authorization: Bearer <token>' \
			 -a 'Accept: application/json' -- URL`

// DownloadCommand configures the command name, flags, and action.
var DownloadCommand = &cli.Command{
	Name:        "download",
	Aliases:     []string{"dl"},
	Usage:       "Imports media from one or more URLs",
	Description: "Download and import media from one or more URLs.\n" + downloadExamples,
	ArgsUsage:   "[url]...",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "dest",
			Aliases: []string{"d"},
			Usage:   "relative originals `PATH` in which new files should be imported",
		},
		&cli.StringFlag{
			Name:    "impersonate",
			Aliases: []string{"i"},
			Usage:   "impersonate browser `IDENTITY` (e.g. chrome, edge or safari; 'none' to disable)",
			Value:   "firefox",
		},
		&cli.StringFlag{
			Name:    "method",
			Aliases: []string{"m"},
			Value:   "pipe",
			Usage:   "download `METHOD` when using external commands: pipe (stdio stream) or file (temporary files)",
		},
		&cli.StringFlag{
			Name:    "remux",
			Aliases: []string{"r"},
			Value:   "auto",
			Usage:   "remux `POLICY` for videos when using --method file: auto (skip if MP4), always, or skip",
		},
		&cli.StringFlag{
			Name:    "sort",
			Aliases: []string{"s"},
			Usage:   "custom `FORMAT` sort expression, e.g. 'quality,res,fps,codec:avc:m4a,size,br,asr,proto,ext,hasaud,source,id'",
		},
		&cli.StringFlag{
			Name:    "cookies",
			Aliases: []string{"c"},
			Usage:   "use Netscape-format cookies.txt `FILE` for HTTP authentication",
		},
		&cli.StringSliceFlag{
			Name:    "header",
			Aliases: []string{"a"},
			Usage:   "add HTTP request `HEADER` in the form 'Name: Value' (repeatable)",
		},
	},
	Action: downloadAction,
}

// downloadAction collects the URLs and flags of the download command and passes them to runDownload.
func downloadAction(ctx *cli.Context) error {
	conf, confErr := InitConfig(ctx)

	if confErr != nil {
		return confErr
	}

	if conf.ReadOnly() {
		return config.ErrReadOnly
	}

	conf.InitDb()
	defer conf.Shutdown()

	// Collect URLs: args or STDIN when no args
	var inputURLs []string
	if ctx.Args().Len() > 0 {
		inputURLs = append(inputURLs, ctx.Args().Slice()...)
	} else {
		// If STDIN is a pipe, read URLs line by line (Phase 1: args take precedence; no --stdin merge)
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 1024*1024)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				inputURLs = append(inputURLs, line)
			}
			if err := scanner.Err(); err != nil {
				return err
			}
		}
	}

	if len(inputURLs) == 0 {
		return fmt.Errorf("no download URLs provided")
	}

	flagMethod := ""

	if ctx.IsSet("method") {
		flagMethod = ctx.String("method")
	}

	return runDownload(conf, DownloadOpts{
		Dest:        ctx.String("dest"),
		Cookies:     strings.TrimSpace(ctx.String("cookies")),
		AddHeaders:  ctx.StringSlice("header"),
		Impersonate: ctx.String("impersonate"),
		Method:      flagMethod,
		FileRemux:   ctx.String("remux"),
		FormatSort:  ctx.String("sort"),
	}, inputURLs)
}
