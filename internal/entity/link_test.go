package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewLink(t *testing.T) {
	ValidateFixtures(t)
	link := NewLink("ss6sg6bxpogaaba1", true, false)
	assert.Equal(t, "ss6sg6bxpogaaba1", link.ShareUID)
	assert.Equal(t, 10, len(link.LinkToken))
	assert.Equal(t, 16, len(link.LinkUID))
}

func TestLink_Expired(t *testing.T) {
	ValidateFixtures(t)
	const oneDay = 60 * 60 * 24

	link := NewLink("ss6sg6bxpogaaba1", true, false)

	link.ModifiedAt = Now().Add(-7 * Day)
	link.LinkExpires = 0

	assert.False(t, link.Expired())

	link.LinkExpires = oneDay

	assert.True(t, link.Expired())

	link.LinkExpires = oneDay * 8

	assert.False(t, link.Expired())

	link.LinkExpires = oneDay * 300
	link.LinkViews = 9
	link.MaxViews = 10

	assert.False(t, link.Expired())

	link.Redeem()

	assert.True(t, link.Expired())
}

func TestLink_Redeem(t *testing.T) {
	ValidateFixtures(t)
	link := NewLink(rnd.GenerateUID(AlbumUID), false, false)

	assert.Equal(t, uint(0), link.LinkViews)

	link.Redeem()

	assert.Equal(t, uint(1), link.LinkViews)

	if err := link.Save(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&link).Error) })

	link.Redeem()

	assert.Equal(t, uint(2), link.LinkViews)
}

func TestLink_SetSlug(t *testing.T) {
	ValidateFixtures(t)
	link := Link{}
	assert.Equal(t, "", link.ShareSlug)
	link.SetSlug("test Slug")
	assert.Equal(t, "test-slug", link.ShareSlug)
}

func TestLink_SetPassword(t *testing.T) {
	ValidateFixtures(t)
	link := Link{LinkUID: "dftjdfkvh"}
	assert.Equal(t, false, link.HasPassword)
	err := link.SetPassword("123")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&Password{}, "uid = ?", link.LinkUID).Error) })
	assert.Equal(t, true, link.HasPassword)
}

func TestLink_InvalidPassword(t *testing.T) {
	ValidateFixtures(t)
	t.Run("NoPassword", func(t *testing.T) {
		link := Link{LinkUID: "dftjdfkvhjh", HasPassword: false}
		assert.False(t, link.InvalidPassword("123"))
	})
	t.Run("InvalidPassword", func(t *testing.T) {
		link := NewLink("dhfjf", false, false)

		err := link.SetPassword("123")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&Password{}, "uid = ?", link.LinkUID).Error) })
		assert.False(t, link.InvalidPassword("123"))
	})
	t.Run("ValidPassword", func(t *testing.T) {
		link := NewLink("dhfjfk", false, false)

		err := link.SetPassword("123kkljgfuA")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&Password{}, "uid = ?", link.LinkUID).Error) })
		assert.True(t, link.InvalidPassword("123"))
	})
}

func TestLink_Save(t *testing.T) {
	ValidateFixtures(t)
	t.Run("InvalidShareUid", func(t *testing.T) {
		link := NewLink("dhfjfjh", false, false)

		assert.Error(t, link.Save())
	})
	t.Run("EmptyToken", func(t *testing.T) {
		link := Link{ShareUID: "ls6sg6bffgtredft", LinkToken: ""}

		assert.Error(t, link.Save())
	})
	t.Run("Success", func(t *testing.T) {
		link := NewLink("ls6sg6bffgtredft", false, false)

		err := link.Save()

		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&link).Error) })
	})
}

func TestLink_Delete(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Success", func(t *testing.T) {
		link := NewLink("ls6sg6bffgtreoft", false, false)

		err := link.Delete()

		if err != nil {
			t.Fatal(err)
		}

	})
	t.Run("EmptyToken", func(t *testing.T) {
		link := Link{ShareUID: "ls6sg6bffgtredft", LinkToken: ""}
		assert.Error(t, link.Delete())
	})
	t.Run("EmptyUid", func(t *testing.T) {
		link := Link{LinkUID: "", ShareUID: "", LinkToken: "abc"}
		assert.Error(t, link.Delete())
	})
}

func TestFindLink(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Success", func(t *testing.T) {
		m := NewLink("ls6sg6bffgtrjoft", false, false)

		link := &m

		if err := link.Save(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&link).Error) })
		uid := link.LinkUID
		t.Logf("%#v", link)
		r := FindLink(uid)
		t.Log(r)
		//TODO Why does it fail? <- because LinkToken is generated random token.
		//assert.Equal(t, "1jxf3jfn2k", r.LinkToken)
	})
	t.Run("Nil", func(t *testing.T) {
		r := FindLink("XXX")
		assert.Nil(t, r)
	})
}

func TestDeleteShareLinks(t *testing.T) {
	ValidateFixtures(t)
	t.Run("EmptyShareUid", func(t *testing.T) {
		assert.Error(t, DeleteShareLinks(""))
	})
}

func TestFindLinks(t *testing.T) {
	ValidateFixtures(t)
	t.Run("FindByToken", func(t *testing.T) {
		r := FindLinks("1jxf3jfn2k", "")
		assert.Equal(t, "as6sg6bxpogaaba8", r[0].ShareUID)
	})
	t.Run("NoTokenAndShare", func(t *testing.T) {
		r := FindLinks("", "")
		assert.Empty(t, r)
	})
	t.Run("InvalidToken", func(t *testing.T) {
		r := FindLinks("lkjh", "")
		assert.Empty(t, r)
	})
	t.Run("FindBySlug", func(t *testing.T) {
		r := FindLinks("", "holiday-2030")
		assert.Equal(t, "as6sg6bxpogaaba8", r[0].ShareUID)
	})
}

func TestFindValidLinksLinks(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Success", func(t *testing.T) {
		r := FindValidLinks("1jxf3jfn2k", "")
		assert.Equal(t, "as6sg6bxpogaaba8", r[0].ShareUID)
	})
}

func TestLink_String(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Success", func(t *testing.T) {
		link := NewLink("jhgko", false, false)
		uid := link.LinkUID
		assert.Equal(t, uid, link.String())
	})
}
