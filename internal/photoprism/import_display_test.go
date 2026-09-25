package photoprism

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestImport_FolderDisplayPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	newImport := func(t *testing.T, cfg *config.Config) *Import {
		t.Helper()

		convert := NewConvert(cfg)

		return NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
	}

	t.Run("UploadWritesNoImportFolderRows", func(t *testing.T) {
		// An upload is imported from a per-session directory under the storage path, which is not
		// part of the import namespace, so it contributes no folder record and names no staging
		// path in the log.
		cfg := config.NewMinimalTestConfigWithDb("import-display", t.TempDir())

		uploadPath := filepath.Join(cfg.StoragePath(), "users", "us6sg6bxpogaaba1", "upload", "sessionrefid")
		nested := filepath.Join(uploadPath, "Vacation 2030")

		require.NoError(t, os.MkdirAll(nested, fs.ModeDir))

		hook := captureLog(t)

		newImport(t, cfg).Start(ImportOptionsUpload(uploadPath, ""))

		found := entity.Folders{}
		require.NoError(t, entity.UnscopedDb().Find(&found, "root = ?", entity.RootImport).Error)
		assert.Empty(t, found, "an upload must not record a folder in the import namespace")

		messages := loggedMessages(hook, logrus.TraceLevel)
		require.NotEmpty(t, messages)

		for _, msg := range messages {
			assert.NotContains(t, msg, cfg.StoragePath(), "a log line must not name the storage root")
			assert.NotContains(t, msg, "sessionrefid", "a log line must not name the session")
		}
	})
	t.Run("SubfolderImportKeepsTheImportNamespace", func(t *testing.T) {
		// Importing a subfolder still records its directories under the configured import path, so
		// the namespace stays the one the rest of the application reads.
		cfg := config.NewMinimalTestConfigWithDb("import-display-subfolder", t.TempDir())

		require.NoError(t, os.MkdirAll(filepath.Join(cfg.ImportPath(), "2030", "05"), fs.ModeDir))

		newImport(t, cfg).Start(ImportOptionsMove(filepath.Join(cfg.ImportPath(), "2030"), ""))

		found := entity.Folders{}
		require.NoError(t, entity.UnscopedDb().Find(&found, "root = ?", entity.RootImport).Error)

		paths := make([]string, 0, len(found))

		for _, folder := range found {
			paths = append(paths, folder.Path)
		}

		assert.Contains(t, paths, "2030/05", "a subfolder walk must record the path from the import root")
		assert.NotContains(t, paths, "05", "a subfolder walk must not record a path from its own root")
	})
	t.Run("SiblingOfTheImportPathIsOutsideTheNamespace", func(t *testing.T) {
		// A directory whose name merely starts with the import path is not inside it.
		cfg := config.NewMinimalTestConfigWithDb("import-display-sibling", t.TempDir())

		sibling := cfg.ImportPath() + "x"
		require.NoError(t, os.MkdirAll(filepath.Join(sibling, "sub"), fs.ModeDir))

		newImport(t, cfg).Start(ImportOptionsMove(sibling, ""))

		found := entity.Folders{}
		require.NoError(t, entity.UnscopedDb().Find(&found, "root = ?", entity.RootImport).Error)
		assert.Empty(t, found, "a sibling directory must not record a folder in the import namespace")
	})
	t.Run("FolderNameIsSanitized", func(t *testing.T) {
		// A directory name reaches the log through the folder record, so control and bidi characters
		// must not survive into it.
		cfg := config.NewMinimalTestConfigWithDb("import-display-clean", t.TempDir())

		importPath := cfg.ImportPath()
		nested := filepath.Join(importPath, "album\u202egpj.exe\a")

		require.NoError(t, os.MkdirAll(nested, fs.ModeDir))

		hook := captureLog(t)

		newImport(t, cfg).Start(ImportOptionsMove(importPath, ""))

		messages := loggedMessages(hook, logrus.TraceLevel)
		require.True(t, slices.ContainsFunc(messages, func(s string) bool {
			return strings.Contains(s, "added folder") && strings.Contains(s, "album")
		}), "the folder line must be among %v", messages)

		for _, msg := range messages {
			assert.NotContains(t, msg, "\u202e", "a log line must not carry a bidi override")
			assert.NotContains(t, msg, "\a", "a log line must not carry a control character")
			assert.False(t, strings.ContainsAny(msg, "\n\r"), "a log line must not be split")
		}
	})
}

func TestMediaFile_ErrorIsMaskable(t *testing.T) {
	// Move and Copy wrap their cause, so the sanitizer can reach the path the error names.
	cfg := config.NewMinimalTestConfig(t.TempDir())

	src := filepath.Join(cfg.ImportPath(), "maskable.jpg")
	require.NoError(t, os.MkdirAll(cfg.ImportPath(), fs.ModeDir))
	require.NoError(t, os.WriteFile(src, []byte("test"), fs.ModeFile))

	// A regular file where the destination needs a directory, so creating it fails by path.
	blocked := filepath.Join(t.TempDir(), "blocked")
	require.NoError(t, os.WriteFile(blocked, []byte("not a directory"), fs.ModeFile))

	dest := filepath.Join(blocked, "sub", "maskable.jpg")

	f, err := NewMediaFile(src)
	require.NoError(t, err)

	t.Run("Copy", func(t *testing.T) {
		copyErr := f.Copy(dest, false)

		require.Error(t, copyErr)
		require.Contains(t, copyErr.Error(), blocked, "the raw error must name the path this masks")
		assert.NotContains(t, clean.Error(copyErr), blocked)
	})
	t.Run("Move", func(t *testing.T) {
		moveErr := f.Move(dest, false)

		require.Error(t, moveErr)
		require.Contains(t, moveErr.Error(), blocked, "the raw error must name the path this masks")
		assert.NotContains(t, clean.Error(moveErr), blocked)
	})
}

func TestMediaFile_CopyReportsAReadFailure(t *testing.T) {
	// Move deletes the source once Copy reports success, so a copy that could not read its source
	// must not report one.
	dir := t.TempDir()

	src := filepath.Join(dir, "unreadable")
	require.NoError(t, os.MkdirAll(src, fs.ModeDir))

	dest := filepath.Join(dir, "dest.jpg")

	m := &MediaFile{fileName: src}

	err := m.Copy(dest, false)

	require.Error(t, err, "a source that cannot be read must not report a successful copy")
	assert.Contains(t, err.Error(), "copy:")
}
