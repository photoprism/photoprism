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
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, true, Empty("0"))
	})
	t.Run("One", func(t *testing.T) {
		assert.Equal(t, true, Empty("-1"))
	})
	t.Run("Date", func(t *testing.T) {
		assert.Equal(t, true, Empty("0000:00:00 00:00:00"))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, true, Empty("nil"))
	})
	t.Run("NaN", func(t *testing.T) {
		assert.Equal(t, true, Empty("NaN"))
	})
	t.Run("NULL", func(t *testing.T) {
		assert.Equal(t, true, Empty("NULL"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.Equal(t, true, Empty("*"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.Equal(t, true, Empty("%"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.True(t, Empty("-"))
	})
	t.Run("Z", func(t *testing.T) {
		assert.True(t, Empty("z"))
	})
	t.Run("Zz", func(t *testing.T) {
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
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("0"))
	})
	t.Run("One", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("-1"))
	})
	t.Run("Date", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("0000:00:00 00:00:00"))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("nil"))
	})
	t.Run("NaN", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("NaN"))
	})
	t.Run("NULL", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("NULL"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("*"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("%"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, NotEmpty("-"))
	})
	t.Run("Z", func(t *testing.T) {
		assert.False(t, NotEmpty("z"))
	})
	t.Run("Zz", func(t *testing.T) {
		assert.True(t, NotEmpty("zz"))
	})
}

func TestEmptyDateTime(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, EmptyDateTime(""))
	})
	t.Run("Case", func(t *testing.T) {
		assert.True(t, EmptyDateTime("-"))
	})
	t.Run("Z", func(t *testing.T) {
		assert.True(t, EmptyDateTime("z"))
	})
	t.Run("Zz", func(t *testing.T) {
		assert.False(t, EmptyDateTime("zz"))
	})
	t.Run("Zero", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0"))
	})
	t.Run("Num00Num00Num00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("00-00-00"))
	})
	t.Run("Num0000Num00Num00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000-00-00"))
	})
	t.Run("Num00Num00Num00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("00:00:00"))
	})
	t.Run("Num0000Num00Num00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000:00:00"))
	})
	t.Run("Num0000Num00Num00Num00Num00Num00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000-00-00 00-00-00"))
	})
	t.Run("Num0000Num00Num00Num00Num00Num00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000:00:00 00:00:00"))
	})
	t.Run("Num0000Num00Num00Num00Num00Num00", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0000-00-00 00:00:00"))
	})
	t.Run("Num0001Num01Num01Num00Num00Num00Num0000Utc", func(t *testing.T) {
		assert.True(t, EmptyDateTime("0001-01-01 00:00:00 +0000 UTC"))
	})
}

func TestDateTimeDefault(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, DateTimeDefault(""))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.True(t, DateTimeDefault("nil"))
	})
	t.Run("Num2002", func(t *testing.T) {
		assert.False(t, DateTimeDefault("2002"))
	})
	t.Run("Num1970Num01Num01", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1970-01-01"))
	})
	t.Run("Num1980Num01Num01", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1980-01-01"))
	})
	t.Run("Num1970Num01Num01Num00Num00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1970-01-01 00:00:00"))
	})
	t.Run("Num1970Num01Num01Num00Num00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1970:01:01 00:00:00"))
	})
	t.Run("Num1980Num01Num01Num00Num00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1980-01-01 00:00:00"))
	})
	t.Run("Num1980Num01Num01Num00Num00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("1980:01:01 00:00:00"))
	})
	t.Run("Num2002TwelveNum08TwelveNum00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("2002-12-08 12:00:00"))
	})
	t.Run("Num2002TwelveNum08TwelveNum00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("2002:12:08 12:00:00"))
	})
	t.Run("Num0000Num00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0000-00-00"))
	})
	t.Run("Num0000Num00Num00Num00Num00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0000-00-00 00-00-00"))
	})
	t.Run("Num0000Num00Num00Num00Num00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0000:00:00 00:00:00"))
	})
	t.Run("Num0000Num00Num00Num00Num00Num00", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0000-00-00 00:00:00"))
	})
	t.Run("Num0001Num01Num01Num00Num00Num00Num0000Utc", func(t *testing.T) {
		assert.True(t, DateTimeDefault("0001-01-01 00:00:00 +0000 UTC"))
	})
}
