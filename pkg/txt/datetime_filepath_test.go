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
	t.Run("Num2016Num08EighteenIPhoneWrniNum2074Jpg", func(t *testing.T) {
		result := DateFromFilePath("2016/08/18 iPhone/WRNI2074.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2016-08-18 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2016Num08EighteenIPhoneOzbjNum8443Jpg", func(t *testing.T) {
		result := DateFromFilePath("2016/08/18 iPhone/OZBJ8443.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2016-08-18 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2018Num04AprilNum2018Num04TwelveNineteenNum24Num49Gif", func(t *testing.T) {
		result := DateFromFilePath("2018/04 - April/2018-04-12 19:24:49.gif")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", result.String())
	})
	t.Run("Num2018", func(t *testing.T) {
		result := DateFromFilePath("2018")
		assert.True(t, result.IsZero())
	})
	t.Run("Num2018Num04TwelveNineteenNum24Num49Gif", func(t *testing.T) {
		result := DateFromFilePath("2018-04-12 19/24/49.gif")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", result.String())
	})
	t.Run("Num2020Num1212Num20130518Num142022ThreeDNum657EbdJpg", func(t *testing.T) {
		result := DateFromFilePath("/2020/1212/20130518_142022_3D657EBD.jpg")
		assert.True(t, result.IsZero(), "\"/2020/1212/20130518_142022_3D657EBD.jpg\" should not generate a valid Date. This is the filename which PhotoPrism generates when importing photos")
	})
	t.Run("Num20130518Num142022ThreeDNum657EbdJpg", func(t *testing.T) {
		result := DateFromFilePath("20130518_142022_3D657EBD.jpg")
		assert.True(t, result.IsZero(), "\"20130518_142022_3D657EBD.jpg\" should not generate a valid Date. This is the filename which PhotoPrism generates when importing photos")
	})
	t.Run("CameraName", func(t *testing.T) {
		for name, want := range map[string]string{
			"IMG_20190101_120000.jpg":                     "2019-01-01 12:00:00 +0000 UTC",
			"/x/IMG_20180318_205851_239.insp":             "2018-03-18 20:58:51 +0000 UTC",
			"card/VID_20201031_094049_00_188.insv":        "2020-10-31 09:40:49 +0000 UTC",
			"x4/LRV_20240415_213145_01_035.lrv":           "2024-04-15 21:31:45 +0000 UTC",
			"VID_20201031_094049":                         "2020-10-31 09:40:49 +0000 UTC",
			"2019/07/IMG_20190101_120000.jpg":             "2019-01-01 12:00:00 +0000 UTC",
			"2020-01-30_09-57-18/IMG_20190101_120000.jpg": "2019-01-01 12:00:00 +0000 UTC",
			"/2020/1212/IMG_20130518_142022_3D657EBD.jpg": "2013-05-18 14:20:22 +0000 UTC",
			"IMG_20190101_120000~2.jpg":                   "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20190101_1200001.jpg":                    "0001-01-01 00:00:00 +0000 UTC",
			"/x/20190101_120000.jpg":                      "0001-01-01 00:00:00 +0000 UTC",
			"XIMG_20190101_120000.jpg":                    "0001-01-01 00:00:00 +0000 UTC",
			"img_20190101_120000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG-20190101_120000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20190101_120000/photo.jpg":               "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20190230_120000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20191301_120000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20190101_246000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_19600101_120000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20190101_126000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20190101_120060.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20190101_240000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_19700101_000000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_19700101_000001.jpg":                     "1970-01-01 00:00:01 +0000 UTC",
			"IMG_19800101_000000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20021208_120000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_20200229_120000.jpg":                     "2020-02-29 12:00:00 +0000 UTC",
			"IMG_20190229_120000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
			"IMG_99990101_120000.jpg":                     "0001-01-01 00:00:00 +0000 UTC",
		} {
			assert.Equal(t, want, DateFromFilePath(name).String(), name)
		}
	})
	t.Run("TelegramNum2020Num01Num30Num09Num57EighteenJpg", func(t *testing.T) {
		result := DateFromFilePath("telegram_2020_01_30_09_57_18.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", result.String())
	})
	t.Run("ScreenshotNum2019Num05Num21AtTenNum45Num52Png", func(t *testing.T) {
		result := DateFromFilePath("Screenshot 2019_05_21 at 10.45.52.png")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 10:45:52 +0000 UTC", result.String())
	})
	t.Run("TelegramNum2020Num01Num30Num09Num57EighteenJpg", func(t *testing.T) {
		result := DateFromFilePath("telegram_2020-01-30_09-57-18.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", result.String())
	})
	t.Run("ScreenshotNum2019Num05Num21AtTenNum45Num52Png", func(t *testing.T) {
		result := DateFromFilePath("Screenshot 2019-05-21 at 10.45.52.png")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 10:45:52 +0000 UTC", result.String())
	})
	t.Run("TelegramNum2020Num01Num30Num09EighteenJpg", func(t *testing.T) {
		result := DateFromFilePath("telegram_2020-01-30_09-18.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2020-01-30 00:00:00 +0000 UTC", result.String())
	})
	t.Run("ScreenshotNum2019Num05Num21AtNum10545Num52Png", func(t *testing.T) {
		result := DateFromFilePath("Screenshot 2019-05-21 at 10545.52.png")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num05Num21File2314Jpg", func(t *testing.T) {
		result := DateFromFilePath("/2019-05-21/file2314.JPG")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num05Num21", func(t *testing.T) {
		result := DateFromFilePath("/2019.05.21")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num05Num21Num2019", func(t *testing.T) {
		result := DateFromFilePath("/05.21.2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num21Num05Num2019", func(t *testing.T) {
		result := DateFromFilePath("/21.05.2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num05Num21Num2019", func(t *testing.T) {
		result := DateFromFilePath("05/21/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num07Num23", func(t *testing.T) {
		result := DateFromFilePath("2019-07-23")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-07-23 00:00:00 +0000 UTC", result.String())
	})
	t.Run("PhotosNum2015Num01Fourteen", func(t *testing.T) {
		result := DateFromFilePath("Photos/2015-01-14")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2015-01-14 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num21Num05Num2019", func(t *testing.T) {
		result := DateFromFilePath("21/05/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num05Num21", func(t *testing.T) {
		result := DateFromFilePath("2019/05/21")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num05Num2145", func(t *testing.T) {
		result := DateFromFilePath("2019/05/2145")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num05Num21Num2019", func(t *testing.T) {
		result := DateFromFilePath("/05/21/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num21Num05Num2019", func(t *testing.T) {
		result := DateFromFilePath("/21/05/2019")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num05Num21Jpeg", func(t *testing.T) {
		result := DateFromFilePath("/2019/05/21.jpeg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num05Num21FooTxt", func(t *testing.T) {
		result := DateFromFilePath("/2019/05/21/foo.txt")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num21Num05", func(t *testing.T) {
		result := DateFromFilePath("2019/21/05")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num05Num21FooJpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/05/21/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-21 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019Num21Num05FooJpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/21/05/foo.jpg")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019FiveFooJpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/5/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-05-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num2019OneThreeFooJpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-01-03 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num1989OneThreeFooJpg", func(t *testing.T) {
		result := DateFromFilePath("/1989/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "1989-01-03 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num1970OneThreeFooJpg", func(t *testing.T) {
		result := DateFromFilePath("/1970/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "1970-01-03 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num1969OneThreeFooJpg", func(t *testing.T) {
		result := DateFromFilePath("/1969/1/3/foo.jpg")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("Num545452019OneThreeFooJpg", func(t *testing.T) {
		result := DateFromFilePath("/2019/1/3/foo.jpg")
		assert.False(t, result.IsZero())
		assert.Equal(t, "2019-01-03 00:00:00 +0000 UTC", result.String())
	})
	t.Run("FoJpg", func(t *testing.T) {
		result := DateFromFilePath("fo.jpg")
		assert.True(t, result.IsZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("NGreaterThanSix", func(t *testing.T) {
		result := DateFromFilePath("2020-01-30_09-87-18-23.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("YearLessThanYearmin", func(t *testing.T) {
		result := DateFromFilePath("1020-01-30_09-57-18.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("HourGreaterThanHourmax", func(t *testing.T) {
		result := DateFromFilePath("2020-01-30_25-57-18.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("InvalidDays", func(t *testing.T) {
		result := DateFromFilePath("2020-01-00.jpg")
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", result.String())
	})
	t.Run("ImgNum20191120WaNum0001Jpg", func(t *testing.T) {
		result := DateFromFilePath("IMG-20191120-WA0001.jpg")
		assert.Equal(t, "2019-11-20 00:00:00 +0000 UTC", result.String())
	})
	t.Run("VidNum20191120WaNum0001Jpg", func(t *testing.T) {
		result := DateFromFilePath("VID-20191120-WA0001.jpg")
		assert.Equal(t, "2019-11-20 00:00:00 +0000 UTC", result.String())
	})
}
