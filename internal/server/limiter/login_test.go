package limiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	clientIp := "192.0.2.42"

	for i := 0; i < 9; i++ {
		t.Logf("tokens now: %f", Login.IP(clientIp).TokensAt(time.Now()))
		assert.True(t, Login.IP(clientIp).Allow())
	}

	assert.True(t, Login.IP(clientIp).Allow())
	assert.False(t, Login.IP(clientIp).Allow())
	assert.False(t, Login.IP(clientIp).Allow())
	assert.False(t, Login.IP(clientIp).Allow())

	t.Logf("tokens now: %f", Login.IP(clientIp).TokensAt(time.Now()))
	t.Logf("tokens +1min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval)))
	t.Logf("tokens +2min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*2)))
	t.Logf("tokens +3min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*3)))
	t.Logf("tokens +4min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*4)))
	t.Logf("tokens +5min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*5)))
	t.Logf("tokens +10min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*10)))
	t.Logf("tokens +15min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*15)))
	t.Logf("tokens +20min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*20)))

	assert.InEpsilon(t, 1, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*1)), 0.1)
	assert.InEpsilon(t, 2, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*2)), 0.1)
	assert.InEpsilon(t, 3, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*3)), 0.1)
	assert.InEpsilon(t, 4, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*4)), 0.1)
	assert.InEpsilon(t, 5, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*5)), 0.1)
	assert.InEpsilon(t, 10, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*10)), 0.1)
	assert.InEpsilon(t, 10, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*20)), 0.01)
	assert.InEpsilon(t, DefaultLoginLimit, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*DefaultLoginLimit*10)), 0.01)

	for i := 0; i < 30; i++ {
		assert.False(t, Login.IP(clientIp).Allow())
	}

	assert.False(t, Login.IP(clientIp).Allow())

	t.Logf("tokens now: %f", Login.IP(clientIp).TokensAt(time.Now()))
	t.Logf("tokens +5min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*5)))
	t.Logf("tokens +10min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*10)))
	t.Logf("tokens +15min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*15)))
	t.Logf("tokens +20min: %f", Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*20)))

	assert.InEpsilon(t, 1, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*1)), 0.1)
	assert.InEpsilon(t, 2, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*2)), 0.1)
	assert.InEpsilon(t, 3, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*3)), 0.1)
	assert.InEpsilon(t, 4, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*4)), 0.1)
	assert.InEpsilon(t, 5, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*5)), 0.1)
	assert.InEpsilon(t, 10, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*10)), 0.1)
	assert.InEpsilon(t, 10, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*20)), 0.01)
	assert.InEpsilon(t, DefaultLoginLimit, Login.IP(clientIp).TokensAt(time.Now().Add(DefaultLoginInterval*DefaultLoginLimit*10)), 0.01)
}
