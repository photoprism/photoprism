package commands

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// newFacesResetContext parses args against the flags the "faces reset" subcommand registers, which
// the shared test context does not do: it applies the app's flags, not a subcommand's.
func newFacesResetContext(t *testing.T, args ...string) *cli.Context {
	t.Helper()

	var cmd *cli.Command

	for _, c := range FacesCommands.Subcommands {
		if c.Name == "reset" {
			cmd = c
			break
		}
	}

	require.NotNil(t, cmd, "faces reset must be registered")

	flagSet := flag.NewFlagSet("reset", flag.ContinueOnError)

	for _, f := range cmd.Flags {
		require.NoError(t, f.Apply(flagSet))
	}

	require.NoError(t, flagSet.Parse(args))

	return cli.NewContext(cli.NewApp(), flagSet, nil)
}

// TestFacesResetFlags covers the flag surface and the combinations the action refuses.
//
// Nothing here reaches the prompt or the database: every case asserts on a refusal that happens
// before either, which is also why there is no positive control for the destructive paths.
func TestFacesResetFlags(t *testing.T) {
	t.Run("AllIsRegistered", func(t *testing.T) {
		ctx := newFacesResetContext(t, "--all")

		assert.True(t, ctx.Bool("all"))
		assert.False(t, ctx.Bool("force"))
	})
	t.Run("AllHasShortAlias", func(t *testing.T) {
		assert.True(t, newFacesResetContext(t, "-a").Bool("all"))
	})
	t.Run("ForceWithAll", func(t *testing.T) {
		err := facesResetAction(newFacesResetContext(t, "--force", "--all"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--all")

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 1, exit.ExitCode())
	})
	t.Run("ForceWithDetector", func(t *testing.T) {
		err := facesResetAction(newFacesResetContext(t, "--force", "--detector=yunet"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--detector")

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 1, exit.ExitCode())
	})
}
