package txt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsTime(t *testing.T) {
	t.Run("Num2020Num1212Num20130518Num142022ThreeDNum657EbdJpg", func(t *testing.T) {
		assert.False(t, IsTime("/2020/1212/20130518_142022_3D657EBD.jpg"))
	})
	t.Run("TelegramNum2020Num01Num30Num09Num57EighteenJpg", func(t *testing.T) {
		assert.False(t, IsTime("telegram_2020_01_30_09_57_18.jpg"))
	})
	t.Run("Case", func(t *testing.T) {
		assert.False(t, IsTime(""))
	})
	t.Run("ScreenshotNum2019Num05Num21AtTenNum45Num52Png", func(t *testing.T) {
		assert.False(t, IsTime("Screenshot 2019_05_21 at 10.45.52.png"))
	})
	t.Run("TelegramNum2020Num01Num30Num09Num57EighteenJpg", func(t *testing.T) {
		assert.False(t, IsTime("telegram_2020-01-30_09-57-18.jpg"))
	})
	t.Run("Num2013Num05Eighteen", func(t *testing.T) {
		assert.True(t, IsTime("2013-05-18"))
	})
	t.Run("Num2013Num05EighteenTwelveNum01Num01", func(t *testing.T) {
		assert.True(t, IsTime("2013-05-18 12:01:01"))
	})
	t.Run("Num20130518Num142022", func(t *testing.T) {
		assert.True(t, IsTime("20130518_142022"))
	})
	t.Run("Num2020Num01Num30Num09Num57Eighteen", func(t *testing.T) {
		assert.True(t, IsTime("2020_01_30_09_57_18"))
	})
	t.Run("Num2019Num05Num21AtTenNum45Num52", func(t *testing.T) {
		assert.True(t, IsTime("2019_05_21 at 10.45.52"))
	})
	t.Run("Num2020Num01Num30Num09Num57Eighteen", func(t *testing.T) {
		assert.True(t, IsTime("2020-01-30_09-57-18"))
	})
}

func TestDateTime(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", DateTime(nil))
	})
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, "", DateTime(&time.Time{}))
	})
	t.Run("Num1665389030", func(t *testing.T) {
		now := time.Unix(1665389030, 0)
		assert.Equal(t, "2022-10-10 08:03:50", DateTime(&now))
	})
}

func TestUnixTime(t *testing.T) {
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, "", UnixTime(0))
	})
	t.Run("Num1665389030", func(t *testing.T) {
		assert.Equal(t, "2022-10-10 08:03:50", UnixTime(1665389030))
	})
}

func TestParseTimeUTC(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		result := ParseTimeUTC("")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num0000Num00Num00Num00Num00Num00", func(t *testing.T) {
		result := ParseTimeUTC("0000-00-00 00:00:00")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num0001Num01Num01Num00Num00Num00Num0000Utc", func(t *testing.T) {
		result := ParseTimeUTC("0001-01-01 00:00:00 +0000 UTC")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2016", func(t *testing.T) {
		result := ParseTimeUTC("2016: :     :  :  ")
		assert.Equal(t, "2016-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2020TwelveNum29", func(t *testing.T) {
		result := ParseTimeUTC("2020-12-29")
		assert.Equal(t, "2020-12-29 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2020TwelveNum29", func(t *testing.T) {
		result := ParseTimeUTC("2020/12/29")
		assert.Equal(t, "2020-12-29 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2020", func(t *testing.T) {
		result := ParseTimeUTC("2020")
		assert.Equal(t, "2020-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2020Num02", func(t *testing.T) {
		result := ParseTimeUTC("2020/02")
		assert.Equal(t, "2020-02-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num29TwelveNum2020", func(t *testing.T) {
		result := ParseTimeUTC("29.12.2020")
		assert.Equal(t, "2020-12-29 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num29TwelveNum2020", func(t *testing.T) {
		result := ParseTimeUTC("29/12/2020")
		assert.Equal(t, "2020-12-29 00:00:00 +0000 UTC", result.String())
	})
}

func TestParseTime(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		result := ParseTime("", "")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num0000Num00Num00Num00Num00Num00", func(t *testing.T) {
		result := ParseTime("0000-00-00 00:00:00", "")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num0001Num01Num01Num00Num00Num00Num0000Utc", func(t *testing.T) {
		result := ParseTime("0001-01-01 00:00:00 +0000 UTC", "")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2016", func(t *testing.T) {
		result := ParseTime("2016: :     :  :  ", "")
		assert.Equal(t, "2016-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2016", func(t *testing.T) {
		result := ParseTime("2016:  :__   :  :  ", "")
		assert.Equal(t, "2016-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28", func(t *testing.T) {
		result := ParseTime("2016:06:28   :  :??", "")
		assert.Equal(t, "2016-06-28 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28Num09Num45Num49", func(t *testing.T) {
		result := ParseTime("2016:06:28 09:45:49", "")
		assert.Equal(t, "2016-06-28 09:45:49 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28Num09Num45Num49TenNum00", func(t *testing.T) {
		result := ParseTime("2016:06:28 09:45:49+10:00", "")
		assert.Equal(t, "2016-06-28 09:45:49 +1000 UTC+10:00", result.String())
	})
	t.Run("Num2016Num06Num28", func(t *testing.T) {
		result := ParseTime("2016:06:28   :  :", "")
		assert.Equal(t, "2016-06-28 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28TNum09Num45Num49", func(t *testing.T) {
		result := ParseTime("2016/06/28T09-45:49", "")
		assert.Equal(t, "2016-06-28 09:45:49 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28TNum09Num45Num49Z", func(t *testing.T) {
		result := ParseTime("2016:06:28T09:45:49Z", "")
		assert.Equal(t, "2016-06-28 09:45:49 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28TNum09Num45Z", func(t *testing.T) {
		result := ParseTime("2016:06:28T09:45:  Z", "")
		assert.Equal(t, "2016-06-28 09:45:00 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28TNum09Num45", func(t *testing.T) {
		result := ParseTime("2016:06:28T09:45:  ", "")
		assert.Equal(t, "2016-06-28 09:45:00 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28TNum09Num45Zabc", func(t *testing.T) {
		result := ParseTime("2016:06:28T09:45:  ZABC", "")
		assert.Equal(t, "2016-06-28 09:45:00 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28TNum09Num45Abc", func(t *testing.T) {
		result := ParseTime("2016:06:28T09:45:  ABC", "")
		assert.Equal(t, "2016-06-28 09:45:00 +0000 UTC", result.String())
	})
	t.Run("Num2016Num06Num28Num09Num45Num49TenNum00Abc", func(t *testing.T) {
		result := ParseTime("2016:06:28 09:45:49+10:00ABC", "")
		assert.Equal(t, "2016-06-28 09:45:49 +1000 UTC+10:00", result.String())
	})
	t.Run("Num2016Num06Num28Num09Num45Num49Num01Num30Abc", func(t *testing.T) {
		result := ParseTime("  2016:06:28 09:45:49-01:30ABC", "")
		assert.Equal(t, "2016-06-28 09:45:49 -0130 UTC-01:30", result.String())
	})
	t.Run("Num2016Num06Num28Num09Num45Num49Num0130", func(t *testing.T) {
		result := ParseTime("2016:06:28 09:45:49-0130", "")
		assert.Equal(t, "2016-06-28 09:45:49 -0130 UTC-01:30", result.String())
	})
	t.Run("UtcNum016Num06Num28Num09Num45Num49Num0130", func(t *testing.T) {
		result := ParseTime("2016:06:28 09:45:49-0130", "UTC")
		assert.Equal(t, "2016-06-28 11:15:49 +0000 UTC", result.String())
	})
	t.Run("UtcNum016Num06Num28Num09Num45Num49Num0130", func(t *testing.T) {
		result := ParseTime("2016:06:28 09:45:49.0130", "UTC")
		assert.Equal(t, "2016-06-28 09:45:49.013 +0000 UTC", result.String())
	})
	t.Run("Num2012Num08Num08Num22Num07Eighteen", func(t *testing.T) {
		result := ParseTime("2012:08:08 22:07:18", "")
		assert.Equal(t, "2012-08-08 22:07:18 +0000 UTC", result.String())
	})
	t.Run("Num2020Num01Num30Num09Num57Eighteen", func(t *testing.T) {
		result := ParseTime("2020-01-30_09-57-18", "")
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", result.String())
	})
	t.Run("BerlinOffset", func(t *testing.T) {
		result := ParseTime("2016:06:28 09:45:49+10:00ABC", "Europe/Berlin")
		assert.Equal(t, "2016-06-28 01:45:49 +0200 CEST", result.String())
	})
	t.Run("BerlinNegativeOffset", func(t *testing.T) {
		result := ParseTime("  2016:06:28 09:45:49-01:30ABC", "Europe/Berlin")
		assert.Equal(t, "2016-06-28 13:15:49 +0200 CEST", result.String())
	})
	t.Run("BerlinNoOffset", func(t *testing.T) {
		result := ParseTime("2012:08:08 22:07:18", "Europe/Berlin")
		assert.Equal(t, "2012-08-08 22:07:18 +0200 CEST", result.String())
	})
	t.Run("BerlinUnderscore", func(t *testing.T) {
		result := ParseTime("2020-01-30_09-57-18", "Europe/Berlin")
		assert.Equal(t, "2020-01-30 09:57:18 +0100 CET", result.String())
	})
	t.Run("UtcSubSec", func(t *testing.T) {
		result := ParseTime("2020:10:17 17:48:24.9508123", "UTC")
		assert.Equal(t, "2020-10-17 17:48:24.9508123 +0000 UTC", result.UTC().String())
		assert.Equal(t, "2020-10-17 17:48:24.9508123", result.Format("2006-01-02 15:04:05.999999999"))
	})
	t.Run("UtcUnknown", func(t *testing.T) {
		assert.True(t, ParseTime("0000:00:00 00:00:00", "UTC").IsZero())
	})
	t.Run("Num2022Num09Num03TSeventeenNum48Num26Num07Num00", func(t *testing.T) {
		result := ParseTime("2022-09-03T17:48:26-07:00", "")
		assert.Equal(t, "2022-09-04 00:48:26 +0000 UTC", result.UTC().String())
		assert.Equal(t, "2022-09-03 17:48:26", result.Format("2006-01-02 15:04:05"))
	})
	t.Run("Num2016Num06Num28Num09Num45Num49UtcTwo", func(t *testing.T) {
		result := ParseTime("2016:06:28 09:45:49 +0000 UTC", "UTC+2")
		assert.Equal(t, "2016-06-28 09:45:49 +0200 UTC+2", result.String())
		assert.Equal(t, "2016-06-28 07:45:49 +0000 UTC", result.UTC().String())
	})
}
