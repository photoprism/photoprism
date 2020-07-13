package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	t.Run("2018/04 - April/2018-04-12 19:24:49.gif", func(t *testing.T) {
		result := Time("2018/04 - April/2018-04-12 19:24:49.gif")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", result.String())
	})

	t.Run("2018-04-12 19/24/49.gif", func(t *testing.T) {
		result := Time("2018-04-12 19/24/49.gif")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", result.String())
	})

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

	t.Run("2019-07-23", func(t *testing.T) {
		result := Time("2019-07-23")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-07-23 00:00:00 +0000 UTC", result.String())
	})

	t.Run("Photos/2015-01-14", func(t *testing.T) {
		result := Time("Photos/2015-01-14")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2015-01-14 00:00:00 +0000 UTC", result.String())
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

	t.Run("fo.jpg", func(t *testing.T) {
		result := Time("fo.jpg")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("n >6", func(t *testing.T) {
		result := Time("2020-01-30_09-87-18-23.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})

	t.Run("year < yearmin", func(t *testing.T) {
		result := Time("1020-01-30_09-57-18.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("hour > hourmax", func(t *testing.T) {
		result := Time("2020-01-30_25-57-18.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("invalid days", func(t *testing.T) {
		result := Time("2020-01-00.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
}

func TestIsTime(t *testing.T) {
	t.Run("/2020/1212/20130518_142022_3D657EBD.jpg", func(t *testing.T) {
		assert.False(t, IsTime("/2020/1212/20130518_142022_3D657EBD.jpg"))
	})

	t.Run("telegram_2020_01_30_09_57_18.jpg", func(t *testing.T) {
		assert.False(t, IsTime("telegram_2020_01_30_09_57_18.jpg"))
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

func TestCountryCode(t *testing.T) {
	t.Run("London", func(t *testing.T) {
		result := CountryCode("London")
		assert.Equal(t, "gb", result)
	})

	t.Run("reunion island", func(t *testing.T) {
		result := CountryCode("Reunion-Island-2019")
		assert.Equal(t, "zz", result)
	})

	t.Run("reunion island france", func(t *testing.T) {
		result := CountryCode("Reunion-Island-france-2019")
		assert.Equal(t, "fr", result)
	})

	t.Run("réunion", func(t *testing.T) {
		result := CountryCode("My-RéunioN-2019")
		assert.Equal(t, "fr", result)
	})

	t.Run("NYC", func(t *testing.T) {
		result := CountryCode("NYC 2019")
		assert.Equal(t, "us", result)
	})

	t.Run("Scuba", func(t *testing.T) {
		result := CountryCode("Scuba 2019")
		assert.Equal(t, "zz", result)
	})

	t.Run("Cuba", func(t *testing.T) {
		result := CountryCode("Cuba 2019")
		assert.Equal(t, "cu", result)
	})

	t.Run("San Francisco", func(t *testing.T) {
		result := CountryCode("San Francisco 2019")
		assert.Equal(t, "us", result)
	})

	t.Run("Los Angeles", func(t *testing.T) {
		result := CountryCode("I was in Los Angeles")
		assert.Equal(t, "us", result)
	})

	t.Run("St Gallen", func(t *testing.T) {
		result := CountryCode("St.----Gallen")
		assert.Equal(t, "ch", result)
	})

	t.Run("Congo Brazzaville", func(t *testing.T) {
		result := CountryCode("Congo Brazzaville")
		assert.Equal(t, "cg", result)
	})

	t.Run("Congo", func(t *testing.T) {
		result := CountryCode("Congo")
		assert.Equal(t, "cd", result)
	})

	t.Run("U.S.A.", func(t *testing.T) {
		result := CountryCode("Born in the U.S.A. is a song written and performed by Bruce Springsteen...")
		assert.Equal(t, "zz", result)
	})

	t.Run("US", func(t *testing.T) {
		result := CountryCode("Somebody help us please!")
		assert.Equal(t, "zz", result)
	})

	t.Run("Never mind Nirvana", func(t *testing.T) {
		result := CountryCode("Never mind Nirvana.")
		assert.Equal(t, "zz", result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := CountryCode("")
		assert.Equal(t, "zz", result)
	})

	t.Run("zz", func(t *testing.T) {
		result := CountryCode("zz")
		assert.Equal(t, "zz", result)
	})
}

func TestYear(t *testing.T) {
	t.Run("London 2002", func(t *testing.T) {
		result := Year("/2002/London 81/")
		assert.Equal(t, 2002, result)
	})

	t.Run("San Francisco 2019", func(t *testing.T) {
		result := Year("San Francisco 2019")
		assert.Equal(t, 2019, result)
	})

	t.Run("string with no number", func(t *testing.T) {
		result := Year("Born in the U.S.A. is a song written and performed by Bruce Springsteen...")
		assert.Equal(t, 0, result)
	})

	t.Run("file name", func(t *testing.T) {
		result := Year("/share/photos/243546/2003/01/myfile.jpg")
		assert.Equal(t, 2003, result)
	})

	t.Run("path", func(t *testing.T) {
		result := Year("/root/1981/London 2005")
		assert.Equal(t, 2005, result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := Year("")
		assert.Equal(t, 0, result)
	})
}

func TestIsUInt(t *testing.T) {
	assert.False(t, IsUInt(""))
	assert.False(t, IsUInt("12 3"))
	assert.True(t, IsUInt("123"))
}
