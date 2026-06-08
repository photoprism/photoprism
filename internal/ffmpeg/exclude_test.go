package ffmpeg

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media/video"
)

func TestDefaultExclude(t *testing.T) {
	t.Run("ContainsMagicYUV", func(t *testing.T) {
		assert.NotEmpty(t, DefaultExclude)
		def := video.NewFormats(DefaultExclude)
		assert.True(t, def.Contains(video.CodecMagicYUV))
		assert.True(t, def.Contains(video.CodecVFW))
		assert.True(t, def.Contains("V_MS/VFW/FOURCC"))
	})
	t.Run("RoundTrips", func(t *testing.T) {
		// DefaultExclude must round-trip through NewFormats.String() so that
		// CLI defaults and config-report output stay in sync.
		assert.Equal(t, DefaultExclude, video.NewFormats(DefaultExclude).String())
	})
}

func TestExclude(t *testing.T) {
	// Snapshot and restore the package state so tests don't leak.
	saved := Exclude()
	t.Cleanup(func() { SetExclude(saved) })

	t.Run("DefaultBlocksMagicYUV", func(t *testing.T) {
		SetExclude(video.NewFormats(DefaultExclude))
		assert.True(t, Exclude().Contains(video.CodecMagicYUV))
		assert.False(t, Exclude().Contains(video.CodecAvc1))
	})
	t.Run("EmptyAllowsEverything", func(t *testing.T) {
		SetExclude(video.NewFormats(""))
		assert.False(t, Exclude().Contains(video.CodecMagicYUV))
	})
	t.Run("CustomList", func(t *testing.T) {
		SetExclude(video.NewFormats("avi, hap"))
		assert.True(t, Exclude().Contains("avi"))
		assert.True(t, Exclude().Contains("hap"))
		assert.False(t, Exclude().Contains(video.CodecMagicYUV))
	})
}

func TestExclude_Concurrent(t *testing.T) {
	// Hammer Exclude()/SetExclude() from multiple goroutines to verify
	// `go test -race` stays clean. Each Store publishes a distinct map.
	saved := Exclude()
	t.Cleanup(func() { SetExclude(saved) })

	const writers, readers, iters = 4, 8, 200
	var wg sync.WaitGroup

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				if seed%2 == j%2 {
					SetExclude(video.NewFormats("magicyuv"))
				} else {
					SetExclude(video.NewFormats("avi, hap"))
				}
			}
		}(i)
	}

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				_ = Exclude().Contains("magicyuv", "avi", "hap")
			}
		}()
	}

	wg.Wait()
}
