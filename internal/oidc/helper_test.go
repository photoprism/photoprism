package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsernameFromUserInfo(t *testing.T) {
	t.Run("PreferredUsername", func(t *testing.T) {
		u := &userinfo{PreferredUsername: "testfest"}
		assert.Equal(t, "testfest", UsernameFromUserInfo(u))
	})
	t.Run("PreferredUsername too short", func(t *testing.T) {
		u := &userinfo{PreferredUsername: "tes"}
		assert.Equal(t, "", UsernameFromUserInfo(u))
	})
	t.Run("EMail", func(t *testing.T) {
		u := &userinfo{Nickname: "tes", Email: "hello@world.com"}
		assert.Equal(t, "hello@world.com", UsernameFromUserInfo(u))
	})
	t.Run("Nickname", func(t *testing.T) {
		u := &userinfo{Nickname: "testofesto", Email: "hel"}
		assert.Equal(t, "testofesto", UsernameFromUserInfo(u))
	})
}

func TestHasRoleAdmin(t *testing.T) {
	t.Run("true case", func(t *testing.T) {
		u := &userinfo{Claim: []interface{}{
			"admin",
			"photoprism_admin",
			"photoprism",
			"random",
		}}
		hasRoleAdmin, err := HasRoleAdmin(u)
		assert.True(t, hasRoleAdmin)
		assert.Nil(t, err)
	})
	t.Run("false case", func(t *testing.T) {
		u := &userinfo{Claim: []interface{}{
			"admin",
			"photoprismo_admin",
			"photoprism",
			"random",
		}}
		hasRoleAdmin, err := HasRoleAdmin(u)
		assert.False(t, hasRoleAdmin)
		assert.Nil(t, err)
	})
	t.Run("false case 2", func(t *testing.T) {
		u := &userinfo{Claim: []interface{}{}}
		hasRoleAdmin, err := HasRoleAdmin(u)
		assert.False(t, hasRoleAdmin)
		assert.Nil(t, err)
	})
	t.Run("error case", func(t *testing.T) {
		u := &userinfo{Claim: nil}
		hasRoleAdmin, err := HasRoleAdmin(u)
		assert.False(t, hasRoleAdmin)
		assert.Error(t, err)
	})
}

type userinfo struct {
	PreferredUsername string
	Nickname          string
	Name              string
	Email             string
	Claim             interface{}
}

func (u *userinfo) GetPreferredUsername() string {
	return u.PreferredUsername
}

func (u *userinfo) GetNickname() string {
	return u.Nickname
}

func (u *userinfo) GetName() string {
	return u.Name
}

func (u *userinfo) GetEmail() string {
	return u.Email
}

func (u *userinfo) GetClaim(key string) interface{} {
	return u.Claim
}
