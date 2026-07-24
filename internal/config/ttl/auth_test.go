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
	t.Run("DownloadTokenMinAge", func(t *testing.T) {
		// Floor sits above the 10-minute client config-poll interval so a token cannot lapse before the
		// next refresh, and stays below the default lifetime.
		assert.Equal(t, Duration(900), DownloadTokenMinAge)
		assert.Greater(t, int(DownloadTokenMinAge), 600)
		assert.Less(t, int(DownloadTokenMinAge), DownloadToken.Int())
	})
}
