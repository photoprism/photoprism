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
	assert.False(t, IsAuthToken("MPkOqm-RtKGOi-ctIvXm-Qv3XhN"))
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

func TestAppPassword(t *testing.T) {
	for n := 0; n < 10; n++ {
		s := AppPassword()
		t.Logf("AppPassword %d: %s", n, s)
		assert.Equal(t, AppPasswordLength, len(s))
	}
}

func BenchmarkAppPassword(b *testing.B) {
	for n := 0; n < b.N; n++ {
		AppPassword()
	}
}

func TestIsAppPasswordt(t *testing.T) {
	t.Run("VerifyChecksum", func(t *testing.T) {
		assert.True(t, IsAppPassword("MPkOqm-RtKGOi-ctIvXm-Qv3XhN", true))
		assert.True(t, IsAppPassword("9q2JHc-P0LzNE-xzvY9j-vMoefj", true))
		assert.False(t, IsAppPassword("MPkOqm-RtKGOi-ctIvXm-Qv3Xha", true))
		assert.False(t, IsAppPassword("9q2JHc-P0LzNE-xzvY9j-vMoef2", true))
		assert.True(t, IsAppPassword(AppPassword(), true))
		assert.True(t, IsAppPassword(AppPassword(), true))
		assert.False(t, IsAppPassword(AuthToken(), true))
		assert.False(t, IsAppPassword(AuthToken(), true))
		assert.False(t, IsAppPassword(SessionID(AuthToken()), true))
		assert.False(t, IsAppPassword(SessionID(AuthToken()), true))
		assert.False(t, IsAppPassword("55785BAC-9H4B-4747-B090-EE123FFEE437", true))
		assert.False(t, IsAppPassword("4B1FEF2D1CF4A5BE38B263E0637EDEAD", true))
		assert.False(t, IsAppPassword("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2", true))
		assert.False(t, IsAppPassword("", true))
	})
	t.Run("IgnoreChecksum", func(t *testing.T) {
		assert.True(t, IsAppPassword("MPkOqm-RtKGOi-ctIvXm-Qv3XhN", false))
		assert.True(t, IsAppPassword("9q2JHc-P0LzNE-xzvY9j-vMoefj", false))
		assert.True(t, IsAppPassword("MPkOqm-RtKGOi-ctIvXm-Qv3Xha", false))
		assert.True(t, IsAppPassword("9q2JHc-P0LzNE-xzvY9j-vMoef2", false))
		assert.True(t, IsAppPassword(AppPassword(), false))
		assert.True(t, IsAppPassword(AppPassword(), false))
		assert.False(t, IsAppPassword(AuthToken(), false))
		assert.False(t, IsAppPassword(AuthToken(), false))
		assert.False(t, IsAppPassword(SessionID(AuthToken()), false))
		assert.False(t, IsAppPassword(SessionID(AuthToken()), false))
		assert.False(t, IsAppPassword("55785BAC-9H4B-4747-B090-EE123FFEE437", false))
		assert.False(t, IsAppPassword("4B1FEF2D1CF4A5BE38B263E0637EDEAD", false))
		assert.False(t, IsAppPassword("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2", false))
		assert.False(t, IsAppPassword("", false))
	})
}

func BenchmarkAppPasswordtVerifyChecksum(b *testing.B) {
	for n := 0; n < b.N; n++ {
		IsAppPassword("MPkOqm-RtKGOi-ctIvXm-Qv3XhN", true)
	}
}

func BenchmarkAppPasswordIgnoreChecksum(b *testing.B) {
	for n := 0; n < b.N; n++ {
		IsAppPassword("MPkOqm-RtKGOi-ctIvXm-Qv3XhN", false)
	}
}

func TestIsAuthAny(t *testing.T) {
	assert.True(t, IsAuthAny("MPkOqm-RtKGOi-ctIvXm-Qv3XhN"))
	assert.True(t, IsAuthAny("9q2JHc-P0LzNE-xzvY9j-vMoefj"))
	assert.True(t, IsAuthAny(AppPassword()))
	assert.True(t, IsAuthAny(AppPassword()))
	assert.True(t, IsAuthAny(AuthToken()))
	assert.True(t, IsAuthAny(AuthToken()))
	assert.False(t, IsAuthAny(SessionID(AuthToken())))
	assert.False(t, IsAuthAny(SessionID(AuthToken())))
	assert.False(t, IsAuthAny("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.False(t, IsAuthAny("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.True(t, IsAuthAny("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2"))
	assert.False(t, IsAuthAny(""))
}

func BenchmarkIsAuthAny(b *testing.B) {
	for n := 0; n < b.N; n++ {
		IsAuthAny("MPkOqm-RtKGOi-ctIvXm-Qv3XhN")
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
