package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeInt64to32(t *testing.T) {
	t.Run("Valid Test 100", func(t *testing.T) {
		assert.Equal(t, int32(100), SafeInt64to32(int64(100)))
	})

	t.Run("Valid Test -100", func(t *testing.T) {
		assert.Equal(t, int32(-100), SafeInt64to32(int64(-100)))
	})

	t.Run("Valid Test 0", func(t *testing.T) {
		assert.Equal(t, int32(0), SafeInt64to32(int64(0)))
	})

	t.Run("Valid Test 2147483645", func(t *testing.T) {
		assert.Equal(t, int32(2147483645), SafeInt64to32(int64(2147483645)))
	})

	t.Run("Valid Test -2147483645", func(t *testing.T) {
		assert.Equal(t, int32(-2147483645), SafeInt64to32(int64(-2147483645)))
	})

	t.Run("Exceed Max Test 3000000000", func(t *testing.T) {
		assert.Equal(t, int32(2147483647), SafeInt64to32(int64(3000000000)))
	})

	t.Run("Exceed Min Test -3000000000", func(t *testing.T) {
		assert.Equal(t, int32(-2147483648), SafeInt64to32(int64(-3000000000)))
	})

}

func TestSafeInt64toint(t *testing.T) {
	t.Run("Valid Test 100", func(t *testing.T) {
		assert.Equal(t, int(100), SafeInt64toint(int64(100)))
	})

	t.Run("Valid Test -100", func(t *testing.T) {
		assert.Equal(t, int(-100), SafeInt64toint(int64(-100)))
	})

	t.Run("Valid Test 0", func(t *testing.T) {
		assert.Equal(t, int(0), SafeInt64toint(int64(0)))
	})

	t.Run("Valid Test 2147483645", func(t *testing.T) {
		assert.Equal(t, int(2147483645), SafeInt64toint(int64(2147483645)))
	})

	t.Run("Valid Test -2147483645", func(t *testing.T) {
		assert.Equal(t, int(-2147483645), SafeInt64toint(int64(-2147483645)))
	})

	t.Run("Exceed Max Test 3000000000", func(t *testing.T) {
		assert.Equal(t, int(2147483647), SafeInt64toint(int64(3000000000)))
	})

	t.Run("Exceed Min Test -3000000000", func(t *testing.T) {
		assert.Equal(t, int(-2147483648), SafeInt64toint(int64(-3000000000)))
	})

}
