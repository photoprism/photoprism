package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthToken(t *testing.T) {
	result := AuthToken()
	assert.Equal(t, AuthTokenLength, len(result))
	assert.True(t, IsAuthToken(result))
	assert.True(t, IsHex(result))

	for n := 0; n < 10; n++ {
		s := AuthToken()
		t.Logf("AuthToken %d: %s", n, s)
		assert.NotEmpty(t, s)
	}
}

func BenchmarkAuthToken(b *testing.B) {
	for n := 0; n < b.N; n++ {
		AuthToken()
	}
}

func TestIsAuthToken(t *testing.T) {
	assert.True(t, IsAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2"))
	assert.True(t, IsAuthToken(AuthToken()))
	assert.True(t, IsAuthToken(AuthToken()))
	assert.False(t, IsAuthToken(SessionID(AuthToken())))
	assert.False(t, IsAuthToken(SessionID(AuthToken())))
	assert.False(t, IsAuthToken("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.False(t, IsAuthToken("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.False(t, IsAuthToken(""))
}

func BenchmarkIsAuthToken(b *testing.B) {
	for n := 0; n < b.N; n++ {
		IsAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2")
	}
}

func TestAuthSecret(t *testing.T) {
	for n := 0; n < 10; n++ {
		s := AuthSecret()
		t.Logf("AuthSecret %d: %s", n, s)
		assert.Equal(t, AuthSecretLength, len(s))
	}
}

func BenchmarkAuthSecret(b *testing.B) {
	for n := 0; n < b.N; n++ {
		AuthSecret()
	}
}

func TestIsAuthSecret(t *testing.T) {
	assert.True(t, IsAuthSecret("f64nl-GQbmZ-CEcLr-Q1VSr"))
	assert.True(t, IsAuthSecret("XRhOW-DM2ol-INove-dAg6m"))
	assert.True(t, IsAuthSecret(AuthSecret()))
	assert.True(t, IsAuthSecret(AuthSecret()))
	assert.False(t, IsAuthSecret(AuthToken()))
	assert.False(t, IsAuthSecret(AuthToken()))
	assert.False(t, IsAuthSecret(SessionID(AuthToken())))
	assert.False(t, IsAuthSecret(SessionID(AuthToken())))
	assert.False(t, IsAuthSecret("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.False(t, IsAuthSecret("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.False(t, IsAuthSecret("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2"))
	assert.False(t, IsAuthSecret(""))
}

func BenchmarkIsAuthSecret(b *testing.B) {
	for n := 0; n < b.N; n++ {
		IsAuthSecret("f64nl-GQbmZ-CEcLr-Q1VSr")
	}
}

func TestSessionID(t *testing.T) {
	result := SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2")
	assert.Equal(t, SessionIdLength, len(result))
	assert.Equal(t, "f22383a703805a031a9835c8c6b6dafb793a21e8f33d0b4887b4ec9bd7ac8cd5", result)

	for n := 0; n < 10; n++ {
		s := SessionID(AuthToken())
		t.Logf("SessionID %d: %s", n, s)
		assert.NotEmpty(t, s)
	}
}

func TestIsSessionID(t *testing.T) {
	assert.False(t, IsSessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2"))
	assert.False(t, IsSessionID(AuthToken()))
	assert.False(t, IsSessionID(AuthToken()))
	assert.True(t, IsSessionID(SessionID(AuthToken())))
	assert.True(t, IsSessionID(SessionID(AuthToken())))
	assert.False(t, IsSessionID("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.False(t, IsSessionID("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.False(t, IsSessionID(""))
}
