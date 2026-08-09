package vision

import (
	"github.com/photoprism/photoprism/pkg/clean"
)

// NormalizeType specifies how label names returned by a model are normalized.
type NormalizeType = string

const (
	// NormalizeAuto uses the default normalization of the model's engine.
	NormalizeAuto NormalizeType = ""
	// NormalizeWord collapses names to a single canonical word.
	NormalizeWord NormalizeType = "single-word"
	// NormalizePhrase keeps multi-word names and matches them against the vocabulary as a whole.
	NormalizePhrase NormalizeType = "phrase"
	// NormalizeFalse keeps the name the model returned, with whitespace and case cleanup only.
	NormalizeFalse NormalizeType = "false"
)

// NormalizeTypes maps configuration strings to standard NormalizeType model settings.
var NormalizeTypes = map[string]NormalizeType{
	NormalizeAuto:   NormalizeAuto,
	"auto":          NormalizeAuto,
	"default":       NormalizeAuto,
	NormalizeWord:   NormalizeWord,
	"singleword":    NormalizeWord,
	"single":        NormalizeWord,
	"word":          NormalizeWord,
	"token":         NormalizeWord,
	NormalizePhrase: NormalizePhrase,
	"phrases":       NormalizePhrase,
	"multi-word":    NormalizePhrase,
	"multiword":     NormalizePhrase,
	NormalizeFalse:  NormalizeFalse,
	"off":           NormalizeFalse,
	"none":          NormalizeFalse,
	"no":            NormalizeFalse,
	"disabled":      NormalizeFalse,
}

// ParseNormalizeType parses a normalize type string into the canonical NormalizeType constant.
// Unknown or empty values default to NormalizeAuto.
func ParseNormalizeType(s string) NormalizeType {
	if t, ok := NormalizeTypes[clean.TypeLowerDash(s)]; ok {
		return t
	}

	return NormalizeAuto
}

// IsNormalizeType reports whether the string is a recognized normalize type, empty included.
func IsNormalizeType(s string) bool {
	_, ok := NormalizeTypes[clean.TypeLowerDash(s)]
	return ok
}

// ReportNormalizeType returns a human-readable string for the normalize type, preserving the
// explicit value when set or "auto" when the engine default applies.
func ReportNormalizeType(t NormalizeType) string {
	if t = ParseNormalizeType(t); t == NormalizeAuto {
		return "auto"
	}

	return t
}

// GetNormalize returns the label name normalization configured for the model, falling back to
// the default of its engine and finally to NormalizeWord. Nil receivers return NormalizeWord.
func (m *Model) GetNormalize() NormalizeType {
	if m == nil {
		return NormalizeWord
	}

	if t := ParseNormalizeType(m.Normalize); t != NormalizeAuto {
		return t
	}

	// A cloud model only uses a compound name when the subject has one.
	if m.IsCloud() {
		return NormalizePhrase
	}

	if info, ok := EngineInfoFor(m.Engine); ok {
		if t := ParseNormalizeType(info.DefaultNormalize); t != NormalizeAuto {
			return t
		}
	}

	return NormalizeWord
}
