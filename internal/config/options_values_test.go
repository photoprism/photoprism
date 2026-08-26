package config

import (
	"math"
	"reflect"
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
