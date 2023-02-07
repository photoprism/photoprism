package commands

import (
	"github.com/photoprism/photoprism/internal/config"
)

// Report represents a report table with title and options.
type Report struct {
	Title  string
	NoWrap bool
	Report func(*config.Config) ([][]string, []string)
}
