package limiter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorize(t *testing.T) {
	t.Run("SeparateFromAuth", func(t *testing.T) {
		clientIp := "192.0.2.77"

		for i := 0; i < DefaultAuthLimit; i++ {
			Authorize.Reserve(clientIp)
		}

		assert.True(t, Authorize.Reject(clientIp))
		assert.False(t, Auth.Reject(clientIp))
	})
}
