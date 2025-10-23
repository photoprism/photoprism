package status

import (
	"github.com/photoprism/photoprism/pkg/clean"
)

// Error returns a sanitized string representation of err for use in audit and
// system logs, for example when an error message should be the final outcome
// token in an `event.Audit*` slice.
func Error(err error) string {
	return clean.Error(err)
}
