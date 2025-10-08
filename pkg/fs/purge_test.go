package fs

import (
	"os"
	"path/filepath"
	"testing"
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
