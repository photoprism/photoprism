package report

// Format represents a report output format.
type Format string

const (
	// Default selects the default format.
	Default Format = ""
	// Markdown renders output as Markdown.
	Markdown Format = "markdown"
	// TSV renders output as tab separated values.
	TSV Format = "tsv"
	// CSV renders output as comma separated values.
	CSV Format = "csv"
	// JSON renders output as JSON.
	JSON Format = "json"
)
