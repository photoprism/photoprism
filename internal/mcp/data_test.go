package mcp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildConfigOptions(t *testing.T) {
	items := buildConfigOptions()
	require.NotEmpty(t, items, "buildConfigOptions must return items")

	for i, item := range items {
		require.NotEmpty(t, item.Environment, "item %d must have an environment variable", i)
		require.NotEmpty(t, item.CLIFlag, "item %d must have a CLI flag", i)
		require.NotEmpty(t, item.Section, "item %d (%s) must have a section", i, item.Environment)
	}
}

func TestBuildSearchFilters(t *testing.T) {
	items := buildSearchFilters()
	require.NotEmpty(t, items, "buildSearchFilters must return items")

	for i, item := range items {
		require.NotEmpty(t, item.Filter, "item %d must have a filter name", i)
		require.NotEmpty(t, item.Type, "item %d must have a type", i)
	}
}

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
