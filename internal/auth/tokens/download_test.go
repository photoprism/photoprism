package tokens

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/pkg/authn/authtoken"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// withKey configures the package Download signer for a test and restores it afterward.
func withKey(t *testing.T, key []byte, path string) {
	t.Helper()
	orig := *Download
	Download.Key, Download.SignaturePath = key, path
	t.Cleanup(func() { *Download = orig })
}

// withPolicy sets the download delivery policy vars for a test and restores them afterward.
func withPolicy(t *testing.T, public bool, coarse string) {
	t.Helper()
	origPublic, origCoarse := PublicMode, CoarseDownload
	PublicMode, CoarseDownload = public, coarse
	t.Cleanup(func() { PublicMode, CoarseDownload = origPublic, origCoarse })
}

func TestDownloadToken(t *testing.T) {
	sid := rnd.SessionID("tokens-delivery-test")

	t.Run("PublicMode", func(t *testing.T) {
		withKey(t, make([]byte, KeyLen), "/api/v1")
		withPolicy(t, true, "coarse-x")
		assert.Equal(t, PublicToken, DownloadToken(sid))
	})
	t.Run("StaticTokenDoesNotOverrideSession", func(t *testing.T) {
		withKey(t, make([]byte, KeyLen), "/api/v1")
		withPolicy(t, false, "static-abc")
		// A configured static token must not opt a session out of per-session scoping.
		v := DownloadToken(sid)
		assert.NotEqual(t, "static-abc", v)
		assert.Equal(t, sid, strings.SplitN(v, ".", 3)[1])
	})
	t.Run("SignedForSession", func(t *testing.T) {
		withKey(t, make([]byte, KeyLen), "/api/v1")
		withPolicy(t, false, "coarse-x")
		v := DownloadToken(sid)
		assert.NotEqual(t, "coarse-x", v)
		assert.Equal(t, sid, strings.SplitN(v, ".", 3)[1])
	})
	t.Run("CoarseForSessionless", func(t *testing.T) {
		withKey(t, make([]byte, KeyLen), "/api/v1")
		withPolicy(t, false, "coarse-x")
		// No session to sign for (public/share config) falls back to the coarse instance token.
		assert.Equal(t, "coarse-x", DownloadToken(""))
	})
}

func TestIsCoarseDownload(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		withPolicy(t, false, "coarse-x")
		assert.True(t, IsCoarseDownload("coarse-x"))
	})
	t.Run("Mismatch", func(t *testing.T) {
		withPolicy(t, false, "coarse-x")
		assert.False(t, IsCoarseDownload("other"))
	})
	t.Run("EmptyCoarseRejectsEverything", func(t *testing.T) {
		withPolicy(t, false, "")
		assert.False(t, IsCoarseDownload(""))
		assert.False(t, IsCoarseDownload("anything"))
	})
}

func TestSignDownload(t *testing.T) {
	key := make([]byte, KeyLen)
	sid := rnd.SessionID("tokens-sign-test")

	t.Run("RoundTrip", func(t *testing.T) {
		withKey(t, key, "/api/v1")
		v := SignDownload(sid)
		parts := strings.SplitN(v, ".", 3)
		assert.Len(t, parts, 3)
		assert.Equal(t, sid, parts[1])
		assert.Contains(t, parts[2], authtoken.Prefix)
		expires, err := strconv.ParseInt(parts[0], 10, 64)
		assert.NoError(t, err)
		got, ok := VerifyDownload(expires, sid, parts[2])
		assert.True(t, ok)
		assert.Equal(t, sid, got)
	})
	t.Run("InvalidSessionID", func(t *testing.T) {
		withKey(t, key, "/api/v1")
		assert.Empty(t, SignDownload("not-a-session-id"))
	})
	t.Run("AbsentKeyRefusesToMint", func(t *testing.T) {
		withKey(t, nil, "/api/v1")
		assert.Empty(t, SignDownload(sid))
	})
	t.Run("HonorsTtlLifetime", func(t *testing.T) {
		withKey(t, key, "/api/v1")
		orig := ttl.DownloadToken
		ttl.DownloadToken = 60
		defer func() { ttl.DownloadToken = orig }()
		parts := strings.SplitN(SignDownload(sid), ".", 3)
		expires, _ := strconv.ParseInt(parts[0], 10, 64)
		assert.InDelta(t, time.Now().Add(60*time.Second).Unix(), expires, 5)
	})
}

func TestVerifyDownload(t *testing.T) {
	key := make([]byte, KeyLen)
	key[0] = 0x42
	sid := rnd.SessionID("tokens-verify-test")

	sign := func() (int64, string) {
		parts := strings.SplitN(SignDownload(sid), ".", 3)
		expires, _ := strconv.ParseInt(parts[0], 10, 64)
		return expires, parts[2]
	}

	t.Run("Valid", func(t *testing.T) {
		withKey(t, key, "/api/v1")
		expires, token := sign()
		got, ok := VerifyDownload(expires, sid, token)
		assert.True(t, ok)
		assert.Equal(t, sid, got)
	})
	t.Run("Forged", func(t *testing.T) {
		withKey(t, key, "/api/v1")
		expires, token := sign()
		_, ok := VerifyDownload(expires, sid, token+"x")
		assert.False(t, ok)
	})
	t.Run("TamperedSid", func(t *testing.T) {
		withKey(t, key, "/api/v1")
		expires, token := sign()
		_, ok := VerifyDownload(expires, rnd.SessionID("other-session"), token)
		assert.False(t, ok)
	})
	t.Run("Expired", func(t *testing.T) {
		withKey(t, key, "/api/v1")
		past := time.Now().Add(-time.Hour).Unix()
		token := authtoken.Sign(Download.Key, Download.SignaturePath, past, url.Values{"sid": {sid}}, "")
		_, ok := VerifyDownload(past, sid, token)
		assert.False(t, ok)
	})
	t.Run("AbsentKeyRejects", func(t *testing.T) {
		withKey(t, nil, "/api/v1")
		_, ok := VerifyDownload(time.Now().Add(time.Hour).Unix(), sid, "HS256-anything")
		assert.False(t, ok)
	})
	t.Run("Empty", func(t *testing.T) {
		withKey(t, key, "/api/v1")
		_, ok := VerifyDownload(0, "", "")
		assert.False(t, ok)
	})
}
