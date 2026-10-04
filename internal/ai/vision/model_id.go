package vision

import (
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

// clippedModelIdWarned holds the shortened model identifiers that were logged.
var clippedModelIdWarned sync.Map

// cleanModelId sanitizes a model identifier with clean.Type and logs a warning once per
// identifier that had to be shortened to clean.LengthType characters. It writes to the
// system log, as a shortened identifier is a configuration issue for the operator.
func cleanModelId(s string) string {
	id := clean.Type(s)

	if id == "" || len(strings.TrimSpace(clean.ASCII(s))) <= clean.LengthType {
		return id
	}

	if _, warned := clippedModelIdWarned.LoadOrStore(id, struct{}{}); !warned {
		event.SystemWarn([]string{"vision", "model identifier exceeds %d characters, using %s"}, clean.LengthType, clean.Log(id))
	}

	return id
}
