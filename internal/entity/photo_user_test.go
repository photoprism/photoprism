package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewPhotoUser(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		uid := rnd.GenerateUID('n')
		userUID := rnd.GenerateUID(UserUID)
		teamUID := rnd.GenerateUID('t')
		perm := uint(15)
		au := NewPhotoUser(uid, userUID, teamUID, perm)

		assert.Equal(t, uid, au.UID)
		assert.Equal(t, userUID, au.UserUID)
		assert.Equal(t, teamUID, au.TeamUID)
		assert.Equal(t, perm, au.Perm)
	})
}

func TestCreatePhotoUser(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		uid := rnd.GenerateUID('n')
		userUID := rnd.GenerateUID(UserUID)
		teamUID := rnd.GenerateUID('t')
		perm := uint(15)
		au := NewPhotoUser(uid, userUID, teamUID, perm)

		err := au.Create()
		assert.Empty(t, err)

		err = Db().Unscoped().Delete(au).Error
		assert.Empty(t, err)
	})
}

func TestSavePhotoUser(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		uid := rnd.GenerateUID('n')
		userUID := rnd.GenerateUID(UserUID)
		teamUID := rnd.GenerateUID('t')
		perm := uint(15)
		au := NewPhotoUser(uid, userUID, teamUID, perm)

		err := au.Create()
		assert.Empty(t, err)

		au.Perm = uint(64)

		err = au.Save()
		assert.Empty(t, err)

		err = Db().Unscoped().Delete(au).Error
		assert.Empty(t, err)
	})
}

func TestFirstOrCreatePhotoUser(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		uid := rnd.GenerateUID('n')
		userUID := rnd.GenerateUID(UserUID)
		teamUID := rnd.GenerateUID('t')
		perm := uint(15)
		au := NewPhotoUser(uid, userUID, teamUID, perm)

		actual := FirstOrCreatePhotoUser(au)

		assert.Equal(t, au, actual)

		find := PhotoUser{UID: uid}
		actual = FirstOrCreatePhotoUser(&find)
		assert.Equal(t, au, actual)

		err := Db().Unscoped().Delete(au).Error
		assert.Empty(t, err)
	})
}
