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
	conf.Options().FaceEngineThreads = 4

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

func TestNewIndexOptions_FacesOnlyOverridesSchedulers(t *testing.T) {
	conf := config.NewMinimalTestConfig(t.TempDir())
	conf.Options().FaceEngineThreads = 1

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
