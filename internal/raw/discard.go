package raw

import (
	"os"
	"sort"
	"strings"
)

// discardOnWarningEnv overrides the RAW extensions whose RawTherapee render is discarded on a decode
// warning, as a comma-separated list that replaces the default set (e.g. "cr3, cr2"). It is an escape
// hatch outside the CLI/config surface for adding a newly found magenta-prone format without a rebuild.
const discardOnWarningEnv = "PHOTOPRISM_RAW_DISCARD_ON_WARNING_EXT"

// discardRenderOnWarning holds the RAW extensions (lowercase, leading dot) whose RawTherapee render is
// discarded when its stderr matches DecoderErrors, so the embedded preview wins (see README.md).
// A gated format must allow preview extraction (PreviewExtAllowed), or its render has no fallback.
var discardRenderOnWarning = defaultDiscardExt()

// init warns when a configured gated format has no usable embedded preview, since a discarded render
// would then have no fallback and the file would fail to index.
func init() {
	for _, ext := range discardExtMissingPreview() {
		log.Warnf("raw: discard-on-warning format %s has no embedded preview to fall back to", ext)
	}
}

// defaultDiscardExt returns the gated-extension set, reading discardOnWarningEnv when set and falling
// back to Canon CR3 otherwise.
func defaultDiscardExt() map[string]bool {
	if env := strings.TrimSpace(os.Getenv(discardOnWarningEnv)); env != "" {
		return parseDiscardExt(env)
	}

	return parseDiscardExt(".cr3")
}

// parseDiscardExt parses a comma-separated extension list into a normalized set (lowercase, leading
// dot). An override that yields no valid entry returns an empty set, disabling the gate entirely.
func parseDiscardExt(s string) map[string]bool {
	set := make(map[string]bool)

	for _, item := range strings.Split(s, ",") {
		ext := strings.ToLower(strings.TrimSpace(item))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		set[ext] = true
	}

	return set
}

// discardExtMissingPreview returns the gated extensions that have no usable embedded preview, sorted, so
// a discarded render would have no fallback. The default set always returns none.
func discardExtMissingPreview() []string {
	var missing []string

	for ext := range discardRenderOnWarning {
		if !PreviewExtAllowed(ext) {
			missing = append(missing, ext)
		}
	}

	sort.Strings(missing)

	return missing
}

// DiscardRenderOnWarning reports whether a RawTherapee render of the extension should be discarded when
// its stderr reports a decode warning (DecoderErrors), preferring the embedded preview. The extension
// must be lowercase with a leading dot (e.g. ".cr3").
func DiscardRenderOnWarning(ext string) bool {
	return discardRenderOnWarning[ext]
}
