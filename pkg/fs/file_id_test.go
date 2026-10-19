package fs

import (
	"os"
	"path/filepath"
	"syscall"
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

	// Times may differ by up to two seconds. The file infos are made up, as a file system may keep whole seconds only.
	stub(1, 2, 42, 42, true)
	second := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	info := func(name string, ms int) os.FileInfo {
		return testFileInfo{name: name, size: 4, modTime: second.Add(time.Duration(ms) * time.Millisecond)}
	}
	assert.True(t, MaybeSameFile(info("a.jpg", 100), info("b.jpg", 900)), "other precision")
	assert.True(t, MaybeSameFile(info("a.jpg", 900), info("b.jpg", 1000)), "rounded")
	assert.True(t, MaybeSameFile(info("a.jpg", 2000), info("b.jpg", 0)), "two seconds")
	assert.True(t, MaybeSameFile(info("a.jpg", 0), info("b.jpg", 2000)), "two seconds")
	assert.False(t, MaybeSameFile(info("a.jpg", 2001), info("b.jpg", 0)), "more than two seconds")
	assert.False(t, MaybeSameFile(info("a.jpg", 0), info("b.jpg", 2001)), "more than two seconds")

	require.NoError(t, os.Chtimes(b, modTime.Add(3*time.Second), modTime.Add(3*time.Second)))
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

	dev, ino, ok := FileID(info)
	assert.True(t, ok)
	assert.NotZero(t, ino)

	st, isStat := info.Sys().(*syscall.Stat_t)
	require.True(t, isStat)
	assert.Equal(t, uint64(st.Dev), dev) //nolint:gosec,unconvert // Dev is int32 on some platforms.
	assert.Equal(t, st.Ino, ino)

	// Another file in the same folder is on the same device, with another inode number.
	other := filepath.Join(filepath.Dir(name), "b.jpg")
	require.NoError(t, os.WriteFile(other, []byte("b"), ModeFile))
	otherInfo, err := os.Stat(other)
	require.NoError(t, err)
	otherDev, otherIno, ok := FileID(otherInfo)
	assert.True(t, ok)
	assert.Equal(t, dev, otherDev)
	assert.NotEqual(t, ino, otherIno)

	_, _, ok = FileID(nil)
	assert.False(t, ok)
}

// testFileInfo is a file info with the given name, size and modification time.
type testFileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (i testFileInfo) Name() string       { return i.name }
func (i testFileInfo) Size() int64        { return i.size }
func (i testFileInfo) Mode() os.FileMode  { return ModeFile }
func (i testFileInfo) ModTime() time.Time { return i.modTime }
func (i testFileInfo) IsDir() bool        { return false }
func (i testFileInfo) Sys() any           { return nil }
