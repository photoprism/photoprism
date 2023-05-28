package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, true, Empty(""))
	})
	t.Run("EnNew", func(t *testing.T) {
		assert.Equal(t, false, Empty(EnNew))
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.Equal(t, false, Empty("     new "))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.Equal(t, false, Empty("NEW"))
	})
	t.Run("Lowercase", func(t *testing.T) {
		assert.Equal(t, false, Empty("new"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, false, Empty("New"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, false, Empty("non"))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, true, Empty("0"))
	})
	t.Run("-1", func(t *testing.T) {
		assert.Equal(t, true, Empty("-1"))
	})
	t.Run("Date", func(t *testing.T) {
		assert.Equal(t, true, Empty("0000:00:00 00:00:00"))
	})
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, true, Empty("nil"))
	})
	t.Run("NaN", func(t *testing.T) {
		assert.Equal(t, true, Empty("NaN"))
	})
	t.Run("NULL", func(t *testing.T) {
		assert.Equal(t, true, Empty("NULL"))
	})
	t.Run("*", func(t *testing.T) {
		assert.Equal(t, true, Empty("*"))
	})
	t.Run("%", func(t *testing.T) {
		assert.Equal(t, true, Empty("%"))
	})
	t.Run("-", func(t *testing.T) {
		assert.True(t, Empty("-"))
	})
	t.Run("z", func(t *testing.T) {
		assert.True(t, Empty("z"))
	})
	t.Run("zz", func(t *testing.T) {
		assert.False(t, Empty("zz"))
	})
}

func TestNotEmpty(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty(""))
	})
	t.Run("EnNew", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty(EnNew))
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("     new "))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("NEW"))
	})
	t.Run("Lowercase", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("new"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("New"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("non"))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("0"))
	})
	t.Run("-1", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("-1"))
	})
	t.Run("Date", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("0000:00:00 00:00:00"))
	})
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("nil"))
	})
	t.Run("NaN", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("NaN"))
	})
	t.Run("NULL", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("NULL"))
	})
	t.Run("*", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("*"))
	})
	t.Run("%", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("%"))
	})
	t.Run("-", func(t *testing.T) {
		assert.False(t, NotEmpty("-"))
	})
	t.Run("z", func(t *testing.T) {
		assert.False(t, NotEmpty("z"))
	})
	t.Run("zz", func(t *testing.T) {
		assert.True(t, NotEmpty("zz"))
	})
}

func TestEmptyDateTime(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, EmptyDateTime(""))
	})
	t.Run("-", func(t *testing.T) {
		assert.True(t, EmptyDateTime("-"))
	})
	t.Run("z", func(t *testing.T) {
		assert.True(t, EmptyDateTime("z"))
	})
	t.Run("zz", func(t *testing.T) {
		assert.False(t, EmptyDateTime("zz"))
	})
	t.Run("0", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0"))
	})
	t.Run("00-00-00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("00-00-00"))
	})
	t.Run("0000-00-00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000-00-00"))
	})
	t.Run("00:00:00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("00:00:00"))
	})
	t.Run("0000:00:00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000:00:00"))
	})
	t.Run("0000-00-00 00-00-00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000-00-00 00-00-00"))
	})
	t.Run("0000:00:00 00:00:00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000:00:00 00:00:00"))
	})
	t.Run("0000-00-00 00:00:00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000-00-00 00:00:00"))
	})
	t.Run("0001-01-01 00:00:00 +0000 UTC", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0001-01-01 00:00:00 +0000 UTC"))
	})
}

func TestDateTimeDefault(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, DateTimeDefault(""))
	})
	t.Run("nil", func(t *testing.T) {
		assert.True(t, DateTimeDefault("nil"))
	})
	t.Run("2002", func(t *testing.T) {
		assert.False(t, DateTimeDefault("2002"))
	})
	t.Run("1970-01-01", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1970-01-01"))
	})
	t.Run("1980-01-01", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1980-01-01"))
	})
	t.Run("1970-01-01 00:00:00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1970-01-01 00:00:00"))
	})
	t.Run("1970:01:01 00:00:00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1970:01:01 00:00:00"))
	})
	t.Run("1980-01-01 00:00:00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1980-01-01 00:00:00"))
	})
	t.Run("1980:01:01 00:00:00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1980:01:01 00:00:00"))
	})
	t.Run("2002-12-08 12:00:00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("2002-12-08 12:00:00"))
	})
	t.Run("2002:12:08 12:00:00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("2002:12:08 12:00:00"))
	})
	t.Run("0000-00-00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0000-00-00"))
	})
	t.Run("0000-00-00 00-00-00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0000-00-00 00-00-00"))
	})
	t.Run("0000:00:00 00:00:00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0000:00:00 00:00:00"))
	})
	t.Run("0000-00-00 00:00:00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0000-00-00 00:00:00"))
	})
	t.Run("0001-01-01 00:00:00 +0000 UTC", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0001-01-01 00:00:00 +0000 UTC"))
	})
}
