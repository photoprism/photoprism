package commands

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShowThumbSizes_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowThumbSizesCommand, []string{"thumb-sizes", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v []map[string]string

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(v) == 0 {
		t.Fatalf("expected at least one item")
	}
	// Expected keys for thumb-sizes detailed report
	for _, k := range []string{"name", "width", "height", "aspect_ratio", "available", "usage"} {
		if _, ok := v[0][k]; !ok {
			t.Fatalf("expected key '%s' in first item", k)
		}
	}
}

func TestShowSources_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowSourcesCommand, []string{"sources", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v []map[string]string

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(v) == 0 {
		t.Fatalf("expected at least one item")
	}
	if _, ok := v[0]["source"]; !ok {
		t.Fatalf("expected key 'source' in first item")
	}
	if _, ok := v[0]["priority"]; !ok {
		t.Fatalf("expected key 'priority' in first item")
	}
}

func TestShowMetadata_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowMetadataCommand, []string{"metadata", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v struct {
		Items []map[string]string `json:"items"`
		Docs  []map[string]string `json:"docs"`
	}

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}

	if len(v.Items) == 0 {
		t.Fatalf("expected items")
	}
}

func TestShowConfig_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigCommand, []string{"config", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v struct {
		Sections []struct {
			Title string              `json:"title"`
			Items []map[string]string `json:"items"`
		} `json:"sections"`
	}

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}

	if len(v.Sections) == 0 || len(v.Sections[0].Items) == 0 {
		t.Fatalf("expected sections with items")
	}
}

func TestShowConfigOptions_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigOptionsCommand, []string{"config-options", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type options = []map[string]string
	var v = options{}

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}

	if len(v) == 0 {
		t.Fatalf("expected sections with items")
	}
}

func TestShowConfigYaml_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigYamlCommand, []string{"config-yaml", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type options = []map[string]string
	var v = options{}

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}

	if len(v) == 0 {
		t.Fatalf("expected sections with items")
	}
}

func TestShowFormatConflict_Error(t *testing.T) {
	out, err := RunWithTestContext(ShowSourcesCommand, []string{"sources", "--json", "--csv"})

	if err == nil {
		t.Fatalf("expected error for conflicting flags, got nil; output=%s", out)
	}

	// Expect an ExitCoder with code 2
	if ec, ok := err.(interface{ ExitCode() int }); ok {
		if ec.ExitCode() != 2 {
			t.Fatalf("expected exit code 2, got %d", ec.ExitCode())
		}
	} else {
		t.Fatalf("expected exit coder error, got %T: %v", err, err)
	}
}

func TestShowConfigOptions_MarkdownSections(t *testing.T) {
	out, err := RunWithTestContext(ShowConfigOptionsCommand, []string{"config-options", "--md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "### Authentication") {
		t.Fatalf("expected Markdown section heading '### Authentication' in output\n%s", out[:min(400, len(out))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestShowFileFormats_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowFileFormatsCommand, []string{"file-formats", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v []map[string]string

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(v) == 0 {
		t.Fatalf("expected at least one item")
	}
	// Keys depend on report settings in command: should include format, description, type, extensions
	if _, ok := v[0]["format"]; !ok {
		t.Fatalf("expected key 'format' in first item")
	}
	if _, ok := v[0]["type"]; !ok {
		t.Fatalf("expected key 'type' in first item")
	}
	if _, ok := v[0]["extensions"]; !ok {
		t.Fatalf("expected key 'extensions' in first item")
	}
}

func TestShowVideoSizes_JSON(t *testing.T) {
	out, err := RunWithTestContext(ShowVideoSizesCommand, []string{"video-sizes", "--json"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var v []map[string]string

	if err = json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(v) == 0 {
		t.Fatalf("expected at least one item")
	}
	if _, ok := v[0]["size"]; !ok {
		t.Fatalf("expected key 'size' in first item")
	}
	if _, ok := v[0]["usage"]; !ok {
		t.Fatalf("expected key 'usage' in first item")
	}
}
