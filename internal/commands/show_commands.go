package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/commands/catalog"
	"github.com/photoprism/photoprism/internal/config"
)

// ShowCommandsCommand configures the command name, flags, and action.
var ShowCommandsCommand = &cli.Command{
	Name:  "commands",
	Usage: "Displays a structured catalog of CLI commands",
	Flags: []cli.Flag{
		JsonFlag(),
		&cli.BoolFlag{Name: "all", Usage: "include hidden commands and flags"},
		&cli.BoolFlag{Name: "short", Usage: "omit flags in Markdown output"},
		&cli.IntFlag{Name: "base-heading", Value: 2, Usage: "base Markdown heading level"},
		&cli.BoolFlag{Name: "nested", Usage: "emit nested JSON structure instead of a flat array"},
	},
	Action: showCommandsAction,
}

type showCommandsOut struct {
	App         catalog.App     `json:"app"`
	GeneratedAt string          `json:"generated_at"`
	GlobalFlags []catalog.Flag  `json:"global_flags,omitempty"`
	Commands    json.RawMessage `json:"commands"`
}

// showCommandsAction displays a structured catalog of CLI commands.
func showCommandsAction(ctx *cli.Context) error {
	// Prefer fast app metadata from the running app; avoid heavy config init in tests
	includeHidden := ctx.Bool("all")
	wantJSON := ctx.Bool("json")
	nested := ctx.Bool("nested")
	baseHeading := ctx.Int("base-heading")
	if baseHeading < 1 {
		baseHeading = 1
	}

	// Collect the app metadata to be included in the output.
	app := catalog.App{}

	if ctx != nil && ctx.App != nil && ctx.App.Metadata != nil {
		if n, ok := ctx.App.Metadata["Name"].(string); ok {
			app.Name = n
		}
		if e, ok := ctx.App.Metadata["Edition"].(string); ok {
			app.Edition = e
		}
		if v, ok := ctx.App.Metadata["Version"].(string); ok {
			app.Version = v
			app.Build = v
		}
	}

	if app.Name == "" || app.Version == "" {
		conf := config.NewConfig(ctx)
		app.Name = conf.Name()
		app.Edition = conf.Edition()
		app.Version = conf.Version()
		app.Build = conf.Version()
	}

	// Collect global flags from the running app.
	var globalFlags []catalog.Flag
	if ctx != nil && ctx.App != nil {
		globalFlags = catalog.FlagsToCatalog(ctx.App.Flags, includeHidden)
	} else {
		globalFlags = catalog.FlagsToCatalog(config.Flags.Cli(), includeHidden)
	}

	// Traverse commands registry using runtime app commands to avoid init cycles.
	var flat []catalog.Command
	var tree []catalog.Node

	var roots []*cli.Command
	if ctx != nil && ctx.App != nil {
		roots = ctx.App.Commands
	}
	for _, c := range roots {
		if c == nil {
			continue
		}
		if c.Hidden && !includeHidden {
			continue
		}
		if nested {
			node := catalog.BuildNode(c, 1, "photoprism", includeHidden, globalFlags)
			tree = append(tree, node)
		} else {
			flat = append(flat, catalog.BuildFlat(c, 1, "photoprism", includeHidden, globalFlags)...)
		}
	}

	// Render JSON output using json.Marshal().
	if wantJSON {
		var cmds json.RawMessage
		var err error
		if nested {
			cmds, err = json.Marshal(tree)
		} else {
			cmds, err = json.Marshal(flat)
		}
		if err != nil {
			return err
		}
		out := showCommandsOut{
			App:         app,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			GlobalFlags: globalFlags,
			Commands:    cmds,
		}
		b, err := json.Marshal(out)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	// Render Markdown using embedded template.
	data := catalog.MarkdownData{
		App:         app,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		BaseHeading: baseHeading,
		Short:       ctx.Bool("short"),
		All:         includeHidden,
		Commands:    flat,
	}

	md, err := catalog.RenderMarkdown(data)
	if err != nil {
		return err
	}
	fmt.Println(md)
	return nil
}
