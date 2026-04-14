package mcp

import (
	"context"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// truncationWarningTemplate is the fmt.Sprintf template used to populate
// the "warnings" field of a tool response when the number of matches
// exceeds the caller's limit. Arguments are: returned count, total count,
// hard upper bound (maxResultLimit).
const truncationWarningTemplate = "results truncated to %d of %d matches; raise the limit parameter (max %d) to retrieve more"

// allowedSearchFilterTypes is the closed allowlist of values accepted by
// the "type" argument of find_search_filters. An empty type defaults to
// "all"; anything outside this set is rejected with a validation error.
var allowedSearchFilterTypes = map[string]struct{}{
	"all":       {},
	"decimal":   {},
	"number":    {},
	"string":    {},
	"switch":    {},
	"timestamp": {},
}

// FindSearchFiltersInput defines the inputs for the search filter lookup tool.
type FindSearchFiltersInput struct {
	Query string `json:"query,omitempty" jsonschema:"optional case-insensitive search query"`
	Type  string `json:"type,omitempty" jsonschema:"optional filter type: decimal, number, string, switch, timestamp, or all"`
	Limit int    `json:"limit,omitempty" jsonschema:"optional maximum number of matches to return"`
}

// FindSearchFiltersOutput defines the structured output for the search filter lookup tool.
type FindSearchFiltersOutput struct {
	Matches      []SearchFilterMatch `json:"matches"`
	TotalMatches int                 `json:"total_matches"`
	Truncated    bool                `json:"truncated"`
	QueryApplied string              `json:"query_applied,omitempty"`
	TypeApplied  string              `json:"type_applied"`
	Warnings     []string            `json:"warnings,omitempty"`
}

// SearchFilterMatch represents a single search filter match.
type SearchFilterMatch struct {
	Filter   string `json:"filter"`
	Type     string `json:"type"`
	Examples string `json:"examples"`
	Notes    string `json:"notes"`
}

// findSearchFilters validates the caller's input, applies the type and
// query filters over data.SearchFilters, and returns at most `limit`
// matches alongside the total match count. When the result is truncated,
// an actionable warning is attached to the response via
// truncationWarningTemplate.
func findSearchFilters(_ context.Context, _ *sdkmcp.CallToolRequest, input FindSearchFiltersInput, data *Dataset) (*sdkmcp.CallToolResult, FindSearchFiltersOutput, error) {
	filterType, err := validateSearchFilterType(input.Type)
	if err != nil {
		return nil, FindSearchFiltersOutput{}, err
	}

	limit, err := validateLimit(input.Limit)
	if err != nil {
		return nil, FindSearchFiltersOutput{}, err
	}

	query, err := validateString("query", input.Query)
	if err != nil {
		return nil, FindSearchFiltersOutput{}, err
	}
	matches := make([]SearchFilterMatch, 0, limit)
	total := 0

	for _, filter := range data.SearchFilters {
		if !matchesSearchFilterType(filter, filterType) {
			continue
		}

		if !matchesSearchFilterQuery(filter, query) {
			continue
		}

		total++

		if len(matches) < limit {
			matches = append(matches, SearchFilterMatch{
				Filter:   filter.Filter,
				Type:     filter.Type,
				Examples: filter.Examples,
				Notes:    filter.Notes,
			})
		}
	}

	result := FindSearchFiltersOutput{
		Matches:      matches,
		TotalMatches: total,
		Truncated:    total > len(matches),
		QueryApplied: strings.TrimSpace(input.Query),
		TypeApplied:  filterType,
	}

	if result.Truncated {
		result.Warnings = append(result.Warnings, fmt.Sprintf(truncationWarningTemplate, len(matches), total, maxResultLimit))
	}

	return nil, result, nil
}

// validateSearchFilterType normalizes the requested filter type and
// checks it against allowedSearchFilterTypes. An empty input defaults to
// "all"; any value outside the allowlist yields a validation error
// surfaced to the caller.
func validateSearchFilterType(filterType string) (string, error) {
	if strings.TrimSpace(filterType) == "" {
		return "all", nil
	}

	normalized := strings.TrimSpace(strings.ToLower(filterType))

	if _, ok := allowedSearchFilterTypes[normalized]; !ok {
		return "", fmt.Errorf("type must be one of decimal, number, string, switch, timestamp, or all")
	}

	return normalized, nil
}

// matchesSearchFilterType reports whether the row matches the requested filter type.
func matchesSearchFilterType(filter SearchFilter, filterType string) bool {
	if filterType == "" || filterType == "all" {
		return true
	}

	return strings.EqualFold(filter.Type, filterType)
}

// matchesSearchFilterQuery reports whether the row matches the requested query.
func matchesSearchFilterQuery(filter SearchFilter, query string) bool {
	if query == "" {
		return true
	}

	haystacks := []string{
		filter.Filter,
		filter.Examples,
		filter.Notes,
	}

	for _, haystack := range haystacks {
		if strings.Contains(strings.ToLower(haystack), query) {
			return true
		}
	}

	return false
}
