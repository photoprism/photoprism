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

// TestConfirmAction covers the confirmation helper, and in particular that a prompt which cannot
// be shown is reported as an error rather than as a declined action.
func TestConfirmAction(t *testing.T) {
	// Cleared explicitly: with it set in the environment every case below would take the
	// non-interactive path, and the subtests that expect a prompt would assert nothing.
	t.Setenv("PHOTOPRISM_CLI", "")

	t.Run("ConfirmedSkipsThePrompt", func(t *testing.T) {
		proceed, err := ConfirmAction(true, "Remove everything?")

		assert.NoError(t, err)
		assert.True(t, proceed)
	})
	t.Run("NoTerminalIsAnError", func(t *testing.T) {
		// The test process has no terminal on stdin, so the prompt cannot run; the helper
		// reports that as an error rather than as a decision.
		proceed, err := ConfirmAction(false, "Remove everything?")

		require.Error(t, err)
		assert.False(t, proceed)
		assert.Contains(t, err.Error(), "--yes")

		// Usage error: the caller fixes it by passing --yes.
		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
	})
	t.Run("NonInteractiveEnvSkipsThePrompt", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", NONINTERACTIVE)

		proceed, err := ConfirmAction(false, "Remove everything?")

		assert.NoError(t, err)
		assert.True(t, proceed)
	})
}

// TestFacesResetRequiresConfirmation pins that a reset without a confirmation reports a failure
// rather than exiting successfully without touching anything.
func TestFacesResetRequiresConfirmation(t *testing.T) {
	// Cleared explicitly: with it set, these would confirm and run the reset for real
	// against the shared package test database.
	t.Setenv("PHOTOPRISM_CLI", "")

	t.Run("Scoped", func(t *testing.T) {
		err := facesResetAction(newFacesResetContext(t))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--yes")
	})
	t.Run("All", func(t *testing.T) {
		err := facesResetAction(newFacesResetContext(t, "--all"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--yes")
	})
	t.Run("Force", func(t *testing.T) {
		err := facesResetAction(newFacesResetContext(t, "--force"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--yes")
	})
	t.Run("YesIsRegistered", func(t *testing.T) {
		ctx := newFacesResetContext(t, "--yes")

		assert.True(t, ctx.Bool("yes"))
		assert.True(t, newFacesResetContext(t, "-y").Bool("yes"))
	})
	t.Run("YesReachesTheConfirmation", func(t *testing.T) {
		// Parsing the flag is not the same as the action reading it. With --yes the
		// confirmation is satisfied, so the run proceeds past it and fails later at the
		// config instead of returning the confirmation error.
		err := facesResetAction(newFacesResetContext(t, "--yes"))

		if err != nil {
			assert.NotContains(t, err.Error(), "could not ask for confirmation")
		}
	})
}
