package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
)

func TestIndexOptionsNone(t *testing.T) {
	opt := IndexOptionsNone(nil)

	assert.Equal(t, "", opt.Path)
	assert.Equal(t, false, opt.Rescan)
	assert.Equal(t, false, opt.Convert)
	assert.Equal(t, false, opt.Stack)
	assert.Equal(t, false, opt.FacesOnly)
}

func TestResolveIndexPath(t *testing.T) {
	base := "/photoprism/originals"

	t.Run("Root", func(t *testing.T) {
		got, err := ResolveIndexPath(base, "/")
		require.NoError(t, err)
		assert.Equal(t, base, got)
	})
	t.Run("Empty", func(t *testing.T) {
		got, err := ResolveIndexPath(base, "")
		require.NoError(t, err)
		assert.Equal(t, base, got)
	})
	t.Run("Dot", func(t *testing.T) {
		got, err := ResolveIndexPath(base, ".")
		require.NoError(t, err)
		assert.Equal(t, base, got)
	})
	t.Run("Subfolder", func(t *testing.T) {
		got, err := ResolveIndexPath(base, "2020/03")
		require.NoError(t, err)
		assert.Equal(t, base+"/2020/03", got)
	})
	t.Run("LeadingSlashSubfolder", func(t *testing.T) {
		got, err := ResolveIndexPath(base, "/2020/03")
		require.NoError(t, err)
		assert.Equal(t, base+"/2020/03", got)
	})
	t.Run("DottedSubfolderAllowed", func(t *testing.T) {
		got, err := ResolveIndexPath(base, "2020/.hidden")
		require.NoError(t, err)
		assert.Equal(t, base+"/2020/.hidden", got)
	})
	t.Run("Traversal", func(t *testing.T) {
		_, err := ResolveIndexPath(base, "../../outside")
		assert.Error(t, err)
	})
	t.Run("InteriorTraversal", func(t *testing.T) {
		_, err := ResolveIndexPath(base, "2020/../../../outside")
		assert.Error(t, err)
	})
	t.Run("LeadingSlashRootsUnderOriginals", func(t *testing.T) {
		// A leading slash is rooted at originals (like filepath.Join), so an
		// "absolute" input stays within the base directory rather than escaping.
		got, err := ResolveIndexPath(base, "/sub/file")
		require.NoError(t, err)
		assert.Equal(t, base+"/sub/file", got)
	})
}

func TestIndexOptions_SkipUnchanged(t *testing.T) {
	opt := IndexOptionsNone(nil)

	assert.True(t, opt.SkipUnchanged())

	opt.Rescan = true

	assert.False(t, opt.SkipUnchanged())
}

func TestIndexOptionsSingle(t *testing.T) {
	opt := IndexOptionsSingle(nil)

	assert.Equal(t, false, opt.Stack)
	assert.Equal(t, true, opt.Convert)
	assert.Equal(t, true, opt.Rescan)
}

func TestIndexOptionsFacesOnly(t *testing.T) {
	opt := IndexOptionsFacesOnly(nil)

	assert.Equal(t, "/", opt.Path)
	assert.Equal(t, true, opt.Rescan)
	assert.Equal(t, true, opt.Convert)
	assert.Equal(t, true, opt.Stack)
	assert.Equal(t, true, opt.FacesOnly)
}

func TestNewIndexOptions_DefaultDetectors(t *testing.T) {
	conf := config.NewMinimalTestConfig(t.TempDir())
	conf.Options().FaceModelThreads = 4

	prevVision := vision.Config
	vision.Config = vision.NewConfig()
	t.Cleanup(func() {
		vision.Config = prevVision
	})

	opts := NewIndexOptions("/", true, true, true, false, false, conf)

	require.True(t, opts.DetectFaces, "face detection should run when enough threads are available")
	assert.True(t, opts.GenerateLabels)
	assert.True(t, opts.DetectNsfw)
}

func TestNewIndexOptions_ImportFaceTags(t *testing.T) {
	conf := config.NewMinimalTestConfig(t.TempDir())

	t.Run("Enabled", func(t *testing.T) {
		conf.Options().XMPFaces = true
		opts := NewIndexOptions("/", true, true, true, false, false, conf)
		assert.True(t, opts.ImportFaceTags)
	})
	t.Run("Disabled", func(t *testing.T) {
		conf.Options().XMPFaces = false
		opts := NewIndexOptions("/", true, true, true, false, false, conf)
		assert.False(t, opts.ImportFaceTags)
	})
}

func TestNewIndexOptions_FacesOnlyOverridesSchedulers(t *testing.T) {
	conf := config.NewMinimalTestConfig(t.TempDir())
	conf.Options().FaceModelThreads = 1

	prevVision := vision.Config
	vision.Config = vision.NewConfig()
	t.Cleanup(func() {
		vision.Config = prevVision
	})

	opts := NewIndexOptions("/", true, true, true, true, false, conf)

	require.True(t, opts.DetectFaces, "faces-only runs must always detect faces")
	assert.False(t, opts.GenerateLabels)
	assert.False(t, opts.DetectNsfw)
}

func TestNewIndexOptions_DisabledModels(t *testing.T) {
	conf := config.NewMinimalTestConfig(t.TempDir())
	conf.Options().DetectNSFW = false
	conf.Options().DisableFaces = true

	prevVision := vision.Config
	vision.Config = &vision.ConfigValues{
		Models: vision.Models{
			&vision.Model{Type: vision.ModelTypeLabels, Run: string(vision.RunManual)},
			&vision.Model{Type: vision.ModelTypeNsfw, Run: string(vision.RunManual)},
		},
	}
	t.Cleanup(func() {
		vision.Config = prevVision
	})

	opts := NewIndexOptions("/", true, true, true, false, false, conf)

	assert.False(t, opts.DetectFaces)
	assert.False(t, opts.GenerateLabels)
	assert.False(t, opts.DetectNsfw)
}
