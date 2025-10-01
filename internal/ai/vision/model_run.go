package vision

import (
	"github.com/photoprism/photoprism/pkg/clean"
)

// RunType specifies when a vision model should be run.
type RunType = string

const (
	RunAuto         RunType = ""              // Automatically decide when to run based on model type and configuration.
	RunNever        RunType = "never"         // Never run the model.
	RunManual       RunType = "manual"        // Only run manually e.g. with the "vision run" command.
	RunAlways       RunType = "always"        // Run manually, on-schedule, on-demand, and on-index.
	RunNewlyIndexed RunType = "newly-indexed" // Run manually amd for newly-indexed pictures.
	RunOnDemand     RunType = "on-demand"     // Run manually, for newly-indexed pictures, and on configured schedule.
	RunOnSchedule   RunType = "on-schedule"   // Run manually and on-schedule.
	RunOnIndex      RunType = "on-index"      // Run manually and on-index.
)

// RunTypes maps configuration strings to standard RunType model settings.
var RunTypes = map[string]RunType{
	RunAuto:         RunAuto,
	"auto":          RunAuto,
	RunNever:        RunNever,
	RunManual:       RunManual,
	"manually":      RunManual,
	"command":       RunManual,
	RunAlways:       RunAlways,
	RunNewlyIndexed: RunNewlyIndexed,
	"indexed":       RunNewlyIndexed,
	"on-indexed":    RunNewlyIndexed,
	"after-index":   RunNewlyIndexed,
	RunOnDemand:     RunOnDemand,
	RunOnSchedule:   RunOnSchedule,
	"schedule":      RunOnSchedule,
	RunOnIndex:      RunOnIndex,
	"index":         RunOnIndex,
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

	switch m.RunType() {
	case RunAuto:
		switch when {
		case RunAuto, RunManual, RunOnDemand, RunOnSchedule:
			return true
		case RunAlways, RunOnIndex:
			return m.IsDefault()
		case RunNewlyIndexed:
			return !m.IsDefault()
		}
	case RunNever:
		return false
	case RunManual:
		return when == RunManual
	case RunAlways:
		return when != RunNever
	case RunNewlyIndexed:
		return when == RunManual || when == RunNewlyIndexed || when == RunOnDemand
	case RunOnDemand:
		return when == RunAuto || when == RunManual || when == RunNewlyIndexed || when == RunOnDemand || when == RunOnSchedule
	case RunOnSchedule:
		return when == RunAuto || when == RunManual || when == RunOnSchedule || when == RunOnDemand
	case RunOnIndex:
		return when == RunManual || when == RunOnIndex
	}

	return false
}
