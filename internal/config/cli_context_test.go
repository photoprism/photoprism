package config

import (
	"flag"
	"os"
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

type siteUrlTarget struct {
	SiteUrl string `flag:"site-url"`
	Trace   bool   `flag:"trace"`
}

// testLocalFlag is a string flag that does not apply to the config.
type testLocalFlag struct {
	*cli.StringFlag
}

// LocalFlag marks the flag as a LocalFlag.
func (*testLocalFlag) LocalFlag() {}

// runSubcommand runs a command whose flags repeat the names of global flags, and returns the values
// ApplyCliContext reads from its context.
func runSubcommand(t *testing.T, local cli.Flag, args ...string) siteUrlTarget {
	t.Helper()

	var target siteUrlTarget

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "site-url", Value: "http://localhost:2342/", EnvVars: []string{"PHOTOPRISM_SITE_URL"}},
		&cli.BoolFlag{Name: "trace", EnvVars: []string{"PHOTOPRISM_TRACE"}},
	}
	app.Commands = []*cli.Command{{
		Name:  "cmd",
		Flags: []cli.Flag{local, &cli.BoolFlag{Name: "trace"}},
		Action: func(ctx *cli.Context) error {
			return ApplyCliContext(&target, ctx)
		},
	}}

	assert.NoError(t, app.Run(append([]string{"photoprism"}, args...)))

	return target
}

// unsetEnv removes an environment variable for the duration of a test, which t.Setenv cannot do.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	assert.NoError(t, os.Unsetenv(key))
}

func TestApplyCliContext_CommandFlags(t *testing.T) {
	localSiteUrl := func() cli.Flag { return &cli.StringFlag{Name: "site-url"} }
	unsetEnv(t, "PHOTOPRISM_TRACE")

	t.Run("Default", func(t *testing.T) {
		unsetEnv(t, "PHOTOPRISM_SITE_URL")
		assert.Equal(t, siteUrlTarget{SiteUrl: "http://localhost:2342/"}, runSubcommand(t, localSiteUrl(), "cmd"))
	})
	t.Run("Environment", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_SITE_URL", "https://env.example.com/")
		t.Setenv("PHOTOPRISM_TRACE", "true")
		assert.Equal(t, siteUrlTarget{SiteUrl: "https://env.example.com/", Trace: true}, runSubcommand(t, localSiteUrl(), "cmd"))
	})
	t.Run("GlobalFlag", func(t *testing.T) {
		unsetEnv(t, "PHOTOPRISM_SITE_URL")
		assert.Equal(t, "https://global.example.com/", runSubcommand(t, localSiteUrl(), "--site-url", "https://global.example.com/", "cmd").SiteUrl)
	})
	t.Run("CommandFlag", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_SITE_URL", "https://env.example.com/")
		assert.Equal(t, siteUrlTarget{SiteUrl: "https://cmd.example.com/", Trace: true},
			runSubcommand(t, localSiteUrl(), "cmd", "--site-url", "https://cmd.example.com/", "--trace"))
	})
	t.Run("LocalFlag", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_SITE_URL", "https://env.example.com/")
		local := &testLocalFlag{&cli.StringFlag{Name: "site-url", Aliases: []string{"s"}}}
		assert.Equal(t, "https://env.example.com/", runSubcommand(t, local, "cmd", "-s", "https://node.example.com/").SiteUrl)
		assert.Equal(t, "https://env.example.com/", runSubcommand(t, local, "cmd").SiteUrl)
	})
}

func TestFlagContext(t *testing.T) {
	t.Run("NoApp", func(t *testing.T) {
		flags := flag.NewFlagSet("test", flag.ContinueOnError)
		flags.String("site-url", "", "doc")
		ctx := cli.NewContext(nil, flags, nil)
		assert.Same(t, ctx, flagContext(ctx, "site-url"))
	})
	t.Run("Nested", func(t *testing.T) {
		parentFlags := flag.NewFlagSet("parent", flag.ContinueOnError)
		parentFlags.String("site-url", "", "doc")
		parent := cli.NewContext(cli.NewApp(), parentFlags, nil)
		childFlags := flag.NewFlagSet("child", flag.ContinueOnError)
		childFlags.String("site-url", "", "doc")
		child := cli.NewContext(parent.App, childFlags, parent)

		assert.Same(t, parent, flagContext(child, "site-url"), "an unset command flag falls back to the outer context")
		assert.NoError(t, parent.Set("site-url", "https://global.example.com/"))
		assert.Same(t, parent, flagContext(child, "site-url"))
		assert.NoError(t, child.Set("site-url", "https://cmd.example.com/"))
		assert.Same(t, child, flagContext(child, "site-url"))
	})
	t.Run("SingleContext", func(t *testing.T) {
		flags := flag.NewFlagSet("test", flag.ContinueOnError)
		flags.String("site-url", "", "doc")
		ctx := cli.NewContext(cli.NewApp(), flags, nil)
		assert.Same(t, ctx, flagContext(ctx, "site-url"))
		assert.NoError(t, ctx.Set("site-url", "https://example.com/"))
		assert.Same(t, ctx, flagContext(ctx, "site-url"))
	})
}

func TestIsLocalFlag(t *testing.T) {
	cmd := &cli.Command{Flags: []cli.Flag{
		&testLocalFlag{&cli.StringFlag{Name: "advertise-url", Aliases: []string{"i"}}},
		&cli.StringFlag{Name: "site-url"},
	}}

	t.Run("Local", func(t *testing.T) {
		assert.True(t, isLocalFlag(cmd, "advertise-url"))
		assert.True(t, isLocalFlag(cmd, "i"))
	})
	t.Run("NotLocal", func(t *testing.T) {
		assert.False(t, isLocalFlag(cmd, "site-url"))
		assert.False(t, isLocalFlag(cmd, "portal-url"))
		assert.False(t, isLocalFlag(nil, "advertise-url"))
	})
}
