package commands

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestTrailingFlagToken(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"NoTail", []string{"alice"}, ""},
		{"OnlyPositionals", []string{"alice", "bob"}, ""},
		{"LongFlagAfterPositional", []string{"alice", "--name", "Alicia"}, "--name"},
		{"EqualsFlagAfterPositional", []string{"alice", "--role=guest"}, "--role=guest"},
		{"ShortFlagAfterPositional", []string{"alice", "-r", "guest"}, "-r"},
		{"BareDashIsPositional", []string{"alice", "-"}, ""},
		{"DoubleDashStopsScan", []string{"alice", "--", "--name", "literal"}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			if err := set.Parse(tc.args); err != nil {
				t.Fatalf("flag parse: %v", err)
			}
			ctx := cli.NewContext(nil, set, nil)
			assert.Equal(t, tc.want, TrailingFlagToken(ctx))
		})
	}
}

func TestRejectTrailingFlags(t *testing.T) {
	t.Run("NoTrailingFlag", func(t *testing.T) {
		set := flag.NewFlagSet("test", flag.ContinueOnError)
		_ = set.Parse([]string{"alice"})
		ctx := cli.NewContext(nil, set, nil)
		assert.NoError(t, RejectTrailingFlags(ctx))
	})

	t.Run("TrailingFlagReturnsError", func(t *testing.T) {
		set := flag.NewFlagSet("test", flag.ContinueOnError)
		_ = set.Parse([]string{"alice", "--name", "Alicia"})
		ctx := cli.NewContext(nil, set, nil)
		err := RejectTrailingFlags(ctx)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "--name")
			assert.Contains(t, err.Error(), "must appear before positional arguments")
		}
	})
}
