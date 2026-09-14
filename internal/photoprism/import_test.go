package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestNewImport(t *testing.T) {
	cfg := config.TestConfig()

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	imp := NewImport(cfg, ind, convert)

	assert.IsType(t, &Import{}, imp)
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
	t.Run("StepsOverASymbolicLink", func(t *testing.T) {
		// A link holds the name whether or not it resolves, and the move that follows refuses one, so
		// a search that overlooked it would hand back a name the import can never take.
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
