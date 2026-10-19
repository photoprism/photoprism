package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/oauth"
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
	t.Run("IdleGrants", func(t *testing.T) {
		idle := &entity.Grant{UserUID: "uqxetse3cy5eo9z2", ClientID: "https://claude.ai/idle.json", RefreshHash: "hash", UsedAt: time.Now().UTC().Add(-oauth.GrantMaxIdle - time.Hour)}
		used := &entity.Grant{UserUID: "uqxetse3cy5eo9z2", ClientID: "https://claude.ai/used.json", RefreshHash: "hash"}
		require.NoError(t, idle.Create())
		require.NoError(t, used.Create())
		t.Cleanup(func() { _ = idle.Delete(); _ = used.Delete() })

		CleanupAction()

		assert.Nil(t, entity.FindGrant(idle.GrantUID))
		assert.NotNil(t, entity.FindGrant(used.GrantUID))
	})
}
