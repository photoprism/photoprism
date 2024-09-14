package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientSecret(t *testing.T) {
	result := ClientSecret()
	assert.Equal(t, ClientSecretLength, len(result))
	assert.NotEqual(t, AuthTokenLength, len(result))
	assert.True(t, IsClientSecret(result))
	assert.False(t, IsAuthToken(result))
	assert.False(t, IsHex(result))

	for n := 0; n < 10; n++ {
		s := ClientSecret()
		t.Logf("ClientSecret %d: %s", n, s)
		assert.True(t, IsClientSecret(s))
	}
}

func BenchmarkClientSecret(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ClientSecret()
	}
}

func TestIsClientSecret(t *testing.T) {
	assert.True(t, IsClientSecret(ClientSecret()))
	assert.True(t, IsClientSecret("69be27ac5ca305b394046a83f6fda181"))
	assert.False(t, IsClientSecret("MPkOqm-RtKGOi-ctIvXm-Qv3XhN"))
	assert.False(t, IsClientSecret("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2"))
	assert.False(t, IsClientSecret(AuthToken()))
	assert.False(t, IsClientSecret(AuthToken()))
	assert.False(t, IsClientSecret(SessionID(AuthToken())))
	assert.False(t, IsClientSecret(SessionID(AuthToken())))
	assert.False(t, IsClientSecret("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.True(t, IsClientSecret("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.False(t, IsClientSecret(""))
}

func BenchmarkIsClientSecret(b *testing.B) {
	for n := 0; n < b.N; n++ {
		IsClientSecret("69be27ac5ca305b394046a83f6fda181")
	}
}
