package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateFromFilePath(t *testing.T) {
	t.Run("NextcloudDateTime", func(t *testing.T) {
		result := DateFromFilePath("nextcloud/2022/04/22-04-06 15-21-03 2160.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2022-04-06 15:21:03 +0000 UTC", result.String())
	})
	t.Run("NextcloudInvalid", func(t *testing.T) {
		result := DateFromFilePath("nextcloud/2022/04/22-04-06 66-22-03 2160.jpg")
		assert.True(t, result.IsZero())
	})
	t.Run("NextcloudNotPlausible", func(t *testing.T) {
		result := DateFromFilePath("nextcloud/2022/04/88-04-06 15-21-03 2160.jpg")
		assert.True(t, result.IsZero())
	})
	t.Run("Nextcloud1990", func(t *testing.T) {
		result := DateFromFilePath("nextcloud/2022/04/90-04-06 15-21-03 2160.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "1990-04-06 15:21:03 +0000 UTC", result.String())
	})
	t.Run("Nextcloud1991", func(t *testing.T) {
		result := DateFromFilePath("nextcloud/2022/04/91-04-06 15-21-03 2160.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "1991-04-06 15:21:03 +0000 UTC", result.String())
	})
	t.Run("Nextcloud2005", func(t *testing.T) {
		result := DateFromFilePath("nextcloud/2022/04/05-04-06 15-21-03 2160.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2005-04-06 15:21:03 +0000 UTC", result.String())
	})
	t.Run("2016/08/18 iPhone/WRNI2074.jpg", func(t *testing.T) {
		result := DateFromFilePath("2016/08/18 iPhone/WRNI2074.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2016-08-18 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2016/08/18 iPhone/OZBJ8443.jpg", func(t *testing.T) {
		result := DateFromFilePath("2016/08/18 iPhone/OZBJ8443.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2016-08-18 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2018/04 - April/2018-04-12 19:24:49.gif", func(t *testing.T) {
		result := DateFromFilePath("2018/04 - April/2018-04-12 19:24:49.gif")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", result.String())
	})
	t.Run("2018", func(t *testing.T) {
		result := DateFromFilePath("2018")
		assert.True(t, result.IsZero())
	})
	t.Run("2018-04-12 19/24/49.gif", func(t *testing.T) {
		result := DateFromFilePath("2018-04-12 19/24/49.gif")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", result.String())
	})
	t.Run("/2020/1212/20130518_142022_3D657EBD.jpg", func(t *testing.T) {
		result := DateFromFilePath("/2020/1212/20130518_142022_3D657EBD.jpg")
		assert.True(t, result.IsZero(), "\"/2020/1212/20130518_142022_3D657EBD.jpg\" should not generate a valid Date. This is the filename which PhotoPrism generates when importing photos")
	})
	t.Run("20130518_142022_3D657EBD.jpg", func(t *testing.T) {
		result := DateFromFilePath("20130518_142022_3D657EBD.jpg")
		assert.True(t, result.IsZero(), "\"20130518_142022_3D657EBD.jpg\" should not generate a valid Date. This is the filename which PhotoPrism generates when importing photos")
	})
	t.Run("telegram_2020_01_30_09_57_18.jpg", func(t *testing.T) {
		result := DateFromFilePath("telegram_2020_01_30_09_57_18.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", result.String())
	})
	t.Run("Screenshot 2019_05_21 at 10.45.52.png", func(t *testing.T) {
		result := DateFromFilePath("Screenshot 2019_05_21 at 10.45.52.png")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 10:45:52 +0000 UTC", result.String())
	})
	t.Run("telegram_2020-01-30_09-57-18.jpg", func(t *testing.T) {
		result := DateFromFilePath("telegram_2020-01-30_09-57-18.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", result.String())
	})
	t.Run("Screenshot 2019-05-21 at 10.45.52.png", func(t *testing.T) {
		result := DateFromFilePath("Screenshot 2019-05-21 at 10.45.52.png")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 10:45:52 +0000 UTC", result.String())
	})
	t.Run("telegram_2020-01-30_09-18.jpg", func(t *testing.T) {
		result := DateFromFilePath("telegram_2020-01-30_09-18.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2020-01-30 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Screenshot 2019-05-21 at 10545.52.png", func(t *testing.T) {
		result := DateFromFilePath("Screenshot 2019-05-21 at 10545.52.png")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/2019-05-21/file2314.JPG", func(t *testing.T) {
		result := DateFromFilePath("/2019-05-21/file2314.JPG")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/2019.05.21", func(t *testing.T) {
		result := DateFromFilePath("/2019.05.21")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/05.21.2019", func(t *testing.T) {
		result := DateFromFilePath("/05.21.2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/21.05.2019", func(t *testing.T) {
		result := DateFromFilePath("/21.05.2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("05/21/2019", func(t *testing.T) {
		result := DateFromFilePath("05/21/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2019-07-23", func(t *testing.T) {
		result := DateFromFilePath("2019-07-23")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-07-23 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Photos/2015-01-14", func(t *testing.T) {
		result := DateFromFilePath("Photos/2015-01-14")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2015-01-14 00:00:00 +0000 UTC", result.String())
	})
	t.Run("21/05/2019", func(t *testing.T) {
		result := DateFromFilePath("21/05/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2019/05/21", func(t *testing.T) {
		result := DateFromFilePath("2019/05/21")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2019/05/2145", func(t *testing.T) {
		result := DateFromFilePath("2019/05/2145")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/05/21/2019", func(t *testing.T) {
		result := DateFromFilePath("/05/21/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/21/05/2019", func(t *testing.T) {
		result := DateFromFilePath("/21/05/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/2019/05/21.jpeg", func(t *testing.T) {
		result := DateFromFilePath("/2019/05/21.jpeg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/2019/05/21/foo.txt", func(t *testing.T) {
		result := DateFromFilePath("/2019/05/21/foo.txt")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("2019/21/05", func(t *testing.T) {
		result := DateFromFilePath("2019/21/05")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/2019/05/21/foo.jpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/05/21/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/2019/21/05/foo.jpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/21/05/foo.jpg")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/2019/5/foo.jpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/5/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/2019/1/3/foo.jpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-01-03 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/1989/1/3/foo.jpg", func(t *testing.T) {
		result := DateFromFilePath("/1989/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "1989-01-03 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/1970/1/3/foo.jpg", func(t *testing.T) {
		result := DateFromFilePath("/1970/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "1970-01-03 00:00:00 +0000 UTC", result.String())
	})
	t.Run("/1969/1/3/foo.jpg", func(t *testing.T) {
		result := DateFromFilePath("/1969/1/3/foo.jpg")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("545452019/1/3/foo.jpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-01-03 00:00:00 +0000 UTC", result.String())
	})
	t.Run("fo.jpg", func(t *testing.T) {
		result := DateFromFilePath("fo.jpg")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("n >6", func(t *testing.T) {
		result := DateFromFilePath("2020-01-30_09-87-18-23.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("year < yearmin", func(t *testing.T) {
		result := DateFromFilePath("1020-01-30_09-57-18.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("hour > hourmax", func(t *testing.T) {
		result := DateFromFilePath("2020-01-30_25-57-18.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("invalid days", func(t *testing.T) {
		result := DateFromFilePath("2020-01-00.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("IMG-20191120-WA0001.jpg", func(t *testing.T) {
		result := DateFromFilePath("IMG-20191120-WA0001.jpg")
		assert.Equal(t, "2019-11-20 00:00:00 +0000 UTC", result.String())
	})
	t.Run("VID-20191120-WA0001.jpg", func(t *testing.T) {
		result := DateFromFilePath("VID-20191120-WA0001.jpg")
		assert.Equal(t, "2019-11-20 00:00:00 +0000 UTC", result.String())
	})
}
