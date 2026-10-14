package vision

import (
	"errors"
	"sync"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

// modelIdWarned holds the reported sanitized model identifiers, keyed by the kind of issue.
var modelIdWarned sync.Map

// cleanModelId sanitizes a model identifier with clean.ModelId and logs a warning once per identifier and
// issue, if it had to be shortened or had characters removed. It writes to the system log, as both are
// configuration issues for the operator.
func cleanModelId(s string) string {
	id, err := clean.ModelId(s)

	if err == nil || id == "" {
		return id
	}

	if errors.Is(err, clean.ErrTooLong) {
		if _, warned := modelIdWarned.LoadOrStore("length:"+id, struct{}{}); !warned {
			event.SystemWarn([]string{"vision", "model identifier exceeds %d characters, using %s"}, clean.LengthType, clean.Log(id))
		}
	}

	if errors.Is(err, clean.ErrInvalidChars) {
		if _, warned := modelIdWarned.LoadOrStore("chars:"+id, struct{}{}); !warned {
			event.SystemWarn([]string{"vision", "model identifier contains unsupported characters, using %s"}, clean.Log(id))
		}
	}

	return id
}

// hasModelId reports whether s holds a model identifier once sanitized.
func hasModelId(s string) bool {
	id, _ := clean.ModelId(s)
	return id != ""
}
