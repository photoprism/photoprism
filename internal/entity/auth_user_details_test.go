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
	t.Run("Success", func(t *testing.T) {
		m := &User{UserUID: "1234"}
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
