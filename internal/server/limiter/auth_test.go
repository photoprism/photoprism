package limiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	clientIp := "192.0.2.42"

	for i := 0; i < 59; i++ {
		t.Logf("tokens now: %f", Auth.IP(clientIp).TokensAt(time.Now()))
		assert.True(t, Auth.IP(clientIp).Allow())
	}

	assert.True(t, Auth.IP(clientIp).Allow())
	assert.False(t, Auth.IP(clientIp).Allow())
	assert.False(t, Auth.IP(clientIp).Allow())
	assert.False(t, Auth.IP(clientIp).Allow())

	t.Logf("tokens now: %f", Auth.IP(clientIp).TokensAt(time.Now()))
	t.Logf("tokens +1min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval)))
	t.Logf("tokens +2min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*2)))
	t.Logf("tokens +3min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*3)))
	t.Logf("tokens +4min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*4)))
	t.Logf("tokens +5min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*5)))
	t.Logf("tokens +10min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*10)))
	t.Logf("tokens +15min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*15)))
	t.Logf("tokens +20min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*20)))

	assert.InEpsilon(t, 1, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*1)), 0.1)
	assert.InEpsilon(t, 2, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*2)), 0.1)
	assert.InEpsilon(t, 3, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*3)), 0.1)
	assert.InEpsilon(t, 4, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*4)), 0.1)
	assert.InEpsilon(t, 5, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*5)), 0.1)
	assert.InEpsilon(t, 10, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*10)), 0.1)
	assert.InEpsilon(t, DefaultAuthLimit, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*DefaultAuthLimit*10)), 0.01)

	for i := 0; i < 30; i++ {
		assert.False(t, Auth.IP(clientIp).Allow())
	}

	assert.False(t, Auth.IP(clientIp).Allow())

	t.Logf("tokens now: %f", Auth.IP(clientIp).TokensAt(time.Now()))
	t.Logf("tokens +5min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*5)))
	t.Logf("tokens +10min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*10)))
	t.Logf("tokens +15min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*15)))
	t.Logf("tokens +20min: %f", Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*20)))

	assert.InEpsilon(t, 1, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*1)), 0.1)
	assert.InEpsilon(t, 2, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*2)), 0.1)
	assert.InEpsilon(t, 3, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*3)), 0.1)
	assert.InEpsilon(t, 4, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*4)), 0.1)
	assert.InEpsilon(t, 5, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*5)), 0.1)
	assert.InEpsilon(t, 10, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*10)), 0.1)
	assert.InEpsilon(t, DefaultAuthLimit, Auth.IP(clientIp).TokensAt(time.Now().Add(DefaultAuthInterval*DefaultAuthLimit*10)), 0.01)
}
