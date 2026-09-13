package entity

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

// newTestLink saves a share link for its own album with the given view limit, zero meaning unlimited.
func newTestLink(t *testing.T, maxViews uint) *Link {
	t.Helper()

	return newTestLinkWithToken(t, "", maxViews)
}

// newTestLinkWithToken saves a share link for its own album under a given token, so a case can put
// two links of different records on one token, which the composite unique index permits.
func newTestLinkWithToken(t *testing.T, token string, maxViews uint) *Link {
	t.Helper()

	link := NewLink(rnd.GenerateUID(AlbumUID), false, false)
	link.MaxViews = maxViews

	if token != "" {
		link.LinkToken = token
	}

	if err := link.Save(); err != nil {
		t.Fatal(err)
	}

	return &link
}

// expireTestLink backdates a saved link so it has passed its expiration time.
func expireTestLink(t *testing.T, link *Link) {
	t.Helper()

	link.LinkExpires = 3600
	link.ModifiedAt = Now().Add(-2 * time.Hour)

	values := Values{"link_expires": link.LinkExpires, "modified_at": link.ModifiedAt}

	if err := Db().Model(&Link{}).Where("link_uid = ?", link.LinkUID).Updates(values).Error; err != nil {
		t.Fatal(err)
	}
}

// linkViews reads the stored view count of a link.
func linkViews(t *testing.T, link *Link) uint {
	t.Helper()

	found := FindLink(link.LinkUID)
	require.NotNil(t, found)

	return found.LinkViews
}

// redeemInNewSession logs in through the production session flow with the share token and the
// optional credentials, and returns the resulting session and login error.
func redeemInNewSession(t *testing.T, token, username, password string) (*Session, error) {
	t.Helper()

	m := NewSession(unix.Day, unix.Hour*6)
	m.SetClientIP("1.2.3.4")

	frm := form.Login{Token: token, Username: username, Password: password}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
	c.Request.RemoteAddr = "1.2.3.4"

	return m, m.LogIn(frm, c)
}

// reloadSession saves a session and reads it back from the database, so its shares are resolved the
// way a later request resolves them.
func reloadSession(t *testing.T, m *Session) *Session {
	t.Helper()

	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	FlushSessionCache()

	found, err := FindSession(m.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	return found
}

func TestLinkRedemption(t *testing.T) {
	t.Run("OneViewGrantsOneAnonymousSession", func(t *testing.T) {
		link := newTestLink(t, 1)

		first, err := redeemInNewSession(t, link.LinkToken, "", "")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, first.Status)
		assert.Equal(t, UIDs{link.ShareUID}, first.SharedUIDs())
		assert.Equal(t, uint(1), linkViews(t, link))

		second, err := redeemInNewSession(t, link.LinkToken, "", "")

		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, second.Status)
		assert.Empty(t, second.SharedUIDs())
		assert.Equal(t, uint(1), linkViews(t, link), "a refused redemption must not count a view")
	})
	t.Run("TwoViewsGrantTwoAnonymousSessions", func(t *testing.T) {
		link := newTestLink(t, 2)

		for i := 1; i <= 2; i++ {
			sess, err := redeemInNewSession(t, link.LinkToken, "", "")

			require.NoError(t, err, "redemption %d", i)
			assert.Equal(t, http.StatusOK, sess.Status)
			assert.Equal(t, UIDs{link.ShareUID}, sess.SharedUIDs())
			assert.Equal(t, uint(i), linkViews(t, link))
		}

		third, err := redeemInNewSession(t, link.LinkToken, "", "")

		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, third.Status)
		assert.Empty(t, third.SharedUIDs())
		assert.Equal(t, uint(2), linkViews(t, link))
	})
	t.Run("ReachedViewLimitKeepsTheSessionItGranted", func(t *testing.T) {
		link := newTestLink(t, 1)

		sess, err := redeemInNewSession(t, link.LinkToken, "", "")
		require.NoError(t, err)

		reloaded := reloadSession(t, sess)

		assert.Equal(t, UIDs{link.ShareUID}, reloaded.SharedUIDs())
		assert.True(t, reloaded.HasShare(link.ShareUID))
	})
	t.Run("DeletedLinkRevokesTheSessionItGranted", func(t *testing.T) {
		link := newTestLink(t, 0)

		sess, err := redeemInNewSession(t, link.LinkToken, "", "")
		require.NoError(t, err)
		require.Equal(t, UIDs{link.ShareUID}, sess.SharedUIDs())

		if err = link.Delete(); err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, reloadSession(t, sess).SharedUIDs())
	})
	t.Run("ExpiredLinkRevokesTheSessionItGranted", func(t *testing.T) {
		link := newTestLink(t, 0)

		sess, err := redeemInNewSession(t, link.LinkToken, "", "")
		require.NoError(t, err)
		require.Equal(t, UIDs{link.ShareUID}, sess.SharedUIDs())

		expireTestLink(t, link)

		assert.Empty(t, reloadSession(t, sess).SharedUIDs())
	})
	t.Run("ExpiredLinkIsNeverRedeemed", func(t *testing.T) {
		link := newTestLink(t, 0)
		expireTestLink(t, link)

		sess, err := redeemInNewSession(t, link.LinkToken, "", "")

		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, sess.Status)
		assert.Equal(t, uint(0), linkViews(t, link))
	})
	t.Run("OneViewGrantsOneRegisteredUser", func(t *testing.T) {
		link := newTestLink(t, 1)

		alice, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, alice.Status)
		assert.True(t, alice.HasShare(link.ShareUID))
		assert.Equal(t, uint(1), linkViews(t, link))

		// Control: the second account's own credentials are accepted without the exhausted token.
		control, err := redeemInNewSession(t, "", "bob", "Bobbob123!")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, control.Status)

		bob, err := redeemInNewSession(t, link.LinkToken, "bob", "Bobbob123!")

		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, bob.Status)
		assert.False(t, bob.HasShare(link.ShareUID))
		assert.Equal(t, uint(1), linkViews(t, link))
	})
	t.Run("RegisteredUserKeepsTheShareTheViewLimitGranted", func(t *testing.T) {
		link := newTestLink(t, 1)

		sess, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(link.ShareUID))

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		assert.True(t, user.RefreshShares().HasShare(link.ShareUID))
	})
	t.Run("DeletedLinkRevokesTheRegisteredShare", func(t *testing.T) {
		link := newTestLink(t, 0)

		sess, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(link.ShareUID))

		if err = link.Delete(); err != nil {
			t.Fatal(err)
		}

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		assert.False(t, user.RefreshShares().HasShare(link.ShareUID))
	})
	t.Run("SessionRedeemsTheSameTokenOnlyOnce", func(t *testing.T) {
		// The sharing page redeems on every load, and the view limit is already reached, so this is
		// the case where re-presenting a held token must be recognized rather than admitted afresh.
		link := newTestLink(t, 1)

		sess, err := redeemInNewSession(t, link.LinkToken, "", "")
		require.NoError(t, err)
		require.Equal(t, uint(1), linkViews(t, link))

		assert.Equal(t, 1, sess.RedeemToken(link.LinkToken))
		assert.Equal(t, uint(1), linkViews(t, link))
		assert.Equal(t, UIDs{link.ShareUID}, sess.SharedUIDs())
	})
	t.Run("RegisteredUserRedeemsTheSameTokenOnlyOnce", func(t *testing.T) {
		// The link carries an expiry so the row it issues holds a value a second write would move.
		link := newTestLink(t, 1)
		link.LinkExpires = 3600

		if err := link.Save(); err != nil {
			t.Fatal(err)
		}

		first, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, first.HasShare(link.ShareUID))
		require.Equal(t, uint(1), linkViews(t, link))

		share := FindUserShare(UserShare{UserUID: first.UserUID, ShareUID: link.ShareUID})
		require.NotNil(t, share)
		require.NotNil(t, share.ExpiresAt)

		issued := *share.ExpiresAt

		// Correct credentials with a token the account already holds must not be refused, must not
		// count another view, and must leave the row as it was issued.
		second, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, second.Status)
		assert.True(t, second.HasShare(link.ShareUID))
		assert.Equal(t, uint(1), linkViews(t, link))

		after := FindUserShare(UserShare{UserUID: first.UserUID, ShareUID: link.ShareUID})
		require.NotNil(t, after)
		require.NotNil(t, after.ExpiresAt)
		assert.WithinDuration(t, issued, *after.ExpiresAt, time.Second)
	})
	t.Run("ExhaustedCoTokenLinkGrantsNoNewSession", func(t *testing.T) {
		// The unique index is (share_uid, link_token), so two records may share one token. The
		// limited one must grant only the sessions it admitted.
		token := rnd.Base36(10)
		limited := newTestLinkWithToken(t, token, 1)
		unlimited := newTestLinkWithToken(t, token, 0)

		first, err := redeemInNewSession(t, token, "", "")

		require.NoError(t, err)
		assert.ElementsMatch(t, UIDs{limited.ShareUID, unlimited.ShareUID}, first.SharedUIDs())
		assert.Equal(t, uint(1), linkViews(t, limited))

		second, err := redeemInNewSession(t, token, "", "")

		require.NoError(t, err)
		assert.Equal(t, UIDs{unlimited.ShareUID}, second.SharedUIDs())
		assert.False(t, second.HasShare(limited.ShareUID))
		assert.Equal(t, uint(1), linkViews(t, limited), "the exhausted link must count no further view")

		// The session the limited link did admit keeps it.
		assert.ElementsMatch(t, UIDs{limited.ShareUID, unlimited.ShareUID}, reloadSession(t, first).SharedUIDs())
	})
	t.Run("ExhaustedLinkDoesNotAdoptAnotherLinksRegisteredShare", func(t *testing.T) {
		// Two links of one record cannot share a token, so this is the second link of the same album
		// under a token of its own. Its budget is spent, so it may not rewrite the row the first
		// link issued, take over its expiration time, or become the link that revocation follows.
		issued := newTestLink(t, 0)
		issued.LinkExpires = 3600

		if err := issued.Save(); err != nil {
			t.Fatal(err)
		}

		exhausted := newTestLinkWithToken(t, "", 1)
		exhausted.ShareUID = issued.ShareUID
		exhausted.Perm = 128

		if err := exhausted.Save(); err != nil {
			t.Fatal(err)
		}

		sess, err := redeemInNewSession(t, issued.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(issued.ShareUID))

		// A stranger spends the second link's only view.
		_, err = redeemInNewSession(t, exhausted.LinkToken, "bob", "Bobbob123!")
		require.NoError(t, err)
		require.Equal(t, uint(1), linkViews(t, exhausted))

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		assert.Equal(t, 0, user.RedeemToken(exhausted.LinkToken))

		share := FindUserShare(UserShare{UserUID: user.GetUID(), ShareUID: issued.ShareUID})
		require.NotNil(t, share)
		assert.Equal(t, issued.LinkUID, share.LinkUID, "the row must still name the link that issued it")
		assert.NotNil(t, share.ExpiresAt, "the row must keep its own expiration time")
		assert.NotEqual(t, uint(128), share.Perm)
		assert.Equal(t, uint(1), linkViews(t, exhausted))

		// Revocation still follows the link the share was redeemed through.
		if err = issued.Delete(); err != nil {
			t.Fatal(err)
		}

		assert.False(t, user.RefreshShares().HasShare(issued.ShareUID))
	})
	t.Run("ExhaustedLinkDoesNotReviveALapsedRegisteredShare", func(t *testing.T) {
		link := newTestLink(t, 1)

		sess, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(link.ShareUID))

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		share := FindUserShare(UserShare{UserUID: user.GetUID(), ShareUID: link.ShareUID})
		require.NotNil(t, share)

		if err = share.Updates(Values{"expires_at": Now().Add(-time.Minute)}); err != nil {
			t.Fatal(err)
		}

		require.False(t, user.RefreshShares().HasShare(link.ShareUID))

		// The link admits no further redemption, so it cannot reinstate the lapsed row.
		assert.Equal(t, 0, user.RedeemToken(link.LinkToken))
		assert.False(t, user.RefreshShares().HasShare(link.ShareUID))
		assert.Equal(t, uint(1), linkViews(t, link))
	})
	t.Run("ExhaustedLinkDoesNotMoveTheExpiryOfTheShareItIssued", func(t *testing.T) {
		// A share is a snapshot of the terms it was redeemed under, and the sharing page re-presents
		// the token on every load, so a link with no budget left counts the row and writes nothing.
		link := newTestLink(t, 1)
		link.LinkExpires = 3600
		link.ModifiedAt = Now().Add(-30 * time.Minute)

		values := Values{"link_expires": link.LinkExpires, "modified_at": link.ModifiedAt}

		if err := Db().Model(&Link{}).Where("link_uid = ?", link.LinkUID).Updates(values).Error; err != nil {
			t.Fatal(err)
		}

		sess, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(link.ShareUID))
		require.Equal(t, uint(1), linkViews(t, link))

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		share := FindUserShare(UserShare{UserUID: user.GetUID(), ShareUID: link.ShareUID})
		require.NotNil(t, share)
		require.NotNil(t, share.ExpiresAt)

		issued := *share.ExpiresAt

		// The owner edits the spent link, which slides the link's own window forward.
		edited := FindLink(link.LinkUID)
		require.NotNil(t, edited)
		edited.Perm = 64

		if err = edited.Save(); err != nil {
			t.Fatal(err)
		}

		require.True(t, edited.ExpiresAt().After(issued), "the edit must move the link's own expiration time")

		assert.Equal(t, 1, user.RedeemToken(link.LinkToken))

		after := FindUserShare(UserShare{UserUID: user.GetUID(), ShareUID: link.ShareUID})
		require.NotNil(t, after)
		require.NotNil(t, after.ExpiresAt)
		assert.WithinDuration(t, issued, *after.ExpiresAt, time.Second, "the share must keep the expiration time it was issued with")
		assert.Equal(t, uint(0), after.Perm, "the share must keep the permissions it was issued with")
		assert.Equal(t, uint(1), linkViews(t, link))
	})
	t.Run("ExhaustedLinkDoesNotMakeTheShareItIssuedPermanent", func(t *testing.T) {
		// Clearing a link's expiry is what an owner does by saving the share dialog without one, so
		// this is the shape in which a spent link must still leave the row bounded.
		link := newTestLink(t, 1)
		link.LinkExpires = 3600

		if err := link.Save(); err != nil {
			t.Fatal(err)
		}

		sess, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(link.ShareUID))

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		share := FindUserShare(UserShare{UserUID: user.GetUID(), ShareUID: link.ShareUID})
		require.NotNil(t, share)
		require.NotNil(t, share.ExpiresAt)

		edited := FindLink(link.LinkUID)
		require.NotNil(t, edited)
		edited.LinkExpires = 0

		if err = edited.Save(); err != nil {
			t.Fatal(err)
		}

		require.Nil(t, edited.ExpiresAt(), "the edited link must no longer expire")

		assert.Equal(t, 1, user.RedeemToken(link.LinkToken))

		after := FindUserShare(UserShare{UserUID: user.GetUID(), ShareUID: link.ShareUID})
		require.NotNil(t, after)
		assert.NotNil(t, after.ExpiresAt, "the share must keep the expiration time it was issued with")
	})
	t.Run("RedeemableLinkRefreshesTheRegisteredShareItTakesOver", func(t *testing.T) {
		// A second link that still admits a redemption may take the row over, and it carries its own
		// terms and spends a view when it does.
		issued := newTestLink(t, 0)
		taking := newTestLinkWithToken(t, "", 0)
		taking.ShareUID = issued.ShareUID
		taking.LinkExpires = 3600
		taking.Perm = 64

		if err := taking.Save(); err != nil {
			t.Fatal(err)
		}

		sess, err := redeemInNewSession(t, issued.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(issued.ShareUID))

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		assert.Equal(t, 1, user.RedeemToken(taking.LinkToken))

		share := FindUserShare(UserShare{UserUID: user.GetUID(), ShareUID: issued.ShareUID})
		require.NotNil(t, share)
		assert.Equal(t, taking.LinkUID, share.LinkUID)
		assert.Equal(t, uint(64), share.Perm)
		require.NotNil(t, share.ExpiresAt)
		assert.WithinDuration(t, *taking.ExpiresAt(), *share.ExpiresAt, time.Minute)
		assert.Equal(t, uint(0), linkViews(t, taking), "a share the account already holds counts no view")

		// Re-presenting the token the row now names counts none either.
		assert.Equal(t, 1, user.RedeemToken(taking.LinkToken))
		assert.Equal(t, uint(0), linkViews(t, taking))
	})
	t.Run("ReinstatingALapsedShareCountsAView", func(t *testing.T) {
		// A lapsed row grants nothing, so the link that writes it again is admitting the account, and
		// its budget bounds how often that may happen.
		link := newTestLink(t, 2)

		sess, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(link.ShareUID))
		require.Equal(t, uint(1), linkViews(t, link))

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		share := FindUserShare(UserShare{UserUID: user.GetUID(), ShareUID: link.ShareUID})
		require.NotNil(t, share)

		if err = share.Updates(Values{"expires_at": Now().Add(-time.Minute)}); err != nil {
			t.Fatal(err)
		}

		require.False(t, user.RefreshShares().HasShare(link.ShareUID))

		assert.Equal(t, 1, user.RedeemToken(link.LinkToken))
		assert.True(t, user.RefreshShares().HasShare(link.ShareUID))
		assert.Equal(t, uint(2), linkViews(t, link), "reinstating a lapsed share must count a view")

		// The budget is now spent, so the next visitor is refused.
		stranger, err := redeemInNewSession(t, link.LinkToken, "", "")

		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, stranger.Status)
		assert.Equal(t, uint(2), linkViews(t, link))
	})
	t.Run("AlternatingTwoTokensForOneRecordCountsNoFurtherView", func(t *testing.T) {
		// Each link records itself on the row it writes, so alternating the two tokens of one record
		// re-points the row on every presentation. The account holds the record throughout, so only
		// the first share counts.
		first := newTestLink(t, 3)
		second := newTestLinkWithToken(t, "", 3)
		second.ShareUID = first.ShareUID

		if err := second.Save(); err != nil {
			t.Fatal(err)
		}

		sess, err := redeemInNewSession(t, first.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(first.ShareUID))
		require.Equal(t, uint(1), linkViews(t, first))

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		for range 3 {
			assert.Equal(t, 1, user.RedeemToken(second.LinkToken))
			assert.Equal(t, 1, user.RedeemToken(first.LinkToken))
		}

		assert.Equal(t, uint(1), linkViews(t, first))
		assert.Equal(t, uint(0), linkViews(t, second))

		// Both budgets are intact, so each still admits its own visitors.
		for _, link := range []*Link{first, second} {
			visitor, visitErr := redeemInNewSession(t, link.LinkToken, "", "")

			require.NoError(t, visitErr)
			assert.Equal(t, UIDs{link.ShareUID}, visitor.SharedUIDs())
		}
	})
	t.Run("ExhaustedLinkIsNotRescuedByACoTokenSibling", func(t *testing.T) {
		// One record reachable through two tokens, the second of which also carries a link on another
		// record. Holding the record through the first token must not rescue the exhausted link.
		shared := rnd.GenerateUID(AlbumUID)

		open := newTestLink(t, 0)
		open.ShareUID = shared

		if err := open.Save(); err != nil {
			t.Fatal(err)
		}

		token := rnd.Base36(10)
		limited := newTestLinkWithToken(t, token, 1)
		limited.ShareUID = shared

		if err := limited.Save(); err != nil {
			t.Fatal(err)
		}

		sibling := newTestLinkWithToken(t, token, 0)

		// A stranger spends the limited link's only view.
		_, err := redeemInNewSession(t, token, "", "")
		require.NoError(t, err)
		require.Equal(t, uint(1), linkViews(t, limited))

		sess, err := redeemInNewSession(t, open.LinkToken, "", "")
		require.NoError(t, err)
		require.Equal(t, 1, sess.RedeemToken(token), "only the sibling admits this session")

		sess.SetData(sess.GetData())

		assert.ElementsMatch(t, UIDs{shared, sibling.ShareUID}, sess.SharedUIDs())
		assert.Equal(t, uint(1), linkViews(t, limited))

		// The record is held through the open link alone, so removing it withdraws the record.
		if err = open.Delete(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, UIDs{sibling.ShareUID}, reloadSession(t, sess).SharedUIDs())
	})
	t.Run("RefreshingSharesCountsNoView", func(t *testing.T) {
		link := newTestLink(t, 0)

		sess, err := redeemInNewSession(t, link.LinkToken, "", "")
		require.NoError(t, err)
		require.Equal(t, uint(1), linkViews(t, link))

		for range 3 {
			sess.GetData().RefreshShares()
		}

		assert.Equal(t, uint(1), linkViews(t, link))
		assert.Equal(t, UIDs{link.ShareUID}, sess.SharedUIDs())
	})
	t.Run("HeldMultiLinkTokenReportsEveryShareItGrants", func(t *testing.T) {
		token := rnd.Base36(10)
		first := newTestLinkWithToken(t, token, 0)
		second := newTestLinkWithToken(t, token, 0)

		sess, err := redeemInNewSession(t, token, "", "")
		require.NoError(t, err)
		require.ElementsMatch(t, UIDs{first.ShareUID, second.ShareUID}, sess.SharedUIDs())

		assert.Equal(t, 2, sess.RedeemToken(token))
	})
	t.Run("RegisteredShareCarriesAndHonorsTheLinkExpiry", func(t *testing.T) {
		// A registered entitlement is a persisted row whose expiration time is taken from the link
		// when it is redeemed, and the lookup filters on that column.
		link := newTestLink(t, 0)
		link.LinkExpires = 3600

		if err := link.Save(); err != nil {
			t.Fatal(err)
		}

		sess, err := redeemInNewSession(t, link.LinkToken, "alice", "Alice123!")
		require.NoError(t, err)
		require.True(t, sess.HasShare(link.ShareUID))

		user := FindUser(User{UserName: "alice"})
		require.NotNil(t, user)

		share := FindUserShare(UserShare{UserUID: user.GetUID(), ShareUID: link.ShareUID})
		require.NotNil(t, share)
		require.NotNil(t, share.ExpiresAt)
		assert.WithinDuration(t, *link.ExpiresAt(), *share.ExpiresAt, time.Minute)

		if err = share.Updates(Values{"expires_at": Now().Add(-time.Minute)}); err != nil {
			t.Fatal(err)
		}

		assert.False(t, user.RefreshShares().HasShare(link.ShareUID))
	})
}
