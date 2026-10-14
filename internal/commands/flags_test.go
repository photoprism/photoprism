package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
)

func TestLocalStringFlag(t *testing.T) {
	t.Run("IsLocalFlag", func(t *testing.T) {
		var f cli.Flag = &LocalStringFlag{&cli.StringFlag{Name: "advertise-url"}}
		_, ok := f.(config.LocalFlag)
		assert.True(t, ok)
		_, ok = cli.Flag(&cli.StringFlag{Name: "advertise-url"}).(config.LocalFlag)
		assert.False(t, ok)
	})
	t.Run("Unwrap", func(t *testing.T) {
		inner := &cli.StringFlag{Name: "advertise-url"}
		assert.Same(t, inner, (&LocalStringFlag{inner}).Unwrap())
	})
	t.Run("NodesModAdvertiseUrl", func(t *testing.T) {
		// The URL of the node being edited must not become the Portal's own advertise URL.
		_, ok := cli.Flag(nodesModInternal).(config.LocalFlag)
		assert.True(t, ok)
		assert.Contains(t, ClusterNodesModCommand.Flags, cli.Flag(nodesModInternal))
	})
	t.Run("RegisterAppName", func(t *testing.T) {
		// The app name reported to the Portal must not become the name of the web app.
		_, ok := cli.Flag(regAppNameFlag).(config.LocalFlag)
		assert.True(t, ok)
		assert.Contains(t, ClusterRegisterCommand.Flags, cli.Flag(regAppNameFlag))
	})
	t.Run("Value", func(t *testing.T) {
		// Runs a command, which resolves the alias the way the real one does.
		for _, args := range [][]string{{"-i", "https://node.example.com/"}, {"--advertise-url", "https://node.example.com/"}} {
			var value string
			app := cli.NewApp()
			app.Commands = []*cli.Command{{Name: "mod", Flags: []cli.Flag{nodesModInternal}, Action: func(ctx *cli.Context) error {
				value = ctx.String("advertise-url")
				return nil
			}}}
			require.NoError(t, app.Run(append([]string{"photoprism", "mod"}, args...)))
			assert.Equal(t, "https://node.example.com/", value, args[0])
		}
	})
}

func TestDeprecatedForceFlag(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		f := DeprecatedForceFlag()

		assert.Equal(t, "force", f.Name)
		assert.Equal(t, []string{"f"}, f.Aliases)
		assert.True(t, f.Hidden, "the alias must not be advertised next to --yes")
		assert.Contains(t, f.Usage, "--yes")
	})
	t.Run("DistinctFromForceFlag", func(t *testing.T) {
		assert.False(t, ForceFlag("").Hidden, "the shared --force flag stays visible")
	})
}

func TestUserRestoreFlag(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		f := UserRestoreFlag()

		assert.Equal(t, "restore", f.Name)
		assert.Empty(t, f.Aliases)
		assert.Equal(t, UserRestoreUsage, f.Usage)
	})
}
