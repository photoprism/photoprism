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

func TestCaseInsensitive(t *testing.T) {
	t.Run("Temp", func(t *testing.T) {
		if result, err := CaseInsensitive(os.TempDir()); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("tmp fs case-insensitive: %t", result)
		}
	})
}

func TestIgnoreCase(t *testing.T) {
	isCS, err := CaseInsensitive(os.TempDir())

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, isCS, ignoreCase)

	restoreCaseMode(t)

	IgnoreCase()
	assert.True(t, ignoreCase)
	assert.Equal(t, ExtensionList.Types(true), FileTypes)
}

// restoreCaseMode restores the case-insensitive lookup settings when the test ends.
func restoreCaseMode(t *testing.T) {
	t.Helper()
	m := GetCaseMode()
	t.Cleanup(func() { RestoreCaseMode(m) })
}

// writeCaseTestFiles creates empty files with the given names below dir.
func writeCaseTestFiles(t *testing.T, dir string, names ...string) {
	t.Helper()

	for _, name := range names {
		fileName := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(fileName), ModeDir))
		require.NoError(t, os.WriteFile(fileName, []byte(name), ModeFile))
	}
}

// requireCaseSensitiveDir skips the test if dir is on a case-insensitive file system.
func requireCaseSensitiveDir(t *testing.T, dir string) {
	t.Helper()

	if insensitive, err := CaseInsensitive(dir); err != nil {
		t.Fatal(err)
	} else if insensitive {
		t.Skip("requires a case-sensitive file system")
	}
}

// caseInsensitiveLstat simulates a case-insensitive file system on a case-sensitive one: a name that is
// not found resolves to the entry whose name differs only in ASCII letter case.
func caseInsensitiveLstat(name string) (os.FileInfo, error) {
	if info, err := os.Lstat(name); err == nil {
		return info, nil
	}

	return os.Lstat(filepath.Join(filepath.Dir(name), swapASCIICase(filepath.Base(name))))
}

// countingCaseProbe returns a probe with the given lstat func that counts the calls it makes.
func countingCaseProbe(lstat func(string) (os.FileInfo, error), readDirs, lstats *int) caseProbe {
	return caseProbe{
		readDir: func(dir string, n int, dev uint64) ([]os.DirEntry, uint64, error) {
			*readDirs++
			return readDirN(dir, n, dev)
		},
		lstat: func(name string) (os.FileInfo, error) {
			*lstats++
			return lstat(name)
		},
	}
}

// assertCaseResult checks the result of a case probe and the reason if it is unknown.
func assertCaseResult(t *testing.T, insensitive bool, err error, wantInsensitive bool, wantErr error) {
	t.Helper()
	assert.Equal(t, wantInsensitive, insensitive, "insensitive")

	assert.Equal(t, wantErr, err, "reason")
}

func TestCaseInsensitiveDir(t *testing.T) {
	t.Run("CaseSensitive", func(t *testing.T) {
		dir := t.TempDir()
		requireCaseSensitiveDir(t, dir)
		writeCaseTestFiles(t, dir, "IMG_0001.CR2")
		insensitive, reason := CaseInsensitiveDir(dir)
		assertCaseResult(t, insensitive, reason, false, nil)
	})
	t.Run("Empty", func(t *testing.T) {
		insensitive, reason := CaseInsensitiveDir(t.TempDir())
		assertCaseResult(t, insensitive, reason, false, errCaseNoName)
		insensitive, reason = CaseInsensitiveDir("")
		assertCaseResult(t, insensitive, reason, false, errCaseRead)
	})
	t.Run("NotFound", func(t *testing.T) {
		insensitive, reason := CaseInsensitiveDir(filepath.Join(t.TempDir(), "missing"))
		assertCaseResult(t, insensitive, reason, false, errCaseRead)
	})
	t.Run("NotDir", func(t *testing.T) {
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "IMG_0001.jpg")
		insensitive, reason := CaseInsensitiveDir(filepath.Join(dir, "IMG_0001.jpg"))
		assertCaseResult(t, insensitive, reason, false, errCaseRead)
	})
	t.Run("ReadOnly", func(t *testing.T) {
		// Tests may run as root, so the listing and modification time show that nothing was written.
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "IMG_0001.jpg", "2024/01/IMG_0002.jpg")
		mtime := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
		require.NoError(t, os.Chtimes(dir, mtime, mtime))
		require.NoError(t, os.Chmod(dir, 0o555))       //nolint:gosec // read-only test folder
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) }) //nolint:gosec // restores the test folder

		before, err := os.ReadDir(dir)
		require.NoError(t, err)

		_, _ = CaseInsensitiveDir(dir)

		after, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Equal(t, before, after)

		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, mtime.Equal(info.ModTime()))
	})
}

func TestCaseProbe_Run(t *testing.T) {
	t.Run("CaseInsensitive", func(t *testing.T) {
		// The host may have no case-insensitive file system, so the lstat func simulates one.
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "IMG_0001.CR2")

		var readDirs, lstats int

		insensitive, reason := countingCaseProbe(caseInsensitiveLstat, &readDirs, &lstats).run(dir)
		assertCaseResult(t, insensitive, reason, true, nil)
		assert.Equal(t, 1, readDirs)
		assert.Equal(t, 2, lstats)
	})
	t.Run("CaseSensitive", func(t *testing.T) {
		// The first name with letters decides, so the others are not looked up.
		dir := t.TempDir()
		requireCaseSensitiveDir(t, dir)
		writeCaseTestFiles(t, dir, "IMG_1.jpg", "IMG_2.jpg", "IMG_3.jpg")

		var readDirs, lstats int

		insensitive, reason := countingCaseProbe(os.Lstat, &readDirs, &lstats).run(dir)
		assertCaseResult(t, insensitive, reason, false, nil)
		assert.Equal(t, 2, lstats)
	})
	t.Run("BothSpellingsListed", func(t *testing.T) {
		// Two names that differ only in case are distinct files, even if lstat reports the same file.
		dir := t.TempDir()
		requireCaseSensitiveDir(t, dir)
		writeCaseTestFiles(t, dir, "a.jpg", "A.JPG")

		var readDirs, lstats int

		sameFile := func(string) (os.FileInfo, error) { return os.Lstat(filepath.Join(dir, "a.jpg")) }

		insensitive, reason := countingCaseProbe(sameFile, &readDirs, &lstats).run(dir)
		assertCaseResult(t, insensitive, reason, false, nil)
		assert.Equal(t, 0, lstats)
	})
	t.Run("BothSpellingsUnlisted", func(t *testing.T) {
		// The swapped name exists but is not among the entries read, so os.SameFile decides. A different file
		// may also be a case-insensitive mount without stable inode numbers, so the result is unknown.
		dir := t.TempDir()
		requireCaseSensitiveDir(t, dir)
		writeCaseTestFiles(t, dir, "a.jpg", "A.JPG")

		p := caseProbe{
			readDir: func(dir string, n int, dev uint64) ([]os.DirEntry, uint64, error) {
				entries, dirDev, err := readDirN(dir, n, dev)

				for _, e := range entries {
					if e.Name() == "a.jpg" {
						return []os.DirEntry{e}, dirDev, err
					}
				}

				return nil, dirDev, err
			},
			lstat: os.Lstat,
		}

		insensitive, reason := p.run(dir)
		assertCaseResult(t, insensitive, reason, false, errCaseOtherFile)
	})
	t.Run("LstatError", func(t *testing.T) {
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "IMG_0001.jpg")

		p := caseProbe{
			readDir: readDirN,
			lstat: func(name string) (os.FileInfo, error) {
				if filepath.Base(name) == "img_0001.JPG" {
					return nil, os.ErrPermission
				}

				return os.Lstat(name)
			},
		}

		insensitive, reason := p.run(dir)
		assertCaseResult(t, insensitive, reason, false, errCaseLookup)
	})
	t.Run("ListedNameError", func(t *testing.T) {
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "IMG_0001.jpg")

		p := caseProbe{
			readDir: readDirN,
			lstat: func(name string) (os.FileInfo, error) {
				if filepath.Base(name) == "IMG_0001.jpg" {
					return nil, os.ErrNotExist
				}

				return caseInsensitiveLstat(name)
			},
		}

		insensitive, reason := p.run(dir)
		assertCaseResult(t, insensitive, reason, false, errCaseLookup)
	})
	t.Run("NoLetters", func(t *testing.T) {
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "20240102_150405.123", "2024/0001.123")

		var readDirs, lstats int

		insensitive, reason := countingCaseProbe(caseInsensitiveLstat, &readDirs, &lstats).run(dir)
		assertCaseResult(t, insensitive, reason, false, errCaseNoName)
		assert.Equal(t, 0, lstats)
	})
	t.Run("ImportLayout", func(t *testing.T) {
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "2024/01/20240102_150405_ABCDEF01.jpg")

		var readDirs, lstats int

		insensitive, reason := countingCaseProbe(caseInsensitiveLstat, &readDirs, &lstats).run(dir)
		assertCaseResult(t, insensitive, reason, true, nil)
		assert.Equal(t, 3, readDirs)
	})
	t.Run("BelowDepth", func(t *testing.T) {
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "2024/01/02/IMG_0001.jpg")

		var readDirs, lstats int

		insensitive, reason := countingCaseProbe(caseInsensitiveLstat, &readDirs, &lstats).run(dir)
		assertCaseResult(t, insensitive, reason, false, errCaseNoName)
		assert.Equal(t, 0, lstats)
	})
	t.Run("Budget", func(t *testing.T) {
		dir := t.TempDir()

		for _, a := range []string{"1", "2", "3", "4", "5"} {
			for _, b := range []string{"1", "2", "3", "4", "5"} {
				for _, c := range []string{"1", "2", "3", "4", "5"} {
					require.NoError(t, os.MkdirAll(filepath.Join(dir, a, b, c), ModeDir))
				}
			}
		}

		var readDirs, lstats, maxEntries int

		p := countingCaseProbe(caseInsensitiveLstat, &readDirs, &lstats)
		readDir := p.readDir
		p.readDir = func(dir string, n int, dev uint64) ([]os.DirEntry, uint64, error) {
			maxEntries = max(maxEntries, n)
			return readDir(dir, n, dev)
		}

		insensitive, reason := p.run(dir)
		assertCaseResult(t, insensitive, reason, false, errCaseNoName)
		assert.Equal(t, caseProbeDirs, readDirs)
		assert.Equal(t, caseProbeEntries, maxEntries)
		assert.Equal(t, 0, lstats)
	})
	t.Run("Symlink", func(t *testing.T) {
		dir := t.TempDir()
		target := t.TempDir()
		writeCaseTestFiles(t, target, "IMG_0001.jpg")
		require.NoError(t, os.Symlink(target, filepath.Join(dir, "2024")))

		var readDirs, lstats int

		insensitive, reason := countingCaseProbe(caseInsensitiveLstat, &readDirs, &lstats).run(dir)
		assertCaseResult(t, insensitive, reason, false, errCaseNoName)
		assert.Equal(t, 1, readDirs)
	})
	t.Run("SwappedFifo", func(t *testing.T) {
		// A listed folder replaced with a FIFO before it is read is not opened for reading.
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "2024/01/IMG_0001.jpg")
		sub := filepath.Join(dir, "2024")

		p := caseProbe{
			readDir: func(name string, n int, dev uint64) ([]os.DirEntry, uint64, error) {
				if name == sub {
					assert.NoError(t, os.RemoveAll(sub))
					assert.NoError(t, syscall.Mkfifo(sub, 0o600))
				}

				return readDirN(name, n, dev)
			},
			lstat: os.Lstat,
		}

		done := make(chan struct{})
		go func() { _, _ = p.run(dir); close(done) }()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("probe blocked opening a FIFO")

			if w, err := os.OpenFile(sub, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil { //nolint:gosec // test FIFO
				_ = w.Close()
			}

			<-done
		}
	})
	t.Run("SwappedSymlink", func(t *testing.T) {
		// A listed folder replaced with a link before it is read is not followed.
		dir := t.TempDir()
		outside := t.TempDir()
		writeCaseTestFiles(t, dir, "2024/0001.123")
		writeCaseTestFiles(t, outside, "IMG_0001.jpg")
		sub := filepath.Join(dir, "2024")

		p := caseProbe{
			readDir: func(name string, n int, dev uint64) ([]os.DirEntry, uint64, error) {
				if name == sub {
					require.NoError(t, os.RemoveAll(sub))
					require.NoError(t, os.Symlink(outside, sub))
				}

				return readDirN(name, n, dev)
			},
			lstat: caseInsensitiveLstat,
		}

		insensitive, reason := p.run(dir)
		assertCaseResult(t, insensitive, reason, false, errCaseNoName)
	})
	t.Run("SubfolderError", func(t *testing.T) {
		// A folder below the root that cannot be read is skipped, so the next one is checked.
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "2023/IMG_0001.jpg", "2024/IMG_0002.jpg")

		var reads int

		p := caseProbe{
			readDir: func(name string, n int, dev uint64) ([]os.DirEntry, uint64, error) {
				if reads++; reads == 2 {
					return nil, 0, os.ErrPermission
				}

				return readDirN(name, n, dev)
			},
			lstat: caseInsensitiveLstat,
		}

		insensitive, reason := p.run(dir)
		assertCaseResult(t, insensitive, reason, true, nil)
		assert.Equal(t, 3, reads)
	})
	t.Run("OtherDevice", func(t *testing.T) {
		// The root device is passed to readDir, so a folder on another device is skipped before it is read.
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "2024/IMG_0001.jpg")

		var devs []uint64

		p := caseProbe{
			readDir: func(name string, n int, dev uint64) ([]os.DirEntry, uint64, error) {
				devs = append(devs, dev)

				if name != dir {
					return nil, dev + 1, errOtherDevice
				}

				return readDirN(name, n, dev)
			},
			lstat: caseInsensitiveLstat,
		}

		insensitive, reason := p.run(dir)
		assertCaseResult(t, insensitive, reason, false, errCaseNoName)
		require.Len(t, devs, 2)
		assert.Equal(t, anyDevice, devs[0])
		assert.NotEqual(t, anyDevice, devs[1])
	})
}

func TestCaseProbeErrors(t *testing.T) {
	assert.Equal(t, "folder not readable", errCaseRead.Error())
	assert.Equal(t, "file name lookup failed", errCaseLookup.Error())
	assert.Equal(t, "file name with swapped case is another file", errCaseOtherFile.Error())
	assert.Equal(t, "no file name with ASCII letters within reach", errCaseNoName.Error())
}

func TestSwapASCIICase(t *testing.T) {
	assert.Equal(t, "img_0001.cr2", swapASCIICase("IMG_0001.CR2"))
	assert.Equal(t, "cAFé.JPG", swapASCIICase("Café.jpg"))
	assert.Equal(t, "2024", swapASCIICase("2024"))
	assert.Equal(t, "", swapASCIICase(""))
}

func TestReadDirN(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "a.jpg", "b.jpg", "c.jpg")

		entries, dev, err := readDirN(dir, 2, anyDevice)
		require.NoError(t, err)
		assert.Len(t, entries, 2)
		assert.NotEqual(t, anyDevice, dev)

		entries, _, err = readDirN(dir, 64, dev)
		require.NoError(t, err)
		assert.Len(t, entries, 3)
	})
	t.Run("Empty", func(t *testing.T) {
		entries, _, err := readDirN(t.TempDir(), 2, anyDevice)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
	t.Run("OtherDevice", func(t *testing.T) {
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "a.jpg")

		_, dev, err := readDirN(dir, 2, anyDevice)
		require.NoError(t, err)

		entries, _, err := readDirN(dir, 2, dev+1)
		assert.ErrorIs(t, err, errOtherDevice)
		assert.Empty(t, entries)
	})
	t.Run("NotDir", func(t *testing.T) {
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "a.jpg")

		_, _, err := readDirN(filepath.Join(dir, "a.jpg"), 2, anyDevice)
		assert.Error(t, err)
	})
	t.Run("Symlink", func(t *testing.T) {
		// A link is followed only at the root, which is opened with anyDevice.
		dir := t.TempDir()
		writeCaseTestFiles(t, dir, "target/a.jpg")
		link := filepath.Join(dir, "link")
		require.NoError(t, os.Symlink(filepath.Join(dir, "target"), link))

		entries, dev, err := readDirN(link, 2, anyDevice)
		require.NoError(t, err)
		assert.Len(t, entries, 1)

		_, _, err = readDirN(link, 2, dev)
		assert.Error(t, err)
	})
	t.Run("Fifo", func(t *testing.T) {
		dir := t.TempDir()
		fifo := filepath.Join(dir, "fifo")
		require.NoError(t, syscall.Mkfifo(fifo, 0o600))

		done := make(chan error, 1)
		go func() { _, _, err := readDirN(fifo, 2, anyDevice); done <- err }()

		select {
		case err := <-done:
			assert.Error(t, err)
		case <-time.After(2 * time.Second):
			t.Error("readDirN blocked opening a FIFO")

			if w, err := os.OpenFile(fifo, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil { //nolint:gosec // test FIFO
				_ = w.Close()
			}

			<-done
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		_, _, err := readDirN(filepath.Join(t.TempDir(), "missing"), 2, anyDevice)
		assert.Error(t, err)
	})
}

func TestTypeExts(t *testing.T) {
	assert.Equal(t, []string{".png", ".PNG"}, typeExts(ImagePng, false)[:2])
	assert.Equal(t, ".png", typeExts(ImagePng, true)[0])
	assert.Equal(t, ExtensionList.Types(true)[ImageJpeg], typeExts(ImageJpeg, true))
	assert.Equal(t, ExtensionList.Types(false)[ImageJpeg], typeExts(ImageJpeg, false))
}

func TestGetCaseMode(t *testing.T) {
	restoreCaseMode(t)
	SetCaseDir("/x/photos", true)

	m := GetCaseMode()
	assert.Equal(t, "/x/photos", m.caseDir)
	assert.True(t, m.caseDirIgnore)
	assert.Equal(t, ignoreCase, m.ignoreCase)
	assert.Equal(t, FileTypes, m.fileTypes)
}

func TestSetCaseDir(t *testing.T) {
	restoreCaseMode(t)

	SetCaseDir("/x/photos/", true)
	assert.Equal(t, "/x/photos", caseDir)
	assert.True(t, caseDirIgnore)

	SetCaseDir("", true)
	assert.Equal(t, "", caseDir)
	assert.False(t, caseDirIgnore)
}

func TestIgnoreCaseIn(t *testing.T) {
	t.Run("Scoped", func(t *testing.T) {
		restoreCaseMode(t)
		ignoreCase = false
		SetCaseDir("/x/photos", true)

		assert.True(t, ignoreCaseIn("/x/photos"))
		assert.True(t, ignoreCaseIn("/x/photos/2024/.photoprism"))
		assert.False(t, ignoreCaseIn("/x/photos2"))
		assert.False(t, ignoreCaseIn("/x/storage/sidecar/2024"))
	})
	t.Run("Reverse", func(t *testing.T) {
		restoreCaseMode(t)
		ignoreCase = true
		SetCaseDir("/x/photos", false)

		assert.False(t, ignoreCaseIn("/x/photos/2024"))
		assert.True(t, ignoreCaseIn("/x/photos2"))
		assert.True(t, ignoreCaseIn("/x/storage/sidecar/2024"))
	})
	t.Run("Unset", func(t *testing.T) {
		restoreCaseMode(t)
		SetCaseDir("", false)

		ignoreCase = false
		assert.False(t, ignoreCaseIn("/x/photos"))

		ignoreCase = true
		assert.True(t, ignoreCaseIn("/x/photos"))
	})
}

func TestRestoreCaseMode(t *testing.T) {
	m := GetCaseMode()

	func() {
		defer RestoreCaseMode(m)
		IgnoreCase()
		SetCaseDir("/x/photos", true)
	}()

	assert.Equal(t, m, GetCaseMode())
}

func TestType_FindEach_CaseDir(t *testing.T) {
	// The temp folder is case-sensitive, so a lookup in case-insensitive mode misses the lowercase names.
	root := t.TempDir()
	requireCaseSensitiveDir(t, root)

	originals := filepath.Join(root, "photos")
	sidecar := filepath.Join(root, "storage", "sidecar")
	importDir := filepath.Join(root, "photos2")

	writeCaseTestFiles(t, originals, "2024/IMG_1234.raw", "2024/img_1234.jpg", "2024/IMG_1234.JPG",
		"2024/.photoprism/img_1234.jpg")
	writeCaseTestFiles(t, sidecar, "2024/img_1234.JPG")
	writeCaseTestFiles(t, importDir, "IMG_5678.raw", "img_5678.jpg")

	dirs := []string{sidecar, ".photoprism"}
	fileName := filepath.Join(originals, "2024", "IMG_1234.raw")
	importName := filepath.Join(importDir, "IMG_5678.raw")

	t.Run("Default", func(t *testing.T) {
		restoreCaseMode(t)
		ignoreCase = false
		SetCaseDir(originals, false)

		assert.Equal(t, []string{
			filepath.Join(originals, "2024", "img_1234.jpg"),
			filepath.Join(originals, "2024", ".photoprism", "img_1234.jpg"),
			filepath.Join(originals, "2024", "IMG_1234.JPG"),
			filepath.Join(sidecar, "2024", "img_1234.JPG"),
		}, ImageJpeg.FindAll(fileName, dirs, originals, false))
		assert.Equal(t, filepath.Join(originals, "2024", "img_1234.jpg"), ImageJpeg.Find(fileName, false))
		assert.Equal(t, filepath.Join(importDir, "img_5678.jpg"), ImageJpeg.FindFirst(importName, dirs, originals, false))
	})
	t.Run("OriginalsInsensitive", func(t *testing.T) {
		restoreCaseMode(t)
		ignoreCase = false
		SetCaseDir(originals, true)

		assert.Equal(t, []string{
			filepath.Join(sidecar, "2024", "img_1234.JPG"),
		}, ImageJpeg.FindAll(fileName, dirs, originals, false))
		assert.Equal(t, filepath.Join(sidecar, "2024", "img_1234.JPG"), ImageJpeg.FindFirst(fileName, dirs, originals, false))
		assert.Equal(t, "", ImageJpeg.Find(fileName, false))
		assert.Equal(t, "", ImageJpeg.FindGenerated(fileName, dirs, originals, false, nil))
		assert.Equal(t, filepath.Join(importDir, "img_5678.jpg"), ImageJpeg.FindFirst(importName, dirs, originals, false))
		assert.Equal(t, filepath.Join(importDir, "img_5678.jpg"), ImageJpeg.Find(importName, false))
	})
	t.Run("StorageInsensitive", func(t *testing.T) {
		restoreCaseMode(t)
		IgnoreCase()
		SetCaseDir(originals, false)

		assert.Equal(t, []string{
			filepath.Join(originals, "2024", "img_1234.jpg"),
			filepath.Join(originals, "2024", ".photoprism", "img_1234.jpg"),
			filepath.Join(originals, "2024", "IMG_1234.JPG"),
		}, ImageJpeg.FindAll(fileName, dirs, originals, false))
		assert.Equal(t, filepath.Join(originals, "2024", "img_1234.jpg"), ImageJpeg.Find(fileName, false))
		assert.Equal(t, "", ImageJpeg.FindFirst(importName, dirs, originals, false))
		assert.Equal(t, "", ImageJpeg.Find(importName, false))
	})
	t.Run("MixedOrder", func(t *testing.T) {
		// Extensions come first, then folders, so the uppercase extension in a case-sensitive folder is found
		// before the next lowercase extension.
		restoreCaseMode(t)
		ignoreCase = false

		root := t.TempDir()
		orig, side := filepath.Join(root, "photos"), filepath.Join(root, "sidecar")
		writeCaseTestFiles(t, orig, "2024/IMG_1234.raw")
		writeCaseTestFiles(t, side, "2024/IMG_1234.JPG", "2024/IMG_1234.jpeg")
		SetCaseDir(orig, true)

		name := filepath.Join(orig, "2024", "IMG_1234.raw")
		assert.Equal(t, filepath.Join(side, "2024", "IMG_1234.JPG"), ImageJpeg.FindFirst(name, []string{side}, orig, false))
	})
}
