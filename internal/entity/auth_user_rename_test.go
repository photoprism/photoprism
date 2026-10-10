package entity

import (
	"errors"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
	"github.com/photoprism/photoprism/pkg/txt"
)

// newRenameTestUser stores an account with the given name and role, and removes it and its sessions
// and clients after the test.
func newRenameTestUser(t *testing.T, name string, role acl.Role) *User {
	t.Helper()

	m := NewUser()
	m.UserName = name
	m.UserRole = role.String()

	require.NoError(t, m.Create())
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(&Session{}, "user_uid = ?", m.UserUID).Error)
		assert.NoError(t, UnscopedDb().Delete(&Client{}, "user_uid = ?", m.UserUID).Error)
		deleteTestUser(t, m)
	})

	return m
}

// newRenameTestSession stores a session of the account.
func newRenameTestSession(t *testing.T, m *User) *Session {
	t.Helper()

	sess, err := AddClientSession("rename-test", unix.Hour, "*", authn.GrantPassword, m)
	require.NoError(t, err)

	return sess
}

// newRenameTestClient stores a client of the account and deletes it if deleted is true.
func newRenameTestClient(t *testing.T, m *User, deleted bool) *Client {
	t.Helper()

	c := NewClient()
	c.ClientName = "rename-test"
	c.SetUser(m)
	require.NoError(t, c.Create())

	if deleted {
		require.NoError(t, Db().Delete(c).Error)
	}

	return c
}

// storedSessionUserName returns the username stored with the session.
func storedSessionUserName(t *testing.T, id string) string {
	t.Helper()

	var found Session
	require.NoError(t, UnscopedDb().Where("id = ?", id).First(&found).Error)

	return found.UserName
}

// storedClientUserName returns the username stored with the client, including a deleted one.
func storedClientUserName(t *testing.T, uid string) string {
	t.Helper()

	var found Client
	require.NoError(t, UnscopedDb().Where("client_uid = ?", uid).First(&found).Error)

	return found.UserName
}

// storedUser returns the account row as stored.
func storedUser(t *testing.T, uid string) User {
	t.Helper()

	var found User
	require.NoError(t, UnscopedDb().Where("user_uid = ?", uid).First(&found).Error)

	return found
}

// setStoredUserName writes a username directly, e.g. a name stored by an earlier version.
func setStoredUserName(t *testing.T, m *User, name string) {
	t.Helper()

	require.NoError(t, UnscopedDb().Model(&User{}).Where("id = ?", m.ID).UpdateColumn("user_name", name).Error)
	m.UserName = name
}

// withRenameHandler registers a rename handler for the duration of a test.
func withRenameHandler(t *testing.T, table string, handler UserRenameHandler) {
	t.Helper()

	orig, found := UserRenameHandlers[table]
	UserRenameHandlers[table] = handler

	t.Cleanup(func() {
		if found {
			UserRenameHandlers[table] = orig
		} else {
			delete(UserRenameHandlers, table)
		}
	})
}

func TestUser_CheckRename(t *testing.T) {
	m := newRenameTestUser(t, "check-rename", acl.RoleAdmin)

	t.Run("Success", func(t *testing.T) {
		name, err := m.CheckRename("Check-Renamed")
		require.NoError(t, err)
		assert.Equal(t, "check-renamed", name)
		assert.Equal(t, "check-rename", m.UserName)
	})
	t.Run("SameName", func(t *testing.T) {
		name, err := m.CheckRename("Check-Rename")
		require.NoError(t, err)
		assert.Equal(t, "", name)
	})
	t.Run("MaxLength", func(t *testing.T) {
		name, err := m.CheckRename(strings.Repeat("a", 30) + "@" + strings.Repeat("b", 29) + ".com")
		require.NoError(t, err)
		assert.Len(t, name, txt.ClipUsername)
	})
	t.Run("SystemUser", func(t *testing.T) {
		_, err := (&Visitor).CheckRename("check-visitor")
		require.EqualError(t, err, "system users cannot be modified")
	})
	t.Run("InvalidUID", func(t *testing.T) {
		_, err := (&User{ID: 999901, UserUID: "invalid"}).CheckRename("check-invalid")
		require.EqualError(t, err, "system users cannot be modified")
	})
	t.Run("Empty", func(t *testing.T) {
		_, err := m.CheckRename("")
		require.EqualError(t, err, "username is empty")
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := m.CheckRename("check rename<")
		require.EqualError(t, err, "username is invalid")
	})
	t.Run("Reserved", func(t *testing.T) {
		_, err := m.CheckRename("visitor")
		require.EqualError(t, err, "username is reserved")
	})
	t.Run("TooLongBytes", func(t *testing.T) {
		_, err := m.CheckRename(strings.Repeat("a", 256))
		require.EqualError(t, err, "username is too long")
	})
	t.Run("TooLongForLogin", func(t *testing.T) {
		_, err := m.CheckRename(strings.Repeat("a", 30) + "@" + strings.Repeat("b", 30) + ".com")
		require.EqualError(t, err, "username is too long")
	})
	t.Run("MultibyteMaxLength", func(t *testing.T) {
		name, err := m.CheckRename("j@" + strings.Repeat("é", 31))
		require.NoError(t, err)
		assert.Len(t, name, txt.ClipUsername)
	})
	t.Run("MultibyteTooLong", func(t *testing.T) {
		_, err := m.CheckRename("j@" + strings.Repeat("é", 31) + "x")
		require.EqualError(t, err, "username is too long")
	})
	t.Run("TooShort", func(t *testing.T) {
		orig := UsernameLength
		UsernameLength = 4
		t.Cleanup(func() { UsernameLength = orig })

		_, err := m.CheckRename("abc")
		require.EqualError(t, err, "username must have at least 4 characters")
	})
	t.Run("NoHandle", func(t *testing.T) {
		_, err := m.CheckRename(".check-rename")
		require.EqualError(t, err, "username '.check-rename' is not supported")
	})
	t.Run("EmptyStoredName", func(t *testing.T) {
		_, err := (&User{ID: 999901, UserUID: rnd.GenerateUID(UserUID)}).CheckRename("")
		require.EqualError(t, err, "username is empty")
	})
	t.Run("Exists", func(t *testing.T) {
		_, err := m.CheckRename("alice")
		require.EqualError(t, err, "user 'alice' already exists")
	})
	t.Run("ExistsDeleted", func(t *testing.T) {
		holder := newRenameTestUser(t, "check-rename-deleted", acl.RoleAdmin)
		require.NoError(t, Db().Delete(holder).Error)

		_, err := m.CheckRename("check-rename-deleted")
		require.EqualError(t, err, "user 'check-rename-deleted' already exists")
	})
	t.Run("ExistsCollation", func(t *testing.T) {
		newRenameTestUser(t, "check-josé", acl.RoleAdmin)

		_, err := m.CheckRename("check-jose")

		if IsDialect(dsn.DriverMySQL) {
			require.EqualError(t, err, "user 'check-jose' already exists")
		} else {
			require.NoError(t, err)
		}
	})
}

func TestUser_Rename(t *testing.T) {
	registerContributorRole(t)

	t.Run("Success", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-success", acl.RoleAdmin)
		sess := newRenameTestSession(t, m)
		client := newRenameTestClient(t, m, false)
		deletedClient := newRenameTestClient(t, m, true)
		other := newRenameTestClient(t, FindUserByName("alice"), false)
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(other).Error) })

		require.NoError(t, m.Rename("Rename-Success-New"))

		assert.Equal(t, "rename-success-new", m.UserName)
		assert.Equal(t, "rename-success-new", storedUser(t, m.UserUID).UserName)
		assert.Equal(t, "rename-success-new", storedSessionUserName(t, sess.ID))
		assert.Equal(t, "rename-success-new", storedClientUserName(t, client.ClientUID))
		assert.Equal(t, "rename-success-new", storedClientUserName(t, deletedClient.ClientUID))
		assert.Equal(t, "alice", storedClientUserName(t, other.ClientUID))
		assert.Nil(t, FindUserByName("rename-success"))
	})
	t.Run("SameName", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-same", acl.RoleAdmin)
		withRenameHandler(t, "zz_rename_test", func(*gorm.DB, *User, string, string) error {
			return errors.New("must not be called")
		})

		require.NoError(t, m.Rename("Rename-Same"))
		assert.Equal(t, "rename-same", storedUser(t, m.UserUID).UserName)
	})
	t.Run("Refused", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-refused", acl.RoleAdmin)

		require.EqualError(t, m.Rename("alice"), "user 'alice' already exists")
		assert.Equal(t, "rename-refused", m.UserName)
		assert.Equal(t, "rename-refused", storedUser(t, m.UserUID).UserName)
	})
	t.Run("Collation", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-jose", acl.RoleAdmin)
		sess := newRenameTestSession(t, m)

		require.NoError(t, m.Rename("rename-josé"))
		assert.Equal(t, "rename-josé", storedUser(t, m.UserUID).UserName)
		assert.Equal(t, "rename-josé", storedSessionUserName(t, sess.ID))
	})
	t.Run("DisplayName", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-display", acl.RoleAdmin)
		require.NoError(t, UnscopedDb().Model(&User{}).Where("id = ?", m.ID).UpdateColumn("display_name", "").Error)
		m.DisplayName = ""

		require.NoError(t, m.Rename("rename-display-new"))
		assert.Equal(t, "Rename-Display-New", storedUser(t, m.UserUID).DisplayName)
	})
	t.Run("EmailDomain", func(t *testing.T) {
		m := newRenameTestUser(t, "jane.rename@old.example", acl.RoleContributor)

		require.NoError(t, m.Rename("jane.rename@new.example"))

		stored := storedUser(t, m.UserUID)
		assert.Equal(t, "jane.rename@new.example", stored.UserName)
		assert.Equal(t, "", stored.BasePath)
		assert.Equal(t, "users/jane.rename", stored.GetBasePath())
	})
}

func TestUser_SaveRename(t *testing.T) {
	t.Run("Rollback", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-rollback", acl.RoleAdmin)
		m.UserEmail = "rename-rollback@example.com"
		require.NoError(t, m.Save())
		sess := newRenameTestSession(t, m)
		client := newRenameTestClient(t, m, false)
		deletedClient := newRenameTestClient(t, m, true)

		withRenameHandler(t, "zz_rename_test", func(*gorm.DB, *User, string, string) error {
			return errors.New("test error")
		})

		oldName := m.UserName
		require.NoError(t, m.SetNewUsername("rename-rollback-new"))
		m.UserEmail = "rename-rollback-new@example.com"
		m.DisplayName = "Rename Rollback New"

		require.EqualError(t, m.SaveRename(oldName), "failed to update zz_rename_test (test error)")

		stored := storedUser(t, m.UserUID)
		assert.Equal(t, "rename-rollback", stored.UserName)
		assert.Equal(t, "rename-rollback@example.com", stored.UserEmail)
		assert.NotEqual(t, "Rename Rollback New", stored.DisplayName)
		assert.Equal(t, "rename-rollback", storedSessionUserName(t, sess.ID))
		assert.Equal(t, "rename-rollback", storedClientUserName(t, client.ClientUID))
		assert.Equal(t, "rename-rollback", storedClientUserName(t, deletedClient.ClientUID))
	})
	t.Run("HandlerOrder", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-order", acl.RoleAdmin)

		var called []string

		for _, table := range []string{"zz_rename_c", "zz_rename_a", "zz_rename_b"} {
			withRenameHandler(t, table, func(tx *gorm.DB, u *User, oldName, newName string) error {
				assert.NotNil(t, tx)
				assert.Equal(t, m.UserUID, u.UserUID)
				assert.Equal(t, "rename-order", oldName)
				assert.Equal(t, "rename-order-new", newName)
				called = append(called, table)
				return nil
			})
		}

		require.NoError(t, m.Rename("rename-order-new"))
		assert.Equal(t, []string{"zz_rename_a", "zz_rename_b", "zz_rename_c"}, called)
	})
	t.Run("FlushesCache", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-cache", acl.RoleAdmin)
		key := rnd.AuthToken()
		CacheWebDAVUser(key, m, CurrentAuthCacheGeneration())
		require.Same(t, m, CachedWebDAVUser(key))

		require.NoError(t, m.Rename("rename-cache-new"))
		assert.Nil(t, CachedWebDAVUser(key))
	})
	t.Run("Restored", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-restored", acl.RoleAdmin)
		require.NoError(t, Db().Delete(m).Error)
		deleted := storedUser(t, m.UserUID)
		require.NotNil(t, deleted.DeletedAt)

		deleted.DeletedAt = nil
		require.NoError(t, deleted.SetNewUsername("rename-restored-new"))
		require.NoError(t, deleted.SaveRename("rename-restored"))

		stored := storedUser(t, m.UserUID)
		assert.Nil(t, stored.DeletedAt)
		assert.Equal(t, "rename-restored-new", stored.UserName)
	})
	t.Run("Unchanged", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-unchanged", acl.RoleAdmin)
		withRenameHandler(t, "zz_rename_test", func(*gorm.DB, *User, string, string) error {
			return errors.New("must not be called")
		})

		m.DisplayName = "Rename Unchanged"
		require.NoError(t, m.SaveRename(m.UserName))
		assert.Equal(t, "Rename Unchanged", storedUser(t, m.UserUID).DisplayName)
	})
	t.Run("SystemUser", func(t *testing.T) {
		m := Visitor
		m.UserName = "visitor-renamed"
		require.EqualError(t, m.SaveRename("visitor"), "system users cannot be modified")
	})
}

func TestUser_SetNormalizedUsername(t *testing.T) {
	registerContributorRole(t)

	t.Run("MixedCase", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-legacy", acl.RoleAdmin)
		sess := newRenameTestSession(t, m)
		setStoredUserName(t, m, "Rename-Legacy")
		require.NoError(t, UnscopedDb().Model(&Session{}).Where("id = ?", sess.ID).UpdateColumn("user_name", "Rename-Legacy").Error)

		oldName := m.UserName
		assert.True(t, m.SetNormalizedUsername())
		assert.Equal(t, "rename-legacy", m.UserName)
		require.NoError(t, m.SaveRename(oldName))

		assert.Equal(t, "rename-legacy", storedUser(t, m.UserUID).UserName)
		assert.Equal(t, "rename-legacy", storedSessionUserName(t, sess.ID))
	})
	t.Run("ChangedHandle", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-plus", acl.RoleContributor)
		setStoredUserName(t, m, "rename+plus")
		loaded := *m
		before := loaded.GetUploadPath()

		assert.True(t, m.SetNormalizedUsername())
		assert.Equal(t, "renameplus", m.UserName)
		assert.Equal(t, "users/rename.plus", m.BasePath)
		assert.Equal(t, before, m.GetUploadPath())
	})
	t.Run("HeldByDeleted", func(t *testing.T) {
		holder := newRenameTestUser(t, "rename-held", acl.RoleAdmin)
		require.NoError(t, Db().Delete(holder).Error)
		m := newRenameTestUser(t, "rename-held-legacy", acl.RoleAdmin)
		setStoredUserName(t, m, "Rename-Held")

		assert.False(t, m.SetNormalizedUsername())
		assert.Equal(t, "Rename-Held", m.UserName)
	})
	t.Run("HeldByActive", func(t *testing.T) {
		newRenameTestUser(t, "rename-active", acl.RoleAdmin)
		m := newRenameTestUser(t, "rename-active-legacy", acl.RoleAdmin)
		setStoredUserName(t, m, "Rename-Active")

		require.NoError(t, m.Rename("rename-active"))
		assert.Equal(t, "Rename-Active", storedUser(t, m.UserUID).UserName)
	})
	t.Run("Unchanged", func(t *testing.T) {
		m := newRenameTestUser(t, "rename-normal", acl.RoleContributor)

		assert.False(t, m.SetNormalizedUsername())
		assert.Equal(t, "rename-normal", m.UserName)
		assert.Equal(t, "", m.BasePath)
	})
	t.Run("SystemUser", func(t *testing.T) {
		m := Visitor
		m.UserName = "Visitor"
		assert.False(t, m.SetNormalizedUsername())
	})
}

func TestUser_SetNewUsername(t *testing.T) {
	registerContributorRole(t)

	// The default base and upload paths each case resolves to must not change with the name.
	cases := []struct {
		name       string
		oldName    string
		newName    string
		role       acl.Role
		basePath   string
		uploadPath string
		wantBase   string
		wantUpload string
	}{
		{"ContributorDefault", "pin-contrib", "pin-contrib-new", acl.RoleContributor, "", "", "users/pin-contrib", ""},
		{"ContributorExplicit", "pin-explicit", "pin-explicit-new", acl.RoleContributor, "shared/team", "inbox", "shared/team", "inbox"},
		{"AdminNoBasePath", "pin-admin", "pin-admin-new", acl.RoleAdmin, "", "", "", ""},
		{"NoHandleUsers", ".pin-nohandle", "pin-nohandle", acl.RoleAdmin, "users", "", "users/+%s", ""},
		{"NoHandleDot", ".pin-dot", "pin-dot", acl.RoleAdmin, ".", "", "users/+%s", ""},
		{"NoHandleContributor", ".pin-nohandle-c", "pin-nohandle-c", acl.RoleContributor, "", "", "users/+%s", ""},
		{"UploadHome", "pin-home", "pin-home-new", acl.RoleAdmin, "", "~", "", "users/pin-home"},
		{"UploadHomeWithBase", "pin-home-base", "pin-home-base-new", acl.RoleAdmin, "photos", "~", "photos", "~"},
		{"UploadHomeContributor", "pin-home-c", "pin-home-c-new", acl.RoleContributor, "", "~", "users/pin-home-c", "~"},
		{"SameHandle", "pin.same@old.example", "pin.same@new.example", acl.RoleContributor, "", "~", "", "~"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &User{
				ID:          999900,
				UserUID:     rnd.GenerateUID(UserUID),
				DisplayName: "Pin Test",
				UserName:    tc.oldName,
				UserRole:    tc.role.String(),
				BasePath:    tc.basePath,
				UploadPath:  tc.uploadPath,
			}

			loaded := *m
			wantBasePath, wantUploadPath := loaded.GetBasePath(), loaded.GetUploadPath()

			require.NoError(t, m.SetNewUsername(tc.newName))
			assert.Equal(t, tc.newName, m.UserName)

			assert.Equal(t, strings.Replace(tc.wantBase, "%s", m.UserUID, 1), m.BasePath)
			assert.Equal(t, tc.wantUpload, m.UploadPath)
			assert.Equal(t, wantBasePath, m.GetBasePath())
			assert.Equal(t, wantUploadPath, m.GetUploadPath())
		})
	}

	t.Run("NormalizedOnly", func(t *testing.T) {
		m := &User{ID: 999900, UserUID: rnd.GenerateUID(UserUID), DisplayName: "Pin Test", UserName: "pin+plus", UserRole: acl.RoleContributor.String()}

		require.NoError(t, m.SetNewUsername("pinplus"))
		assert.Equal(t, "pinplus", m.UserName)
		assert.Equal(t, "users/pin.plus", m.BasePath)
	})
	t.Run("Refused", func(t *testing.T) {
		m := &User{ID: 999900, UserUID: rnd.GenerateUID(UserUID), DisplayName: "Pin Test", UserName: "pin-refused", UserRole: acl.RoleContributor.String()}

		require.Error(t, m.SetNewUsername("alice"))
		assert.Equal(t, "pin-refused", m.UserName)
		assert.Equal(t, "", m.BasePath)
	})
}

func TestUser_NormalizedUsername(t *testing.T) {
	t.Run("Canonical", func(t *testing.T) {
		name, err := (&User{ID: 1000, UserName: "jane"}).normalizedUsername()
		require.NoError(t, err)
		assert.Equal(t, "jane", name)
	})
	t.Run("StoredMixedCase", func(t *testing.T) {
		name, err := (&User{ID: 1000, UserName: "Jane"}).normalizedUsername()
		require.NoError(t, err)
		assert.Equal(t, "jane", name)
	})
	t.Run("StoredInvalid", func(t *testing.T) {
		name, err := (&User{ID: 1000, UserName: "jane+doe"}).normalizedUsername()
		require.NoError(t, err)
		assert.Equal(t, "janedoe", name)
	})
	t.Run("NewInvalid", func(t *testing.T) {
		_, err := (&User{UserName: "jane+doe"}).normalizedUsername()
		require.EqualError(t, err, "username is invalid")
	})
	t.Run("Empty", func(t *testing.T) {
		_, err := (&User{ID: 1000}).normalizedUsername()
		require.EqualError(t, err, "username is empty")
	})
}

func TestUser_SetRenamedUsername(t *testing.T) {
	registerContributorRole(t)

	t.Run("DefaultChanges", func(t *testing.T) {
		m := &User{UserUID: rnd.GenerateUID(UserUID), UserName: "jane", UserRole: acl.RoleContributor.String(), UploadPath: "~"}
		m.setRenamedUsername("joan")
		assert.Equal(t, "joan", m.UserName)
		assert.Equal(t, "users/jane", m.BasePath)
		assert.Equal(t, "~", m.UploadPath)
	})
	t.Run("UploadHome", func(t *testing.T) {
		m := &User{UserUID: rnd.GenerateUID(UserUID), UserName: "jane", UserRole: acl.RoleAdmin.String(), UploadPath: "~"}
		m.setRenamedUsername("joan")
		assert.Equal(t, "", m.BasePath)
		assert.Equal(t, "users/jane", m.UploadPath)
	})
	t.Run("DefaultUnchanged", func(t *testing.T) {
		m := &User{UserUID: rnd.GenerateUID(UserUID), UserName: "jane@old.example", UserRole: acl.RoleContributor.String(), UploadPath: "~"}
		m.setRenamedUsername("jane@new.example")
		assert.Equal(t, "jane@new.example", m.UserName)
		assert.Equal(t, "", m.BasePath)
		assert.Equal(t, "~", m.UploadPath)
	})
}

func TestRenameSessionUser(t *testing.T) {
	m := newRenameTestUser(t, "rename-sessions", acl.RoleAdmin)
	sess := newRenameTestSession(t, m)
	other := newRenameTestSession(t, FindUserByName("alice"))
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&Session{}, "id = ?", other.ID).Error) })

	require.NoError(t, Db().Transaction(func(tx *gorm.DB) error {
		return renameSessionUser(tx, m, m.UserName, "rename-sessions-new")
	}))

	assert.Equal(t, "rename-sessions-new", storedSessionUserName(t, sess.ID))
	assert.Equal(t, "alice", storedSessionUserName(t, other.ID))
}

func TestRenameClientUser(t *testing.T) {
	m := newRenameTestUser(t, "rename-clients", acl.RoleAdmin)
	client := newRenameTestClient(t, m, false)
	deleted := newRenameTestClient(t, m, true)

	require.NoError(t, Db().Transaction(func(tx *gorm.DB) error {
		return renameClientUser(tx, m, m.UserName, "rename-clients-new")
	}))

	assert.Equal(t, "rename-clients-new", storedClientUserName(t, client.ClientUID))
	assert.Equal(t, "rename-clients-new", storedClientUserName(t, deleted.ClientUID))
}

func TestUser_DeletedUsernameHolder(t *testing.T) {
	holder := newRenameTestUser(t, "rename-deleted-holder", acl.RoleAdmin)
	m := newRenameTestUser(t, "rename-deleted-holder-legacy", acl.RoleAdmin)

	t.Run("Active", func(t *testing.T) {
		assert.False(t, m.deletedUsernameHolder("rename-deleted-holder"))
	})
	t.Run("Deleted", func(t *testing.T) {
		require.NoError(t, Db().Delete(holder).Error)
		assert.True(t, m.deletedUsernameHolder("rename-deleted-holder"))
		assert.False(t, holder.deletedUsernameHolder("rename-deleted-holder"))
	})
	t.Run("ValidateKeepsStoredName", func(t *testing.T) {
		setStoredUserName(t, m, "Rename-Deleted-Holder")
		require.NoError(t, m.Validate())
		assert.Equal(t, "Rename-Deleted-Holder", m.UserName)
	})
	t.Run("ValidateActiveAndDeleted", func(t *testing.T) {
		deleted := newRenameTestUser(t, "rename-both-holders", acl.RoleAdmin)
		require.NoError(t, Db().Delete(deleted).Error)
		active := newRenameTestUser(t, "rename-both-holders-active", acl.RoleAdmin)
		setStoredUserName(t, active, "rename-both-holders")
		legacy := newRenameTestUser(t, "rename-both-holders-legacy", acl.RoleAdmin)
		setStoredUserName(t, legacy, "Rename-Both-Holders")

		require.EqualError(t, legacy.Validate(), "user 'rename-both-holders' already exists")
	})
	t.Run("ValidateNormalizes", func(t *testing.T) {
		other := newRenameTestUser(t, "rename-validate-normal", acl.RoleAdmin)
		setStoredUserName(t, other, "Rename-Validate-Normal")
		require.NoError(t, other.Validate())
		assert.Equal(t, "rename-validate-normal", other.UserName)
	})
}
