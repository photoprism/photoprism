package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateTime(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		result := DateTime("", "")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("0000-00-00 00:00:00", func(t *testing.T) {
		result := DateTime("0000-00-00 00:00:00", "")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("0001-01-01 00:00:00 +0000 UTC", func(t *testing.T) {
		result := DateTime("0001-01-01 00:00:00 +0000 UTC", "")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2016: :     :  :  ", func(t *testing.T) {
		result := DateTime("2016: :     :  :  ", "")
		assert.Equal(t, "2016-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2016:  :__    :  :  ", func(t *testing.T) {
		result := DateTime("2016:  :__   :  :  ", "")
		assert.Equal(t, "2016-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2016:06:28   :  :??", func(t *testing.T) {
		result := DateTime("2016:06:28   :  :??", "")
		assert.Equal(t, "2016-06-28 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2016:06:28 09:45:49", func(t *testing.T) {
		result := DateTime("2016:06:28 09:45:49", "")
		assert.Equal(t, "2016-06-28 09:45:49 +0000 UTC", result.String())
	})
	t.Run("2016:06:28 09:45:49+10:00", func(t *testing.T) {
		result := DateTime("2016:06:28 09:45:49+10:00", "")
		assert.Equal(t, "2016-06-28 09:45:49 +1000 UTC+10:00", result.String())
	})
	t.Run("2016:06:28   :  :", func(t *testing.T) {
		result := DateTime("2016:06:28   :  :", "")
		assert.Equal(t, "2016-06-28 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2016/06/28T09-45:49", func(t *testing.T) {
		result := DateTime("2016/06/28T09-45:49", "")
		assert.Equal(t, "2016-06-28 09:45:49 +0000 UTC", result.String())
	})
	t.Run("2016:06:28T09:45:49Z", func(t *testing.T) {
		result := DateTime("2016:06:28T09:45:49Z", "")
		assert.Equal(t, "2016-06-28 09:45:49 +0000 UTC", result.String())
	})
	t.Run("2016:06:28T09:45:  Z", func(t *testing.T) {
		result := DateTime("2016:06:28T09:45:  Z", "")
		assert.Equal(t, "2016-06-28 09:45:00 +0000 UTC", result.String())
	})
	t.Run("2016:06:28T09:45:  ", func(t *testing.T) {
		result := DateTime("2016:06:28T09:45:  ", "")
		assert.Equal(t, "2016-06-28 09:45:00 +0000 UTC", result.String())
	})
	t.Run("2016:06:28T09:45:  ZABC", func(t *testing.T) {
		result := DateTime("2016:06:28T09:45:  ZABC", "")
		assert.Equal(t, "2016-06-28 09:45:00 +0000 UTC", result.String())
	})
	t.Run("2016:06:28T09:45:  ABC", func(t *testing.T) {
		result := DateTime("2016:06:28T09:45:  ABC", "")
		assert.Equal(t, "2016-06-28 09:45:00 +0000 UTC", result.String())
	})
	t.Run("2016:06:28 09:45:49+10:00ABC", func(t *testing.T) {
		result := DateTime("2016:06:28 09:45:49+10:00ABC", "")
		assert.Equal(t, "2016-06-28 09:45:49 +1000 UTC+10:00", result.String())
	})
	t.Run("  2016:06:28 09:45:49-01:30ABC", func(t *testing.T) {
		result := DateTime("  2016:06:28 09:45:49-01:30ABC", "")
		assert.Equal(t, "2016-06-28 09:45:49 -0130 UTC-01:30", result.String())
	})
	t.Run("2016:06:28 09:45:49-0130", func(t *testing.T) {
		result := DateTime("2016:06:28 09:45:49-0130", "")
		assert.Equal(t, "2016-06-28 09:45:49 -0130 UTC-01:30", result.String())
	})
	t.Run("UTC/016:06:28 09:45:49-0130", func(t *testing.T) {
		result := DateTime("2016:06:28 09:45:49-0130", "UTC")
		assert.Equal(t, "2016-06-28 11:15:49 +0000 UTC", result.String())
	})
	t.Run("UTC/016:06:28 09:45:49-0130", func(t *testing.T) {
		result := DateTime("2016:06:28 09:45:49.0130", "UTC")
		assert.Equal(t, "2016-06-28 09:45:49.013 +0000 UTC", result.String())
	})
	t.Run("2012:08:08 22:07:18", func(t *testing.T) {
		result := DateTime("2012:08:08 22:07:18", "")
		assert.Equal(t, "2012-08-08 22:07:18 +0000 UTC", result.String())
	})
	t.Run("2020-01-30_09-57-18", func(t *testing.T) {
		result := DateTime("2020-01-30_09-57-18", "")
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", result.String())
	})
	t.Run("EuropeBerlin/2016:06:28 09:45:49+10:00ABC", func(t *testing.T) {
		result := DateTime("2016:06:28 09:45:49+10:00ABC", "Europe/Berlin")
		assert.Equal(t, "2016-06-28 01:45:49 +0200 CEST", result.String())
	})
	t.Run("EuropeBerlin/  2016:06:28 09:45:49-01:30ABC", func(t *testing.T) {
		result := DateTime("  2016:06:28 09:45:49-01:30ABC", "Europe/Berlin")
		assert.Equal(t, "2016-06-28 13:15:49 +0200 CEST", result.String())
	})
	t.Run("EuropeBerlin/2012:08:08 22:07:18", func(t *testing.T) {
		result := DateTime("2012:08:08 22:07:18", "Europe/Berlin")
		assert.Equal(t, "2012-08-08 22:07:18 +0200 CEST", result.String())
	})
	t.Run("EuropeBerlin/2020-01-30_09-57-18", func(t *testing.T) {
		result := DateTime("2020-01-30_09-57-18", "Europe/Berlin")
		assert.Equal(t, "2020-01-30 09:57:18 +0100 CET", result.String())
	})
	t.Run("EuropeBerlin/2020-10-17-48-24.950488", func(t *testing.T) {
		result := DateTime("2020:10:17 17:48:24.9508123", "UTC")
		assert.Equal(t, "2020-10-17 17:48:24.9508123 +0000 UTC", result.UTC().String())
		assert.Equal(t, "2020-10-17 17:48:24.9508123", result.Format("2006-01-02 15:04:05.999999999"))
	})
	t.Run("UTC/0000:00:00 00:00:00", func(t *testing.T) {
		assert.True(t, DateTime("0000:00:00 00:00:00", "UTC").IsZero())
	})
	t.Run("2022-09-03T17:48:26-07:00", func(t *testing.T) {
		result := DateTime("2022-09-03T17:48:26-07:00", "")
		assert.Equal(t, "2022-09-04 00:48:26 +0000 UTC", result.UTC().String())
		assert.Equal(t, "2022-09-03 17:48:26", result.Format("2006-01-02 15:04:05"))
	})
}

func TestIsTime(t *testing.T) {
	t.Run("/2020/1212/20130518_142022_3D657EBD.jpg", func(t *testing.T) {
		assert.False(t, IsTime("/2020/1212/20130518_142022_3D657EBD.jpg"))
	})

	t.Run("telegram_2020_01_30_09_57_18.jpg", func(t *testing.T) {
		assert.False(t, IsTime("telegram_2020_01_30_09_57_18.jpg"))
	})

	t.Run("", func(t *testing.T) {
		assert.False(t, IsTime(""))
	})

	t.Run("Screenshot 2019_05_21 at 10.45.52.png", func(t *testing.T) {
		assert.False(t, IsTime("Screenshot 2019_05_21 at 10.45.52.png"))
	})

	t.Run("telegram_2020-01-30_09-57-18.jpg", func(t *testing.T) {
		assert.False(t, IsTime("telegram_2020-01-30_09-57-18.jpg"))
	})

	t.Run("2013-05-18", func(t *testing.T) {
		assert.True(t, IsTime("2013-05-18"))
	})

	t.Run("2013-05-18 12:01:01", func(t *testing.T) {
		assert.True(t, IsTime("2013-05-18 12:01:01"))
	})

	t.Run("20130518_142022", func(t *testing.T) {
		assert.True(t, IsTime("20130518_142022"))
	})

	t.Run("2020_01_30_09_57_18", func(t *testing.T) {
		assert.True(t, IsTime("2020_01_30_09_57_18"))
	})

	t.Run("2019_05_21 at 10.45.52", func(t *testing.T) {
		assert.True(t, IsTime("2019_05_21 at 10.45.52"))
	})

	t.Run("2020-01-30_09-57-18", func(t *testing.T) {
		assert.True(t, IsTime("2020-01-30_09-57-18"))
	})
}
