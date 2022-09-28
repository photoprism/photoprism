package react

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReactions(t *testing.T) {
	assert.Equal(t, Heart, Reactions["heart"])
}

func TestReaction_Name(t *testing.T) {
	assert.Equal(t, "heart", Heart.Name())
	assert.Equal(t, "cheers", Emoji("🥂").Name())
	assert.Equal(t, "cat-love", Emoji("😻").Name())
	assert.Equal(t, "see-no-evil", Reactions["see-no-evil"].Name())
}

func TestReaction_Unknown(t *testing.T) {
	assert.True(t, Unknown.Unknown())
	assert.True(t, Emoji("A").Unknown())
	assert.True(t, Emoji("23").Unknown())
	assert.False(t, Heart.Unknown())
	assert.False(t, Emoji("🥂").Unknown())
	assert.False(t, Emoji("😻").Unknown())
	assert.False(t, Reactions["see-no-evil"].Unknown())
}

func TestReaction_String(t *testing.T) {
	assert.Equal(t, "❤️", Heart.String())
}

func TestReaction_Bytes(t *testing.T) {
	assert.Equal(t, []byte{0xe2, 0x9d, 0xa4, 0xef, 0xb8, 0x8f}, Heart.Bytes())
}
