package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVideoSize(t *testing.T) {
	assert.Equal(t, Sizes[Fit720], VideoSize(720))
	assert.Equal(t, Sizes[Fit720], VideoSize(1279))
	assert.Equal(t, Sizes[Fit1280], VideoSize(1280))
	assert.Equal(t, Sizes[Fit1280], VideoSize(1281))
	assert.Equal(t, Sizes[Fit1920], VideoSize(1920))
	assert.Equal(t, Sizes[Fit1920], VideoSize(2000))
	assert.Equal(t, Sizes[Fit2048], VideoSize(2048))
	assert.Equal(t, Sizes[Fit2560], VideoSize(3000))
	assert.Equal(t, Sizes[Fit3840], VideoSize(0))
	assert.Equal(t, Sizes[Fit3840], VideoSize(4000))
	assert.Equal(t, Sizes[Fit7680], VideoSize(8000))
	assert.Equal(t, Sizes[Fit7680], VideoSize(-1))
}
