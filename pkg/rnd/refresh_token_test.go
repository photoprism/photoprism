package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefreshToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		uid := GenerateUID('g')
		token := RefreshToken(uid)
		assert.Len(t, token, len(uid)+1+AuthTokenLength)
		assert.NotEqual(t, token, RefreshToken(uid))
	})
}

func TestParseRefreshToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		uid := GenerateUID('g')
		token := RefreshToken(uid)
		grantUID, secret, ok := ParseRefreshToken(token, 'g')
		assert.True(t, ok)
		assert.Equal(t, uid, grantUID)
		assert.Equal(t, token[len(uid)+1:], secret)
	})
	t.Run("WrongPrefix", func(t *testing.T) {
		_, _, ok := ParseRefreshToken(RefreshToken(GenerateUID('u')), 'g')
		assert.False(t, ok)
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, s := range []string{"", "abc", GenerateUID('g'), GenerateUID('g') + ".", GenerateUID('g') + ".xyz", AuthToken()} {
			_, _, ok := ParseRefreshToken(s, 'g')
			assert.False(t, ok, s)
		}
	})
	t.Run("NotAnAuthToken", func(t *testing.T) {
		token := RefreshToken(GenerateUID('g'))
		assert.False(t, IsAuthAny(token))
		assert.False(t, IsRefID(token))
		assert.False(t, IsSessionID(token))
	})
}
