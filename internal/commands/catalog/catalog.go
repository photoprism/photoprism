/*
Package catalog provides DTOs, builders, and a templated renderer to export the PhotoPrism CLI command tree (and flags) as Markdown or JSON for documentation purposes.

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package catalog

import (
	"bytes"
	"sort"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"
)

// Flag describes a CLI flag.
type Flag struct {
	Name     string   `json:"name"`
	Aliases  []string `json:"aliases,omitempty"`
	Type     string   `json:"type"`
	Required bool     `json:"required,omitempty"`
	Default  string   `json:"default,omitempty"`
	Env      []string `json:"env,omitempty"`
	Category string   `json:"category,omitempty"`
	Usage    string   `json:"usage,omitempty"`
	Hidden   bool     `json:"hidden"`
}

// Command describes a CLI command (flat form).
type Command struct {
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	Parent      string   `json:"parent,omitempty"`
	Depth       int      `json:"depth"`
	Usage       string   `json:"usage,omitempty"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	ArgsUsage   string   `json:"args_usage,omitempty"`
	Hidden      bool     `json:"hidden"`
	Flags       []Flag   `json:"flags,omitempty"`
}

// Node is a nested representation of commands.
type Node struct {
	Command
	Subcommands []Node `json:"subcommands,omitempty"`
}

// App carries app metadata for top-level JSON/MD.
type App struct {
	Name    string `json:"name"`
	Edition string `json:"edition"`
	Version string `json:"version"`
	Build   string `json:"build,omitempty"`
}

// MarkdownData is the data model used by the Markdown template.
type MarkdownData struct {
	App         App
	GeneratedAt string
	BaseHeading int
	Short       bool
	All         bool
	Commands    []Command
}

// BuildFlat returns a depth-first flat list of commands starting at c.
func BuildFlat(c *cli.Command, depth int, parentFull string, includeHidden bool, global []Flag) []Command {
	// Omit nested 'help' subcommands; keep only top-level 'photoprism help'
	if skipHelp(c, parentFull) {
		return nil
	}
	var out []Command
	info := CommandInfo(c, depth, parentFull, includeHidden, global)
	out = append(out, info)
	for _, sub := range c.Subcommands {
		if sub == nil || (sub.Hidden && !includeHidden) || skipHelp(sub, info.FullName) {
			continue
		}
		out = append(out, BuildFlat(sub, depth+1, info.FullName, includeHidden, global)...)
	}
	return out
}

// BuildNode returns a nested representation of c and its subcommands.
func BuildNode(c *cli.Command, depth int, parentFull string, includeHidden bool, global []Flag) Node {
	info := CommandInfo(c, depth, parentFull, includeHidden, global)
	node := Node{Command: info}
	for _, sub := range c.Subcommands {
		if sub == nil || (sub.Hidden && !includeHidden) || skipHelp(sub, info.FullName) {
			continue
		}
		node.Subcommands = append(node.Subcommands, BuildNode(sub, depth+1, info.FullName, includeHidden, global))
	}
	return node
}

// skipHelp returns true for nested 'help' commands so they are omitted from output.
// Top-level 'photoprism help' remains included.
func skipHelp(c *cli.Command, parentFull string) bool {
	if c == nil {
		return false
	}
	if strings.EqualFold(c.Name, "help") {
		// Keep only at the root where parent is 'photoprism'
		return parentFull != "photoprism"
	}
	return false
}

// CommandInfo converts a cli.Command to a Command DTO.
func CommandInfo(c *cli.Command, depth int, parentFull string, includeHidden bool, global []Flag) Command {
	pathName := c.Name
	fullName := strings.TrimSpace(parentFull + " " + pathName)
	parent := parentFull

	cmd := Command{
		Name:        pathName,
		FullName:    fullName,
		Parent:      parent,
		Depth:       depth,
		Usage:       c.Usage,
		Description: strings.TrimSpace(c.Description),
		Category:    c.Category,
		Aliases:     c.Aliases,
		ArgsUsage:   c.ArgsUsage,
		Hidden:      c.Hidden,
	}

	// Build set of canonical global flag names to exclude from per-command flags
	globalSet := map[string]struct{}{}
	for _, gf := range global {
		globalSet[strings.TrimLeft(gf.Name, "-")] = struct{}{}
	}

	// Convert flags and optionally filter hidden/global
	flags := FlagsToCatalog(c.Flags, includeHidden)
	keep := make([]Flag, 0, len(flags))
	for _, f := range flags {
		name := strings.TrimLeft(f.Name, "-")
		if _, isGlobal := globalSet[name]; isGlobal {
			continue
		}
		if !includeHidden && f.Hidden {
			continue
		}
		keep = append(keep, f)
	}
	sort.Slice(keep, func(i, j int) bool { return keep[i].Name < keep[j].Name })
	cmd.Flags = keep
	return cmd
}

// FlagsToCatalog converts cli flags to Flag DTOs, filtering hidden if needed.
func FlagsToCatalog(flags []cli.Flag, includeHidden bool) []Flag {
	out := make([]Flag, 0, len(flags))
	for _, f := range flags {
		if f == nil {
			continue
		}
		cf := DescribeFlag(f)
		if !includeHidden && cf.Hidden {
			continue
		}
		out = append(out, cf)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// DescribeFlag inspects a cli.Flag and returns a Flag with metadata.
func DescribeFlag(f cli.Flag) Flag {
	// Names and aliases
	names := f.Names()
	primary := ""
	aliases := make([]string, 0, len(names))
	for i, n := range names {
		if i == 0 {
			primary = n
		}
		if primary == "" || (len(n) > 1 && primary != n) {
			primary = n
		}
	}
	for _, n := range names {
		if n == primary {
			continue
		}
		if len(n) == 1 {
			aliases = append(aliases, "-"+n)
		} else {
			aliases = append(aliases, "--"+n)
		}
	}

	type hasUsage interface{ GetUsage() string }
	type hasCategory interface{ GetCategory() string }
	type hasEnv interface{ GetEnvVars() []string }
	type hasDefault interface{ GetDefaultText() string }
	type hasRequired interface{ IsRequired() bool }
	type hasVisible interface{ IsVisible() bool }

	usage, category, def := "", "", ""
	env := []string{}
	required, hidden := false, false
	if hf, ok := f.(hasUsage); ok {
		usage = hf.GetUsage()
	}
	if hf, ok := f.(hasCategory); ok {
		category = hf.GetCategory()
	}
	if hf, ok := f.(hasEnv); ok {
		env = append(env, hf.GetEnvVars()...)
	}
	if hf, ok := f.(hasDefault); ok {
		def = hf.GetDefaultText()
	}
	if hf, ok := f.(hasRequired); ok {
		required = hf.IsRequired()
	}
	if hv, ok := f.(hasVisible); ok {
		hidden = !hv.IsVisible()
	}

	t := flagTypeString(f)
	name := primary
	if len(name) == 1 {
		name = "-" + name
	} else {
		name = "--" + name
	}

	return Flag{
		Name: name, Aliases: aliases, Type: t, Required: required, Default: def,
		Env: env, Category: category, Usage: usage, Hidden: hidden,
	}
}

func flagTypeString(f cli.Flag) string {
	switch f.(type) {
	case *cli.BoolFlag:
		return "bool"
	case *cli.StringFlag:
		return "string"
	case *cli.IntFlag:
		return "int"
	case *cli.Int64Flag:
		return "int64"
	case *cli.UintFlag:
		return "uint"
	case *cli.Uint64Flag:
		return "uint64"
	case *cli.Float64Flag:
		return "float64"
	case *cli.DurationFlag:
		return "duration"
	case *cli.TimestampFlag:
		return "timestamp"
	case *cli.PathFlag:
		return "path"
	case *cli.StringSliceFlag:
		return "stringSlice"
	case *cli.IntSliceFlag:
		return "intSlice"
	case *cli.Int64SliceFlag:
		return "int64Slice"
	case *cli.Float64SliceFlag:
		return "float64Slice"
	case *cli.GenericFlag:
		return "generic"
	default:
		return "unknown"
	}
}

// Default Markdown template (adjustable in source via rebuild).
var commandsMDTemplate = `# {{ .App.Name }} CLI Commands ({{ .App.Edition }}) — {{ .App.Version }}

_Generated: {{ .GeneratedAt }}_

{{- $base := .BaseHeading -}}
{{- range .Commands }}

{{ heading (add $base (dec .Depth)) }} {{ .FullName }}

**Usage:** {{ .Usage }}
{{- if .Description }}

**Description:** {{ .Description }}
{{- end }}
{{- if .Aliases }}

**Aliases:** {{ join .Aliases ", " }}
{{- end }}
{{- if .ArgsUsage }}

**Args:** ` + "`" + `{{ .ArgsUsage }}` + "`" + `
{{- end }}
{{- if and (not $.Short) .Flags }}

| Flag | Aliases | Type | Default | Env | Required |{{ if $.All }} Hidden |{{ end }} Usage |
|:-----|:--------|:-----|:--------|:----|:---------|{{ if $.All }}:------:|{{ end }}:------|
{{- range .Flags }}
| ` + "`" + `{{ .Name }}` + "`" + ` | {{ join .Aliases ", " }} | {{ .Type }} | {{ .Default }} | {{ join .Env ", " }} | {{ .Required }} |{{ if $.All }} {{ .Hidden }} |{{ end }} {{ .Usage }} |
{{- end }}
{{- end }}

{{- end }}`

// RenderMarkdown renders the catalog to Markdown using the embedded template.
func RenderMarkdown(data MarkdownData) (string, error) {
	tmpl, err := template.New("commands").Funcs(template.FuncMap{
		"heading": func(n int) string {
			if n < 1 {
				n = 1
			} else if n > 6 {
				n = 6
			}
			return strings.Repeat("#", n)
		},
		"join": strings.Join,
		"add":  func(a, b int) int { return a + b },
		"dec": func(a int) int {
			if a > 0 {
				return a - 1
			}
			return 0
		},
	}).Parse(commandsMDTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
