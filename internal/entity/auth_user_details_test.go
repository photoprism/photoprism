package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserDetails(t *testing.T) {
	t.Run("Empty UID", func(t *testing.T) {
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

	m.UserDetails.Updates(UserDetails{GivenName: "Jane"})
	assert.Equal(t, "Jane", m.UserDetails.GivenName)
}

func TestUserDetails_DisplayName(t *testing.T) {
	t.Run("Dr. John Doe", func(t *testing.T) {
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
