package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPurgeTestDbFiles_Recursive(t *testing.T) {
	dir := t.TempDir()

	toCreate := []string{
		filepath.Join(dir, ".alpha.db"),         // match '.*.db'
		filepath.Join(dir, ".BETA.DB"),          // case-insensitive
		filepath.Join(dir, ".gamma.db-journal"), // match '.*.db-journal'
		filepath.Join(dir, ".DELTA.DB-JOURNAL"), // case-insensitive
		filepath.Join(dir, ".test.sqlite"),      // match '.test.*'
		filepath.Join(dir, ".test.anything"),    // match '.test.*'
		filepath.Join(dir, "epsilon.db"),        // no leading dot → keep
		filepath.Join(dir, "zeta"),              // no extension → keep
	}

	nestedDir := filepath.Join(dir, "nested")
	if err := os.MkdirAll(nestedDir, ModeDir); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	toCreate = append(toCreate,
		filepath.Join(nestedDir, ".theta.db"),
		filepath.Join(nestedDir, "iota.db-journal"), // no leading dot → keep
	)

	for _, f := range toCreate {
		if err := os.WriteFile(f, []byte("x"), ModeSecretFile); err != nil {
			t.Fatalf("create file %s: %v", f, err)
		}
	}

	PurgeTestDbFiles(dir, true)

	// Expect deletions.
	deleted := []string{
		filepath.Join(dir, ".alpha.db"),
		filepath.Join(dir, ".BETA.DB"),
		filepath.Join(dir, ".gamma.db-journal"),
		filepath.Join(dir, ".DELTA.DB-JOURNAL"),
		filepath.Join(dir, ".test.sqlite"),
		filepath.Join(dir, ".test.anything"),
		filepath.Join(nestedDir, ".theta.db"),
	}
	for _, f := range deleted {
		if FileExists(f) {
			t.Fatalf("expected %s to be deleted", f)
		}
	}

	// Expect survivors.
	survivors := []string{
		filepath.Join(dir, "epsilon.db"),
		filepath.Join(dir, "zeta"),
		filepath.Join(nestedDir, "iota.db-journal"),
	}
	for _, f := range survivors {
		if !FileExists(f) {
			t.Fatalf("expected %s to remain", f)
		}
	}
}

func TestPurgeTestDbFiles_NonRecursive(t *testing.T) {
	dir := t.TempDir()

	// Top-level files
	files := []string{
		filepath.Join(dir, ".a.db"),
		filepath.Join(dir, ".b.db-journal"),
		filepath.Join(dir, ".test.c"),
		filepath.Join(dir, "should-stay.db"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("x"), ModeSecretFile); err != nil {
			t.Fatalf("create %s: %v", f, err)
		}
	}

	// Nested files
	nested := filepath.Join(dir, "sub")
	if err := os.MkdirAll(nested, ModeDir); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	nestedFiles := []string{
		filepath.Join(nested, ".nested.db"),
		filepath.Join(nested, ".test.nested"),
	}
	for _, f := range nestedFiles {
		if err := os.WriteFile(f, []byte("x"), ModeSecretFile); err != nil {
			t.Fatalf("create %s: %v", f, err)
		}
	}

	PurgeTestDbFiles(dir, false)

	// Top-level deleted
	for _, f := range []string{filepath.Join(dir, ".a.db"), filepath.Join(dir, ".b.db-journal"), filepath.Join(dir, ".test.c")} {
		if FileExists(f) {
			t.Fatalf("expected %s to be deleted", f)
		}
	}
	// Top-level survivor
	if !FileExists(filepath.Join(dir, "should-stay.db")) {
		t.Fatalf("expected top-level survivor to remain")
	}
	// Nested survivors (non-recursive should not touch these)
	for _, f := range nestedFiles {
		if !FileExists(f) {
			t.Fatalf("expected nested file to remain: %s", f)
		}
	}
}

func TestPurgeExpired(t *testing.T) {
	// stamp creates a file in dir and backdates its modification time by age.
	stamp := func(t *testing.T, dir, name string, age time.Duration) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("test"), ModeFile); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		modTime := time.Now().Add(-age)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("chtimes %s: %v", name, err)
		}
		return path
	}
	t.Run("RemovesExpiredOnly", func(t *testing.T) {
		dir := t.TempDir()
		expired := stamp(t, dir, "photoprism-download-20260727-094439-zihqtuw4.zip", 2*time.Hour)
		fresh := stamp(t, dir, "photoprism-download-20260727-101500-abcdefgh.zip", time.Minute)
		other := stamp(t, dir, "notes.txt", 2*time.Hour)
		removed, remaining, failed := PurgeExpired(dir, ".zip", time.Hour)
		assert.Equal(t, 1, removed)
		assert.Equal(t, 1, remaining)
		assert.Equal(t, 0, failed)
		assert.False(t, FileExists(expired))
		assert.True(t, FileExists(fresh))
		assert.True(t, FileExists(other))
	})
	t.Run("CaseInsensitiveExt", func(t *testing.T) {
		dir := t.TempDir()
		expired := stamp(t, dir, "ARCHIVE.ZIP", 2*time.Hour)
		removed, remaining, failed := PurgeExpired(dir, ".zip", time.Hour)
		assert.Equal(t, 1, removed)
		assert.Equal(t, 0, remaining)
		assert.Equal(t, 0, failed)
		assert.False(t, FileExists(expired))
	})
	t.Run("EmptyExtMatchesAll", func(t *testing.T) {
		dir := t.TempDir()
		expired := stamp(t, dir, "notes.txt", 2*time.Hour)
		removed, _, _ := PurgeExpired(dir, "", time.Hour)
		assert.Equal(t, 1, removed)
		assert.False(t, FileExists(expired))
	})
	t.Run("SkipsDirectories", func(t *testing.T) {
		dir := t.TempDir()
		nested := filepath.Join(dir, "nested.zip")
		if err := os.MkdirAll(nested, ModeDir); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		modTime := time.Now().Add(-2 * time.Hour)
		if err := os.Chtimes(nested, modTime, modTime); err != nil {
			t.Fatalf("chtimes: %v", err)
		}
		removed, remaining, failed := PurgeExpired(dir, ".zip", time.Hour)
		assert.Equal(t, 0, removed)
		assert.Equal(t, 0, remaining)
		assert.Equal(t, 0, failed)
		assert.True(t, PathExists(nested))
	})
	t.Run("ZeroMaxAgeRemovesNothing", func(t *testing.T) {
		dir := t.TempDir()
		expired := stamp(t, dir, "archive.zip", 2*time.Hour)
		removed, remaining, failed := PurgeExpired(dir, ".zip", 0)
		assert.Equal(t, 0, removed)
		assert.Equal(t, 0, remaining)
		assert.Equal(t, 0, failed)
		assert.True(t, FileExists(expired))
	})
	t.Run("EmptyDir", func(t *testing.T) {
		removed, remaining, failed := PurgeExpired("", ".zip", time.Hour)
		assert.Equal(t, 0, removed)
		assert.Equal(t, 0, remaining)
		assert.Equal(t, 0, failed)
	})
	t.Run("MissingDir", func(t *testing.T) {
		removed, remaining, failed := PurgeExpired(filepath.Join(t.TempDir(), "missing"), ".zip", time.Hour)
		assert.Equal(t, 0, removed)
		assert.Equal(t, 0, remaining)
		assert.Equal(t, 0, failed)
	})
}
