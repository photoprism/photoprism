package duf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupMounts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Get slice of mounted file systems.
		m, warnings, err := Mounts()

		// No warnings or errors are expected.
		assert.NoError(t, err)
		assert.Empty(t, warnings)

		filters := FilterOptions{
			HiddenDevices:     parseCommaSeparatedValues(hideDevices),
			OnlyDevices:       parseCommaSeparatedValues(onlyDevices),
			HiddenFilesystems: parseCommaSeparatedValues(hideFs),
			OnlyFilesystems:   parseCommaSeparatedValues(onlyFs),
			HiddenMountPoints: parseCommaSeparatedValues(hideMp),
			OnlyMountPoints:   parseCommaSeparatedValues(onlyMp),
		}

		results := GroupMounts(m, filters)

		t.Logf("results, %#v", results)

		// At least one mount returned?
		if len(results) < 1 {
			t.Error("at least one result expected")
		} else if local, found := results[LocalDevice]; found {
			for _, d := range local {
				if d.Total <= 0 {
					t.Error("total should be a positive integer")
				} else {
					t.Logf("%s is mounted at %s: %d of %d GB used (%.1f%%)", d.Device, d.Mountpoint, d.Used/GB, d.Total/GB, (float64(d.Used)/float64(d.Total))*100)
				}
			}
		} else {
			t.Error("no local devices found")
		}
	})
}
