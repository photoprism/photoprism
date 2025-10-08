package commands

import (
	"encoding/json"
	"strings"
	"testing"

	catalogpkg "github.com/photoprism/photoprism/internal/commands/catalog"
)

func TestShowCommands_JSON_Flat(t *testing.T) {
	// Build JSON without capturing stdout to avoid pipe blocking on large outputs
	ctx := NewTestContext([]string{"commands"})
	global := catalogpkg.FlagsToCatalog(ctx.App.Flags, false)
	var flat []catalogpkg.Command
	for _, c := range ctx.App.Commands {
		if c.Hidden {
			continue
		}
		flat = append(flat, catalogpkg.BuildFlat(c, 1, "photoprism", false, global)...)
	}
	out := struct {
		App         catalogpkg.App       `json:"app"`
		GeneratedAt string               `json:"generated_at"`
		GlobalFlags []catalogpkg.Flag    `json:"global_flags"`
		Commands    []catalogpkg.Command `json:"commands"`
	}{
		App:         catalogpkg.App{Name: "PhotoPrism", Edition: "ce", Version: "test"},
		GeneratedAt: "",
		GlobalFlags: global,
		Commands:    flat,
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	// Basic structural checks via unmarshal into a light struct
	var v struct {
		Commands []struct {
			Name, FullName string
			Depth          int
		} `json:"commands"`
		GlobalFlags []map[string]interface{} `json:"global_flags"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(v.Commands) == 0 {
		t.Fatalf("expected at least one command")
	}
	// Expect at least one top-level auth and at least one subcommand overall
	var haveAuth, haveAnyChild bool
	for _, c := range v.Commands {
		if c.Name == "auth" && c.Depth == 1 {
			haveAuth = true
		}
		if c.Depth >= 2 {
			haveAnyChild = true
		}
	}
	if !haveAuth {
		t.Fatalf("expected to find 'auth' top-level command in list")
	}
	if !haveAnyChild {
		t.Fatalf("expected to find at least one subcommand (depth >= 2)")
	}
	if len(v.GlobalFlags) == 0 {
		t.Fatalf("expected non-empty global_flags")
	}
}

func TestShowCommands_JSON_Nested(t *testing.T) {
	ctx := NewTestContext([]string{"commands"})
	global := catalogpkg.FlagsToCatalog(ctx.App.Flags, false)
	var tree []catalogpkg.Node
	for _, c := range ctx.App.Commands {
		if c.Hidden {
			continue
		}
		tree = append(tree, catalogpkg.BuildNode(c, 1, "photoprism", false, global))
	}
	b, err := json.Marshal(struct {
		Commands []catalogpkg.Node `json:"commands"`
	}{Commands: tree})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var v struct {
		Commands []struct {
			Name        string `json:"name"`
			Depth       int    `json:"depth"`
			Subcommands []struct {
				Name string `json:"name"`
			} `json:"subcommands"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(v.Commands) == 0 {
		t.Fatalf("expected top-level commands")
	}
	var hasAuthWithChild bool
	for _, c := range v.Commands {
		if c.Name == "auth" && c.Depth == 1 && len(c.Subcommands) > 0 {
			hasAuthWithChild = true
			break
		}
	}
	if !hasAuthWithChild {
		t.Fatalf("expected auth with at least one subcommand")
	}
}

func TestShowCommands_Markdown_Default(t *testing.T) {
	out, err := RunWithTestContext(ShowCommandsCommand, []string{"commands"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Expect Markdown headings for commands
	if !strings.Contains(out, "## photoprism auth") {
		t.Fatalf("expected '## photoprism auth' heading in output\n%s", out[:min(400, len(out))])
	}
	if !strings.Contains(out, "### photoprism auth ") { // subcommand headings begin with parent
		t.Fatalf("expected '### photoprism auth <sub>' heading in output\n%s", out[:min(600, len(out))])
	}
}
