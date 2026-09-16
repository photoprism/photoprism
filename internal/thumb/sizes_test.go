package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxSize(t *testing.T) {
	SizeCached = 7680
	SizeOnDemand = 1024

	assert.Equal(t, MaxSize(), 7680)

	SizeCached = 2048
	SizeOnDemand = 7680
}

// TestMaxRenderSize covers the bound rendering obeys, which the face crop source may raise above
// the delivered sizes: it is rendered once per indexed file rather than per request.
func TestMaxRenderSize(t *testing.T) {
	cached, onDemand, faceSize := SizeCached, SizeOnDemand, SizeFace
	t.Cleanup(func() { SizeCached, SizeOnDemand, SizeFace = cached, onDemand, faceSize })

	t.Run("AboveWhatIsDelivered", func(t *testing.T) {
		SizeCached, SizeOnDemand, SizeFace = 720, 720, 4096

		assert.Equal(t, 720, MaxSize(), "what is delivered must not follow it")
		assert.Equal(t, 4096, MaxRenderSize())
		assert.False(t, InvalidSize(4096))
		assert.True(t, InvalidSize(4097))
		assert.True(t, Sizes[Fit4096].ExceedsLimit(), "a request must still be clamped to 720")
	})
	t.Run("BelowWhatIsDelivered", func(t *testing.T) {
		SizeCached, SizeOnDemand, SizeFace = 720, 7680, 0

		assert.Equal(t, 7680, MaxRenderSize())
	})
}

func TestSize_ExceedsLimit(t *testing.T) {
	SizeCached = 1024
	SizeOnDemand = 2048

	fit4096 := Sizes[Fit4096]
	assert.True(t, fit4096.ExceedsLimit())

	fit2048 := Sizes[Fit2048]
	assert.False(t, fit2048.ExceedsLimit())

	tile500 := Sizes[Tile500]
	assert.False(t, tile500.ExceedsLimit())

	SizeCached = 2048
	SizeOnDemand = 7680
}

func TestSize_Limit(t *testing.T) {
	t.Run("Larger", func(t *testing.T) {
		assert.Equal(t, Fit720, Sizes[Fit15360].Limit(SizeFit720).Name)
		assert.Equal(t, Fit720, Sizes[Tile1080].Limit(SizeFit720).Name)
	})
	t.Run("Smaller", func(t *testing.T) {
		assert.Equal(t, Tile500, Sizes[Tile500].Limit(SizeFit720).Name)
		assert.Equal(t, Tile50, Sizes[Tile50].Limit(SizeFit720).Name)
	})
	t.Run("Equal", func(t *testing.T) {
		assert.Equal(t, Fit720, Sizes[Fit720].Limit(SizeFit720).Name)
	})
}

func TestSize_Clamp(t *testing.T) {
	t.Cleanup(func() {
		SizeCached = 2048
		SizeOnDemand = 7680
	})

	setLimits := func(cached, onDemand int) {
		SizeCached, SizeOnDemand = cached, onDemand
	}

	t.Run("ExceedsLimit", func(t *testing.T) {
		setLimits(1024, 2048)
		assert.Equal(t, Fit1920, Sizes[Fit4096].Clamp().Name)
		assert.Equal(t, Fit1920, Sizes[Fit15360].Clamp().Name)
	})
	t.Run("WithinLimit", func(t *testing.T) {
		setLimits(1024, 2048)
		assert.Equal(t, Fit1920, Sizes[Fit1920].Clamp().Name)
		assert.Equal(t, Tile500, Sizes[Tile500].Clamp().Name)
	})
	t.Run("ResolvesToFitSize", func(t *testing.T) {
		// Names lists tile_500 before fit_720, so a plain lookup would return a center crop.
		setLimits(720, 720)
		assert.Equal(t, Fit720, Sizes[Fit1920].Clamp().Name)
		assert.Equal(t, Fit720, Sizes[Tile1080].Clamp().Name)
	})
	t.Run("EveryLimit", func(t *testing.T) {
		for _, limit := range []int{720, 1279, 1280, 1919, 2048, 4095, 7680, 15360} {
			setLimits(limit, limit)

			for name, size := range Sizes {
				clamped := size.Clamp()
				assert.NotEmpty(t, clamped.Name, "%s at %d", name, limit)
				assert.False(t, clamped.ExceedsLimit(), "%s at %d resolves to %s", name, limit, clamped.Name)
			}
		}
	})
	t.Run("NoRenderableSize", func(t *testing.T) {
		setLimits(1, 1)
		assert.Equal(t, Fit720, Sizes[Fit15360].Clamp().Name)
	})
}

func TestSize_Uncached(t *testing.T) {
	SizeCached = 1024
	SizeOnDemand = 2048

	fit4096 := Sizes[Fit4096]
	assert.True(t, fit4096.Uncached())

	fit2048 := Sizes[Fit2048]
	assert.True(t, fit2048.Uncached())

	tile500 := Sizes[Tile500]
	assert.False(t, tile500.Uncached())

	SizeCached = 2048
	SizeOnDemand = 7680
}
