package mcp

import (
	"context"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListMetadataFieldsInput defines the inputs for the metadata field lookup tool.
type ListMetadataFieldsInput struct {
	Query  string `json:"query,omitempty" jsonschema:"optional case-insensitive search in field and tag names"`
	Limit  int    `json:"limit,omitempty" jsonschema:"optional maximum number of matches to return (default 20, max 50)"`
	Offset int    `json:"offset,omitempty" jsonschema:"optional number of matches to skip"`
}

// ListMetadataFieldsOutput defines the structured output for the metadata field lookup tool.
type ListMetadataFieldsOutput struct {
	Items     []MetadataField `json:"items"`
	Total     int             `json:"total"`
	Limit     int             `json:"limit"`
	Offset    int             `json:"offset"`
	Truncated bool            `json:"truncated"`
	Warnings  []string        `json:"warnings,omitempty"`
}

// listMetadataFields validates the caller's input, applies the query filter
// over data.MetadataFields, and returns the requested page of matches.
func listMetadataFields(_ context.Context, _ *sdkmcp.CallToolRequest, input ListMetadataFieldsInput, data *Dataset) (*sdkmcp.CallToolResult, ListMetadataFieldsOutput, error) {
	limit, err := validateLimit(input.Limit)
	if err != nil {
		return nil, ListMetadataFieldsOutput{}, err
	}

	offset, err := validateOffset(input.Offset)
	if err != nil {
		return nil, ListMetadataFieldsOutput{}, err
	}

	query, err := validateString("query", input.Query)
	if err != nil {
		return nil, ListMetadataFieldsOutput{}, err
	}

	matches := make([]MetadataField, 0, len(data.MetadataFields))

	for _, field := range data.MetadataFields {
		if matchesMetadataFieldQuery(field, query) {
			matches = append(matches, field)
		}
	}

	items, truncated := pageOf(matches, limit, offset)

	return nil, ListMetadataFieldsOutput{
		Items:     items,
		Total:     len(matches),
		Limit:     limit,
		Offset:    offset,
		Truncated: truncated,
		Warnings:  pageWarnings(len(items), len(matches), offset, truncated),
	}, nil
}

// matchesMetadataFieldQuery reports whether the field name or one of its tags matches the requested query.
func matchesMetadataFieldQuery(field MetadataField, query string) bool {
	if query == "" || strings.Contains(strings.ToLower(field.Field), query) {
		return true
	}

	for _, tags := range [][]string{field.Exiftool, field.XMP, field.DCMI} {
		for _, tag := range tags {
			if strings.Contains(strings.ToLower(tag), query) {
				return true
			}
		}
	}

	return false
}
