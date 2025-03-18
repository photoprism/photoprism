//go:build linux

package duf

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinuxMounts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Get slice of mounted file systems.
		results, warnings, err := mounts()

		// No warnings or errors are expected.
		assert.NoError(t, err)
		assert.Empty(t, warnings)

		// At least one mount returned?
		if len(results) < 1 {
			t.Error("at least one result expected")
		} else {
			// If so, check the first mount for plausibility.
			result := results[0]
			assert.NotEmpty(t, result.Device)
			assert.Equal(t, "local", result.DeviceType)
			assert.Equal(t, "/", result.Mountpoint)
			assert.NotEmpty(t, result.Fstype)
			assert.NotEmpty(t, result.Opts)
			assert.NotEmpty(t, result.Total)
			assert.NotEmpty(t, result.Used)
			assert.NotEmpty(t, result.Free)
			assert.NotEmpty(t, result.Inodes)
			assert.NotEmpty(t, result.InodesFree)
			assert.NotEmpty(t, result.InodesUsed)
			assert.NotEmpty(t, result.Blocks)
			assert.NotEmpty(t, result.BlockSize)
			assert.NotEmpty(t, result.Metadata)
		}
	})
}

func TestGetFields(t *testing.T) {
	var tt = []struct {
		input    string
		number   int
		expected [11]string
	}{
		// Empty lines
		{
			input:  "",
			number: 0,
		},
		{
			input:  " ",
			number: 0,
		},
		{
			input:  "   ",
			number: 0,
		},
		{
			input:  "	",
			number: 0,
		},

		// Comments
		{
			input:  "#",
			number: 0,
		},
		{
			input:  "# ",
			number: 0,
		},
		{
			input:  "#	",
			number: 0,
		},
		{
			input:  "# I'm a lazy dog",
			number: 0,
		},

		// Bad fields
		{
			input:    "1 2",
			number:   2,
			expected: [11]string{"1", "2"},
		},
		{
			input:    "1	2",
			number:   2,
			expected: [11]string{"1", "2"},
		},
		{
			input:    "1	2		3",
			number:   3,
			expected: [11]string{"1", "2", "3"},
		},
		{
			input:    "1	2		3   4",
			number:   4,
			expected: [11]string{"1", "2", "3", "4"},
		},

		// No optional separator or no options
		{
			input:    "1 2 3 4 5 6 7 NotASeparator 9 10 11",
			number:   6,
			expected: [11]string{"1", "2", "3", "4", "5", "6", "7 NotASeparator 9 10 11"},
		},
		{
			input:    "1 2 3 4 5 6 7 8 9 10 11",
			number:   6,
			expected: [11]string{"1", "2", "3", "4", "5", "6", "7 8 9 10 11"},
		},
		{
			input:    "1 2 3 4 5 6 - 9 10 11",
			number:   11,
			expected: [11]string{"1", "2", "3", "4", "5", "6", "", "-", "9", "10", "11"},
		},

		// Normal mount table line
		{
			input:    "22 27 0:21 / /proc rw,nosuid,nodev,noexec,relatime shared:5 - proc proc rw",
			number:   11,
			expected: [11]string{"22", "27", "0:21", "/", "/proc", "rw,nosuid,nodev,noexec,relatime", "shared:5", "-", "proc", "proc", "rw"},
		},
		{
			input:    "31 23 0:27 / /sys/fs/cgroup rw,nosuid,nodev,noexec,relatime shared:9 - cgroup2 cgroup2 rw,nsdelegate,memory_recursiveprot",
			number:   11,
			expected: [11]string{"31", "23", "0:27", "/", "/sys/fs/cgroup", "rw,nosuid,nodev,noexec,relatime", "shared:9", "-", "cgroup2", "cgroup2", "rw,nsdelegate,memory_recursiveprot"},
		},
		{
			input:    "40 27 0:33 / /tmp rw,nosuid,nodev shared:18 - tmpfs tmpfs",
			number:   10,
			expected: [11]string{"40", "27", "0:33", "/", "/tmp", "rw,nosuid,nodev", "shared:18", "-", "tmpfs", "tmpfs"},
		},
		{
			input:    "40 27 0:33 / /tmp rw,nosuid,nodev shared:18 shared:22 - tmpfs tmpfs",
			number:   10,
			expected: [11]string{"40", "27", "0:33", "/", "/tmp", "rw,nosuid,nodev", "shared:18 shared:22", "-", "tmpfs", "tmpfs"},
		},
		{
			input:    "50 27 0:33 / /tmp rw,nosuid,nodev - tmpfs tmpfs",
			number:   10,
			expected: [11]string{"50", "27", "0:33", "/", "/tmp", "rw,nosuid,nodev", "", "-", "tmpfs", "tmpfs"},
		},

		// Exceptional mount table lines
		{
			input:    "328 27 0:73 / /mnt/a rw,relatime shared:206 - tmpfs - rw,inode64",
			number:   11,
			expected: [11]string{"328", "27", "0:73", "/", "/mnt/a", "rw,relatime", "shared:206", "-", "tmpfs", "-", "rw,inode64"},
		},
		{
			input:    "330 27 0:73 / /mnt/a rw,relatime shared:206 - tmpfs 👾 rw,inode64",
			number:   11,
			expected: [11]string{"330", "27", "0:73", "/", "/mnt/a", "rw,relatime", "shared:206", "-", "tmpfs", "👾", "rw,inode64"},
		},
		{
			input:    "335 27 0:73 / /mnt/👾 rw,relatime shared:206 - tmpfs 👾 rw,inode64",
			number:   11,
			expected: [11]string{"335", "27", "0:73", "/", "/mnt/👾", "rw,relatime", "shared:206", "-", "tmpfs", "👾", "rw,inode64"},
		},
		{
			input:    "509 27 0:78 / /mnt/- rw,relatime shared:223 - tmpfs 👾 rw,inode64",
			number:   11,
			expected: [11]string{"509", "27", "0:78", "/", "/mnt/-", "rw,relatime", "shared:223", "-", "tmpfs", "👾", "rw,inode64"},
		},
		{
			input:    "362 27 0:76 / /mnt/a\\040b rw,relatime shared:215 - tmpfs 👾 rw,inode64",
			number:   11,
			expected: [11]string{"362", "27", "0:76", "/", "/mnt/a b", "rw,relatime", "shared:215", "-", "tmpfs", "👾", "rw,inode64"},
		},
		{
			input:    "1 2 3:3 / /mnt/\\011 rw shared:7 - tmpfs - rw,inode64",
			number:   11,
			expected: [11]string{"1", "2", "3:3", "/", "/mnt/\t", "rw", "shared:7", "-", "tmpfs", "-", "rw,inode64"},
		},
		{
			input:    "11 2 3:3 / /mnt/a\\012b rw shared:7 - tmpfs - rw,inode64",
			number:   11,
			expected: [11]string{"11", "2", "3:3", "/", "/mnt/a\nb", "rw", "shared:7", "-", "tmpfs", "-", "rw,inode64"},
		},
		{
			input:    "111 2 3:3 / /mnt/a\\134b rw shared:7 - tmpfs - rw,inode64",
			number:   11,
			expected: [11]string{"111", "2", "3:3", "/", "/mnt/a\\b", "rw", "shared:7", "-", "tmpfs", "-", "rw,inode64"},
		},
		{
			input:    "1111 2 3:3 / /mnt/a\\042b rw shared:7 - tmpfs - rw,inode64",
			number:   11,
			expected: [11]string{"1111", "2", "3:3", "/", "/mnt/a\"b", "rw", "shared:7", "-", "tmpfs", "-", "rw,inode64"},
		},
	}

	for _, tc := range tt {
		nb, actual := parseMountInfoLine(tc.input)
		if nb != tc.number || !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("\nparseMountInfoLine(%q) == \n(%d) %q, \nexpected (%d) %q", tc.input, nb, actual, tc.number, tc.expected)
		}
	}
}
