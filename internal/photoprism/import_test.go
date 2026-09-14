package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestNewImport(t *testing.T) {
	cfg := config.TestConfig()

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	imp := NewImport(cfg, ind, convert)

	assert.IsType(t, &Import{}, imp)
}

func TestStoredCopyOf(t *testing.T) {
	cfg := config.TestConfig()
	require.NoError(t, cfg.InitializeTestData())

	mediaFile, err := NewMediaFile(cfg.ImportPath() + "/raw/IMG_2567.CR2")
	require.NoError(t, err)
	require.NotEmpty(t, mediaFile.Hash())

	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, StoredCopyOf(nil))
	})
	t.Run("NoHash", func(t *testing.T) {
		dir := t.TempDir()
		name := filepath.Join(dir, "gone.jpg")
		require.NoError(t, os.WriteFile(name, []byte("x"), fs.ModeFile))

		unreadable, newErr := NewMediaFile(name)
		require.NoError(t, newErr)
		require.NoError(t, os.Remove(name))
		require.Empty(t, unreadable.Hash())

		assert.Nil(t, StoredCopyOf(unreadable))
	})
	t.Run("NotIndexed", func(t *testing.T) {
		assert.Nil(t, StoredCopyOf(mediaFile), "no row records this content")
	})
	t.Run("IndexedButNotOnDisk", func(t *testing.T) {
		// A row on its own is not the content. Answering otherwise lets a caller remove the copy it
		// is holding.
		photo := entity.PhotoFixtures.Get("Photo01")
		file := &entity.File{
			PhotoID:  photo.ID,
			PhotoUID: photo.PhotoUID,
			FileRoot: entity.RootOriginals,
			FileName: "zz-stored/absent.cr2",
			FileHash: mediaFile.Hash(),
		}
		require.NoError(t, file.Save())

		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(file).Error })

		assert.Nil(t, StoredCopyOf(mediaFile))
	})
	t.Run("IndexedAndOnDisk", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo01")
		stored := "zz-stored/present.cr2"
		storedPath := filepath.Join(cfg.OriginalsPath(), stored)

		require.NoError(t, os.MkdirAll(filepath.Dir(storedPath), fs.ModeDir))
		require.NoError(t, os.WriteFile(storedPath, []byte("content"), fs.ModeFile))

		file := &entity.File{
			PhotoID:  photo.ID,
			PhotoUID: photo.PhotoUID,
			FileRoot: entity.RootOriginals,
			FileName: stored,
			FileHash: mediaFile.Hash(),
		}
		require.NoError(t, file.Save())

		t.Cleanup(func() {
			_ = entity.UnscopedDb().Delete(file).Error
			_ = os.RemoveAll(filepath.Dir(storedPath))
		})

		found := StoredCopyOf(mediaFile)

		require.NotNil(t, found)
		assert.Equal(t, stored, found.FileName)
	})
}

func TestImport_DestinationFilename(t *testing.T) {
	cfg := config.TestConfig()

	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

	imp := NewImport(cfg, ind, convert)

	rawFile, err := NewMediaFile(cfg.ImportPath() + "/raw/IMG_2567.CR2")

	if err != nil {
		t.Fatal(err)
	}

	t.Run("NoBasePath", func(t *testing.T) {
		fileName, err := imp.DestinationFilename(rawFile, rawFile, "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, cfg.OriginalsPath()+"/2019/07/20190705_153230_C167C6FD.cr2", fileName)
	})
	t.Run("WithBasePath", func(t *testing.T) {
		fileName, err := imp.DestinationFilename(rawFile, rawFile, "users/guest")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, cfg.OriginalsPath()+"/users/guest/2019/07/20190705_153230_C167C6FD.cr2", fileName)
	})
	t.Run("ARecordedNameHeldByALinkIsNotReportedAsADuplicate", func(t *testing.T) {
		// A duplicate is reported only when the recorded file is there. The worker deletes the file
		// it is importing on that answer, so a name that resolves to nothing takes the search
		// instead.
		hash := rawFile.Hash()
		require.NotEmpty(t, hash)

		recorded := "zz-linked/recorded.cr2"
		recordedPath := filepath.Join(cfg.OriginalsPath(), recorded)

		require.NoError(t, os.MkdirAll(filepath.Dir(recordedPath), fs.ModeDir))
		require.NoError(t, os.Symlink(filepath.Join(cfg.OriginalsPath(), "zz-linked", "gone.cr2"), recordedPath))

		photo := entity.PhotoFixtures.Get("Photo01")

		file := &entity.File{
			PhotoID:  photo.ID,
			PhotoUID: photo.PhotoUID,
			FileRoot: entity.RootOriginals,
			FileName: recorded,
			FileHash: hash,
		}
		require.NoError(t, file.Save())

		t.Cleanup(func() {
			_ = entity.UnscopedDb().Delete(file).Error
			_ = os.RemoveAll(filepath.Dir(recordedPath))
		})

		require.False(t, fs.FileExists(recordedPath), "the recorded name must resolve to nothing")

		fileName, err := imp.DestinationFilename(rawFile, rawFile, "")

		require.NoError(t, err, "a name that resolves to nothing is not a duplicate")
		assert.NotEqual(t, recordedPath, fileName, "and must not be handed back as the destination")
	})
	t.Run("AnUnreadableSourceDoesNotMatchALinkWithNoTarget", func(t *testing.T) {
		// Both sides hash to the empty string when they cannot be read, so a hash is compared only
		// once there is one on each side.
		dir := t.TempDir()

		name := filepath.Join(dir, "unreadable.jpg")
		require.NoError(t, os.WriteFile(name, []byte("x"), fs.ModeFile))

		unreadable, err := NewMediaFile(name)
		require.NoError(t, err)

		// The hash is computed once and kept, so the source goes before anything asks for it.
		require.NoError(t, os.Remove(name))
		require.Empty(t, unreadable.Hash(), "the source must not be readable, which is the case at issue")

		// Learn the name the search settles on before anything holds it.
		free, err := imp.DestinationFilename(unreadable, unreadable, "")
		require.NoError(t, err)

		require.NoError(t, os.MkdirAll(filepath.Dir(free), fs.ModeDir))
		require.NoError(t, os.Symlink(filepath.Join(dir, "no-such-target.jpg"), free))

		t.Cleanup(func() { _ = os.Remove(free) })

		require.Empty(t, fs.Hash(free), "and the name must be held by something that cannot be read either")

		fileName, err := imp.DestinationFilename(unreadable, unreadable, "")

		require.NoError(t, err, "two unreadable names are not the same file")
		assert.NotEqual(t, free, fileName, "the held name must be stepped over")
	})
	t.Run("StepsOverASymbolicLink", func(t *testing.T) {
		// The search reads the name rather than what it resolves to, because the move that follows
		// refuses a link.
		taken := cfg.OriginalsPath() + "/2019/07/20190705_153230_C167C6FD.cr2"

		require.NoError(t, os.MkdirAll(filepath.Dir(taken), fs.ModeDir))
		require.NoError(t, os.Symlink(filepath.Join(cfg.OriginalsPath(), "no-such-target.cr2"), taken))

		t.Cleanup(func() { _ = os.Remove(taken) })

		require.False(t, fs.FileExists(taken), "the link must not resolve, which is the case at issue")

		fileName, err := imp.DestinationFilename(rawFile, rawFile, "")

		require.NoError(t, err)
		assert.NotEqual(t, taken, fileName, "the name a link holds must not be handed back")
		assert.Equal(t, cfg.OriginalsPath()+"/2019/07/20190705_153230_C167C6FD.00001.cr2", fileName)
	})
}

func TestImport_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()

	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

	imp := NewImport(cfg, ind, convert)

	opt := ImportOptionsMove(cfg.ImportPath(), "")

	imp.Start(opt)
}
