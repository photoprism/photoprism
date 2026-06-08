package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCliFlag_Skip(t *testing.T) {
	withTags := CliFlag{
		Flag: &cli.StringFlag{
			Name:    "with-tags",
			Usage:   "`STRING`",
			EnvVars: EnvVars("WITH_TAGS"),
		},
		Tags: []string{"foo", "bar"},
	}

	noTags := CliFlag{
		Flag: &cli.StringFlag{
			Name:    "no-tags",
			Usage:   "`STRING`",
			EnvVars: EnvVars("NO_TAGS"),
		},
		Tags: []string{},
	}

	t.Run("True", func(t *testing.T) {
		assert.True(t, withTags.Skip([]string{"baz"}))
		assert.False(t, noTags.Skip([]string{"baz"}))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, withTags.Skip([]string{"foo"}))
		assert.False(t, noTags.Skip([]string{"foo"}))
	})
}

func TestCliFlag_EnvVars(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		testFlag := CliFlag{
			Flag: &cli.StringFlag{
				Name:    "test",
				Usage:   "`STRING`",
				EnvVars: nil,
			},
			Tags: []string{"foo", "bar"},
		}

		assert.Equal(t, "test", testFlag.Name())
		assert.Equal(t, []string{"test"}, testFlag.Names())
		assert.Equal(t, "test", testFlag.String())
		assert.Equal(t, "", testFlag.EnvVar())
		assert.Equal(t, []string{}, testFlag.EnvVars())
	})
	t.Run("One", func(t *testing.T) {
		testFlag := CliFlag{
			Flag: &cli.StringFlag{
				Name:    "test",
				Usage:   "`STRING`",
				EnvVars: EnvVars("BAR_BAZ"),
			},
			Tags: []string{"foo", "bar"},
		}

		assert.Equal(t, "test", testFlag.Name())
		assert.Equal(t, []string{"test"}, testFlag.Names())
		assert.Equal(t, "test", testFlag.String())
		assert.Equal(t, "PHOTOPRISM_BAR_BAZ", testFlag.EnvVar())
		assert.Equal(t, []string{"PHOTOPRISM_BAR_BAZ"}, testFlag.EnvVars())
	})
	t.Run("Multiple", func(t *testing.T) {
		testFlag := CliFlag{
			Flag: &cli.StringFlag{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "`STRING`",
				EnvVars: EnvVars("FOO_1", "ORIGINALS_PATH"),
			},
			Tags: []string{"foo", "bar"},
		}

		assert.Equal(t, "test", testFlag.Name())
		assert.Equal(t, []string{"test", "t"}, testFlag.Names())
		assert.Equal(t, "test, t", testFlag.String())
		assert.Equal(t, "PHOTOPRISM_FOO_1, PHOTOPRISM_ORIGINALS_PATH", testFlag.EnvVar())
		assert.Equal(t, []string{"PHOTOPRISM_FOO_1", "PHOTOPRISM_ORIGINALS_PATH"}, testFlag.EnvVars())
	})
}

func TestCliFlag_Hidden(t *testing.T) {
	hidden := CliFlag{
		Flag: &cli.StringFlag{
			Name:    "is-hidden",
			Usage:   "`STRING`",
			EnvVars: []string{"PHOTOPRISM_HIDDEN"},
			Hidden:  true,
		},
		Tags: []string{"foo", "bar"},
	}

	visible := CliFlag{
		Flag: &cli.StringFlag{
			Name:    "is-visible",
			Usage:   "`STRING`",
			EnvVars: []string{"PHOTOPRISM_VISIBLE"},
			Hidden:  false,
		},
		Tags: []string{},
	}

	t.Run("True", func(t *testing.T) {
		assert.True(t, hidden.Hidden())
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, visible.Hidden())
	})
}

func TestCliFlag_Default(t *testing.T) {
	hasdefault := CliFlag{
		Flag: &cli.StringFlag{
			Name:    "flag-with-default",
			Usage:   "`STRING`",
			EnvVars: []string{"PHOTOPRISM_DEFAULT"},
		},
		DocDefault: "default-value",
		Tags:       []string{"foo", "bar"},
	}

	nodefault := CliFlag{
		Flag: &cli.StringFlag{
			Name:    "flag-without-default",
			Usage:   "`STRING`",
			EnvVars: []string{"PHOTOPRISM_NODEFAULT"},
		},
		Tags: []string{},
	}

	assert.Equal(t, "default-value", hasdefault.Default())
	assert.Equal(t, "", nodefault.Default())
}

// TestCliFlag_DefaultStringDefault verifies that the literal source-code
// default of a StringFlag is returned with quotes stripped, not the %q
// representation urfave/cli uses for documentation output.
func TestCliFlag_DefaultStringDefault(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		f := CliFlag{
			Flag: &cli.StringFlag{
				Name:    "flag-empty-default",
				EnvVars: []string{"PHOTOPRISM_TEST_EMPTY"},
			},
		}

		assert.Equal(t, "", f.Default())
	})
	t.Run("Literal", func(t *testing.T) {
		f := CliFlag{
			Flag: &cli.StringFlag{
				Name:    "flag-literal-default",
				EnvVars: []string{"PHOTOPRISM_TEST_LITERAL"},
				Value:   "hello",
			},
		}

		assert.Equal(t, "hello", f.Default())
	})
}

// TestCliFlag_DefaultIgnoresEnvOverride verifies that an env-supplied
// runtime value never replaces the documented default.
func TestCliFlag_DefaultIgnoresEnvOverride(t *testing.T) {
	t.Setenv("PHOTOPRISM_TEST_OVERRIDE", "secret-runtime-value")

	flag := &cli.StringFlag{
		Name:    "flag-env-override",
		EnvVars: []string{"PHOTOPRISM_TEST_OVERRIDE"},
		Value:   "documented-default",
	}

	app := cli.NewApp()
	app.Flags = []cli.Flag{flag}
	app.Action = func(*cli.Context) error { return nil }
	require.NoError(t, app.Run([]string{"app"}))

	wrapped := CliFlag{Flag: flag}

	assert.Equal(t, "secret-runtime-value", flag.GetValue(), "sanity: urfave/cli should expose the env value via GetValue")
	assert.Equal(t, "documented-default", wrapped.Default(), "Default() must keep returning the documented default after env override")
}

// TestCliFlag_DefaultSecret verifies that flags carrying secret data
// never echo the runtime value (or the source-code default) through
// Default(). Operators can attach a placeholder via DocDefault when one
// is desirable; otherwise Default() returns the empty string.
func TestCliFlag_DefaultSecret(t *testing.T) {
	t.Setenv("PHOTOPRISM_TEST_SECRET", "hunter2-runtime-secret")

	flag := &cli.StringFlag{
		Name:    "flag-secret",
		EnvVars: []string{"PHOTOPRISM_TEST_SECRET"},
		Value:   "compiled-default",
	}

	app := cli.NewApp()
	app.Flags = []cli.Flag{flag}
	app.Action = func(*cli.Context) error { return nil }
	require.NoError(t, app.Run([]string{"app"}))

	t.Run("EmptyPlaceholder", func(t *testing.T) {
		f := CliFlag{Flag: flag, Secret: true}

		assert.Equal(t, "", f.Default())
	})
	t.Run("DocDefaultPlaceholder", func(t *testing.T) {
		f := CliFlag{Flag: flag, Secret: true, DocDefault: "[redacted]"}

		assert.Equal(t, "[redacted]", f.Default())
	})
}

func TestCliFlag_EnvVar(t *testing.T) {
	hasDefault := CliFlag{
		Flag: &cli.StringFlag{
			Name:    "flag-with-default",
			Usage:   "`STRING`",
			EnvVars: []string{"PHOTOPRISM_DEFAULT"},
		},
		DocDefault: "default-value",
		Tags:       []string{"foo", "bar"},
	}

	assert.Equal(t, "PHOTOPRISM_DEFAULT", hasDefault.EnvVar())
}

func TestCliFlag_CommandFlag(t *testing.T) {
	hasdefault := CliFlag{
		Flag: &cli.StringFlag{
			Name:    "flag-with-default",
			Usage:   "`STRING`",
			EnvVars: []string{"PHOTOPRISM_DEFAULT"},
		},
		DocDefault: "default-value",
		Tags:       []string{"foo", "bar"},
	}

	assert.Equal(t, "--flag-with-default", hasdefault.CommandFlag())
}

func TestCliFlag_Usage(t *testing.T) {
	community := CliFlag{
		Flag: &cli.StringFlag{
			Name:  "flag-community",
			Usage: "`STRING`",
		},
		DocDefault: "default-value",
		Tags:       []string{"foo", "bar"},
	}

	essentials := CliFlag{
		Flag: &cli.StringFlag{
			Name:  "flag-essentials",
			Usage: "`STRING`",
		},
		Tags: []string{"essentials"},
	}

	plus := CliFlag{
		Flag: &cli.StringFlag{
			Name:  "flag-plus",
			Usage: "`STRING`",
		},
		Tags: []string{"plus"},
	}

	pro := CliFlag{
		Flag: &cli.StringFlag{
			Name:  "flag-pro",
			Usage: "`STRING`",
		},
		Tags: []string{"pro"},
	}

	assert.Contains(t, community.Usage(), "STRING")
	assert.Contains(t, essentials.Usage(), "*essentials*")
	assert.Contains(t, plus.Usage(), "*plus*")
	assert.Contains(t, pro.Usage(), "*pro*")
}
