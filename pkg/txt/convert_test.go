package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	t.Run("/2020/1212/20130518_142022_3D657EBD.jpg", func(t *testing.T) {
		result := Time("/2020/1212/20130518_142022_3D657EBD.jpg")
		//assert.False(t, result.IsZero())
		assert.Equal(t, "2013-05-18 14:20:22 +0000 UTC", result.String())
	})

	t.Run("telegram_2020_01_30_09_57_18.jpg", func(t *testing.T) {
		result := Time("telegram_2020_01_30_09_57_18.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", result.String())
	})

	t.Run("Screenshot 2019_05_21 at 10.45.52.png", func(t *testing.T) {
		result := Time("Screenshot 2019_05_21 at 10.45.52.png")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 10:45:52 +0000 UTC", result.String())
	})

	t.Run("telegram_2020-01-30_09-57-18.jpg", func(t *testing.T) {
		result := Time("telegram_2020-01-30_09-57-18.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", result.String())
	})

	t.Run("Screenshot 2019-05-21 at 10.45.52.png", func(t *testing.T) {
		result := Time("Screenshot 2019-05-21 at 10.45.52.png")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 10:45:52 +0000 UTC", result.String())
	})

	t.Run("telegram_2020-01-30_09-18.jpg", func(t *testing.T) {
		result := Time("telegram_2020-01-30_09-18.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2020-01-30 00:00:00 +0000 UTC", result.String())
	})

	t.Run("Screenshot 2019-05-21 at 10545.52.png", func(t *testing.T) {
		result := Time("Screenshot 2019-05-21 at 10545.52.png")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/2019-05-21/file2314.JPG", func(t *testing.T) {
		result := Time("/2019-05-21/file2314.JPG")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/2019.05.21", func(t *testing.T) {
		result := Time("/2019.05.21")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/05.21.2019", func(t *testing.T) {
		result := Time("/05.21.2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/21.05.2019", func(t *testing.T) {
		result := Time("/21.05.2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("05/21/2019", func(t *testing.T) {
		result := Time("05/21/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("21/05/2019", func(t *testing.T) {
		result := Time("21/05/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("2019/05/21", func(t *testing.T) {
		result := Time("2019/05/21")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})

	t.Run("2019/05/2145", func(t *testing.T) {
		result := Time("2019/05/2145")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/05/21/2019", func(t *testing.T) {
		result := Time("/05/21/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/21/05/2019", func(t *testing.T) {
		result := Time("/21/05/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/2019/05/21.jpeg", func(t *testing.T) {
		result := Time("/2019/05/21.jpeg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/2019/05/21/foo.txt", func(t *testing.T) {
		result := Time("/2019/05/21/foo.txt")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})

	t.Run("2019/21/05", func(t *testing.T) {
		result := Time("2019/21/05")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/2019/05/21/foo.jpg", func(t *testing.T) {
		result := Time("/2019/05/21/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/2019/21/05/foo.jpg", func(t *testing.T) {
		result := Time("/2019/21/05/foo.jpg")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/2019/5/foo.jpg", func(t *testing.T) {
		result := Time("/2019/5/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/2019/1/3/foo.jpg", func(t *testing.T) {
		result := Time("/2019/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-01-03 00:00:00 +0000 UTC", result.String())
	})

	t.Run("/1989/1/3/foo.jpg", func(t *testing.T) {
		result := Time("/1989/1/3/foo.jpg")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("545452019/1/3/foo.jpg", func(t *testing.T) {
		result := Time("/2019/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-01-03 00:00:00 +0000 UTC", result.String())
	})
}

func TestInt(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		result := Int("")
		assert.Equal(t, 0, result)
	})

	t.Run("non-numeric", func(t *testing.T) {
		result := Int("Screenshot")
		assert.Equal(t, 0, result)
	})

	t.Run("zero", func(t *testing.T) {
		result := Int("0")
		assert.Equal(t, 0, result)
	})

	t.Run("int", func(t *testing.T) {
		result := Int("123")
		assert.Equal(t, 123, result)
	})

	t.Run("negative int", func(t *testing.T) {
		result := Int("-123")
		assert.Equal(t, -123, result)
	})
}
