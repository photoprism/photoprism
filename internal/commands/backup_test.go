package commands

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

// newBackupContext parses args against the backup command's own flags. The command is attached
// because Context.IsSet resolves a name through Command.Flags. Pass the long form: urfave copies
// an alias onto its other names in normalizeFlags during a real run, which this set does not do,
// so "-r 2" would read as set with a zero value here.
func newBackupContext(t *testing.T, args ...string) *cli.Context {
	t.Helper()

	set := flag.NewFlagSet("backup", flag.ContinueOnError)

	for _, f := range BackupCommand.Flags {
		if err := f.Apply(set); err != nil {
			t.Fatal(err)
		}
	}

	if err := set.Parse(args); err != nil {
		t.Fatal(err)
	}

	ctx := cli.NewContext(cli.NewApp(), set, nil)
	ctx.Command = BackupCommand

	return ctx
}

func TestBackupRetain(t *testing.T) {
	t.Run("UnsetTakesConfiguredValue", func(t *testing.T) {
		assert.Equal(t, 7, backupRetain(newBackupContext(t), 7))
	})
	t.Run("UnsetTakesConfiguredKeepAll", func(t *testing.T) {
		assert.Equal(t, -1, backupRetain(newBackupContext(t), -1))
	})
	t.Run("FlagOverridesConfiguredValue", func(t *testing.T) {
		assert.Equal(t, 2, backupRetain(newBackupContext(t, "--retain", "2"), 7))
	})
	t.Run("FlagKeepAllOverridesConfiguredValue", func(t *testing.T) {
		assert.Equal(t, -1, backupRetain(newBackupContext(t, "--retain", "-1"), 3))
	})
	t.Run("FlagZeroOverridesConfiguredValue", func(t *testing.T) {
		assert.Equal(t, 0, backupRetain(newBackupContext(t, "--retain", "0"), 3))
	})
}
