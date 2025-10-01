package vision

import "testing"

func TestParseRunType(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  RunType
	}{
		{name: "EmptyIsAuto", in: "", out: RunAuto},
		{name: "WhitespaceTrim", in: "  manual  ", out: RunManual},
		{name: "SynonymManually", in: "manually", out: RunManual},
		{name: "UppercaseSchedule", in: "ON-SCHEDULE", out: RunOnSchedule},
		{name: "IndexAlias", in: "index", out: RunOnIndex},
		{name: "ExplicitOnIndex", in: "on-index", out: RunOnIndex},
		{name: "AfterIndexAlias", in: "after-index", out: RunNewlyIndexed},
		{name: "UnknownFallsBack", in: "something", out: RunAuto},
	}

	for _, tc := range cases {
		tc := tc

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
		tc := tc

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
