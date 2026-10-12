package entity

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewLink(t *testing.T) {
	link := NewLink("ss6sg6bxpogaaba1", true, false)
	assert.Equal(t, "ss6sg6bxpogaaba1", link.ShareUID)
	assert.Equal(t, 10, len(link.LinkToken))
	assert.Equal(t, 16, len(link.LinkUID))
}

func TestLink_Expired(t *testing.T) {
	const oneDay = 60 * 60 * 24

	link := NewLink("ss6sg6bxpogaaba1", true, false)

	link.ModifiedAt = Now().Add(-7 * Day)
	link.LinkExpires = 0

	assert.False(t, link.Expired())

	link.LinkExpires = oneDay

	assert.True(t, link.Expired())

	link.LinkExpires = oneDay * 8

	assert.False(t, link.Expired())

	t.Run("ViewLimitIsNotExpiry", func(t *testing.T) {
		reached := NewLink("ss6sg6bxpogaaba1", true, false)
		reached.LinkViews = 10
		reached.MaxViews = 10

		assert.False(t, reached.Expired())
	})
}

func TestLink_Redeemable(t *testing.T) {
	const oneDay = 60 * 60 * 24

	t.Run("NoLimits", func(t *testing.T) {
		link := NewLink("ss6sg6bxpogaaba1", true, false)
		assert.True(t, link.Redeemable())
	})
	t.Run("ViewsRemaining", func(t *testing.T) {
		link := NewLink("ss6sg6bxpogaaba1", true, false)
		link.MaxViews = 2
		link.LinkViews = 1

		assert.True(t, link.Redeemable())
	})
	t.Run("ViewLimitReached", func(t *testing.T) {
		link := NewLink("ss6sg6bxpogaaba1", true, false)
		link.MaxViews = 2
		link.LinkViews = 2

		assert.False(t, link.Redeemable())
	})
	t.Run("ViewLimitExceeded", func(t *testing.T) {
		link := NewLink("ss6sg6bxpogaaba1", true, false)
		link.MaxViews = 2
		link.LinkViews = 3

		assert.False(t, link.Redeemable())
	})
	t.Run("UnlimitedViews", func(t *testing.T) {
		link := NewLink("ss6sg6bxpogaaba1", true, false)
		link.MaxViews = 0
		link.LinkViews = 500

		assert.True(t, link.Redeemable())
	})
	t.Run("Expired", func(t *testing.T) {
		link := NewLink("ss6sg6bxpogaaba1", true, false)
		link.ModifiedAt = Now().Add(-7 * Day)
		link.LinkExpires = oneDay

		assert.False(t, link.Redeemable())
	})
}

func TestLink_Redeem(t *testing.T) {
	link := NewLink(rnd.GenerateUID(AlbumUID), false, false)

	assert.Equal(t, uint(0), link.LinkViews)

	link.Redeem()

	assert.Equal(t, uint(1), link.LinkViews)

	if err := link.Save(); err != nil {
		t.Fatal(err)
	}

	link.Redeem()

	assert.Equal(t, uint(2), link.LinkViews)
}

func TestLink_SetSlug(t *testing.T) {
	link := Link{}
	assert.Equal(t, "", link.ShareSlug)
	link.SetSlug("test Slug")
	assert.Equal(t, "test-slug", link.ShareSlug)
}

func TestLink_Save(t *testing.T) {
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
	})
}

func TestLink_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		link := NewLink("ls6sg6bffgtreoft", false, false)

		err := link.Delete()

		if err != nil {
			t.Fatal(err)
		}

	})
	t.Run("RemovesPassword", func(t *testing.T) {
		link := newTestLink(t, 0)
		other := newTestLink(t, 0)
		saveTestLinkPassword(t, link.LinkUID)
		saveTestLinkPassword(t, other.LinkUID)

		require.NoError(t, link.Delete())

		assert.Nil(t, FindLink(link.LinkUID))
		assert.Nil(t, FindPassword(link.LinkUID))
		assert.NotNil(t, FindLink(other.LinkUID))
		assert.NotNil(t, FindPassword(other.LinkUID))
		assert.NotNil(t, FindPassword(UserFixtures.Pointer("alice").UserUID))
	})
	t.Run("NoPassword", func(t *testing.T) {
		link := newTestLink(t, 0)

		require.NoError(t, link.Delete())

		assert.Nil(t, FindLink(link.LinkUID))
	})
	t.Run("PasswordDeleteFails", func(t *testing.T) {
		link := newTestLink(t, 0)
		saveTestLinkPassword(t, link.LinkUID)
		failTestDeletes(t, Password{}.TableName())

		assert.Error(t, link.Delete())

		assert.NotNil(t, FindLink(link.LinkUID))
		assert.NotNil(t, FindPassword(link.LinkUID))
	})
	t.Run("LinkDeleteFails", func(t *testing.T) {
		link := newTestLink(t, 0)
		saveTestLinkPassword(t, link.LinkUID)
		failTestDeletes(t, Link{}.TableName())

		assert.Error(t, link.Delete())

		assert.NotNil(t, FindLink(link.LinkUID))
		assert.NotNil(t, FindPassword(link.LinkUID))
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

// saveTestLinkPassword stores a password for a link UID and removes it when the test ends.
func saveTestLinkPassword(t *testing.T, linkUid string) {
	t.Helper()

	pw := NewPassword(linkUid, "LinkPassword123!", false)
	require.NoError(t, pw.Save())
	t.Cleanup(func() { _ = UnscopedDb().Delete(&Password{}, "uid = ?", linkUid).Error })
}

// failTestDeletes makes every delete from the named table fail until the test ends.
func failTestDeletes(t *testing.T, table string) {
	t.Helper()

	fail := func(tx *gorm.DB) {
		if tx.Statement.Table == table {
			_ = tx.AddError(errors.New("delete control"))
		}
	}

	require.NoError(t, Db().Callback().Delete().Before("gorm:begin_transaction").Register("test:link-delete", fail))
	t.Cleanup(func() { _ = Db().Callback().Delete().Remove("test:link-delete") })
}

func TestFindLink(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewLink("ls6sg6bffgtrjoft", false, false)

		link := &m

		if err := link.Save(); err != nil {
			t.Fatal(err)
		}
		uid := link.LinkUID
		t.Logf("%#v", link)
		r := FindLink(uid)
		t.Log(r)
		//TODO Why does it fail?
		//assert.Equal(t, "1jxf3jfn2k", r.LinkToken)
	})
	t.Run("Nil", func(t *testing.T) {
		r := FindLink("XXX")
		assert.Nil(t, r)
	})
}

func TestDeleteShareLinks(t *testing.T) {
	t.Run("RemovesPasswords", func(t *testing.T) {
		shareUid := rnd.GenerateUID(AlbumUID)
		first := NewLink(shareUid, false, false)
		require.NoError(t, first.Save())
		second := NewLink(shareUid, false, false)
		require.NoError(t, second.Save())
		saveTestLinkPassword(t, first.LinkUID)
		saveTestLinkPassword(t, second.LinkUID)
		other := newTestLink(t, 0)
		saveTestLinkPassword(t, other.LinkUID)

		require.NoError(t, DeleteShareLinks(shareUid))

		assert.Nil(t, FindLink(first.LinkUID))
		assert.Nil(t, FindLink(second.LinkUID))
		assert.Nil(t, FindPassword(first.LinkUID))
		assert.Nil(t, FindPassword(second.LinkUID))
		assert.NotNil(t, FindLink(other.LinkUID))
		assert.NotNil(t, FindPassword(other.LinkUID))
		assert.NotNil(t, FindPassword(UserFixtures.Pointer("alice").UserUID))
	})
	t.Run("NoLinks", func(t *testing.T) {
		assert.NoError(t, DeleteShareLinks(rnd.GenerateUID(AlbumUID)))
	})
	t.Run("PasswordDeleteFails", func(t *testing.T) {
		link := newTestLink(t, 0)
		saveTestLinkPassword(t, link.LinkUID)
		failTestDeletes(t, Password{}.TableName())

		assert.Error(t, DeleteShareLinks(link.ShareUID))

		assert.NotNil(t, FindLink(link.LinkUID))
		assert.NotNil(t, FindPassword(link.LinkUID))
	})
	t.Run("LinkDeleteFails", func(t *testing.T) {
		link := newTestLink(t, 0)
		saveTestLinkPassword(t, link.LinkUID)
		failTestDeletes(t, Link{}.TableName())

		assert.Error(t, DeleteShareLinks(link.ShareUID))

		assert.NotNil(t, FindLink(link.LinkUID))
		assert.NotNil(t, FindPassword(link.LinkUID))
	})
	t.Run("EmptyShareUid", func(t *testing.T) {
		assert.Error(t, DeleteShareLinks(""))
	})
}

func TestFindLinks(t *testing.T) {
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
	t.Run("NumericToken", func(t *testing.T) {
		link := NewLink("as6sg6bxpogaaba8", false, false)
		require.NoError(t, link.Save())
		t.Cleanup(func() { _ = UnscopedDb().Delete(&link).Error })
		require.NoError(t, UnscopedDb().Model(&link).UpdateColumn("link_token", "12345").Error)

		assert.Len(t, FindLinks("12345", ""), 1)

		// SQLite schemas of earlier releases stored the token 012345 as the number 12345.
		if DbDialect() == dsn.DialectSQLite {
			assert.Len(t, FindLinks("012345", ""), 1)
		} else {
			assert.Empty(t, FindLinks("012345", ""))
		}

		assert.Empty(t, FindLinks("012346", ""))
	})
}

func TestSqliteNumericToken(t *testing.T) {
	if DbDialect() != dsn.DialectSQLite {
		t.Run("OtherDialect", func(t *testing.T) {
			assert.Equal(t, "012345", sqliteNumericToken("012345"))
		})
		return
	}

	t.Run("LeadingZeros", func(t *testing.T) {
		assert.Equal(t, "12345", sqliteNumericToken("0012345"))
	})
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, "0", sqliteNumericToken("000"))
	})
	t.Run("NotNumeric", func(t *testing.T) {
		assert.Equal(t, "01a", sqliteNumericToken("01a"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", sqliteNumericToken(""))
	})
}

func TestFindRedeemableLinks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r := FindRedeemableLinks("1jxf3jfn2k", "")
		assert.Equal(t, "as6sg6bxpogaaba8", r[0].ShareUID)
	})
}

func TestFindRedeemableLinksByToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r := FindRedeemableLinksByToken("1jxf3jfn2k", "holiday-2030")
		assert.Equal(t, "as6sg6bxpogaaba8", r[0].ShareUID)
	})
	t.Run("WrongToken", func(t *testing.T) {
		r := FindRedeemableLinksByToken("wrongtoken", "holiday-2030")
		assert.Empty(t, r)
	})
	t.Run("EmptyToken", func(t *testing.T) {
		r := FindRedeemableLinksByToken("", "holiday-2030")
		assert.Empty(t, r)
	})
	t.Run("RejectsOversizedToken", func(t *testing.T) {
		r := FindRedeemableLinksByToken(strings.Repeat("a", 161), "holiday-2030")
		assert.Empty(t, r)
	})
	t.Run("RejectsUnusableToken", func(t *testing.T) {
		r := FindRedeemableLinksByToken("....", "as6sg6bxpogaaba8")
		assert.Empty(t, r)
	})
	t.Run("ViewLimitReached", func(t *testing.T) {
		link := newTestLink(t, 1)
		link.Redeem()

		assert.Empty(t, FindRedeemableLinksByToken(link.LinkToken, ""))
	})
}

func TestFindRedeemedLinks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r := FindRedeemedLinks("1jxf3jfn2k", "")
		assert.Equal(t, "as6sg6bxpogaaba8", r[0].ShareUID)
	})
	t.Run("ViewLimitReached", func(t *testing.T) {
		// The view limit bounds new redemptions, so a link that has reached it still resolves here.
		link := newTestLink(t, 1)
		link.Redeem()

		r := FindRedeemedLinks(link.LinkToken, "")

		require.Len(t, r, 1)
		assert.Equal(t, link.ShareUID, r[0].ShareUID)
	})
	t.Run("Expired", func(t *testing.T) {
		link := newTestLink(t, 0)
		expireTestLink(t, link)

		assert.Empty(t, FindRedeemedLinks(link.LinkToken, ""))
	})
	t.Run("Deleted", func(t *testing.T) {
		link := newTestLink(t, 0)

		if err := link.Delete(); err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, FindRedeemedLinks(link.LinkToken, ""))
	})
	t.Run("WrongShareUID", func(t *testing.T) {
		link := newTestLink(t, 0)
		assert.Empty(t, FindRedeemedLinks(link.LinkToken, "as6sg6bxpogaaba7"))
	})
}

func TestFindRedeemedLinksByToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r := FindRedeemedLinksByToken("  1jxf3jfn2k  ", "holiday-2030")
		assert.Equal(t, "as6sg6bxpogaaba8", r[0].ShareUID)
	})
	t.Run("EmptyToken", func(t *testing.T) {
		assert.Empty(t, FindRedeemedLinksByToken("", "holiday-2030"))
	})
	t.Run("RejectsUnusableToken", func(t *testing.T) {
		assert.Empty(t, FindRedeemedLinksByToken("....", "as6sg6bxpogaaba8"))
	})
	t.Run("RejectsOversizedToken", func(t *testing.T) {
		assert.Empty(t, FindRedeemedLinksByToken(strings.Repeat("a", 161), "holiday-2030"))
	})
}

func TestLink_String(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		link := NewLink("jhgko", false, false)
		uid := link.LinkUID
		assert.Equal(t, uid, link.String())
	})
}
