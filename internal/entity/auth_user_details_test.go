package entity

import (
	"testing"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestCreateUserDetails(t *testing.T) {
	t.Run("EmptyUid", func(t *testing.T) {
		m := &User{UserUID: ""}
		assert.Error(t, CreateUserDetails(m))
		assert.Nil(t, m.UserDetails)
	})
	t.Run("InvalidUid", func(t *testing.T) {
		m := &User{UserUID: "pqxetse3cy5eo9z2"}
		assert.ErrorIs(t, CreateUserDetails(m), ErrInvalidUserUID)
		assert.Equal(t, int64(0), countUserDetails(t, m.UserUID))
	})
	t.Run("Success", func(t *testing.T) {
		m := &User{UserUID: "usg73p55zwgr1gbq"}
		err := CreateUserDetails(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, m.UserDetails)
	})
}

func TestUserDetails_HasID(t *testing.T) {
	u := FindUserByName("alice")
	assert.True(t, u.UserDetails.HasID())
}

// countUserDetails returns the number of details rows with the given user UID.
func countUserDetails(t *testing.T, uid string) (n int64) {
	t.Helper()
	require.NoError(t, UnscopedDb().Model(&UserDetails{}).Where("user_uid = ?", uid).Count(&n).Error)
	return n
}

func TestUserDetails_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewUserDetails(rnd.GenerateUID(UserUID))
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })
		assert.Equal(t, int64(1), countUserDetails(t, m.UserUID))
	})
	t.Run("InvalidUid", func(t *testing.T) {
		for _, uid := range invalidUserUIDs {
			before := countUserDetails(t, uid)
			assert.ErrorIs(t, NewUserDetails(uid).Create(), ErrInvalidUserUID, "uid %q", uid)
			assert.Equal(t, before, countUserDetails(t, uid), "uid %q", uid)
		}
	})
}

func TestUserDetails_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewUserDetails(rnd.GenerateUID(UserUID))
		m.GivenName = "John"
		require.NoError(t, m.Save())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		m.GivenName = "Jane"
		require.NoError(t, m.Save())
		assert.Equal(t, int64(1), countUserDetails(t, m.UserUID))

		found := UserDetails{}
		require.NoError(t, UnscopedDb().Where("user_uid = ?", m.UserUID).First(&found).Error)
		assert.Equal(t, "Jane", found.GivenName)
	})
	t.Run("InvalidUid", func(t *testing.T) {
		for _, uid := range invalidUserUIDs {
			before := countUserDetails(t, uid)
			assert.ErrorIs(t, NewUserDetails(uid).Save(), ErrInvalidUserUID, "uid %q", uid)
			assert.Equal(t, before, countUserDetails(t, uid), "uid %q", uid)
		}
	})
}

func TestUserDetails_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewUserDetails(rnd.GenerateUID(UserUID))
		m.GivenName, m.FamilyName = "John", "Doe"
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		assert.NoError(t, m.Updates(UserDetails{GivenName: "Jane"}))
		assert.Equal(t, "Jane", m.GivenName)

		found := UserDetails{}
		require.NoError(t, UnscopedDb().Where("user_uid = ?", m.UserUID).First(&found).Error)
		assert.Equal(t, "Jane", found.GivenName)
		assert.Equal(t, "Doe", found.FamilyName)
	})
	t.Run("NoKey", func(t *testing.T) {
		m := &UserDetails{GivenName: "John"}

		assert.ErrorIs(t, m.Updates(UserDetails{GivenName: "Jane"}), gorm.ErrMissingWhereClause)
	})
}

func TestUserDetails_DisplayName(t *testing.T) {
	t.Run("DrJohnDoe", func(t *testing.T) {
		m := &User{
			UserUID: "1234",
			UserDetails: &UserDetails{
				BirthYear:  1999,
				BirthMonth: 3,
				NameTitle:  "Dr.",
				GivenName:  "John",
				MiddleName: "Wulfrick",
				FamilyName: "Doe",
			}}

		assert.Equal(t, "Dr. John Doe", m.UserDetails.DisplayName())
	})
	t.Run("Empty", func(t *testing.T) {
		m := &User{}
		assert.Equal(t, "", m.UserDetails.DisplayName())
	})

}
