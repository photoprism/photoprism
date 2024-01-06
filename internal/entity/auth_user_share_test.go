package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestUserShares_Contains(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		m := UserShares{UserShare{UserUID: "uqxetse3cy5eo9z2", ShareUID: "as6sg6bxpogaaba9"}}
		assert.False(t, m.Contains("as6sg6bxpogaaxxx"))
	})
	t.Run("True", func(t *testing.T) {
		m := UserShares{UserShare{UserUID: "uqxetse3cy5eo9z2", ShareUID: "as6sg6bxpogaaba9"}}
		assert.True(t, m.Contains("as6sg6bxpogaaba9"))
	})
	t.Run("Empty", func(t *testing.T) {
		m := UserShares{}
		assert.False(t, m.Contains("as6sg6bxpogaaxxx"))
	})
}

func TestNewUserShare(t *testing.T) {
	expires := TimeStamp().Add(time.Hour * 48)
	m := NewUserShare(Admin.UID(), AlbumFixtures.Get("berlin-2019").AlbumUID, PermReact, &expires)

	assert.True(t, m.HasID())
	assert.True(t, rnd.IsRefID(m.RefID))
	assert.True(t, rnd.IsUID(m.UserUID, UserUID))
	assert.True(t, rnd.IsUID(m.ShareUID, AlbumUID))
	assert.Equal(t, PermReact, m.Perm)
	assert.Equal(t, expires, *m.ExpiresAt)
	assert.Equal(t, "", m.Comment)
	assert.Equal(t, "", m.LinkUID)
}

func TestPerm(t *testing.T) {
	assert.Equal(t, uint(0), PermDefault)
	assert.Equal(t, uint(1), PermNone)
	assert.Equal(t, uint(2), PermView)
	assert.Equal(t, uint(4), PermReact)
	assert.Equal(t, uint(8), PermComment)
	assert.Equal(t, uint(16), PermUpload)
	assert.Equal(t, uint(32), PermEdit)
	assert.Equal(t, uint(64), PermShare)
	assert.Equal(t, uint(128), PermAll)
}

func TestFindUserShare(t *testing.T) {
	t.Run("AliceAlbum", func(t *testing.T) {
		m := FindUserShare(UserShare{UserUID: "uqxetse3cy5eo9z2", ShareUID: "as6sg6bxpogaaba9"})

		expected := UserShareFixtures.Get("AliceAlbum")

		assert.NotNil(t, m)
		assert.True(t, m.HasID())
		assert.True(t, rnd.IsRefID(m.RefID))
		assert.True(t, rnd.IsUID(m.UserUID, UserUID))
		assert.True(t, rnd.IsUID(m.ShareUID, AlbumUID))
		assert.Equal(t, expected.Perm, m.Perm)
		assert.Equal(t, expected.ExpiresAt, m.ExpiresAt)
		assert.Equal(t, expected.Comment, m.Comment)
		assert.Equal(t, expected.LinkUID, m.LinkUID)
		assert.Equal(t, expected.UserUID, m.UserUID)
		assert.Equal(t, expected.ShareUID, m.ShareUID)
	})
	t.Run("Empty", func(t *testing.T) {
		m := FindUserShare(UserShare{})
		assert.Nil(t, m)
	})
}

func TestFindUserShares(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		found := FindUserShares(UserFixtures.Pointer("alice").UID())
		assert.NotNil(t, found)
		assert.Len(t, found, 1)

		m := found[0]
		expected := UserShareFixtures.Get("AliceAlbum")

		assert.NotNil(t, m)
		assert.True(t, m.HasID())
		assert.True(t, rnd.IsRefID(m.RefID))
		assert.True(t, rnd.IsUID(m.UserUID, UserUID))
		assert.True(t, rnd.IsUID(m.ShareUID, AlbumUID))
		assert.Equal(t, expected.Perm, m.Perm)
		assert.Equal(t, expected.ExpiresAt, m.ExpiresAt)
		assert.Equal(t, expected.Comment, m.Comment)
		assert.Equal(t, expected.LinkUID, m.LinkUID)
		assert.Equal(t, expected.UserUID, m.UserUID)
		assert.Equal(t, expected.ShareUID, m.ShareUID)
	})
	t.Run("Invalid uid", func(t *testing.T) {
		found := FindUserShares("123")
		assert.IsType(t, UserShares{}, found)
		assert.Empty(t, found)
	})
}

func TestUserShare_Create(t *testing.T) {
	m := UserShare{}
	err := m.Create()

	if err != nil {
		t.Fatal(err)
	}
}

func TestUserShare_UpdateLink(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := UserShare{
			ShareUID: "as6sg6bxpogaaba9",
		}

		assert.Equal(t, "", m.LinkUID)
		assert.Equal(t, "", m.Comment)

		l := Link{
			LinkUID:  "ss62xpryd1ob8xxx",
			ShareUID: "as6sg6bxpogaaba9",
			Comment:  "Wedding",
		}
		err := m.UpdateLink(l)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "ss62xpryd1ob8xxx", m.LinkUID)
		assert.Equal(t, "Wedding", m.Comment)
	})
	t.Run("UID mismatch", func(t *testing.T) {
		m := UserShare{
			ShareUID: "as6sg6bxpogaaba9",
		}

		assert.Equal(t, "", m.LinkUID)
		assert.Equal(t, "", m.Comment)

		l := Link{
			LinkUID:  "ss62xpryd1ob8xxx",
			ShareUID: "as6sg6bxpogaaba8",
			Comment:  "Wedding",
		}
		err := m.UpdateLink(l)

		assert.Error(t, err)
		assert.Equal(t, "", m.LinkUID)
		assert.Equal(t, "", m.Comment)
	})
}
