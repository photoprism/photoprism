## PhotoPrism MCP Prototype

**Last Updated:** March 25, 2026

### Current capabilities

- CLI entrypoint: `photoprism mcp serve`
- Transport: stdio only
- Read-only resources:
  - `photoprism://config-options`
  - `photoprism://search-filters`
- Read-only tools:
  - `list_config_keys`
  - `find_search_filters`
- Documented prompt templates:
  - Support answer draft
  - Config troubleshooting checklist
  - Search filter composer
  - Developer onboarding helper

### Goals and non-goals

Goals:

- Prove the MCP model works end-to-end inside the PhotoPrism codebase
- Reuse internal reference data instead of maintaining a separate copy
- Keep outputs concise enough for LLM use
- Make the prototype easy for another team member to run locally

Non-Goals:

- No write-capable tools
- No direct database access
- No live PhotoPrism instance or API queries
- No production authentication model
- No broad autonomous troubleshooting
- No streamable HTTP transport in Week 2

### Internal data sources

- Config options: `internal/config.Flags.Report()` plus `internal/config.OptionsReportSections`
- Search filters: `internal/form.Report(&form.SearchPhotos{})`

### Run locally

Build the CLI:

```bash
go build ./cmd/photoprism
```

Start the MCP server:

```bash
./photoprism mcp serve
```

The process waits for an MCP client on stdin/stdout. Logs are written to stderr so the MCP message stream stays valid.

### Test with MCP inspector

Launch the Inspector directly against the CLI command:

```bash
npx @modelcontextprotocol/inspector ./photoprism mcp serve
```

Useful smoke tests:

- List resources
- Read `photoprism://config-options`
- Read `photoprism://search-filters`
- Call `list_config_keys` with `{"query":"http","limit":3}`
- Call `find_search_filters` with `{"query":"Berlin","limit":5}`

### Available resources

`photoprism://config-options`

- JSON payload with `edition` and `items`
- Each item includes `section`, `environment`, `cli_flag`, `default`, and `description`

`photoprism://search-filters`

- JSON payload with `edition` and `items`
- Each item includes `filter`, `type`, `examples`, and `notes`

### Available tools

`list_config_keys`

- Inputs: `section`, `query`, `edition`, `limit`
- Returns matching config rows with environment variables, CLI flags, defaults, descriptions, and a conservative `edition_support` hint
- Validation rejects unsupported `edition` values

`find_search_filters`

- Inputs: `query`, `type`, `limit`
- Returns matching search filters with examples and notes
- Validation rejects unsupported filter `type` values
