package tokens

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// withPreviewKey configures the package Download and Preview signers for a test and restores them afterward.
func withPreviewKey(t *testing.T, key []byte, path string) {
	t.Helper()
	origDownload, origPreview := *Download, *Preview
	Download.Key, Download.SignaturePath = key, path
	Preview.Key, Preview.SignaturePath = key, path
	t.Cleanup(func() { *Download, *Preview = origDownload, origPreview })
}

// splitPreview parses a compact preview token and fails the test if it has another shape.
func splitPreview(t *testing.T, value string) (int64, string, string) {
	t.Helper()
	expires, sid, token, ok := ParseCompact(value)
	require.True(t, ok, value)
	return expires, sid, token
}

func TestPreviewExpires(t *testing.T) {
	now := time.Date(2026, 10, 21, 7, 30, 0, 0, time.UTC)
	endOfDay := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC).Unix()

	t.Run("SessionExpiry", func(t *testing.T) {
		exp := now.Add(time.Hour)
		assert.Equal(t, exp.Unix(), PreviewExpires(exp, now))
	})
	t.Run("NoExpiry", func(t *testing.T) {
		assert.Equal(t, endOfDay, PreviewExpires(time.Time{}, now))
		assert.Equal(t, endOfDay, PreviewExpires(time.Time{}, now.Add(16*time.Hour)))
	})
	t.Run("DayBoundary", func(t *testing.T) {
		midnight := endOfDay - PreviewDay
		assert.Equal(t, endOfDay, PreviewExpires(time.Time{}, time.Unix(midnight-1, 0)))
		assert.Equal(t, endOfDay+PreviewDay, PreviewExpires(time.Time{}, time.Unix(midnight, 0)))
		assert.GreaterOrEqual(t, PreviewExpires(time.Time{}, time.Unix(midnight-1, 0))-(midnight-1), PreviewDay)
	})
}

func TestPreviewToken(t *testing.T) {
	sid := rnd.SessionID("tokens-preview-test")

	t.Run("SessionExpiry", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		exp := time.Now().Add(time.Hour).Truncate(time.Second)
		v := PreviewToken(sid, exp)
		expires, gotSid, token := splitPreview(t, v)
		assert.Equal(t, exp.Unix(), expires)
		assert.Equal(t, sid, gotSid)
		id, ok := VerifyPreview(expires, gotSid, token)
		assert.True(t, ok)
		assert.Equal(t, sid, id)

		// The value is the same for every mint of the same session.
		assert.Equal(t, v, PreviewToken(sid, exp))
	})
	t.Run("NoExpiry", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		v := PreviewToken(sid, time.Time{})
		expires, _, token := splitPreview(t, v)
		assert.Equal(t, PreviewExpires(time.Time{}, time.Now()), expires)
		_, ok := VerifyPreview(expires, sid, token)
		assert.True(t, ok)
		assert.Equal(t, v, PreviewToken(sid, time.Time{}))
	})
	t.Run("InvalidSession", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		assert.Equal(t, "", PreviewToken("not-a-session", time.Now().Add(time.Hour)))
	})
	t.Run("Unconfigured", func(t *testing.T) {
		withPreviewKey(t, nil, "/api/v1")
		assert.Equal(t, "", PreviewToken(sid, time.Now().Add(time.Hour)))
	})
}

func TestSignPreview(t *testing.T) {
	sid := rnd.SessionID("tokens-sign-preview-test")
	exp := time.Now().Add(time.Hour).Unix()

	t.Run("Success", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		v := SignPreview(sid, exp)
		assert.True(t, strings.HasPrefix(v, strconv.FormatInt(exp, 10)+"."+sid+".HS256-"))
		assert.LessOrEqual(t, len(v), 256)
	})
	t.Run("InvalidSession", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		assert.Equal(t, "", SignPreview("", exp))
	})
	t.Run("Unconfigured", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen-1), "/api/v1")
		assert.Equal(t, "", SignPreview(sid, exp))
	})
}

func TestVerifyPreview(t *testing.T) {
	sid := rnd.SessionID("tokens-verify-preview-test")
	other := rnd.SessionID("tokens-verify-preview-other")
	exp := time.Now().Add(time.Hour).Unix()

	t.Run("Success", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		_, _, token := splitPreview(t, SignPreview(sid, exp))
		id, ok := VerifyPreview(exp, sid, token)
		assert.True(t, ok)
		assert.Equal(t, sid, id)
	})
	t.Run("Tampered", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		_, _, token := splitPreview(t, SignPreview(sid, exp))
		_, ok := VerifyPreview(exp, other, token)
		assert.False(t, ok, "sid")
		_, ok = VerifyPreview(exp+60, sid, token)
		assert.False(t, ok, "expiry")
		last := token[len(token)-1]
		flipped := byte('A')
		if last == 'A' {
			flipped = 'B'
		}
		_, ok = VerifyPreview(exp, sid, token[:len(token)-1]+string(flipped))
		assert.False(t, ok, "signature")
	})
	t.Run("Expired", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		past := time.Now().Add(-time.Minute).Unix()
		_, _, token := splitPreview(t, SignPreview(sid, past))
		_, ok := VerifyPreview(past, sid, token)
		assert.False(t, ok)
	})
	t.Run("NoKey", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		_, _, token := splitPreview(t, SignPreview(sid, exp))
		Preview.Key = nil
		_, ok := VerifyPreview(exp, sid, token)
		assert.False(t, ok)
	})
	t.Run("DownloadTokenRefused", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		expires, gotSid, token := splitPreview(t, SignDownload(sid))
		_, ok := VerifyPreview(expires, gotSid, token)
		assert.False(t, ok)
	})
	t.Run("PreviewTokenNotDownload", func(t *testing.T) {
		withPreviewKey(t, make([]byte, KeyLen), "/api/v1")
		_, _, token := splitPreview(t, SignPreview(sid, exp))
		_, ok := VerifyDownload(exp, sid, token)
		assert.False(t, ok)
	})
}

func TestPreviewParams(t *testing.T) {
	sid := rnd.SessionID("tokens-preview-params-test")
	params := previewParams(sid)
	assert.Equal(t, sid, params.Get("sid"))
	assert.Equal(t, PurposePreview, params.Get("kind"))
}

func TestParseCompact(t *testing.T) {
	sid := rnd.SessionID("tokens-parse-compact-test")

	t.Run("Success", func(t *testing.T) {
		expires, gotSid, token, ok := ParseCompact("1792570707." + sid + ".HS256-abc")
		assert.True(t, ok)
		assert.Equal(t, int64(1792570707), expires)
		assert.Equal(t, sid, gotSid)
		assert.Equal(t, "HS256-abc", token)
	})
	t.Run("InvalidShape", func(t *testing.T) {
		for _, v := range []string{
			"",
			"2ud3qfpu",
			"..",
			"1792570707." + sid,
			"1792570707." + sid + ".HS256.abc",
			"1792570707" + sid + "HS256-abc",
			"x." + sid + ".HS256-abc",
			"-1." + sid + ".HS256-abc",
			"01792570707." + sid + ".HS256-abc",
			"+1792570707." + sid + ".HS256-abc",
			"1792570707.notasession.HS256-abc",
			"1792570707." + sid + ".",
		} {
			_, _, _, ok := ParseCompact(v)
			assert.False(t, ok, v)
		}
	})
}
