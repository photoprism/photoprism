package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestCleanupAction(t *testing.T) {
	t.Run("ExpiredAuthCodes", func(t *testing.T) {
		m := &entity.AuthCode{
			CodeHash:  rnd.Sha256([]byte(rnd.AuthToken())),
			CodeType:  entity.AuthCodeCode,
			ExpiresAt: time.Now().UTC().Add(-time.Minute),
		}
		require.NoError(t, m.Create())

		CleanupAction()

		found, err := entity.FindAuthCode(m.CodeHash, entity.AuthCodeCode)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}
