package status

import (
	"errors"
	"testing"

	"github.com/photoprism/photoprism/pkg/clean"
)

func TestError(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "Nil",
			err:  nil,
		},
		{
			name: "SanitizeSpecialCharacters",
			err:  errors.New("permission denied { DROP TABLE users; }"),
		},
		{
			name: "WhitespaceOnly",
			err:  errors.New("   "),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, want := Error(tt.err), clean.Error(tt.err); got != want {
				t.Fatalf("Error(%v) = %q, want %q", tt.err, got, want)
			}
		})
	}
}
