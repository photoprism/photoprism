package vision

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// RunType specifies when a vision model should be run.
type RunType = string

const (
	// RunAuto automatically decides when to run based on model type and configuration.
	RunAuto RunType = ""
	// RunNever disables the model entirely.
	RunNever RunType = "never"
	// RunManual runs only when explicitly invoked (e.g., via the "vision run" command).
	RunManual RunType = "manual"
	// RunAlways runs manually, on-schedule, on-demand, and on-index.
	RunAlways RunType = "always"
	// RunNewlyIndexed runs manually and for newly indexed pictures.
	RunNewlyIndexed RunType = "newly-indexed"
	// RunOnDemand runs manually, for newly indexed pictures, and on configured schedule.
	RunOnDemand RunType = "on-demand"
	// RunOnSchedule runs manually and on-schedule.
	RunOnSchedule RunType = "on-schedule"
	// RunOnIndex runs manually and after indexing.
	RunOnIndex RunType = "on-index"
)

// ReportRunType returns a human-readable string for the run type, preserving the
// explicit value when set or "auto" when delegation is in effect.
func ReportRunType(when RunType) string {
	when = ParseRunType(when)

	if when == RunAuto {
		return "auto"
	}

	return when
}

// KnownRunType reports whether s names a supported run type. ParseRunType cannot answer this:
// it returns RunAuto both for the values that ask for it and for the ones it does not know.
func KnownRunType(s string) bool {
	_, found := RunTypes[clean.TypeLowerDash(s)]

	return found
}

// RunTypeUsageString lists the canonical run types for use in CLI help text. It leaves out the
// aliases RunTypes accepts, which would make the list unreadable without adding a choice.
func RunTypeUsageString() string {
	return strings.Join([]string{"auto", RunAlways, RunOnIndex, RunNewlyIndexed,
		RunOnSchedule, RunOnDemand, RunManual, RunNever}, ", ")
}

// RunTypes maps configuration strings to standard RunType model settings.
var RunTypes = map[string]RunType{
	RunAuto:            RunAuto,
	"auto":             RunAuto,
	RunNever:           RunNever,
	RunManual:          RunManual,
	"manually":         RunManual,
	"command":          RunManual,
	RunAlways:          RunAlways,
	RunNewlyIndexed:    RunNewlyIndexed,
	"on-newly-indexed": RunNewlyIndexed,
	"indexed":          RunNewlyIndexed,
	"on-indexed":       RunNewlyIndexed,
	"after-index":      RunNewlyIndexed,
	RunOnDemand:        RunOnDemand,
	RunOnSchedule:      RunOnSchedule,
	"schedule":         RunOnSchedule,
	RunOnIndex:         RunOnIndex,
	"index":            RunOnIndex,
}

// ParseRunType parses a run type string into the canonical RunType constant.
// Unknown or empty values default to RunAuto.
func ParseRunType(s string) RunType {
	if t, ok := RunTypes[clean.TypeLowerDash(s)]; ok {
		return t
	}

	return RunAuto
}

// RunType returns the normalized run type configured for the model. Nil
// receivers default to RunAuto.
func (m *Model) RunType() RunType {
	if m == nil {
		return RunAuto
	}

	return ParseRunType(m.Run)
}

// ShouldRun reports whether the model should execute in the specified
// scheduling context. Nil receivers always return false.
func (m *Model) ShouldRun(when RunType) bool {
	if m == nil {
		return false
	}

	when = ParseRunType(when)

	if should, decided := ShouldRunAt(m.RunType(), when); decided {
		return should
	}

	switch when {
	case RunAuto, RunManual, RunOnDemand, RunOnSchedule:
		return true
	case RunAlways, RunOnIndex:
		return m.IsDefault()
	case RunNewlyIndexed:
		return !m.IsDefault()
	}

	return false
}

// ShouldRunAt reports whether work configured to run at "run" executes in the "when" context, and
// whether that could be decided here at all.
//
// RunAuto is left undecided, because what "auto" means belongs to the caller: a vision model asks
// whether it is the default one, face detection whether the host is fast enough to detect inline.
// Everything else is one table rather than one per caller - a second copy dropped a term from
// RunOnDemand, which left FACE_RUN=on-demand skipping the scheduled pass it names.
func ShouldRunAt(run, when RunType) (should, decided bool) {
	when = ParseRunType(when)

	switch ParseRunType(run) {
	case RunNever:
		return false, true
	case RunManual:
		return when == RunManual, true
	case RunAlways:
		return when != RunNever, true
	case RunNewlyIndexed:
		return when == RunManual || when == RunNewlyIndexed || when == RunOnDemand, true
	case RunOnDemand:
		return when == RunAuto || when == RunManual || when == RunNewlyIndexed || when == RunOnDemand || when == RunOnSchedule, true
	case RunOnSchedule:
		return when == RunAuto || when == RunManual || when == RunOnSchedule || when == RunOnDemand, true
	case RunOnIndex:
		return when == RunManual || when == RunOnIndex, true
	}

	return false, false
}
