package fs

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stagedNames returns the names of the temporary files a copy may have left in a directory.
func stagedNames(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	var found []string

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") && strings.Contains(entry.Name(), ExtTmp) {
			found = append(found, entry.Name())
		}
	}

	return found
}

// sameDevice reports whether two paths are on the same file system.
func sameDevice(t *testing.T, a, b string) bool {
	t.Helper()

	var infoA, infoB syscall.Stat_t

	require.NoError(t, syscall.Stat(a, &infoA))
	require.NoError(t, syscall.Stat(b, &infoB))

	return infoA.Dev == infoB.Dev
}

func TestDestReplaceable(t *testing.T) {
	dir := t.TempDir()

	t.Run("Absent", func(t *testing.T) {
		assert.False(t, destReplaceable(filepath.Join(dir, "absent")))
	})
	t.Run("EmptyFile", func(t *testing.T) {
		name := filepath.Join(dir, "empty")
		require.NoError(t, os.WriteFile(name, nil, ModeFile))
		assert.True(t, destReplaceable(name))
	})
	t.Run("NonEmptyFile", func(t *testing.T) {
		name := filepath.Join(dir, "full")
		require.NoError(t, os.WriteFile(name, []byte("x"), ModeFile))
		assert.False(t, destReplaceable(name))
	})
	t.Run("Directory", func(t *testing.T) {
		assert.False(t, destReplaceable(dir))
	})
	t.Run("SymlinkToEmptyFile", func(t *testing.T) {
		// The name is read rather than what it resolves to, so a link is never an empty file.
		target := filepath.Join(dir, "link-target-empty")
		require.NoError(t, os.WriteFile(target, nil, ModeFile))

		name := filepath.Join(dir, "link-to-empty")
		require.NoError(t, os.Symlink(target, name))

		assert.False(t, destReplaceable(name))
	})
}

func TestDestExists(t *testing.T) {
	err := destExists("/originals/2030/05/photo.jpg")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "photo.jpg")
	assert.NotContains(t, err.Error(), "/originals", "the message names the file rather than its path")
}

func TestStageFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "photo.jpg")

	t.Run("Success", func(t *testing.T) {
		f, err := stageFile(dest)
		require.NoError(t, err)

		defer func() {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}()

		base := filepath.Base(f.Name())
		assert.True(t, strings.HasPrefix(base, ".photo.jpg."), "the staged name must be hidden: %s", base)
		assert.True(t, strings.HasSuffix(base, ".tmp.jpg"), "the staged name must carry the temporary extension: %s", base)
		assert.Equal(t, dir, filepath.Dir(f.Name()), "the staged file must be a sibling of the destination")
		assert.NoFileExists(t, dest, "staging must not create the destination")

		// A second call must not collide with the first.
		second, err := stageFile(dest)
		require.NoError(t, err)
		assert.NotEqual(t, f.Name(), second.Name())
		_ = second.Close()
		_ = os.Remove(second.Name())
	})
	t.Run("MissingDirectory", func(t *testing.T) {
		_, err := stageFile(filepath.Join(dir, "absent", "photo.jpg"))
		assert.Error(t, err)
	})
}

func TestPublishFile(t *testing.T) {
	staged := func(t *testing.T, dir, content string) string {
		t.Helper()

		f, err := stageFile(filepath.Join(dir, "photo.jpg"))
		require.NoError(t, err)
		_, err = f.WriteString(content)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		return f.Name()
	}

	t.Run("FreeName", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "photo.jpg")

		require.NoError(t, publishFile(staged(t, dir, "new"), dest, false))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "new", string(b))
		assert.Empty(t, stagedNames(t, dir), "the staged file must not be left behind")
	})
	t.Run("TakenNameWithoutForce", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "photo.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("old"), ModeFile))

		err := publishFile(staged(t, dir, "new"), dest, false)

		require.Error(t, err)
		b, readErr := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, readErr)
		assert.Equal(t, "old", string(b), "the destination must be left as it was")
	})
	t.Run("TakenNameWithForce", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "photo.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("old"), ModeFile))

		require.NoError(t, publishFile(staged(t, dir, "new"), dest, true))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "new", string(b))
		assert.Empty(t, stagedNames(t, dir))
	})
	t.Run("EmptyDestinationWithoutForce", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "photo.jpg")
		require.NoError(t, os.WriteFile(dest, nil, ModeFile))

		require.NoError(t, publishFile(staged(t, dir, "new"), dest, false))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "new", string(b))
	})
	t.Run("ExportedFreeName", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "photo.jpg")

		require.NoError(t, PublishFile(staged(t, dir, "new"), dest, false))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "new", string(b))
	})
	t.Run("ExportedTakenName", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "photo.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("old"), ModeFile))

		err := PublishFile(staged(t, dir, "new"), dest, false)

		// The sentinel is what lets a caller tell a taken name from a failure.
		assert.ErrorIs(t, err, os.ErrExist)

		b, readErr := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, readErr)
		assert.Equal(t, "old", string(b))
	})
}

// TestCopyMove_Symlinks covers the destinations an operator's library actually contains. Symbolic
// links to files and directories inside the originals folder are a supported layout, so a link is
// left to its owner: it is neither written through nor replaced, whatever the force flag says.
func TestCopyMove_Symlinks(t *testing.T) {
	t.Run("DanglingLinkIsRefused", func(t *testing.T) {
		for _, force := range []bool{false, true} {
			dir := t.TempDir()
			outside := filepath.Join(t.TempDir(), "outside.jpg")

			src := filepath.Join(dir, "src.jpg")
			require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

			dest := filepath.Join(dir, "dangling.jpg")
			require.NoError(t, os.Symlink(outside, dest))

			err := Copy(src, dest, force)

			require.Error(t, err, "force=%v", force)
			assert.Contains(t, err.Error(), "symbolic link")
			assert.NoFileExists(t, outside, "the link target must not be created")
			assert.Empty(t, stagedNames(t, dir))

			info, statErr := os.Lstat(dest)
			require.NoError(t, statErr)
			assert.NotZero(t, info.Mode()&os.ModeSymlink, "the link must be left in place")
		}
	})
	t.Run("LiveLinkIsRefusedAndItsTargetLeftAlone", func(t *testing.T) {
		for _, force := range []bool{false, true} {
			dir := t.TempDir()
			outside := filepath.Join(t.TempDir(), "outside.jpg")
			require.NoError(t, os.WriteFile(outside, []byte("target"), ModeFile))

			src := filepath.Join(dir, "src.jpg")
			require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

			dest := filepath.Join(dir, "linked.jpg")
			require.NoError(t, os.Symlink(outside, dest))

			require.Error(t, Copy(src, dest, force), "force=%v", force)

			b, err := os.ReadFile(outside) //nolint:gosec // test helper reads temp file
			require.NoError(t, err)
			assert.Equal(t, "target", string(b), "the link target must be left as it was")
		}
	})
	t.Run("LinkedDirectoryStillReceivesTheCopy", func(t *testing.T) {
		// The layout an operator creates by linking an external drive into the library, and the one
		// a git-annex library has throughout, so it must keep working.
		dir := t.TempDir()
		real := filepath.Join(t.TempDir(), "external")
		require.NoError(t, os.MkdirAll(real, ModeDir))

		linked := filepath.Join(dir, "external")
		require.NoError(t, os.Symlink(real, linked))

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		require.NoError(t, Copy(src, filepath.Join(linked, "photo.jpg"), false))

		b, err := os.ReadFile(filepath.Join(real, "photo.jpg")) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "payload", string(b), "the copy must land in the linked directory")
	})
	t.Run("MoveRefusesALinkedDestination", func(t *testing.T) {
		for _, force := range []bool{false, true} {
			dir := t.TempDir()
			outside := filepath.Join(t.TempDir(), "outside.jpg")

			src := filepath.Join(dir, "src.jpg")
			require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

			dest := filepath.Join(dir, "dangling.jpg")
			require.NoError(t, os.Symlink(outside, dest))

			require.Error(t, Move(src, dest, force), "force=%v", force)
			assert.NoFileExists(t, outside, "the link target must not be created")
			assert.FileExists(t, src, "a refused move must keep its source")
		}
	})
}

// TestCopyMove_Publication covers what a call leaves behind: the destination it was asked for, the
// permissions a regular file carries, and nothing else.
func TestCopyMove_Publication(t *testing.T) {
	t.Run("CopyLeavesNoStagedFile", func(t *testing.T) {
		dir := t.TempDir()

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		require.NoError(t, Copy(src, filepath.Join(dir, "dest.jpg"), false))

		assert.Empty(t, stagedNames(t, dir))
	})
	t.Run("FailedCopyRemovesOnlyWhatItCreated", func(t *testing.T) {
		dir := t.TempDir()

		// A directory cannot be read as a file, so the copy fails after staging.
		src := filepath.Join(dir, "src")
		require.NoError(t, os.MkdirAll(src, ModeDir))

		dest := filepath.Join(dir, "dest.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("keep"), ModeFile))

		require.Error(t, Copy(src, dest, true))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "keep", string(b), "a failed copy must leave the destination alone")
		assert.Empty(t, stagedNames(t, dir), "a failed copy must remove the file it staged")
	})
	t.Run("NewDestinationCarriesTheSharedFileMode", func(t *testing.T) {
		// The staged file is created with the shared regular-file mode, so publishing it does not
		// hand the destination a private one.
		dir := t.TempDir()

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		// A file the package writes directly is the reference, so the umask applies to both.
		reference := filepath.Join(dir, "reference.jpg")
		require.NoError(t, os.WriteFile(reference, nil, ModeFile))

		dest := filepath.Join(dir, "dest.jpg")
		require.NoError(t, Copy(src, dest, false))

		info, err := os.Stat(dest)
		require.NoError(t, err)

		refInfo, err := os.Stat(reference)
		require.NoError(t, err)

		assert.Equal(t, refInfo.Mode().Perm(), info.Mode().Perm(), "a copy must carry the mode a written file does")
	})
	t.Run("ReplacedDestinationKeepsItsMode", func(t *testing.T) {
		// Publishing replaces the destination rather than writing into it, so the mode it already
		// carried is applied to the staged file first.
		dir := t.TempDir()

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		dest := filepath.Join(dir, "dest.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("old"), 0o600))
		require.NoError(t, os.Chmod(dest, 0o600))

		require.NoError(t, Copy(src, dest, true))

		info, err := os.Stat(dest)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "a replaced destination must keep its own mode")

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "payload", string(b))
	})
	t.Run("ReplacedDestinationKeepsItsGroup", func(t *testing.T) {
		// An unprivileged process may set a group it belongs to, so the group is the half of the
		// ownership carry an ordinary run can pin. The uid half needs privilege the suite lacks.
		groups, err := os.Getgroups()
		require.NoError(t, err)

		other := -1

		for _, gid := range groups {
			if gid != os.Getegid() {
				other = gid
				break
			}
		}

		if other < 0 {
			t.Skip("the process belongs to no group besides its own")
		}

		dir := t.TempDir()

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		dest := filepath.Join(dir, "dest.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("old"), ModeFile))
		require.NoError(t, os.Chown(dest, -1, other))

		require.NoError(t, Copy(src, dest, true))

		info, err := os.Stat(dest)
		require.NoError(t, err)

		owner, ok := info.Sys().(*syscall.Stat_t)
		require.True(t, ok)
		assert.EqualValues(t, other, owner.Gid, "a replaced destination must keep the group it had")
	})
	t.Run("ReplacedDestinationLeavesItsOtherNamesAlone", func(t *testing.T) {
		// Publishing replaces the destination rather than writing into it, so another name for the
		// same content keeps what it had instead of being rewritten through the shared inode.
		dir := t.TempDir()

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		dest := filepath.Join(dir, "dest.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("shared"), ModeFile))

		peer := filepath.Join(dir, "peer.jpg")
		require.NoError(t, os.Link(dest, peer))

		require.NoError(t, Copy(src, dest, true))

		b, err := os.ReadFile(peer) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "shared", string(b), "another name for the replaced content must keep it")

		b, err = os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "payload", string(b))
	})
	t.Run("PanicLeavesNothingBehind", func(t *testing.T) {
		// Cleanup is tracked explicitly rather than inferred from the error, so it also covers the
		// way out a panic takes.
		original := linkFile

		defer func() { linkFile = original }()

		linkFile = func(string, string) error { panic("probe") }

		dir := t.TempDir()

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		dest := filepath.Join(dir, "dest.jpg")
		err := Copy(src, dest, false)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "panic")
		assert.NoFileExists(t, dest, "a panic must not publish a destination")
		assert.Empty(t, stagedNames(t, dir), "a panic must remove the file the call staged")
	})
	t.Run("MoveReportsASourceItCouldNotRemove", func(t *testing.T) {
		// A link publishes without needing write access to the source directory, so the removal that
		// completes the move has to be reported rather than discarded.
		if os.Geteuid() == 0 {
			t.Skip("running as root ignores the directory mode")
		}

		base := t.TempDir()

		srcDir := filepath.Join(base, "drop")
		require.NoError(t, os.MkdirAll(srcDir, ModeDir))

		src := filepath.Join(srcDir, "photo.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))
		require.NoError(t, os.Chmod(srcDir, 0o555)) //nolint:gosec // the test needs a directory it cannot write

		t.Cleanup(func() { _ = os.Chmod(srcDir, 0o755) }) //nolint:gosec // restores the directory so the temp dir can be removed

		err := Move(src, filepath.Join(base, "photo.jpg"), false)

		require.Error(t, err, "a move that leaves its source behind must not report success")
		assert.FileExists(t, src)
	})
	t.Run("MoveWithoutLinksStaysARename", func(t *testing.T) {
		// A file system without hard links keeps the rename it would otherwise have taken, rather
		// than reading and writing the whole file.
		original := linkFile

		defer func() { linkFile = original }()

		linkFile = func(string, string) error { return syscall.EPERM }

		dir := t.TempDir()

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		before, err := os.Stat(src)
		require.NoError(t, err)

		dest := filepath.Join(dir, "dest.jpg")
		require.NoError(t, Move(src, dest, false))

		after, err := os.Stat(dest)
		require.NoError(t, err)

		assert.True(t, os.SameFile(before, after), "a same-volume move must not copy the file")
		assert.NoFileExists(t, src)
	})
	t.Run("MoveAcrossDevicesCopiesAndRemoves", func(t *testing.T) {
		// Separate mounts for originals and import are the ordinary container layout, so a rename
		// that cannot cross the boundary must fall back rather than fail.
		other := "/dev/shm"

		if _, err := os.Stat(other); err != nil {
			t.Skipf("no second file system available: %s", err)
		}

		dir := t.TempDir()

		if sameDevice(t, dir, other) {
			t.Skip("the temporary directory is on the same file system")
		}

		crossDir, err := os.MkdirTemp(other, "zzfs")
		require.NoError(t, err)

		t.Cleanup(func() { _ = os.RemoveAll(crossDir) })

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		dest := filepath.Join(crossDir, "dest.jpg")
		require.NoError(t, Move(src, dest, false))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "payload", string(b))
		assert.NoFileExists(t, src, "a completed move must remove its source")
		assert.Empty(t, stagedNames(t, crossDir))
	})
	t.Run("MoveAcrossDevicesWithForce", func(t *testing.T) {
		// The shape both production callers use.
		other := "/dev/shm"

		if _, err := os.Stat(other); err != nil {
			t.Skipf("no second file system available: %s", err)
		}

		dir := t.TempDir()

		if sameDevice(t, dir, other) {
			t.Skip("the temporary directory is on the same file system")
		}

		crossDir, err := os.MkdirTemp(other, "zzfs")
		require.NoError(t, err)

		t.Cleanup(func() { _ = os.RemoveAll(crossDir) })

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		dest := filepath.Join(crossDir, "dest.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("old"), ModeFile))

		require.NoError(t, Move(src, dest, true))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "payload", string(b))
		assert.NoFileExists(t, src)
		assert.Empty(t, stagedNames(t, crossDir))
	})
	t.Run("FailedPublishRemovesTheStagedFile", func(t *testing.T) {
		// A destination that becomes unwritable between the check and the publish leaves nothing
		// behind, which the copy step's own failure path cannot show.
		dir := t.TempDir()

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		dest := filepath.Join(dir, "dest.jpg")
		require.NoError(t, os.MkdirAll(dest, ModeDir))

		require.Error(t, Copy(src, dest, true), "a directory cannot be replaced by a copy")
		assert.Empty(t, stagedNames(t, dir), "a failed publish must remove the file it staged")
	})
	t.Run("MoveAcrossDevicesKeepsAnExistingDestination", func(t *testing.T) {
		other := "/dev/shm"

		if _, err := os.Stat(other); err != nil {
			t.Skipf("no second file system available: %s", err)
		}

		dir := t.TempDir()

		if sameDevice(t, dir, other) {
			t.Skip("the temporary directory is on the same file system")
		}

		crossDir, err := os.MkdirTemp(other, "zzfs")
		require.NoError(t, err)

		t.Cleanup(func() { _ = os.RemoveAll(crossDir) })

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		dest := filepath.Join(crossDir, "dest.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("keep"), ModeFile))

		require.Error(t, Move(src, dest, false))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "keep", string(b))
		assert.FileExists(t, src, "a refused move must keep its source")
	})
}

func TestStageName(t *testing.T) {
	t.Run("Hidden", func(t *testing.T) {
		name := stageName("/originals/2030/photo.jpg")

		assert.Equal(t, "/originals/2030", filepath.Dir(name))
		assert.True(t, strings.HasPrefix(filepath.Base(name), ".photo.jpg."), "name: %s", name)
		assert.True(t, strings.HasSuffix(name, ".tmp.jpg"), "name: %s", name)
	})
	t.Run("LongName", func(t *testing.T) {
		// A directory entry is bounded, so a destination at the limit must still be stageable.
		base := strings.Repeat("a", 250) + ".jpg"

		name := stageName(filepath.Join("/originals", base))

		assert.LessOrEqual(t, len(filepath.Base(name)), 255, "the staged name must fit a directory entry")
		assert.True(t, strings.HasSuffix(name, ".tmp.jpg"))
	})
	t.Run("LongNameOnARuneBoundary", func(t *testing.T) {
		// A file system that validates names refuses a split character, so the clamp must not cut
		// one in half.
		base := strings.Repeat("\u3042", 81) + ".jpg"

		name := stageName(filepath.Join("/originals", base))

		assert.LessOrEqual(t, len(filepath.Base(name)), 255)
		assert.True(t, utf8.ValidString(name), "the staged name must stay valid UTF-8: %q", name)
	})
	t.Run("Unique", func(t *testing.T) {
		assert.NotEqual(t, stageName("/originals/photo.jpg"), stageName("/originals/photo.jpg"))
	})
	t.Run("LongExtension", func(t *testing.T) {
		// An extension long enough to fill the entry on its own is dropped rather than clamped,
		// so the staged name still fits.
		base := "photo." + strings.Repeat("a", 250)

		name := stageName(filepath.Join("/originals", base))

		assert.LessOrEqual(t, len(filepath.Base(name)), 255)
		assert.True(t, strings.HasSuffix(name, ExtTmp), "name: %s", name)
	})
	t.Run("KeepsDestinationExtension", func(t *testing.T) {
		// Media tools detect a staged file's type from its name, so the destination's extension
		// has to survive staging.
		name := stageName("/sidecar/2030/video.m2ts.mp4")

		assert.True(t, strings.HasPrefix(filepath.Base(name), ".video.m2ts.mp4."), "name: %s", name)
		assert.True(t, strings.HasSuffix(name, ExtTmp+ExtMp4), "name: %s", name)
	})
}

func TestCreateStageFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "video.mp4")

		name, err := CreateStageFile(dest)
		require.NoError(t, err)

		assert.FileExists(t, name)
		assert.NotEqual(t, dest, name, "the staged file must not take the destination name")
		assert.True(t, strings.HasSuffix(name, ExtMp4), "name: %s", name)
		assert.NoFileExists(t, dest, "staging must not create the destination")

		// A second call must reserve a different name, so concurrent callers cannot share one.
		other, err := CreateStageFile(dest)
		require.NoError(t, err)
		assert.NotEqual(t, name, other)
	})
	t.Run("UnwritableDir", func(t *testing.T) {
		name, err := CreateStageFile(filepath.Join(t.TempDir(), "missing", "video.mp4"))

		assert.Error(t, err)
		assert.Empty(t, name)
	})
}

func TestCopy_LongDestinationName(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src.jpg")
	require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

	dest := filepath.Join(dir, strings.Repeat("a", 250)+".jpg")

	require.NoError(t, Copy(src, dest, false))

	b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
	require.NoError(t, err)
	assert.Equal(t, "payload", string(b))
	assert.Empty(t, stagedNames(t, dir))
}

// TestPublishFile_WithoutLinks covers a file system that cannot hard link, where the name is checked
// with a stat instead. SMB shares without the Unix extensions and FAT volumes behave this way.
func TestPublishFile_WithoutLinks(t *testing.T) {
	original := linkFile

	defer func() { linkFile = original }()

	linkFile = func(string, string) error { return syscall.EPERM }

	staged := func(t *testing.T, dir, content string) string {
		t.Helper()

		f, err := stageFile(filepath.Join(dir, "photo.jpg"))
		require.NoError(t, err)
		_, err = f.WriteString(content)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		return f.Name()
	}

	t.Run("FreeName", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "photo.jpg")

		require.NoError(t, publishFile(staged(t, dir, "new"), dest, false))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "new", string(b))
		assert.Empty(t, stagedNames(t, dir))
	})
	t.Run("TakenName", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "photo.jpg")
		require.NoError(t, os.WriteFile(dest, []byte("old"), ModeFile))

		require.Error(t, publishFile(staged(t, dir, "new"), dest, false))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "old", string(b), "the destination must be left as it was")
	})
	t.Run("EmptyDestination", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "photo.jpg")
		require.NoError(t, os.WriteFile(dest, nil, ModeFile))

		require.NoError(t, publishFile(staged(t, dir, "new"), dest, false))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "new", string(b))
	})
	t.Run("CopyStillPublishes", func(t *testing.T) {
		dir := t.TempDir()

		src := filepath.Join(dir, "src.jpg")
		require.NoError(t, os.WriteFile(src, []byte("payload"), ModeFile))

		dest := filepath.Join(dir, "dest.jpg")
		require.NoError(t, Copy(src, dest, false))

		b, err := os.ReadFile(dest) //nolint:gosec // test helper reads temp file
		require.NoError(t, err)
		assert.Equal(t, "payload", string(b))
		assert.Empty(t, stagedNames(t, dir))
	})
}
