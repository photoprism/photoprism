package mcp

import (
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
)

// Resource URIs advertised to MCP clients. Each one maps to a JSON
// payload built from internal PhotoPrism reference data and registered
// in resources.go.
const (
	// configOptionsURI is the URI of the "config-options" resource,
	// which mirrors the `photoprism show config-options` CLI report.
	configOptionsURI = "photoprism://config-options"
	// searchFiltersURI is the URI of the "search-filters" resource,
	// which mirrors the `photoprism show search-filters` CLI report.
	searchFiltersURI = "photoprism://search-filters"
	// cliCommandsURI is the URI of the "cli-commands" resource, which mirrors
	// the `photoprism show commands` CLI report of the running edition.
	cliCommandsURI = "photoprism://cli-commands"
	// fileFormatsURI is the URI of the "file-formats" resource,
	// which mirrors the `photoprism show file-formats` CLI report.
	fileFormatsURI = "photoprism://file-formats"
	// metadataFieldsURI is the URI of the "metadata-fields" resource,
	// which mirrors the `photoprism show metadata` CLI report.
	metadataFieldsURI = "photoprism://metadata-fields"
	// thumbnailSizesURI is the URI of the "thumbnail-sizes" resource, which
	// mirrors the `photoprism show thumb-sizes` and `show video-sizes` reports.
	thumbnailSizesURI = "photoprism://thumbnail-sizes"
)

// Dataset caches the static MCP data returned by resources and tools.
type Dataset struct {
	CurrentEdition  string
	ConfigOptions   []ConfigOption
	SearchFilters   []SearchFilter
	CliCommands     []CliCommand
	FileFormats     []FileFormat
	FileFormatTypes []string
	MetadataFields  []MetadataField
	MetadataDocs    []MetadataDoc
	ThumbnailSizes  []ThumbnailSize
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

// NewDataset builds the static MCP dataset for the current build. CliCommands is nil if
// opt.Commands is nil.
func NewDataset(currentEdition string, opt Options) *Dataset {
	fileFormats := buildFileFormats()

	return &Dataset{
		CurrentEdition:  normalizeEdition(currentEdition),
		ConfigOptions:   buildConfigOptions(),
		SearchFilters:   buildSearchFilters(),
		CliCommands:     buildCliCommands(opt.Commands, opt.GlobalFlags),
		FileFormats:     fileFormats,
		FileFormatTypes: fileFormatTypes(fileFormats),
		MetadataFields:  buildMetadataFields(),
		MetadataDocs:    buildMetadataDocs(),
		ThumbnailSizes:  buildThumbnailSizes(),
	}
}

// buildConfigOptions walks config.Flags and returns one ConfigOption per
// non-hidden flag, with the originating section title attached. Section
// titles come from config.OptionsReportSections; each section declares one
// or more "start" env vars that mark where the section begins in the flag
// list. Flags encountered before the first known start env var inherit an
// empty section, matching the layout of the `photoprism show config-*`
// CLI reports.
func buildConfigOptions() []ConfigOption {
	// Index section start env vars by the exact env var name for O(1)
	// lookup while iterating flags below.
	sectionStarts := make(map[string]string)

	for _, section := range config.OptionsReportSections {
		for key := range strings.SplitSeq(section.Start, ", ") {
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

		// Update the section title whenever we cross a section boundary.
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

// buildSearchFilters runs the existing search-filter report against
// form.SearchPhotos and returns the rows sorted by (type, filter) for
// deterministic MCP output. Rows with fewer than four columns are
// dropped defensively in case the report shape ever changes.
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
