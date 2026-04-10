package mcp

import (
	"context"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

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
}

// SearchFilterMatch represents a single search filter match.
type SearchFilterMatch struct {
	Filter   string `json:"filter"`
	Type     string `json:"type"`
	Examples string `json:"examples"`
	Notes    string `json:"notes"`
}

// findSearchFilters validates the request and returns matching filters.
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

	return nil, FindSearchFiltersOutput{
		Matches:      matches,
		TotalMatches: total,
		Truncated:    total > len(matches),
		QueryApplied: strings.TrimSpace(input.Query),
		TypeApplied:  filterType,
	}, nil
}

// validateSearchFilterType validates and normalizes the optional filter type.
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
