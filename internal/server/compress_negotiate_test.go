package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNegotiateEncoding(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		prefs  []string
		want   string
	}{
		{"PrefersFirstConfigured", "zstd, gzip", []string{"zstd", "gzip"}, "zstd"},
		{"FallsBackToSecondConfigured", "gzip", []string{"zstd", "gzip"}, "gzip"},
		{"OperatorPrefsWinOverClientOrder", "zstd, gzip", []string{"gzip", "zstd"}, "gzip"},
		{"QZeroIsExplicitRefusal", "zstd;q=0, gzip", []string{"zstd", "gzip"}, "gzip"},
		{"QZeroBeatsStarFallback", "zstd;q=0, *", []string{"zstd", "gzip"}, "gzip"},
		{"StarMatchesAnyConfigured", "*", []string{"zstd", "gzip"}, "zstd"},
		{"StarWithZeroQualityYieldsIdentity", "*;q=0", []string{"zstd", "gzip"}, "identity"},
		{"UnknownAcceptYieldsIdentity", "br", []string{"zstd", "gzip"}, "identity"},
		{"EmptyAcceptYieldsIdentity", "", []string{"zstd", "gzip"}, "identity"},
		{"MissingHeaderYieldsIdentity", "", []string{"gzip"}, "identity"},
		{"EmptyPrefsYieldsIdentity", "gzip, zstd", nil, "identity"},
		{"WhitespaceTolerant", "  zstd  ,  gzip  ", []string{"zstd", "gzip"}, "zstd"},
		{"AcceptIsCaseInsensitive", "ZSTD, GZIP", []string{"zstd", "gzip"}, "zstd"},
		{"PrefsAreCaseInsensitive", "gzip", []string{"GZIP"}, "gzip"},
		{"MalformedQDefaultsToOne", "gzip;q=abc", []string{"gzip"}, "gzip"},
		{"FractionalQAboveZeroAccepted", "gzip;q=0.1", []string{"gzip"}, "gzip"},
		{"MultipleParametersIgnoreNonQ", "gzip;foo=bar;q=0.8", []string{"gzip"}, "gzip"},
		{"DropsEmptyTokens", " , gzip , ", []string{"gzip"}, "gzip"},
		// The next two cases document a deliberate pragmatic deviation from a
		// strict reading of RFC 9110 §12.5.3: when the client refuses every
		// configured coding AND identity, RFC says the server SHOULD return
		// 406 Not Acceptable. We instead fall back to identity, matching the
		// long-standing behavior of gin-contrib/gzip and most Go HTTP
		// middleware. Returning 406 from compression middleware would
		// short-circuit handler logic, audit events, and authn checks — too
		// invasive for a SHOULD-level requirement that real browsers never
		// trigger (they always include identity implicitly with q=1).
		{"AllRefusedFallsBackToIdentity", "gzip;q=0, zstd;q=0, identity;q=0", []string{"zstd", "gzip"}, "identity"},
		{"AllConfiguredRefusedNoIdentityRefused", "gzip;q=0, zstd;q=0", []string{"zstd", "gzip"}, "identity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NegotiateEncoding(tt.accept, tt.prefs))
		})
	}
}

func TestParseAcceptEncoding(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, parseAcceptEncoding(""))
	})

	t.Run("SingleToken", func(t *testing.T) {
		got := parseAcceptEncoding("gzip")
		assert.Equal(t, map[string]float64{"gzip": 1.0}, got)
	})

	t.Run("MultipleWithQuality", func(t *testing.T) {
		got := parseAcceptEncoding("zstd;q=0.9, gzip;q=0.5, br;q=0")
		assert.InDelta(t, 0.9, got["zstd"], 1e-9)
		assert.InDelta(t, 0.5, got["gzip"], 1e-9)
		assert.InDelta(t, 0.0, got["br"], 1e-9)
	})

	t.Run("Wildcard", func(t *testing.T) {
		got := parseAcceptEncoding("*")
		assert.Equal(t, 1.0, got["*"])
	})

	t.Run("MalformedQualityFallsBackToOne", func(t *testing.T) {
		got := parseAcceptEncoding("gzip;q=not-a-number")
		assert.InDelta(t, 1.0, got["gzip"], 1e-9)
	})

	t.Run("LowercasesTokens", func(t *testing.T) {
		got := parseAcceptEncoding("GZIP, ZSTD")
		_, gzipOK := got["gzip"]
		_, zstdOK := got["zstd"]
		assert.True(t, gzipOK)
		assert.True(t, zstdOK)
	})
}
