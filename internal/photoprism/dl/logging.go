package dl

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// redactArgs returns a copy of args for the trace: the value of an --add-header "Name: Value" pair
// becomes the shared marker, and a URI has its credentials removed, which re-encodes it. Arguments
// stay unquoted, as the trace prints them as a list.
func redactArgs(args []string) []string {
	out := make([]string, len(args))
	copy(out, args)
	for i := 0; i < len(out); i++ {
		if out[i] == "--add-header" && i+1 < len(out) {
			hv := out[i+1]
			if idx := strings.Index(hv, ":"); idx > 0 {
				name := strings.TrimSpace(hv[:idx])
				out[i+1] = name + ": " + txt.Masked
			} else {
				out[i+1] = txt.Masked
			}
			i++
			continue
		}

		out[i] = clean.UriRedactedText(out[i])
	}
	return out
}
