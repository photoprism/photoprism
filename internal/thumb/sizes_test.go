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
