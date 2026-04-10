package mcp

import (
	"context"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultResultLimit = 20
	maxResultLimit     = 50
	maxQueryLength     = 200
)

var allowedEditions = map[string]struct{}{
	"all":    {},
	"ce":     {},
	"plus":   {},
	"pro":    {},
	"portal": {},
}

// ListConfigKeysInput defines the supported inputs for the config lookup tool.
type ListConfigKeysInput struct {
	Section string `json:"section,omitempty" jsonschema:"optional section title filter"`
	Query   string `json:"query,omitempty" jsonschema:"optional case-insensitive search query"`
	Edition string `json:"edition,omitempty" jsonschema:"optional edition: ce, plus, pro, portal, or all"`
	Limit   int    `json:"limit,omitempty" jsonschema:"optional maximum number of matches to return"`
}

// ListConfigKeysOutput defines the structured output for the config lookup tool.
type ListConfigKeysOutput struct {
	Matches        []ConfigKeyMatch `json:"matches"`
	TotalMatches   int              `json:"total_matches"`
	Truncated      bool             `json:"truncated"`
	QueryApplied   string           `json:"query_applied,omitempty"`
	SectionApplied string           `json:"section_applied,omitempty"`
	EditionApplied string           `json:"edition_applied"`
	Warnings       []string         `json:"warnings,omitempty"`
}

// ConfigKeyMatch represents a single config option match returned by the tool.
type ConfigKeyMatch struct {
	Section        string `json:"section"`
	Environment    string `json:"environment"`
	CLIFlag        string `json:"cli_flag"`
	Default        string `json:"default"`
	Description    string `json:"description"`
	EditionSupport string `json:"edition_support"`
}

// registerTools adds the read-only config lookup tool to the server.
func registerTools(server *sdkmcp.Server, data *Dataset) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "list_config_keys",
		Description: "Lists read-only PhotoPrism config keys from the existing config report.",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, input ListConfigKeysInput) (*sdkmcp.CallToolResult, ListConfigKeysOutput, error) {
		return listConfigKeys(ctx, req, input, data)
	})

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "find_search_filters",
		Description: "Finds PhotoPrism search filters from the existing search filter reference.",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, input FindSearchFiltersInput) (*sdkmcp.CallToolResult, FindSearchFiltersOutput, error) {
		return findSearchFilters(ctx, req, input, data)
	})
}

// listConfigKeys validates the request and returns compact config matches.
func listConfigKeys(_ context.Context, _ *sdkmcp.CallToolRequest, input ListConfigKeysInput, data *Dataset) (*sdkmcp.CallToolResult, ListConfigKeysOutput, error) {
	edition, err := validateEdition(input.Edition)
	if err != nil {
		return nil, ListConfigKeysOutput{}, err
	}

	limit, err := validateLimit(input.Limit)
	if err != nil {
		return nil, ListConfigKeysOutput{}, err
	}

	query, err := validateString("query", input.Query)
	if err != nil {
		return nil, ListConfigKeysOutput{}, err
	}

	section, err := validateString("section", input.Section)
	if err != nil {
		return nil, ListConfigKeysOutput{}, err
	}
	matches := make([]ConfigKeyMatch, 0, limit)
	total := 0

	for _, option := range data.ConfigOptions {
		if !matchesSection(option, section) {
			continue
		}

		if !matchesQuery(option, query) {
			continue
		}

		total++

		if len(matches) < limit {
			matches = append(matches, ConfigKeyMatch{
				Section:        option.Section,
				Environment:    option.Environment,
				CLIFlag:        option.CLIFlag,
				Default:        option.Default,
				Description:    option.Description,
				EditionSupport: editionSupportFor(option, data.CurrentEdition),
			})
		}
	}

	result := ListConfigKeysOutput{
		Matches:        matches,
		TotalMatches:   total,
		Truncated:      total > len(matches),
		QueryApplied:   strings.TrimSpace(input.Query),
		SectionApplied: strings.TrimSpace(input.Section),
		EditionApplied: edition,
	}

	if edition != "all" && edition != data.CurrentEdition {
		result.Warnings = []string{
			fmt.Sprintf("edition filtering is advisory in this prototype; results come from the current %s build metadata", data.CurrentEdition),
		}
	}

	return nil, result, nil
}

// validateEdition validates and normalizes the requested edition.
func validateEdition(edition string) (string, error) {
	if strings.TrimSpace(edition) == "" {
		return "all", nil
	}

	normalized := normalizeEdition(edition)

	if _, ok := allowedEditions[normalized]; !ok {
		return "", fmt.Errorf("edition must be one of ce, plus, pro, portal, or all")
	}

	return normalized, nil
}

// validateLimit validates and normalizes the requested result limit.
func validateLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultResultLimit, nil
	}

	if limit < 1 {
		return 0, fmt.Errorf("limit must be greater than 0")
	}

	if limit > maxResultLimit {
		return maxResultLimit, nil
	}

	return limit, nil
}

// validateString validates and normalizes a free-text input field, using the
// supplied field name in error messages so callers can reuse it for query,
// section, or any other string parameter.
func validateString(field, value string) (string, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	if len(value) > maxQueryLength {
		return "", fmt.Errorf("%s must not exceed %d characters", field, maxQueryLength)
	}

	return value, nil
}

// matchesSection reports whether an option matches the requested section filter.
func matchesSection(option ConfigOption, section string) bool {
	if section == "" {
		return true
	}

	return strings.Contains(strings.ToLower(option.Section), section)
}

// matchesQuery reports whether an option matches the requested free-text query.
func matchesQuery(option ConfigOption, query string) bool {
	if query == "" {
		return true
	}

	haystacks := []string{
		option.Environment,
		option.CLIFlag,
		option.Description,
	}

	for _, haystack := range haystacks {
		if strings.Contains(strings.ToLower(haystack), query) {
			return true
		}
	}

	return false
}

// editionSupportFor returns an edition annotation based on a config option's tags.
func editionSupportFor(option ConfigOption, currentEdition string) string {
	if currentEdition == "unknown" {
		return "unknown"
	}

	if len(option.Tags) == 0 {
		return "all"
	}

	// Check tags in priority order (most restrictive first).
	for _, tag := range []string{"portal", "pro", "plus", "essentials"} {
		for _, t := range option.Tags {
			if t == tag {
				return tag
			}
		}
	}

	return "all"
}
