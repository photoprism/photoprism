package workers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
)

// TestNsfwPrivateFlag verifies that unavailable decisions preserve the private flag.
func TestNsfwPrivateFlag(t *testing.T) {
	t.Run("UnsafeFlagsPublicPhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(false, nsfw.NewResult(0.99, nsfw.DefaultThreshold))
		assert.True(t, flag)
		assert.True(t, write)
	})
	t.Run("UnsafeKeepsPrivatePhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(true, nsfw.NewResult(0.99, nsfw.DefaultThreshold))
		assert.True(t, flag)
		assert.False(t, write, "nothing to write when the flag already matches")
	})
	t.Run("SafeClearsPrivatePhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(true, nsfw.NewResult(0.01, nsfw.DefaultThreshold))
		assert.False(t, flag)
		assert.True(t, write, "a decided safe result must still be able to clear the flag")
	})
	t.Run("SafeKeepsPublicPhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(false, nsfw.NewResult(0.01, nsfw.DefaultThreshold))
		assert.False(t, flag)
		assert.False(t, write)
	})
	t.Run("UnavailableKeepsPrivatePhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(true, nsfw.Unavailable("model is missing"))
		assert.True(t, flag, "an undecided result must not un-private a photo")
		assert.False(t, write)
	})
	t.Run("UnavailableKeepsPublicPhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(false, nsfw.Unavailable("thumbnail is missing"))
		assert.False(t, flag)
		assert.False(t, write)
	})
	// The zero value reaches this function whenever a caller forgets to decide, so it has to
	// behave like any other undecided result rather than like a clearance.
	t.Run("ZeroResultKeepsPrivatePhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(true, nsfw.Result{})
		assert.True(t, flag)
		assert.False(t, write)
	})
}

// TestLabelsPrivateFlag verifies the shared worker policy never applies labels outside labels mode.
func TestLabelsPrivateFlag(t *testing.T) {
	previous := vision.Config
	vision.Config = vision.NewConfig()
	t.Cleanup(func() { vision.Config = previous })
	labels := classify.Labels{{Name: "custom prompt", NSFW: true}}
	for _, mode := range []string{"auto", "none", "labels"} {
		t.Run(mode, func(t *testing.T) {
			conf := config.NewMinimalTestConfig(t.TempDir())
			conf.Options().NsfwModel = mode
			conf.Options().DetectNSFW = true
			flag, write := labelsPrivateFlag(conf, false, labels)
			assert.Equal(t, mode == "labels", flag)
			assert.Equal(t, mode == "labels", write)
			flag, write = labelsPrivateFlag(conf, true, labels)
			assert.True(t, flag)
			assert.False(t, write)
			conf.Options().DetectNSFW = false
			flag, write = labelsPrivateFlag(conf, false, labels)
			assert.False(t, flag)
			assert.False(t, write)
		})
	}
	conf := config.NewMinimalTestConfig(t.TempDir())
	conf.Options().NsfwModel, conf.Options().DetectNSFW = "labels", true
	flag, write := labelsPrivateFlag(conf, false, nil)
	assert.False(t, flag)
	assert.False(t, write)
	t.Run("LabelsThreshold", func(t *testing.T) {
		index := 90
		vision.Config.Thresholds.NSFW = 40
		vision.Config.Thresholds.NSFWIndex = &index
		t.Cleanup(func() { vision.Config.Thresholds = vision.DefaultThresholds })
		scored := classify.Labels{{Name: "beach", NSFWConfidence: 50}}
		flag, write := labelsPrivateFlag(conf, false, scored)
		assert.True(t, flag)
		assert.True(t, write)
		vision.Config.Thresholds.NSFW = 60
		flag, write = labelsPrivateFlag(conf, false, scored)
		assert.False(t, flag)
		assert.False(t, write)
	})
	vision.Config = nil
	flag, write = labelsPrivateFlag(conf, false, labels)
	assert.False(t, flag)
	assert.False(t, write)
}
