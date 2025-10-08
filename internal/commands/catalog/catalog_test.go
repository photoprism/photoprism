package catalog

import (
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestFlagsToCatalog_Visibility(t *testing.T) {
	flags := []cli.Flag{
		&cli.StringFlag{Name: "config-path", Aliases: []string{"c"}, Usage: "config path", EnvVars: []string{"PHOTOPRISM_CONFIG_PATH"}},
		&cli.BoolFlag{Name: "trace", Usage: "enable trace", Hidden: true},
		&cli.IntFlag{Name: "count", Usage: "max results", Value: 5, Required: true},
	}

	vis := FlagsToCatalog(flags, false)
	if len(vis) != 2 {
		t.Fatalf("expected 2 visible flags, got %d", len(vis))
	}
	all := FlagsToCatalog(flags, true)
	if len(all) != 3 {
		t.Fatalf("expected 3 flags with --all, got %d", len(all))
	}
	// Check hidden is marked correctly when included
	var hiddenOK bool
	for _, f := range all {
		if f.Name == "--trace" && f.Hidden {
			hiddenOK = true
		}
	}
	if !hiddenOK {
		t.Fatalf("expected hidden flag '--trace' with hidden=true")
	}
}

func TestCommandInfo_GlobalFlagElimination(t *testing.T) {
	// Define a command with a global-like flag and a local one
	cmd := &cli.Command{
		Name: "auth",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json"}, // should be filtered out as global
			&cli.BoolFlag{Name: "force"},
		},
	}
	globals := FlagsToCatalog([]cli.Flag{&cli.BoolFlag{Name: "json"}}, false)
	info := CommandInfo(cmd, 1, "photoprism", false, globals)
	if len(info.Flags) != 1 || info.Flags[0].Name != "--force" {
		t.Fatalf("expected only '--force' flag, got %+v", info.Flags)
	}
}

func TestBuildFlatAndNode(t *testing.T) {
	add := &cli.Command{Name: "add"}
	help := &cli.Command{Name: "help"}
	rmHidden := &cli.Command{Name: "rm", Hidden: true}
	auth := &cli.Command{Name: "auth", Subcommands: []*cli.Command{add, rmHidden, help}}

	globals := FlagsToCatalog(nil, false)

	// Flat without hidden
	flat := BuildFlat(auth, 1, "photoprism", false, globals)
	if len(flat) != 2 { // auth + add (help omitted)
		t.Fatalf("expected 2 commands (auth, add), got %d", len(flat))
	}
	if flat[0].FullName != "photoprism auth" || flat[0].Depth != 1 {
		t.Fatalf("unexpected root entry: %+v", flat[0])
	}
	if flat[1].FullName != "photoprism auth add" || flat[1].Depth != 2 {
		t.Fatalf("unexpected child entry: %+v", flat[1])
	}

	// Nested with hidden
	node := BuildNode(auth, 1, "photoprism", true, globals)
	if len(node.Subcommands) != 2 { // add + rm (help omitted)
		t.Fatalf("expected 2 subcommands when including hidden, got %d", len(node.Subcommands))
	}
}

func TestRenderMarkdown_Headings(t *testing.T) {
	data := MarkdownData{
		App:         App{Name: "PhotoPrism", Edition: "ce", Version: "test"},
		GeneratedAt: "",
		BaseHeading: 2,
		Short:       true, // hide flags table to focus on headings
		All:         false,
		Commands: []Command{
			{Name: "auth", FullName: "photoprism auth", Depth: 1},
			{Name: "auth add", FullName: "photoprism auth add", Depth: 2},
		},
	}
	out, err := RenderMarkdown(data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(out, "## photoprism auth") {
		t.Fatalf("expected '## photoprism auth' heading, got:\n%s", out)
	}
	if !strings.Contains(out, "### photoprism auth add") {
		t.Fatalf("expected '### photoprism auth add' heading, got:\n%s", out)
	}
}

func TestRenderMarkdown_HiddenColumn(t *testing.T) {
	cmd := Command{
		Name:     "auth",
		FullName: "photoprism auth",
		Depth:    1,
		Flags: []Flag{
			{Name: "--visible", Type: "bool", Hidden: false, Usage: "visible flag"},
			{Name: "--secret", Type: "bool", Hidden: true, Usage: "hidden flag"},
		},
	}
	base := MarkdownData{
		App:         App{Name: "PhotoPrism", Edition: "ce", Version: "test"},
		GeneratedAt: "",
		BaseHeading: 2,
		Short:       false,
		Commands:    []Command{cmd},
	}

	// Default: no Hidden column
	base.All = false
	out, err := RenderMarkdown(base)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if strings.Contains(out, " Hidden ") {
		t.Fatalf("did not expect 'Hidden' column without --all:\n%s", out)
	}

	// With --all: Hidden column and boolean value present
	base.All = true
	out, err = RenderMarkdown(base)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(out, " Hidden ") {
		t.Fatalf("expected 'Hidden' column with --all:\n%s", out)
	}
	if !strings.Contains(out, "hidden flag") || !strings.Contains(out, " true ") {
		t.Fatalf("expected hidden flag row to include 'true':\n%s", out)
	}
}

func TestRenderMarkdown_ShortOmitsFlags(t *testing.T) {
	cmd := Command{
		Name:     "auth",
		FullName: "photoprism auth",
		Depth:    1,
		Flags: []Flag{
			{Name: "--json", Type: "bool", Hidden: false, Usage: "json output"},
		},
	}
	data := MarkdownData{
		App:         App{Name: "PhotoPrism", Edition: "ce", Version: "test"},
		GeneratedAt: "",
		BaseHeading: 2,
		Short:       true,
		All:         false,
		Commands:    []Command{cmd},
	}
	out, err := RenderMarkdown(data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if strings.Contains(out, "| Flag | Aliases | Type |") {
		t.Fatalf("did not expect flags table when Short=true:\n%s", out)
	}
}
