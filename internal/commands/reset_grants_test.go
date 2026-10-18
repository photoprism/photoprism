package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// seedAppGrant stores an app grant and an authorization code, so a test can check that a reset removes them.
func seedAppGrant(t *testing.T) {
	t.Helper()

	grant := &entity.Grant{UserUID: entity.UserFixtures.Get("alice").UserUID, ClientID: "https://claude.ai/reset.json", RefreshHash: "hash"}
	require.NoError(t, grant.Create())
	require.NoError(t, (&entity.AuthCode{CodeHash: rnd.AuthToken(), CodeType: entity.AuthCodeCode, ExpiresAt: time.Now().Add(time.Minute)}).Create())
}

// countAppGrants returns the number of stored app grants and authorization codes.
func countAppGrants(t *testing.T, db *gorm.DB) (grants, codes int64) {
	t.Helper()

	require.NoError(t, db.Model(&entity.Grant{}).Count(&grants).Error)
	require.NoError(t, db.Model(&entity.AuthCode{}).Count(&codes).Error)

	return grants, codes
}

func TestResetAppGrants(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := resetConfigAndOpenDB(t)
		db := c.Db()

		seedAppGrant(t)
		require.NoError(t, ResetAppGrants(db))

		grants, codes := countAppGrants(t, db)
		assert.Zero(t, grants)
		assert.Zero(t, codes)
	})
	t.Run("ClosedDatabase", func(t *testing.T) {
		c := resetConfigAndOpenDB(t)
		db := c.Db()
		sqlDb, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDb.Close())

		assert.Error(t, ResetAppGrants(db))
	})
}
