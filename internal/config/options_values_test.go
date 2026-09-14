package config

import (
	"math"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionFields(t *testing.T) {
	fields := optionFields()
	t.Run("KnownOptions", func(t *testing.T) {
		assert.Equal(t, reflect.TypeFor[int](), fields["JpegQuality"].Type)
		assert.Equal(t, reflect.TypeFor[int64](), fields["SessionMaxAge"].Type)
		assert.Equal(t, reflect.TypeFor[float64](), fields["StorageFree"].Type)
		assert.Equal(t, reflect.TypeFor[string](), fields["SiteUrl"].Type)
		assert.Equal(t, reflect.TypeFor[time.Duration](), fields["WakeupInterval"].Type)
	})
	t.Run("Exposed", func(t *testing.T) {
		assert.True(t, fields["JpegQuality"].Exposed)
		assert.True(t, fields["SiteUrl"].Exposed)
		assert.False(t, fields["VisionKey"].Exposed)
		assert.False(t, fields["AdminPassword"].Exposed)
		assert.False(t, fields["JoinToken"].Exposed)
	})
	t.Run("InlineStructIsFlattened", func(t *testing.T) {
		// The deprecated DSN is stored inline, so it is patched by its own name.
		field, found := fields["DatabaseDsn"]
		require.True(t, found)
		assert.False(t, field.Exposed)
		_, found = fields["Deprecated"]
		assert.False(t, found)
	})
	t.Run("SkipsUnnamedOptions", func(t *testing.T) {
		// Sponsor is tagged yaml:"-" and is never persisted.
		_, found := fields["Sponsor"]
		assert.False(t, found)
		_, found = fields["-"]
		assert.False(t, found)
	})
	t.Run("Cached", func(t *testing.T) {
		assert.True(t, reflect.ValueOf(optionFields()).Pointer() == reflect.ValueOf(fields).Pointer())
	})
}

func TestRemoveUnsupportedOptionValues(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		v := Values{
			"JpegQuality":   85,
			"SiteTitle":     "Test",
			"VisionKey":     "test-key",
			"VisionUri":     "https://example.com/",
			"AdminPassword": "secret",
		}
		removed := RemoveUnsupportedOptionValues(v)
		assert.Equal(t, []string{"AdminPassword", "VisionKey", "VisionUri"}, removed)
		assert.Equal(t, Values{"JpegQuality": 85, "SiteTitle": "Test"}, v)
	})
	t.Run("KeepsExposedOptions", func(t *testing.T) {
		v := Values{"JpegQuality": 85, "SiteUrl": "https://example.com/"}
		assert.Empty(t, RemoveUnsupportedOptionValues(v))
		assert.Len(t, v, 2)
	})
	t.Run("RemovesNamesThatAreNotOptions", func(t *testing.T) {
		v := Values{"SiteUrl": "https://example.com/", "NotAnOption": "value"}
		assert.Equal(t, []string{"NotAnOption"}, RemoveUnsupportedOptionValues(v))
		assert.Equal(t, Values{"SiteUrl": "https://example.com/"}, v)
	})
	t.Run("RemovesInlineDeprecatedDsn", func(t *testing.T) {
		v := Values{"DatabaseDsn": "user:pass@tcp(host)/db"}
		assert.Equal(t, []string{"DatabaseDsn"}, RemoveUnsupportedOptionValues(v))
		assert.Empty(t, v)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, RemoveUnsupportedOptionValues(Values{}))
		assert.Empty(t, RemoveUnsupportedOptionValues(nil))
	})
}

func TestCoerceOptionValues(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		v := Values{"JpegQuality": 85.61960784313726, "JpegSize": 17020.0, "ThumbSize": 3008.0}
		require.NoError(t, CoerceOptionValues(v))
		assert.Equal(t, int64(86), v["JpegQuality"])
		assert.Equal(t, int64(17020), v["JpegSize"])
		assert.Equal(t, int64(3008), v["ThumbSize"])
	})
	t.Run("RoundsToNearest", func(t *testing.T) {
		v := Values{"JpegQuality": 85.4}
		require.NoError(t, CoerceOptionValues(v))
		assert.Equal(t, int64(85), v["JpegQuality"])
	})
	t.Run("LeavesNonNumericValues", func(t *testing.T) {
		v := Values{"SiteUrl": "https://example.com/", "ReadOnly": true, "WakeupInterval": "1h"}
		require.NoError(t, CoerceOptionValues(v))
		assert.Equal(t, "https://example.com/", v["SiteUrl"])
		assert.Equal(t, true, v["ReadOnly"])
		assert.Equal(t, "1h", v["WakeupInterval"])
	})
	t.Run("LeavesFloatOptions", func(t *testing.T) {
		v := Values{"StorageFree": 12.5}
		require.NoError(t, CoerceOptionValues(v))
		assert.Equal(t, 12.5, v["StorageFree"])
	})
	t.Run("LeavesUnknownOptions", func(t *testing.T) {
		v := Values{"NotAnOption": 12.5}
		require.NoError(t, CoerceOptionValues(v))
		assert.Equal(t, 12.5, v["NotAnOption"])
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		v := Values{"JpegQuality": math.NaN()}
		err := CoerceOptionValues(v)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
		assert.Contains(t, err.Error(), "JpegQuality")
	})
	t.Run("Empty", func(t *testing.T) {
		assert.NoError(t, CoerceOptionValues(Values{}))
		assert.NoError(t, CoerceOptionValues(nil))
	})
}

func TestCoerceOptionValue(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		got, err := coerceOptionValue("JpegQuality", reflect.TypeFor[int](), 85.6)
		require.NoError(t, err)
		assert.Equal(t, int64(86), got)
	})
	t.Run("Uint", func(t *testing.T) {
		got, err := coerceOptionValue("Test", reflect.TypeFor[uint32](), 42.0)
		require.NoError(t, err)
		assert.Equal(t, uint64(42), got)
	})
	t.Run("Duration", func(t *testing.T) {
		got, err := coerceOptionValue("WakeupInterval", reflect.TypeFor[time.Duration](), 3600.0)
		require.NoError(t, err)
		assert.Equal(t, int64(3600), got)
	})
	t.Run("String", func(t *testing.T) {
		got, err := coerceOptionValue("SiteUrl", reflect.TypeFor[string](), 12.5)
		require.NoError(t, err)
		assert.Equal(t, 12.5, got)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := coerceOptionValue("JpegQuality", reflect.TypeFor[int](), math.Inf(1))
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
}

func TestCoerceOptionInt(t *testing.T) {
	t.Run("FromFloat", func(t *testing.T) {
		got, err := coerceOptionInt("Test", reflect.TypeFor[int](), 85.61960784313726)
		require.NoError(t, err)
		assert.Equal(t, int64(86), got)
	})
	t.Run("FromInt", func(t *testing.T) {
		got, err := coerceOptionInt("Test", reflect.TypeFor[int](), 42)
		require.NoError(t, err)
		assert.Equal(t, int64(42), got)
	})
	t.Run("FromUint", func(t *testing.T) {
		got, err := coerceOptionInt("Test", reflect.TypeFor[int](), uint64(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), got)
	})
	t.Run("Negative", func(t *testing.T) {
		got, err := coerceOptionInt("Test", reflect.TypeFor[int](), -3.5)
		require.NoError(t, err)
		assert.Equal(t, int64(-4), got)
	})
	t.Run("PassesThroughString", func(t *testing.T) {
		got, err := coerceOptionInt("Test", reflect.TypeFor[int64](), "1h")
		require.NoError(t, err)
		assert.Equal(t, "1h", got)
	})
	t.Run("OverflowsNarrowField", func(t *testing.T) {
		_, err := coerceOptionInt("Test", reflect.TypeFor[int8](), 300.0)
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
	t.Run("OverflowsInt64", func(t *testing.T) {
		_, err := coerceOptionInt("Test", reflect.TypeFor[int64](), 1e19)
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
	t.Run("OverflowsFromUint", func(t *testing.T) {
		_, err := coerceOptionInt("Test", reflect.TypeFor[int64](), uint64(math.MaxUint64))
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := coerceOptionInt("Test", reflect.TypeFor[int](), math.NaN())
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
}

func TestCoerceOptionUint(t *testing.T) {
	t.Run("FromFloat", func(t *testing.T) {
		got, err := coerceOptionUint("Test", reflect.TypeFor[uint](), 42.6)
		require.NoError(t, err)
		assert.Equal(t, uint64(43), got)
	})
	t.Run("FromUint", func(t *testing.T) {
		got, err := coerceOptionUint("Test", reflect.TypeFor[uint](), uint8(7))
		require.NoError(t, err)
		assert.Equal(t, uint64(7), got)
	})
	t.Run("FromInt", func(t *testing.T) {
		got, err := coerceOptionUint("Test", reflect.TypeFor[uint](), 7)
		require.NoError(t, err)
		assert.Equal(t, uint64(7), got)
	})
	t.Run("PassesThroughBool", func(t *testing.T) {
		got, err := coerceOptionUint("Test", reflect.TypeFor[uint](), true)
		require.NoError(t, err)
		assert.Equal(t, true, got)
	})
	t.Run("NegativeInt", func(t *testing.T) {
		_, err := coerceOptionUint("Test", reflect.TypeFor[uint](), -1)
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
	t.Run("NegativeFloat", func(t *testing.T) {
		_, err := coerceOptionUint("Test", reflect.TypeFor[uint](), -0.6)
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
	t.Run("OverflowsNarrowField", func(t *testing.T) {
		_, err := coerceOptionUint("Test", reflect.TypeFor[uint8](), 300.0)
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := coerceOptionUint("Test", reflect.TypeFor[uint](), math.Inf(-1))
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
}

func TestRoundOptionFloat(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		f, err := roundOptionFloat("Test", 85.61960784313726)
		require.NoError(t, err)
		assert.Equal(t, float64(86), f)
	})
	t.Run("RoundsHalfAway", func(t *testing.T) {
		f, err := roundOptionFloat("Test", 2.5)
		require.NoError(t, err)
		assert.Equal(t, float64(3), f)
	})
	t.Run("NaN", func(t *testing.T) {
		_, err := roundOptionFloat("Test", math.NaN())
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
	t.Run("Inf", func(t *testing.T) {
		_, err := roundOptionFloat("Test", math.Inf(1))
		assert.ErrorIs(t, err, ErrInvalidOptionValue)
	})
}

func TestRedactedOptions(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("CredentialIsReplaced", func(t *testing.T) {
		orig := c.Options().HttpsProxy
		t.Cleanup(func() { c.Options().HttpsProxy = orig })

		const configured = "https://proxy-user:proxy-pass@proxy.example.com:3128"

		c.Options().HttpsProxy = configured
		out := c.RedactedOptions()

		assert.NotContains(t, out.HttpsProxy, "proxy-pass")
		assert.Contains(t, out.HttpsProxy, "proxy.example.com")
		// The stored value is untouched, so the copy cannot become the configuration.
		assert.Equal(t, configured, c.Options().HttpsProxy)
	})
	t.Run("Unset", func(t *testing.T) {
		orig := c.Options().HttpsProxy
		t.Cleanup(func() { c.Options().HttpsProxy = orig })

		c.Options().HttpsProxy = ""
		assert.Empty(t, c.RedactedOptions().HttpsProxy)
	})
	t.Run("UnreadableValueIsNotReportedAsUnset", func(t *testing.T) {
		// A bare percent sign is enough for url.Parse to refuse the value. Reporting "" for it
		// would tell a reader the proxy is not configured, and a full-object write would then
		// store that answer.
		orig := c.Options().HttpsProxy
		t.Cleanup(func() { c.Options().HttpsProxy = orig })

		c.Options().HttpsProxy = "https://proxy-user:pa%ss@proxy.example.com:3128"
		out := c.RedactedOptions().HttpsProxy

		assert.Equal(t, RedactedOptionMarker, out)
		assert.NotEmpty(t, out)
	})
}

func TestRemoveRedactedOptionValues(t *testing.T) {
	c := NewConfig(CliTestContext())
	orig := c.Options().HttpsProxy
	t.Cleanup(func() { c.Options().HttpsProxy = orig })

	c.Options().HttpsProxy = "https://proxy-user:proxy-pass@proxy.example.com:3128"
	redacted := c.RedactedOptions().HttpsProxy

	t.Run("UnchangedValueIsDropped", func(t *testing.T) {
		v := Values{"HttpsProxy": redacted, "SiteTitle": "Example"}
		assert.Equal(t, []string{"HttpsProxy"}, c.RemoveRedactedOptionValues(v))
		assert.NotContains(t, v, "HttpsProxy")
		assert.Contains(t, v, "SiteTitle")
	})
	t.Run("NewValueIsKept", func(t *testing.T) {
		v := Values{"HttpsProxy": "https://other:pass@proxy.example.net:3128"}
		assert.Empty(t, c.RemoveRedactedOptionValues(v))
		assert.Contains(t, v, "HttpsProxy")
	})
	t.Run("ClearingIsKept", func(t *testing.T) {
		// An operator removing the proxy sends an empty string, which is not a rendered form.
		v := Values{"HttpsProxy": ""}
		assert.Empty(t, c.RemoveRedactedOptionValues(v))
		assert.Contains(t, v, "HttpsProxy")
	})
	t.Run("StaleRedactedValueIsDropped", func(t *testing.T) {
		// Read before the proxy changed, posted after. The test is on the value's own shape, not
		// on what is stored now, so it is dropped rather than written back as the password.
		v := Values{"HttpsProxy": "https://someone:***@proxy.example.net:3128"}
		assert.Equal(t, []string{"HttpsProxy"}, c.RemoveRedactedOptionValues(v))
		assert.NotContains(t, v, "HttpsProxy")
	})
	t.Run("MarkerIsDropped", func(t *testing.T) {
		v := Values{"HttpsProxy": RedactedOptionMarker}
		assert.Equal(t, []string{"HttpsProxy"}, c.RemoveRedactedOptionValues(v))
		assert.NotContains(t, v, "HttpsProxy")
	})
}

func TestExposedOptionNames(t *testing.T) {
	// ExposedOptionCount pins how many options the API returns. Adding a field without json:"-"
	// changes this number, so the author is asked once whether it can carry a credential and
	// therefore belongs in RedactedOptionNames.
	const ExposedOptionCount = 96

	fields := optionFields()

	var exposed []string

	for name, field := range fields {
		if field.Exposed {
			exposed = append(exposed, name)
		}
	}

	sort.Strings(exposed)

	assert.Len(t, exposed, ExposedOptionCount,
		"the set of options the API returns changed; if the new one can carry a credential, add it to RedactedOptionNames")
	assert.Contains(t, exposed, "HttpsProxy")
	assert.NotContains(t, exposed, "StoragePath", `a field tagged json:"-" is not returned`)

	for _, name := range RedactedOptionNames {
		assert.Contains(t, exposed, name, "redacting an option the API never returns would do nothing")
	}
}

func TestRedactOptionValue(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		assert.Empty(t, redactOptionValue(""))
	})
	t.Run("Credentials", func(t *testing.T) {
		out := redactOptionValue("https://proxy-user:proxy-pass@proxy.example.com:3128")
		assert.NotContains(t, out, "proxy-pass")
		assert.Contains(t, out, "proxy.example.com")
	})
	t.Run("Unreadable", func(t *testing.T) {
		assert.Equal(t, RedactedOptionMarker, redactOptionValue("https://u:pa%ss@proxy.example.com"))
	})
}

func TestIsRedactedOptionValue(t *testing.T) {
	t.Run("Marker", func(t *testing.T) {
		assert.True(t, isRedactedOptionValue(RedactedOptionMarker))
	})
	t.Run("RenderedPassword", func(t *testing.T) {
		assert.True(t, isRedactedOptionValue("https://u:***@proxy.example.com:3128"))
	})
	t.Run("RealValue", func(t *testing.T) {
		assert.False(t, isRedactedOptionValue("https://proxy-user:proxy-pass@proxy.example.com:3128"))
	})
	t.Run("NoCredentials", func(t *testing.T) {
		assert.False(t, isRedactedOptionValue("https://proxy.example.com:3128"))
	})
	t.Run("RenderedQuery", func(t *testing.T) {
		// A credential can sit in the query, which is rendered as the marker when it does not parse.
		assert.True(t, isRedactedOptionValue("https://proxy.example.com:3128/?token=***"))
		assert.True(t, isRedactedOptionValue("https://proxy.example.com:3128/?***"))
	})
	t.Run("RealQuery", func(t *testing.T) {
		assert.False(t, isRedactedOptionValue("https://proxy.example.com:3128/?tier=fast"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, isRedactedOptionValue(""))
	})
}
