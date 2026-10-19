package mcp

import (
	"context"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// cliCommandsUnavailable is the warning returned if the server was created without a command list.
const cliCommandsUnavailable = "the command list of this edition is not available on this server"

// ListCliCommandsInput defines the inputs for the CLI command lookup tool.
type ListCliCommandsInput struct {
	Query  string `json:"query,omitempty" jsonschema:"optional case-insensitive search in command paths, aliases, usage, descriptions, and flag names"`
	Limit  int    `json:"limit,omitempty" jsonschema:"optional maximum number of matches to return (default 20, max 50)"`
	Offset int    `json:"offset,omitempty" jsonschema:"optional number of matches to skip"`
}

// ListCliCommandsOutput defines the structured output for the CLI command lookup tool.
type ListCliCommandsOutput struct {
	Items     []CliCommand `json:"items"`
	Total     int          `json:"total"`
	Limit     int          `json:"limit"`
	Offset    int          `json:"offset"`
	Truncated bool         `json:"truncated"`
	Warnings  []string     `json:"warnings,omitempty"`
}

// listCliCommands validates the caller's input, applies the query filter over
// data.CliCommands, and returns the requested page of matches.
func listCliCommands(_ context.Context, _ *sdkmcp.CallToolRequest, input ListCliCommandsInput, data *Dataset) (*sdkmcp.CallToolResult, ListCliCommandsOutput, error) {
	limit, err := validateLimit(input.Limit)
	if err != nil {
		return nil, ListCliCommandsOutput{}, err
	}

	offset, err := validateOffset(input.Offset)
	if err != nil {
		return nil, ListCliCommandsOutput{}, err
	}

	query, err := validateString("query", input.Query)
	if err != nil {
		return nil, ListCliCommandsOutput{}, err
	}

	matches := make([]CliCommand, 0, len(data.CliCommands))

	for _, cmd := range data.CliCommands {
		if matchesCliCommandQuery(cmd, query) {
			matches = append(matches, cmd)
		}
	}

	items, truncated := pageOf(matches, limit, offset)

	result := ListCliCommandsOutput{
		Items:     items,
		Total:     len(matches),
		Limit:     limit,
		Offset:    offset,
		Truncated: truncated,
		Warnings:  pageWarnings(len(items), len(matches), offset, truncated),
	}

	if data.CliCommands == nil {
		result.Warnings = append(result.Warnings, cliCommandsUnavailable)
	}

	return nil, result, nil
}

// matchesCliCommandQuery reports whether the command path, an alias, the usage, the description,
// or a flag name matches the requested query.
func matchesCliCommandQuery(cmd CliCommand, query string) bool {
	if query == "" {
		return true
	}

	haystacks := append([]string{cmd.Command, cmd.Usage, cmd.Description}, cmd.Aliases...)

	for _, f := range cmd.Flags {
		haystacks = append(haystacks, f.Name)
		haystacks = append(haystacks, f.Aliases...)
	}

	for _, haystack := range haystacks {
		if strings.Contains(strings.ToLower(haystack), query) {
			return true
		}
	}

	return false
}
