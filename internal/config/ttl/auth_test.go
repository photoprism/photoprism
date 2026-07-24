package ttl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthTokenDefaults(t *testing.T) {
	t.Run("DownloadToken", func(t *testing.T) {
		// Kept short (one hour) so a leaked token expires quickly.
		assert.Equal(t, Duration(3600), DownloadToken)
	})
}
