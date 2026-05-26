package config

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type durationTarget struct {
	Interval time.Duration `flag:"interval"`
}

type indexWorkersTarget struct {
	IndexWorkers string `flag:"index-workers"`
}

func TestApplyCliContext_Duration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{name: "WithUnits", input: "1h30m", expected: 90 * time.Minute},
		{name: "NumericSeconds", input: "30", expected: 30 * time.Second},
		{name: "Invalid", input: "not-a-duration", expected: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flags := flag.NewFlagSet("test", flag.ContinueOnError)
			flags.String("interval", "", "doc")
			app := cli.NewApp()
			ctx := cli.NewContext(app, flags, nil)
			_ = ctx.Set("interval", tc.input)

			target := &durationTarget{}
			err := ApplyCliContext(target, ctx)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, target.Interval)
		})
	}
}

// TestApplyCliContext_IndexWorkersString verifies that the index-workers
// option, recently changed from int to string with "auto" as the default,
// propagates correctly through the reflection-based CLI binder. Both the
// "auto" sentinel and a numeric override must land verbatim in the
// Options.IndexWorkers field for IndexWorkers() to interpret later.
func TestApplyCliContext_IndexWorkersString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "Auto", input: IndexWorkersAuto},
		{name: "Numeric", input: "4"},
		{name: "Empty", input: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := flag.NewFlagSet("test", flag.ContinueOnError)
			flags.String("index-workers", "", "doc")
			app := cli.NewApp()
			ctx := cli.NewContext(app, flags, nil)
			_ = ctx.Set("index-workers", tc.input)

			target := &indexWorkersTarget{}
			err := ApplyCliContext(target, ctx)

			assert.NoError(t, err)
			assert.Equal(t, tc.input, target.IndexWorkers)
		})
	}
}
