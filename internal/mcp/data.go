package mcp

import (
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
)

const (
	configOptionsURI = "photoprism://config-options"
	searchFiltersURI = "photoprism://search-filters"
)

// Dataset caches the static MCP prototype data returned by resources and tools.
type Dataset struct {
	CurrentEdition string
	ConfigOptions  []ConfigOption
	SearchFilters  []SearchFilter
}

// ConfigOption represents a single config option report row with section metadata.
type ConfigOption struct {
	Section     string   `json:"section"`
	Environment string   `json:"environment"`
	CLIFlag     string   `json:"cli_flag"`
	Default     string   `json:"default"`
	Description string   `json:"description"`
	Tags        []string `json:"-"`
}

// SearchFilter represents a single search filter report row.
type SearchFilter struct {
	Filter   string `json:"filter"`
	Type     string `json:"type"`
	Examples string `json:"examples"`
	Notes    string `json:"notes"`
}

// ConfigOptionsResource contains the JSON payload for the config-options resource.
type ConfigOptionsResource struct {
	Edition string         `json:"edition"`
	Items   []ConfigOption `json:"items"`
}

// SearchFiltersResource contains the JSON payload for the search-filters resource.
type SearchFiltersResource struct {
	Edition string         `json:"edition"`
	Items   []SearchFilter `json:"items"`
}

// NewDataset builds the static MCP prototype dataset for the current build.
func NewDataset(currentEdition string) *Dataset {
	return &Dataset{
		CurrentEdition: normalizeEdition(currentEdition),
		ConfigOptions:  buildConfigOptions(),
		SearchFilters:  buildSearchFilters(),
	}
}

// buildConfigOptions returns config options with section titles attached.
func buildConfigOptions() []ConfigOption {
	// Build a map from section start env vars to section titles.
	sectionStarts := make(map[string]string)

	for _, section := range config.OptionsReportSections {
		for _, key := range strings.Split(section.Start, ", ") {
			sectionStarts[key] = section.Title
		}
	}

	items := make([]ConfigOption, 0, len(config.Flags))
	currentSection := ""

	for _, flag := range config.Flags {
		if flag.Hidden() {
			continue
		}

		envVar := flag.EnvVar()

		if title, ok := sectionStarts[envVar]; ok {
			currentSection = title
		}

		items = append(items, ConfigOption{
			Section:     currentSection,
			Environment: envVar,
			CLIFlag:     flag.CommandFlag(),
			Default:     flag.Default(),
			Description: flag.Usage(),
			Tags:        flag.Tags,
		})
	}

	return items
}

// buildSearchFilters returns search filters sorted like the existing CLI report.
func buildSearchFilters() []SearchFilter {
	rows, _ := form.Report(&form.SearchPhotos{})

	sort.Slice(rows, func(i, j int) bool {
		if rows[i][1] == rows[j][1] {
			return rows[i][0] < rows[j][0]
		}

		return rows[i][1] < rows[j][1]
	})

	items := make([]SearchFilter, 0, len(rows))

	for _, row := range rows {
		if len(row) < 4 {
			continue
		}

		items = append(items, SearchFilter{
			Filter:   row[0],
			Type:     row[1],
			Examples: row[2],
			Notes:    row[3],
		})
	}

	return items
}

// normalizeEdition returns a lowercase edition identifier suitable for MCP output.
func normalizeEdition(edition string) string {
	edition = strings.TrimSpace(strings.ToLower(edition))

	if edition == "" {
		return "unknown"
	}

	return edition
}
