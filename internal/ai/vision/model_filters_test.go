package vision

import (
	"testing"
)

func TestFilterModels(t *testing.T) {
	cases := []struct {
		name     string
		models   []string
		when     RunType
		allow    func(ModelType, RunType) bool
		expected []string
	}{
		{
			name:     "NilPredicate",
			models:   []string{"caption", "labels"},
			when:     RunManual,
			allow:    nil,
			expected: []string{"caption", "labels"},
		},
		{
			name:   "SkipUnknown",
			models: []string{"caption", "", "unknown", "labels"},
			when:   RunManual,
			allow: func(mt ModelType, when RunType) bool {
				return mt == ModelTypeLabels
			},
			expected: []string{"labels"},
		},
		{
			name:   "ContextAware",
			models: []string{"caption", "labels"},
			when:   RunOnSchedule,
			allow: func(mt ModelType, when RunType) bool {
				return mt == ModelTypeCaption
			},
			expected: []string{"caption"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := FilterModels(tc.models, tc.when, tc.allow)
			if len(got) != len(tc.expected) {
				t.Fatalf("expected %d models, got %d", len(tc.expected), len(got))
			}
			for i := range got {
				if got[i] != tc.expected[i] {
					t.Fatalf("expected %v, got %v", tc.expected, got)
				}
			}
		})
	}
}
