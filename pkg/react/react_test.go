package react

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	assert.Equal(t, Cheers, Find("cheers"))
	assert.Equal(t, CatLove, Find("cat-love"))
	assert.Equal(t, Emoji(""), Find("alien"))
}

func TestKnown(t *testing.T) {
	assert.True(t, Known("🥂"))
	assert.True(t, Known("😻"))
	assert.False(t, Known("👽"))
}
