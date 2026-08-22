package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewAlbumUser(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		uid := rnd.GenerateUID('n')
		userUID := rnd.GenerateUID(UserUID)
		teamUID := rnd.GenerateUID('t')
		perm := uint(15)
		au := NewAlbumUser(uid, userUID, teamUID, perm)

		assert.Equal(t, uid, au.UID)
		assert.Equal(t, userUID, au.UserUID)
		assert.Equal(t, teamUID, au.TeamUID)
		assert.Equal(t, perm, au.Perm)
	})
}

func TestCreateAlbumUser(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		uid := rnd.GenerateUID('n')
		userUID := rnd.GenerateUID(UserUID)
		teamUID := rnd.GenerateUID('t')
		perm := uint(15)
		au := NewAlbumUser(uid, userUID, teamUID, perm)

		err := au.Create()
		assert.Empty(t, err)

		err = Db().Unscoped().Delete(au).Error
		assert.Empty(t, err)
	})
}

func TestSaveAlbumUser(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		uid := rnd.GenerateUID('n')
		userUID := rnd.GenerateUID(UserUID)
		teamUID := rnd.GenerateUID('t')
		perm := uint(15)
		au := NewAlbumUser(uid, userUID, teamUID, perm)

		err := au.Create()
		assert.Empty(t, err)

		au.Perm = uint(64)

		err = au.Save()
		assert.Empty(t, err)

		err = Db().Unscoped().Delete(au).Error
		assert.Empty(t, err)
	})
}

func TestFirstOrCreateAlbumUser(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		uid := rnd.GenerateUID('n')
		userUID := rnd.GenerateUID(UserUID)
		teamUID := rnd.GenerateUID('t')
		perm := uint(15)
		au := NewAlbumUser(uid, userUID, teamUID, perm)

		actual := FirstOrCreateAlbumUser(au)

		assert.Equal(t, au, actual)

		find := AlbumUser{UID: uid}
		actual = FirstOrCreateAlbumUser(&find)
		assert.Equal(t, au, actual)

		err := Db().Unscoped().Delete(au).Error
		assert.Empty(t, err)
	})
}
