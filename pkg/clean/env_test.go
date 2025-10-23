package clean

import "testing"

func TestEnvVar(t *testing.T) {
	cases := []struct {
		flag     string
		expected string
	}{
		{"cluster-bootstrap-max-attempts", "PHOTOPRISM_CLUSTER_BOOTSTRAP_MAX_ATTEMPTS"},
		{"theme-path", "PHOTOPRISM_THEME_PATH"},
		{"debug", "PHOTOPRISM_DEBUG"},
	}

	for _, tc := range cases {
		t.Run(tc.flag, func(t *testing.T) {
			if got := EnvVar(tc.flag); got != tc.expected {
				t.Fatalf("EnvVar(%q) = %q, expected %q", tc.flag, got, tc.expected)
			}
		})
	}
}

func TestEnvVars(t *testing.T) {
	input := []string{"debug", "trace"}
	expected := []string{"PHOTOPRISM_DEBUG", "PHOTOPRISM_TRACE"}

	got := EnvVars(input...)

	if len(got) != len(expected) {
		t.Fatalf("EnvVars returned %d elements, expected %d", len(got), len(expected))
	}

	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("EnvVars[%d] = %q, expected %q", i, got[i], expected[i])
		}
	}
}

func TestEnvVarIdempotent(t *testing.T) {
	if EnvVar("already_upper") != "PHOTOPRISM_ALREADY_UPPER" {
		t.Fatalf("EnvVar should upper-case and replace hyphen/underscore consistently")
	}
}
