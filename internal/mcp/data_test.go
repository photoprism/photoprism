package mcp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
)

// TestBuildConfigOptions asserts that every non-hidden config flag
// surfaced by buildConfigOptions carries a section title, environment
// variable, and CLI flag.
func TestBuildConfigOptions(t *testing.T) {
	items := buildConfigOptions()
	require.NotEmpty(t, items, "buildConfigOptions must return items")

	for i, item := range items {
		require.NotEmpty(t, item.Environment, "item %d must have an environment variable", i)
		require.NotEmpty(t, item.CLIFlag, "item %d must have a CLI flag", i)
		require.NotEmpty(t, item.Section, "item %d (%s) must have a section", i, item.Environment)
	}
}

// TestBuildSearchFilters asserts that every search filter row surfaced
// by buildSearchFilters carries a filter name and type.
func TestBuildSearchFilters(t *testing.T) {
	items := buildSearchFilters()
	require.NotEmpty(t, items, "buildSearchFilters must return items")

	for i, item := range items {
		require.NotEmpty(t, item.Filter, "item %d must have a filter name", i)
		require.NotEmpty(t, item.Type, "item %d must have a type", i)
	}
}

// TestNormalizeEdition covers trimming, case-folding, and the empty-input
// fallback to "unknown" in normalizeEdition.
func TestNormalizeEdition(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "unknown"},
		{" ", "unknown"},
		{"CE", "ce"},
		{"pro", "pro"},
		{"  Plus  ", "plus"},
		{"Portal", "portal"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			require.Equal(t, tc.expected, normalizeEdition(tc.input))
		})
	}
}

// TestBuildConfigOptionsReturnsDocumentedDefaults pins the dataset
// contract that backs photoprism://config-options and list_config_keys:
// the rows reflect the documented defaults compiled into the binary,
// never the runtime values an operator supplies through environment
// variables. Each marker below maps to one Secret-annotated flag whose
// Default field must remain empty regardless of the env override.
func TestBuildConfigOptionsReturnsDocumentedDefaults(t *testing.T) {
	const (
		adminMarker    = "TestBuildConfigOptionsReturnsDocumentedDefaults-admin"
		dbMarker       = "TestBuildConfigOptionsReturnsDocumentedDefaults-db"
		oidcMarker     = "TestBuildConfigOptionsReturnsDocumentedDefaults-oidc"
		joinMarker     = "TestBuildConfigOptionsReturnsDocumentedDefaults-join"
		downloadMarker = "TestBuildConfigOptionsReturnsDocumentedDefaults-download"
		previewMarker  = "TestBuildConfigOptionsReturnsDocumentedDefaults-preview"
		visionMarker   = "TestBuildConfigOptionsReturnsDocumentedDefaults-vision"
	)

	env := map[string]string{
		"PHOTOPRISM_ADMIN_PASSWORD":    adminMarker,
		"PHOTOPRISM_DATABASE_PASSWORD": dbMarker,
		"PHOTOPRISM_OIDC_SECRET":       oidcMarker,
		"PHOTOPRISM_JOIN_TOKEN":        joinMarker,
		"PHOTOPRISM_DOWNLOAD_TOKEN":    downloadMarker,
		"PHOTOPRISM_PREVIEW_TOKEN":     previewMarker,
		"PHOTOPRISM_VISION_KEY":        visionMarker,
	}

	for k, v := range env {
		t.Setenv(k, v)
	}

	app := cli.NewApp()
	app.Flags = config.Flags.Cli()
	app.Action = func(*cli.Context) error { return nil }
	require.NoError(t, app.Run([]string{"photoprism"}))

	items := buildConfigOptions()
	require.NotEmpty(t, items)

	markers := []string{adminMarker, dbMarker, oidcMarker, joinMarker, downloadMarker, previewMarker, visionMarker}

	for _, item := range items {
		fields := []struct {
			name  string
			value string
		}{
			{"Environment", item.Environment},
			{"CLIFlag", item.CLIFlag},
			{"Default", item.Default},
			{"Description", item.Description},
		}

		for _, f := range fields {
			for _, marker := range markers {
				require.NotContains(t, f.value, marker,
					"item %s field %s echoed runtime value %q", item.Environment, f.name, marker)
			}
		}
	}

	expectEmpty := map[string]struct{}{
		"PHOTOPRISM_ADMIN_PASSWORD":    {},
		"PHOTOPRISM_DATABASE_PASSWORD": {},
		"PHOTOPRISM_OIDC_SECRET":       {},
		"PHOTOPRISM_JOIN_TOKEN":        {},
		"PHOTOPRISM_DOWNLOAD_TOKEN":    {},
		"PHOTOPRISM_PREVIEW_TOKEN":     {},
		"PHOTOPRISM_VISION_KEY":        {},
	}

	for _, item := range items {
		for envVar := range expectEmpty {
			if !strings.Contains(item.Environment, envVar) {
				continue
			}

			require.Empty(t, item.Default,
				"flag %s must surface an empty Default; got %q", item.Environment, item.Default)
		}
	}
}

// TestEditionSupportFor exercises the tag-to-edition mapping that drives
// the edition_support hint returned by list_config_keys, including the
// "unknown" short-circuit and the priority order (portal > pro > plus >
// essentials > all).
func TestEditionSupportFor(t *testing.T) {
	tests := []struct {
		name           string
		tags           []string
		currentEdition string
		expected       string
	}{
		{"NoTags", nil, "ce", "all"},
		{"EmptyTags", []string{}, "ce", "all"},
		{"Portal", []string{"portal"}, "pro", "portal"},
		{"Pro", []string{"pro"}, "pro", "pro"},
		{"Plus", []string{"plus"}, "pro", "plus"},
		{"Essentials", []string{"essentials"}, "pro", "essentials"},
		{"UnknownEdition", []string{"pro"}, "unknown", "unknown"},
		{"UnrelatedTag", []string{"sponsor"}, "ce", "all"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			option := ConfigOption{Tags: tc.tags}
			require.Equal(t, tc.expected, editionSupportFor(option, tc.currentEdition))
		})
	}
}
