package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase36(t *testing.T) {
	token := Base36(10)
	t.Logf("Base36 Token: %s", token)
	assert.NotEmpty(t, token)
	assert.True(t, IsRefID(token))
	assert.False(t, InvalidRefID(token))
	assert.Equal(t, 10, len(token))

	for n := 0; n < 10; n++ {
		token = Base36(10)
		t.Logf("Base36 %d: %s", n, token)
		assert.NotEmpty(t, token)
	}
}

func TestBase62(t *testing.T) {
	token := Base62(23)
	t.Logf("Base62 Token: %s", token)
	assert.NotEmpty(t, token)
	assert.False(t, IsRefID(token))
	assert.True(t, InvalidRefID(token))
	assert.Equal(t, 23, len(token))

	for n := 0; n < 10; n++ {
		token = Base62(10)
		t.Logf("Base62 %d: %s", n, token)
		assert.NotEmpty(t, token)
	}
}

func TestCharset(t *testing.T) {
	token := Charset(23, CharsetBase62)
	t.Logf("Charset Token: %s", token)
	assert.NotEmpty(t, token)
	assert.False(t, IsRefID(token))
	assert.True(t, InvalidRefID(token))
	assert.Equal(t, 23, len(token))
}
