package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	t.Run("Auth", func(t *testing.T) {
		assert.Equal(t, "X-Auth-Token", XAuthToken)
		assert.Equal(t, "X-Session-ID", XSessionID)
		assert.Equal(t, "Authorization", Auth)
		assert.Equal(t, "Basic", AuthBasic)
		assert.Equal(t, "Bearer", AuthBearer)
	})
}
