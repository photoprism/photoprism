package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaybeSameFile(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.jpg"), filepath.Join(dir, "b.jpg")
	require.NoError(t, os.WriteFile(a, []byte("same"), ModeFile))
	require.NoError(t, os.WriteFile(b, []byte("same"), ModeFile))
	modTime := time.Now().Add(-time.Hour).Truncate(time.Second)
	require.NoError(t, os.Chtimes(a, modTime, modTime))
	require.NoError(t, os.Chtimes(b, modTime, modTime))

	stat := func(name string) os.FileInfo {
		info, err := os.Stat(name)
		require.NoError(t, err)
		return info
	}

	// stub reports the device and inode numbers of a and b.
	stub := func(aDev, bDev, aIno, bIno uint64, ok bool) {
		fileID = func(info os.FileInfo) (dev, ino uint64, found bool) {
			if info.Name() == "a.jpg" {
				return aDev, aIno, ok
			}
			return bDev, bIno, ok
		}
	}

	saved := fileID
	t.Cleanup(func() { fileID = saved })

	stub(1, 2, 42, 42, true)
	assert.True(t, MaybeSameFile(stat(a), stat(b)))
	assert.False(t, MaybeSameFile(nil, stat(b)), "nil")
	assert.False(t, MaybeSameFile(stat(a), nil), "nil")
	stub(1, 1, 42, 42, true)
	assert.False(t, MaybeSameFile(stat(a), stat(b)), "same device")
	stub(1, 2, 42, 43, true)
	assert.False(t, MaybeSameFile(stat(a), stat(b)), "other inode")
	stub(1, 2, 42, 42, false)
	assert.False(t, MaybeSameFile(stat(a), stat(b)), "no inode")

	// One side without inode information is not enough either.
	fileID = func(info os.FileInfo) (dev, ino uint64, found bool) {
		if info.Name() == "a.jpg" {
			return 1, 42, true
		}
		return 2, 42, false
	}
	assert.False(t, MaybeSameFile(stat(a), stat(b)), "no inode for b")
	assert.False(t, MaybeSameFile(stat(b), stat(a)), "no inode for a")

	// Times are compared in whole seconds.
	stub(1, 2, 42, 42, true)
	require.NoError(t, os.Chtimes(b, modTime.Add(300*time.Nanosecond), modTime.Add(300*time.Nanosecond)))
	assert.True(t, MaybeSameFile(stat(a), stat(b)), "other precision")

	stub(1, 2, 42, 42, true)
	require.NoError(t, os.Chtimes(b, modTime.Add(time.Second), modTime.Add(time.Second)))
	assert.False(t, MaybeSameFile(stat(a), stat(b)), "other time")
	require.NoError(t, os.Chtimes(b, modTime, modTime))
	require.NoError(t, os.WriteFile(b, []byte("other"), ModeFile))
	require.NoError(t, os.Chtimes(b, modTime, modTime))
	assert.False(t, MaybeSameFile(stat(a), stat(b)), "other size")
}

func TestFileID(t *testing.T) {
	name := filepath.Join(t.TempDir(), "a.jpg")
	require.NoError(t, os.WriteFile(name, []byte("a"), ModeFile))
	info, err := os.Stat(name)
	require.NoError(t, err)

	_, ino, ok := FileID(info)
	assert.True(t, ok)
	assert.NotZero(t, ino)

	_, _, ok = FileID(nil)
	assert.False(t, ok)
}
