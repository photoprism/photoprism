package commands

import (
	"errors"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
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
	t.Run("TraceIsRegistered", func(t *testing.T) {
		assert.True(t, newFacesResetContext(t, "-t").Bool("trace"))
	})
	t.Run("UnknownDetector", func(t *testing.T) {
		err := facesResetAction(newFacesResetContext(t, "--detector=unknown"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported face detector")

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
	})
	t.Run("ForceWithAll", func(t *testing.T) {
		err := facesResetAction(newFacesResetContext(t, "--force", "--all"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--all")

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
	})
	t.Run("ForceWithDetector", func(t *testing.T) {
		err := facesResetAction(newFacesResetContext(t, "--force", "--detector=yunet"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--detector")

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
	})
}

func TestFacesResetDetector(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		assert.Equal(t, "", facesResetDetector(newFacesResetContext(t)))
	})
	t.Run("Detector", func(t *testing.T) {
		assert.Equal(t, "yunet", facesResetDetector(newFacesResetContext(t, "--detector= yunet ")))
	})
	t.Run("DeprecatedEngine", func(t *testing.T) {
		assert.Equal(t, "auto", facesResetDetector(newFacesResetContext(t, "--engine=onnx")))
	})
	t.Run("DeprecatedEngineNone", func(t *testing.T) {
		assert.Equal(t, "", facesResetDetector(newFacesResetContext(t, "--engine=none")))
	})
	t.Run("DetectorWinsOverEngine", func(t *testing.T) {
		assert.Equal(t, "none", facesResetDetector(newFacesResetContext(t, "--detector=none", "--engine=onnx")))
	})
}

func TestFacesResetDetectorName(t *testing.T) {
	// The package's config, since a second one would replace the database the other tests share.
	c := get.Config()
	require.NotNil(t, c)

	detector := c.Options().FaceDetector
	t.Cleanup(func() { c.Options().FaceDetector = detector })
	c.Options().FaceDetector = face.DetectorYuNet

	t.Run("Auto", func(t *testing.T) {
		assert.Equal(t, "yunet", facesResetDetectorName(c, "auto"))
	})
	t.Run("Named", func(t *testing.T) {
		assert.Equal(t, "scrfd", facesResetDetectorName(c, "scrfd"))
	})
	t.Run("None", func(t *testing.T) {
		assert.Equal(t, "", facesResetDetectorName(c, ""))
		assert.Equal(t, "none", facesResetDetectorName(c, "none"))
	})
	t.Run("NoConfig", func(t *testing.T) {
		assert.Equal(t, "auto", facesResetDetectorName(nil, "auto"))
	})
}

func TestFacesResetLabel(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "Remove automatically recognized faces, matches, and people left without faces?", facesResetLabel(false, ""))
	})
	t.Run("All", func(t *testing.T) {
		assert.Equal(t, "Remove all faces and matches, including names and unverified people, keeping the markers?", facesResetLabel(true, ""))
	})
	t.Run("Detector", func(t *testing.T) {
		assert.Equal(t, "Remove all faces and automatic matches, then detect faces in all pictures with yunet "+
			"and remove the unnamed markers it does not find again?", facesResetLabel(false, "yunet"))
	})
	t.Run("AllWithDetector", func(t *testing.T) {
		assert.Equal(t, "Remove all faces, matches, names, and unverified people, then detect faces in all pictures with yunet "+
			"and remove the unnamed markers it does not find again?", facesResetLabel(true, "yunet"))
	})
	t.Run("DetectorNone", func(t *testing.T) {
		assert.Equal(t, facesResetLabel(true, ""), facesResetLabel(true, "none"))
		assert.Equal(t, facesResetLabel(false, ""), facesResetLabel(false, "NONE"))
	})
	t.Run("DetectorAuto", func(t *testing.T) {
		assert.Equal(t, "Remove all faces and automatic matches, then detect faces in all pictures with the configured detector "+
			"and remove the unnamed markers it does not find again?", facesResetLabel(false, "auto"))
	})
}

func TestFacesResetDescription(t *testing.T) {
	for _, flag := range []string{"--all", "--force", "--detector", "faces update", "faces index"} {
		assert.Contains(t, FacesResetDescription, flag)
	}
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
	t.Run("AnsweredYes", func(t *testing.T) {
		pipeResetAnswers(t, "y\n")

		proceed, err := ConfirmAction(false, "Remove everything?")

		assert.NoError(t, err)
		assert.True(t, proceed)
	})
	t.Run("AnsweredNo", func(t *testing.T) {
		pipeResetAnswers(t, "n\n")

		proceed, err := ConfirmAction(false, "Remove everything?")

		assert.NoError(t, err)
		assert.False(t, proceed)
	})
	t.Run("EndOfInputOnTerminalDeclines", func(t *testing.T) {
		// Ctrl-D on a terminal ends the input, which declines like "n".
		pipeResetAnswers(t, "")
		prev := confirmTerminal
		confirmTerminal = func() bool { return true }
		t.Cleanup(func() { confirmTerminal = prev })

		proceed, err := ConfirmAction(false, "Remove everything?")

		assert.NoError(t, err)
		assert.False(t, proceed)
	})
	t.Run("EndOfInputWithoutTerminalIsAnError", func(t *testing.T) {
		pipeResetAnswers(t, "")

		proceed, err := ConfirmAction(false, "Remove everything?")

		require.Error(t, err)
		assert.False(t, proceed)
		assert.Contains(t, err.Error(), "no terminal")
		assert.NotContains(t, err.Error(), "^D")

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

// TestConfirmLabel covers the label cleanup, since the prompt appends its own question mark.
func TestConfirmLabel(t *testing.T) {
	t.Run("TrailingQuestionMark", func(t *testing.T) {
		assert.Equal(t, "Delete user alice", confirmLabel("Delete user alice?"))
	})
	t.Run("SeveralQuestionMarks", func(t *testing.T) {
		assert.Equal(t, "Delete user alice", confirmLabel("Delete user alice??"))
	})
	t.Run("SpaceBeforeQuestionMark", func(t *testing.T) {
		assert.Equal(t, "Delete user alice", confirmLabel("Delete user alice ?"))
	})
	t.Run("NoQuestionMark", func(t *testing.T) {
		assert.Equal(t, "Delete user alice", confirmLabel(" Delete user alice "))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", confirmLabel(""))
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
		// Parsing the flag is not the same as the action reading it. The config is stubbed to
		// fail, so the run stops right after the confirmation: the sentinel proves --yes
		// satisfied it, and nothing reaches the database.
		sentinel := errors.New("config unavailable")

		restore := InitConfig
		InitConfig = func(*cli.Context) (*config.Config, error) { return nil, sentinel }
		t.Cleanup(func() { InitConfig = restore })

		assert.ErrorIs(t, facesResetAction(newFacesResetContext(t, "--yes")), sentinel)
		assert.ErrorIs(t, facesResetAction(newFacesResetContext(t, "--yes", "--all")), sentinel)
		assert.ErrorIs(t, facesResetAction(newFacesResetContext(t, "--yes", "--force")), sentinel)
		assert.ErrorIs(t, facesResetAction(newFacesResetContext(t, "--yes", "--detector=auto")), sentinel)
	})
	t.Run("NoTerminalBeforeConfig", func(t *testing.T) {
		// The prompt comes first, so a run that cannot be confirmed never loads the config.
		restore := InitConfig
		InitConfig = func(*cli.Context) (*config.Config, error) {
			t.Fatal("the config must not be loaded before the confirmation")
			return nil, nil
		}
		t.Cleanup(func() { InitConfig = restore })

		err := facesResetAction(newFacesResetContext(t, "--detector=auto"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--yes")
	})
	t.Run("AutoWithoutDetector", func(t *testing.T) {
		c := get.Config()
		detector := c.Options().FaceDetector
		c.Options().FaceDetector = face.DetectorNone

		restore := InitCoreConfig
		InitCoreConfig = func(*cli.Context, bool) (*config.Config, error) { return c, nil }

		t.Cleanup(func() {
			InitCoreConfig = restore
			c.Options().FaceDetector = detector
		})

		err := facesResetAction(newFacesResetContext(t, "--yes", "--detector=auto"))

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
		assert.Contains(t, err.Error(), "no face detector can be used")
	})
}
