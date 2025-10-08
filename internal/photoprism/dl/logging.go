package dl

import "strings"

// redactArgs returns a copy of args with sensitive header values masked.
// It looks for patterns: --add-header "Name: Value" and rewrites Value as ****.
func redactArgs(args []string) []string {
	out := make([]string, len(args))
	copy(out, args)
	for i := 0; i < len(out); i++ {
		if out[i] == "--add-header" && i+1 < len(out) {
			hv := out[i+1]
			if idx := strings.Index(hv, ":"); idx > 0 {
				name := strings.TrimSpace(hv[:idx])
				out[i+1] = name + ": ****"
			} else {
				out[i+1] = "****"
			}
			i++
		}
	}
	return out
}
