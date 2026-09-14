package photoprism

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/proc"
)

func TestNewConvertBudget(t *testing.T) {
	t.Run("Limited", func(t *testing.T) {
		b := NewConvertBudget(time.Minute)
		assert.Equal(t, time.Minute, b.Remaining())
		assert.False(t, b.Exhausted())
	})
	t.Run("Unlimited", func(t *testing.T) {
		b := NewConvertBudget(0)
		assert.Equal(t, time.Duration(0), b.Remaining())
		assert.False(t, b.Exhausted(), "an unlimited budget is never exhausted")
	})
	t.Run("Nil", func(t *testing.T) {
		var b *ConvertBudget
		assert.Equal(t, time.Duration(0), b.Remaining())
		assert.False(t, b.Exhausted())
	})
}

func TestConvertBudget_Run(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		b := NewConvertBudget(time.Minute)
		assert.NoError(t, b.Run(exec.Command("/bin/sh", "-c", "true"), nil))
	})
	t.Run("ChargesWhatItUsed", func(t *testing.T) {
		b := NewConvertBudget(2 * time.Second)
		before := b.Remaining()
		require.NoError(t, b.Run(exec.Command("/bin/sh", "-c", "sleep 0.3"), nil))
		assert.Less(t, b.Remaining(), before, "a completed command must reduce the budget")
	})
	t.Run("SharedAcrossCommands", func(t *testing.T) {
		// The budget belongs to the file, not to the command, so a chain of candidates cannot
		// occupy the converter for the configured time once per candidate.
		b := NewConvertBudget(700 * time.Millisecond)
		assert.True(t, errors.Is(b.Run(exec.Command("/bin/sh", "-c", "sleep 30"), nil), proc.ErrTimeout))
		assert.True(t, b.Exhausted())
		err := b.Run(exec.Command("/bin/sh", "-c", "sleep 30"), nil)
		assert.True(t, errors.Is(err, proc.ErrTimeout), "the next command must not get a fresh budget")
	})
	t.Run("UnlimitedRunsWithoutDeadline", func(t *testing.T) {
		b := NewConvertBudget(0)
		assert.NoError(t, b.Run(exec.Command("/bin/sh", "-c", "sleep 0.2"), nil))
		assert.False(t, b.Exhausted())
	})
	t.Run("KeepsTheDeadlineWhenStderrIsPresent", func(t *testing.T) {
		// A stopped command has usually written to stderr too, so the diagnostic is added to
		// the deadline rather than replacing it.
		b := NewConvertBudget(300 * time.Millisecond)
		err := b.Run(exec.Command("/bin/sh", "-c", "echo broken 1>&2; sleep 30"),
			func() string { return "broken" })
		assert.True(t, errors.Is(err, proc.ErrTimeout))
		assert.Contains(t, err.Error(), "broken")
	})
	t.Run("ReportsExitStatus", func(t *testing.T) {
		b := NewConvertBudget(time.Minute)
		err := b.Run(exec.Command("/bin/sh", "-c", "exit 5"), func() string { return "" })
		assert.Error(t, err)
		assert.False(t, errors.Is(err, proc.ErrTimeout))
	})
}

func TestRemoveConvertOutput(t *testing.T) {
	t.Run("RemovesPartialOutput", func(t *testing.T) {
		name := filepath.Join(t.TempDir(), "partial.jpg")
		require.NoError(t, os.WriteFile(name, []byte("truncated"), 0o600))
		RemoveConvertOutput(name, nil)
		assert.NoFileExists(t, name)
	})
	t.Run("Missing", func(t *testing.T) {
		assert.NotPanics(t, func() { RemoveConvertOutput(filepath.Join(t.TempDir(), "none.jpg"), nil) })
	})
	t.Run("Empty", func(t *testing.T) {
		assert.NotPanics(t, func() { RemoveConvertOutput("", nil) })
	})
}

func TestLogConvertError(t *testing.T) {
	t.Run("Timeout", func(t *testing.T) {
		assert.NotPanics(t, func() {
			LogConvertError(proc.ErrTimeout, exec.Command("/bin/sh", "-c", "true"), "example.jpg")
		})
	})
	t.Run("Other", func(t *testing.T) {
		assert.NotPanics(t, func() {
			LogConvertError(errors.New("broken"), exec.Command("/bin/sh", "-c", "true"), "example.jpg")
		})
	})
	t.Run("NoError", func(t *testing.T) {
		assert.NotPanics(t, func() { LogConvertError(nil, nil, "example.jpg") })
	})
	t.Run("NoCommand", func(t *testing.T) {
		assert.NotPanics(t, func() { LogConvertError(proc.ErrTimeout, nil, "example.jpg") })
	})
}
