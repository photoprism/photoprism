package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFilterParams checks which DSN parameters are kept, dropped, or refused.
func TestFilterParams(t *testing.T) {
	rules := ParamRules{"parseTime": IsTrue, "timeout": ValidDuration, "tls": ValidBool}
	t.Run("Success", func(t *testing.T) {
		params, dropped, err := FilterParams("parseTime=true&timeout=15s", rules)
		require.NoError(t, err)
		assert.Equal(t, "parseTime=true&timeout=15s", params)
		assert.Empty(t, dropped)
	})
	t.Run("Dropped", func(t *testing.T) {
		params, dropped, err := FilterParams("option=value&timeout=15s&other&&x=/y&=x&a=b=c&", rules)
		require.NoError(t, err)
		assert.Equal(t, "timeout=15s", params)
		assert.Equal(t, []string{"option", "", "x", "", "a"}, dropped)
	})
	t.Run("Unnamed", func(t *testing.T) {
		params, dropped, err := FilterParams("parseTime=true&Zx8Qm2Lp9Rt4Vw6Yb1Nc3Kd5Hf7Gj0Sa&option=v", rules)
		require.NoError(t, err)
		assert.Equal(t, "parseTime=true", params)
		assert.Equal(t, []string{"", "option"}, dropped)
	})
	t.Run("Empty", func(t *testing.T) {
		params, dropped, err := FilterParams("", rules)
		require.NoError(t, err)
		assert.Equal(t, "", params)
		assert.Empty(t, dropped)
	})
	t.Run("InvalidValue", func(t *testing.T) {
		for _, query := range []string{"tls=custom", "parseTime=false", "timeout=-5s", "tls", "timeout=15s=x"} {
			_, _, err := FilterParams(query, rules)
			assert.Error(t, err, query)
		}
	})
	t.Run("Duplicate", func(t *testing.T) {
		_, _, err := FilterParams("tls=false&tls=true", rules)
		assert.EqualError(t, err, "duplicate dsn parameter tls")
	})
}

// TestMergeParams checks that missing default DSN parameters are appended.
func TestMergeParams(t *testing.T) {
	defaults := "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true"
	t.Run("AddMissing", func(t *testing.T) {
		assert.Equal(t, "tls=true&charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true", MergeParams("tls=true", defaults))
		assert.Equal(t, "charset=utf8&collation=utf8mb4_unicode_ci&parseTime=true", MergeParams("charset=utf8", defaults))
	})
	t.Run("RepeatedDefault", func(t *testing.T) {
		assert.Equal(t, "parseTime=true&timeout=5s", MergeParams("parseTime=true", "timeout=5s&timeout=10s"))
	})
	t.Run("Unchanged", func(t *testing.T) {
		assert.Equal(t, defaults, MergeParams(defaults, defaults))
		assert.Equal(t, "timeout=5s", MergeParams("timeout=5s", ""))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, defaults, MergeParams("", defaults))
	})
}

// TestHasParam checks DSN parameter lookups by name.
func TestHasParam(t *testing.T) {
	assert.True(t, HasParam("charset=utf8&parseTime=true", "charset"))
	assert.True(t, HasParam("tls", "tls"))
	assert.False(t, HasParam("charset=utf8", "collation"))
	assert.False(t, HasParam("", "charset"))
	assert.False(t, HasParam("xcharset=utf8", "charset"))
}

// TestQuery checks that DSN parameters are read as the driver reads them.
func TestQuery(t *testing.T) {
	assert.Equal(t, "charset=utf8mb4&parseTime=true", Query("user:secret@tcp(proxysql:6032)/?charset=utf8mb4&parseTime=true"))
	assert.Equal(t, "", Query("user:secret@tcp(proxysql:6032)/"))
	assert.Equal(t, "b=c", Query("user:p?w@tcp(proxysql:6032)/db?b=c"))
	assert.Equal(t, "charset=latin1", Query("user:secret@tcp(proxysql:6032)/db?x=/y?charset=latin1"))
	assert.Equal(t, "", Query(""))
}

// TestUtf8Params checks the charset and collation parameters of a DSN query.
func TestUtf8Params(t *testing.T) {
	assert.True(t, Utf8Params(""))
	assert.True(t, Utf8Params("charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true"))
	assert.False(t, Utf8Params("charset=latin1"))
	assert.False(t, Utf8Params("collation=latin1_swedish_ci"))
	assert.False(t, Utf8Params("charset=utf8mb4&charset=latin1"))
	assert.False(t, Utf8Params("charset"))
}
