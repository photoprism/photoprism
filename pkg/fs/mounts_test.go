package fs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testMountInfo is a mountinfo table with mounts below, at, and next to /photoprism/originals.
const testMountInfo = `22 1 8:2 / / rw,relatime shared:1 - ext4 /dev/sda2 rw
590 22 0:52 / /photoprism/originals rw,relatime - cifs //nas/photos rw,vers=3.1.1
591 590 8:17 / /photoprism/originals/usb rw,relatime - ext4 /dev/sdb1 rw
592 591 0:60 / /photoprism/originals/usb/My\040Photos rw,relatime - vfat /dev/sdc1 rw
593 22 8:18 / /photoprism/originals2 rw,relatime - ext4 /dev/sdb2 rw
594 22 0:61 / /photoprism/storage rw,relatime - ext4 /dev/sdd1 rw
595 590 8:17 /backup /photoprism/originals/usb rw,relatime - ext4 /dev/sdb1 rw
596 590 0:62 / /photoprism/originals/net rw,relatime shared:5 - autofs systemd-1 rw
597 596 0:63 / /photoprism/originals/net rw,relatime shared:6 master:1 - nfs4 nas:/photos rw
598 590 0:64 / /photoprism/originals/auto rw,relatime - autofs systemd-1 rw
invalid line
`

func TestMountPointsFrom(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := mountPointsFrom(strings.NewReader(testMountInfo), "/photoprism/originals", "/photoprism/originals")
		require.NoError(t, err)
		assert.Equal(t, []string{
			"/photoprism/originals/usb",
			"/photoprism/originals/usb/My Photos",
			"/photoprism/originals/net",
			"/photoprism/originals/auto",
		}, result)
	})
	t.Run("Link", func(t *testing.T) {
		result, err := mountPointsFrom(strings.NewReader(testMountInfo), "/originals", "/photoprism/originals")
		require.NoError(t, err)
		require.Len(t, result, 4)
		assert.Equal(t, "/originals/usb", result[0])
		assert.Equal(t, "/originals/usb/My Photos", result[1])
	})
	t.Run("None", func(t *testing.T) {
		result, err := mountPointsFrom(strings.NewReader(testMountInfo), "/photoprism/storage", "/photoprism/storage")
		require.NoError(t, err)
		assert.Empty(t, result)
	})
	t.Run("UnicodeSpace", func(t *testing.T) {
		// The kernel escapes only space, tab, newline, and backslash, so other whitespace is part of a path.
		table := "591 590 8:17 / /photoprism/originals/\u5199\u771f\u3000NAS rw,relatime - ext4 /dev/sdb1 rw\n" +
			"592 590 8:18 /srv/a\u00a0b /photoprism/originals/nas rw,relatime - ext4 /dev/sdc1 rw\n"

		result, err := mountPointsFrom(strings.NewReader(table), "/photoprism/originals", "/photoprism/originals")
		require.NoError(t, err)
		assert.Equal(t, []string{"/photoprism/originals/\u5199\u771f\u3000NAS", "/photoprism/originals/nas"}, result)
	})
	t.Run("Sibling", func(t *testing.T) {
		result, err := mountPointsFrom(strings.NewReader(testMountInfo), "/photoprism/originals/usb", "/photoprism/originals/usb")
		require.NoError(t, err)
		assert.Equal(t, []string{"/photoprism/originals/usb/My Photos"}, result)
	})
	t.Run("LongLine", func(t *testing.T) {
		// A line longer than the read buffer is parsed from its leading part, and the following lines still are.
		long := "594 22 0:61 / /photoprism/originals/overlay rw - overlay overlay rw,lowerdir=" + strings.Repeat("/l", mountInfoBuffer) + " 1 2 3 /photoprism/originals/tail x\n"
		cut := "595 22 0:62 / /photoprism/originals/" + strings.Repeat("x", mountInfoBuffer) + "\n"
		table := "593 22 8:18 / /photoprism/originals/before rw - ext4 /dev/sdb2 rw\n" + long + cut +
			"596 22 8:19 / /photoprism/originals/after rw - ext4 /dev/sdb3 rw\n"

		result, err := mountPointsFrom(strings.NewReader(table), "/photoprism/originals", "/photoprism/originals")
		require.NoError(t, err)
		assert.Equal(t, []string{"/photoprism/originals/before", "/photoprism/originals/overlay", "/photoprism/originals/after"}, result)
	})
	t.Run("NoNewline", func(t *testing.T) {
		result, err := mountPointsFrom(strings.NewReader("593 22 8:18 / /photoprism/originals/usb rw - ext4 /dev/sdb2 rw"), "/photoprism/originals", "/photoprism/originals")
		require.NoError(t, err)
		assert.Equal(t, []string{"/photoprism/originals/usb"}, result)
	})
	t.Run("Error", func(t *testing.T) {
		r := io.MultiReader(strings.NewReader("593 22 8:18 / /photoprism/originals/usb rw - ext4 /dev/sdb2 rw\n"), iotest.ErrReader(errors.New("read error")))
		result, err := mountPointsFrom(r, "/photoprism/originals", "/photoprism/originals")
		assert.EqualError(t, err, "read error")
		assert.Equal(t, []string{"/photoprism/originals/usb"}, result)
	})
	t.Run("LongLineError", func(t *testing.T) {
		// A one-time error while the rest of a long line is discarded is returned.
		line := "593 22 8:18 / /photoprism/originals/usb rw - ext4 /dev/sdb2 rw"
		line += strings.Repeat("x", mountInfoBuffer-len(line))
		r := io.MultiReader(strings.NewReader(line), &onceErrReader{err: errors.New("read error")})
		result, err := mountPointsFrom(r, "/photoprism/originals", "/photoprism/originals")
		assert.EqualError(t, err, "read error")
		assert.Equal(t, []string{"/photoprism/originals/usb"}, result)
	})
	t.Run("Empty", func(t *testing.T) {
		result, err := mountPointsFrom(strings.NewReader(""), "/photoprism/originals", "/photoprism/originals")
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

// onceErrReader returns err on the first read and io.EOF after that.
type onceErrReader struct{ err error }

// Read implements io.Reader.
func (r *onceErrReader) Read([]byte) (int, error) {
	err := r.err
	r.err = nil

	if err == nil {
		return 0, io.EOF
	}

	return 0, err
}

func TestMountPoints(t *testing.T) {
	if _, err := os.Stat(MountInfoFile); err != nil {
		t.Skip("requires a mount table")
	}

	result, err := MountPoints("")
	require.NoError(t, err)
	assert.Empty(t, result)

	result, err = MountPoints(t.TempDir())
	require.NoError(t, err)
	assert.Empty(t, result)

	// A link to the root folder lists the mounts below it as paths below the link, e.g. "<link>/proc".
	link := filepath.Join(t.TempDir(), "root")
	require.NoError(t, os.Symlink("/", link))

	result, err = MountPoints(link)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	for _, p := range result {
		assert.True(t, strings.HasPrefix(p, link+"/"), p)
	}
}

func TestUnescapeMountPath(t *testing.T) {
	assert.Equal(t, "/mnt/My Photos", unescapeMountPath(`/mnt/My\040Photos`))
	assert.Equal(t, "/mnt/a\tb\\c", unescapeMountPath(`/mnt/a\011b\134c`))
	assert.Equal(t, "/mnt/plain", unescapeMountPath("/mnt/plain"))
	assert.Equal(t, `/mnt/x\9`, unescapeMountPath(`/mnt/x\9`))
	assert.Equal(t, `/mnt/x\999`, unescapeMountPath(`/mnt/x\999`))
}
