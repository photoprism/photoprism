package vision

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRunType(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  RunType
	}{
		{name: "EmptyIsAuto", in: "", out: RunAuto},
		{name: "WhitespaceTrim", in: "  manual  ", out: RunManual},
		{name: "SynonymManually", in: "manually", out: RunManual},
		{name: "SynonymCommand", in: "command", out: RunManual},
		{name: "UppercaseSchedule", in: "ON-SCHEDULE", out: RunOnSchedule},
		{name: "IndexAlias", in: "index", out: RunOnIndex},
		{name: "ExplicitOnIndex", in: "on-index", out: RunOnIndex},
		{name: "NewlyIndexedAlias", in: "on-newly-indexed", out: RunNewlyIndexed},
		{name: "AfterIndexAlias", in: "after-index", out: RunNewlyIndexed},
		{name: "UnknownFallsBack", in: "something", out: RunAuto},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {
			if got := ParseRunType(tc.in); got != tc.out {
				t.Fatalf("ParseRunType(%q) = %q, want %q", tc.in, got, tc.out)
			}
		})
	}
}

func TestModel_RunType(t *testing.T) {
	cases := []struct {
		name  string
		model *Model
		want  RunType
	}{
		{
			name:  "Nil",
			model: nil,
			want:  RunAuto,
		},
		{
			name:  "Manual",
			model: &Model{Run: "manual"},
			want:  RunManual,
		},
		{
			name:  "AfterIndex",
			model: &Model{Run: "after-index"},
			want:  RunNewlyIndexed,
		},
		{
			name:  "DefaultAuto",
			model: &Model{Run: ""},
			want:  RunAuto,
		},
		{
			name:  "UnknownString",
			model: &Model{Run: "custom"},
			want:  RunAuto,
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {
			if got := tc.model.RunType(); got != tc.want {
				t.Fatalf("(*Model).RunType() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestModel_ShouldRun_AutoDefault(t *testing.T) {
	model := NasnetModel.Clone()
	model.Run = ""

	assertShouldRun(t, model, RunManual, true)
	assertShouldRun(t, model, RunOnDemand, true)
	assertShouldRun(t, model, RunOnSchedule, true)
	assertShouldRun(t, model, RunAlways, true)
	assertShouldRun(t, model, RunOnIndex, true)
	assertShouldRun(t, model, RunNewlyIndexed, false)
	assertShouldRun(t, model, RunNever, false)
}

func TestModel_ShouldRun_AutoCustom(t *testing.T) {
	model := &Model{Run: "", Type: ModelTypeLabels, Name: "custom"}

	assertShouldRun(t, model, RunManual, true)
	assertShouldRun(t, model, RunOnDemand, true)
	assertShouldRun(t, model, RunOnSchedule, true)
	assertShouldRun(t, model, RunAlways, false)
	assertShouldRun(t, model, RunOnIndex, false)
	assertShouldRun(t, model, RunNewlyIndexed, true)
}

func TestModel_ShouldRun_RunNewlyIndexed(t *testing.T) {
	model := &Model{Run: string(RunNewlyIndexed)}

	assertShouldRun(t, model, RunManual, true)
	assertShouldRun(t, model, RunNewlyIndexed, true)
	assertShouldRun(t, model, RunOnDemand, true)
	assertShouldRun(t, model, RunOnSchedule, false)
}

func TestModel_ShouldRun_RunOnSchedule(t *testing.T) {
	model := &Model{Run: string(RunOnSchedule)}

	assertShouldRun(t, model, RunManual, true)
	assertShouldRun(t, model, RunOnSchedule, true)
	assertShouldRun(t, model, RunOnDemand, true)
	assertShouldRun(t, model, RunNewlyIndexed, false)
}

func TestModel_ShouldRun_RunAlways(t *testing.T) {
	model := &Model{Run: string(RunAlways)}

	assertShouldRun(t, model, RunManual, true)
	assertShouldRun(t, model, RunOnSchedule, true)
	assertShouldRun(t, model, RunNewlyIndexed, true)
	assertShouldRun(t, model, RunOnDemand, true)
	assertShouldRun(t, model, RunNever, false)
}

func TestModel_ShouldRun_RunManual(t *testing.T) {
	model := &Model{Run: string(RunManual)}

	assertShouldRun(t, model, RunManual, true)
	assertShouldRun(t, model, RunOnDemand, false)
	assertShouldRun(t, model, RunOnIndex, false)
}

func TestModel_ShouldRun_RunNever(t *testing.T) {
	model := &Model{Run: string(RunNever)}

	assertShouldRun(t, model, RunManual, false)
	assertShouldRun(t, model, RunOnDemand, false)
}

func TestModel_ShouldRun_NilModel(t *testing.T) {
	var model *Model
	if model.ShouldRun(RunManual) {
		t.Fatalf("expected nil model to never run")
	}
}

func TestModel_ShouldRun_RunOnIndex(t *testing.T) {
	model := &Model{Run: string(RunOnIndex)}

	assertShouldRun(t, model, RunManual, true)
	assertShouldRun(t, model, RunOnIndex, true)
	assertShouldRun(t, model, RunOnSchedule, false)
	assertShouldRun(t, model, RunOnDemand, false)
}

func assertShouldRun(t *testing.T, m *Model, when RunType, want bool) {
	if got := m.ShouldRun(when); got != want {
		t.Fatalf("ShouldRun(%q) = %v, want %v (model run=%q)", when, got, want, m.RunType())
	}
}

func TestRunTypeUsageString(t *testing.T) {
	usage := RunTypeUsageString()

	t.Run("NamesTheCanonicalTypes", func(t *testing.T) {
		for _, run := range []RunType{RunAlways, RunOnIndex, RunNewlyIndexed, RunOnSchedule, RunOnDemand, RunManual, RunNever} {
			assert.Contains(t, usage, run)
		}

		assert.True(t, strings.HasPrefix(usage, "auto, "))
	})
	t.Run("OmitsAliases", func(t *testing.T) {
		// The aliases add no choice and would make the list unreadable in a help listing.
		assert.NotContains(t, usage, "after-index")
		assert.NotContains(t, usage, "manually")
	})
	t.Run("EveryNameParses", func(t *testing.T) {
		for _, name := range strings.Split(usage, ", ") {
			assert.True(t, KnownRunType(name), name)
		}
	})
}

// TestKnownRunType pins what ParseRunType cannot answer: it returns RunAuto both for the values
// that ask for it and for the ones it does not recognize, so a typo could not be reported.
func TestKnownRunType(t *testing.T) {
	t.Run("Known", func(t *testing.T) {
		assert.True(t, KnownRunType("auto"))
		assert.True(t, KnownRunType(""))
		assert.True(t, KnownRunType("On-Schedule"))
		assert.True(t, KnownRunType("after-index"))
		// Underscores and spaces normalize to dashes, as they do for model and detector names.
		assert.True(t, KnownRunType("on_schedule"))
		assert.True(t, KnownRunType("on schedule"))
	})
	t.Run("Unknown", func(t *testing.T) {
		// ParseRunType answers RunAuto for these as well as for "auto", which is why the report
		// of an unsupported value could never fire while it was derived from the parsed result.
		assert.False(t, KnownRunType("whenever"))
		assert.False(t, KnownRunType("bogus"))
		assert.Equal(t, RunAuto, ParseRunType("whenever"))
	})
}
