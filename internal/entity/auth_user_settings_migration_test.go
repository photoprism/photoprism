package entity

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// errRollbackSettings rolls back the transaction of a settings migration test.
var errRollbackSettings = errors.New("rollback")

// TestUserSettings_OwnerMigration checks that the SQLite migration removes settings rows without an
// account and keeps the rows of every existing account, whatever storage class their UID has.
func TestUserSettings_OwnerMigration(t *testing.T) {
	conn, ok := dbConn.(*DbConn)
	require.True(t, ok, "test database provider")

	if conn.Driver != dsn.DriverSQLite3 {
		t.Skip("settings rows without an account are removed on SQLite only")
	}

	var statements []string

	for _, m := range migrate.Dialects[conn.Driver] {
		if m.ID == "20261012-000001" {
			statements = m.Statements
		}
	}

	require.NotEmpty(t, statements)

	alice := UserFixtures.Pointer("alice").UserUID
	bob := UserFixtures.Pointer("bob").UserUID
	deleted := UserFixtures.Pointer("deleted").UserUID

	// owned lists every account UID together with the bytes and storage class of its settings rows.
	owned := func(t *testing.T, tx *gorm.DB) map[string][]string {
		var uids []string
		require.NoError(t, tx.Raw("SELECT user_uid FROM auth_users WHERE LENGTH(user_uid) > 0").Scan(&uids).Error)

		result := make(map[string][]string, len(uids))

		for _, uid := range uids {
			var rows []string
			require.NoError(t, tx.Raw("SELECT typeof(user_uid) || ':' || hex(user_uid) || ':' || hex(ui_language) FROM auth_users_settings WHERE hex(user_uid) = hex(?)", uid).Scan(&rows).Error)
			result[uid] = rows
		}

		return result
	}

	// orphaned counts the settings rows whose UID is missing, empty, or matches no account byte for byte.
	orphaned := func(t *testing.T, tx *gorm.DB) (n int64) {
		require.NoError(t, tx.Raw("SELECT COUNT(*) FROM auth_users_settings s WHERE s.user_uid IS NULL OR LENGTH(s.user_uid) = 0 OR NOT EXISTS (SELECT 1 FROM auth_users u WHERE hex(u.user_uid) = hex(s.user_uid))").Scan(&n).Error)
		return n
	}

	err := UnscopedDb().Transaction(func(tx *gorm.DB) error {
		// Store the settings UIDs of alice and the deleted user as BLOB, and the account UIDs of bob and the visitor.
		require.NoError(t, tx.Exec("UPDATE auth_users_settings SET user_uid = CAST(user_uid AS BLOB), ui_language = 'nl' WHERE user_uid = ?", alice).Error)
		require.NoError(t, tx.Exec("UPDATE auth_users_settings SET user_uid = CAST(user_uid AS BLOB), ui_language = 'fr' WHERE user_uid = ?", deleted).Error)
		require.NoError(t, tx.Exec("UPDATE auth_users SET user_uid = CAST(user_uid AS BLOB) WHERE user_uid IN (?, ?)", bob, Visitor.UserUID).Error)

		// Add accounts without a UID, which own no settings.
		require.NoError(t, tx.Exec("INSERT INTO auth_users (id, user_uid, user_name) VALUES (-90, NULL, 'null-uid'), (-91, '', 'empty-uid')").Error)

		for _, uid := range []string{"NULL", "NULL", "''", "x''", "char(1)", "char(3)", "'uxxxxxxxxxxxxxxx'", "CAST('uyyyyyyyyyyyyyyy' AS BLOB)"} {
			require.NoError(t, tx.Exec("INSERT INTO auth_users_settings (user_uid, ui_language) VALUES ("+uid+", 'de')").Error, uid)
		}

		before := owned(t, tx)

		for _, uid := range []string{alice, bob, deleted, UnknownUser.UserUID, Visitor.UserUID} {
			require.Len(t, before[uid], 1, uid)
		}

		assert.Contains(t, before[alice][0], "blob:")
		assert.Contains(t, before[deleted][0], "blob:")
		assert.Contains(t, before[bob][0], "text:")
		require.GreaterOrEqual(t, orphaned(t, tx), int64(8))

		for i := range 2 {
			for _, s := range statements {
				res := tx.Exec(s)
				require.NoError(t, res.Error)

				if i > 0 {
					assert.Equal(t, int64(0), res.RowsAffected, "second run")
				}
			}

			assert.Equal(t, int64(0), orphaned(t, tx))
			assert.Equal(t, before, owned(t, tx))
		}

		return errRollbackSettings
	})

	require.ErrorIs(t, err, errRollbackSettings)
}
