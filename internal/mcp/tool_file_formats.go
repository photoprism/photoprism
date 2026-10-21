package mcp

import (
	"context"
	"fmt"
	"slices"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/pkg/txt"
)

// ListFileFormatsInput defines the inputs for the file format lookup tool.
type ListFileFormatsInput struct {
	Query  string `json:"query,omitempty" jsonschema:"optional case-insensitive search in format names, extensions, and descriptions"`
	Type   string `json:"type,omitempty" jsonschema:"optional format type, e.g. image, raw, video, or all"`
	Limit  int    `json:"limit,omitempty" jsonschema:"optional maximum number of matches to return (default 20, max 50)"`
	Offset int    `json:"offset,omitempty" jsonschema:"optional number of matches to skip"`
}

// ListFileFormatsOutput defines the structured output for the file format lookup tool.
type ListFileFormatsOutput struct {
	Items     []FileFormat `json:"items"`
	Total     int          `json:"total"`
	Limit     int          `json:"limit"`
	Offset    int          `json:"offset"`
	Truncated bool         `json:"truncated"`
	Warnings  []string     `json:"warnings,omitempty"`
}

// listFileFormats validates the caller's input, applies the type and query
// filters over data.FileFormats, and returns the requested page of matches.
func listFileFormats(_ context.Context, _ *sdkmcp.CallToolRequest, input ListFileFormatsInput, data *Dataset) (*sdkmcp.CallToolResult, ListFileFormatsOutput, error) {
	formatType, err := validateFileFormatType(input.Type, data.FileFormatTypes)
	if err != nil {
		return nil, ListFileFormatsOutput{}, err
	}

	limit, err := validateLimit(input.Limit)
	if err != nil {
		return nil, ListFileFormatsOutput{}, err
	}

	offset, err := validateOffset(input.Offset)
	if err != nil {
		return nil, ListFileFormatsOutput{}, err
	}

	query, err := validateString("query", input.Query)
	if err != nil {
		return nil, ListFileFormatsOutput{}, err
	}

	matches := make([]FileFormat, 0, len(data.FileFormats))

	for _, format := range data.FileFormats {
		if formatType != "all" && format.Type != formatType {
			continue
		}

		if !matchesFileFormatQuery(format, query) {
			continue
		}

		matches = append(matches, format)
	}

	items, truncated := pageOf(matches, limit, offset)

	return nil, ListFileFormatsOutput{
		Items:     items,
		Total:     len(matches),
		Limit:     limit,
		Offset:    offset,
		Truncated: truncated,
		Warnings:  pageWarnings(len(items), len(matches), offset, truncated),
	}, nil
}

// validateFileFormatType normalizes the requested format type and checks it
// against the types of the supported formats. An empty input defaults to "all".
func validateFileFormatType(formatType string, types []string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(formatType))

	if normalized == "" || normalized == "all" {
		return "all", nil
	}

	if !slices.Contains(types, normalized) {
		return "", fmt.Errorf("type must be one of %s", txt.JoinOr(append(slices.Clone(types), "all")))
	}

	return normalized, nil
}

// matchesFileFormatQuery reports whether the format matches the requested query.
func matchesFileFormatQuery(format FileFormat, query string) bool {
	if query == "" {
		return true
	}

	if strings.Contains(strings.ToLower(format.Format), query) ||
		strings.Contains(strings.ToLower(format.Description), query) {
		return true
	}

	for _, ext := range format.Extensions {
		if strings.Contains(strings.ToLower(ext), query) {
			return true
		}
	}

	return false
}
