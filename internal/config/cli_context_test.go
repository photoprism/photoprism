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
		tc := tc
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
