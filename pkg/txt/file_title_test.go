package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileTitle(t *testing.T) {
	t.Run("桥", func(t *testing.T) {
		assert.Equal(t, "桥", FileTitle("桥"))
	})
	t.Run("桥船", func(t *testing.T) {
		result := FileTitle("桥船")
		assert.Equal(t, "桥船", result)
	})
	t.Run("桥船猫", func(t *testing.T) {
		result := FileTitle("桥船猫")
		assert.Equal(t, "桥船猫", result)
	})
	t.Run("谢谢！", func(t *testing.T) {
		result := FileTitle("谢谢！")
		assert.Equal(t, "谢谢！", result)
	})
	t.Run("i_love_you!", func(t *testing.T) {
		assert.Equal(t, "Love You!", FileTitle("i_love_you!"))
	})
	t.Run("photoprism", func(t *testing.T) {
		assert.Equal(t, "PhotoPrism: Browse Your Life in Pictures", FileTitle("photoprism: Browse your life in pictures"))
	})
	t.Run("dash", func(t *testing.T) {
		assert.Equal(t, "Photo Lover", FileTitle("photo-lover"))
	})
	t.Run("nyc", func(t *testing.T) {
		assert.Equal(t, "BRIDGE in, or by, NYC", FileTitle("BRIDGE in, or by, nyc"))
	})
	t.Run("apple", func(t *testing.T) {
		assert.Equal(t, "Phil Unveils iPhone, iPad, iPod, 'airpods', Airpod, AirPlay, iMac or MacBook Pro and Max", FileTitle("phil unveils iphone, ipad, ipod, 'airpods', airpod, airplay, imac or macbook 11 pro and max"))
	})
	t.Run("IMG_4568", func(t *testing.T) {
		assert.Equal(t, "", FileTitle("IMG_4568"))
	})
	t.Run("mr-kitty_life.svg", func(t *testing.T) {
		assert.Equal(t, "Mr Kitty Life", FileTitle("mr-kitty_life.svg"))
	})
	t.Run("mr-kitty--life.svg", func(t *testing.T) {
		assert.Equal(t, "Mr Kitty / Life", FileTitle("mr-kitty--life.svg"))
	})
	t.Run("queen-city-yacht-club--toronto-island_7999432607_o.jpg", func(t *testing.T) {
		assert.Equal(t, "Queen City Yacht Club / Toronto Island", FileTitle("queen-city-yacht-club--toronto-island_7999432607_o.jpg"))
	})
	t.Run("tim-robbins--tiff-2012_7999233420_o.jpg", func(t *testing.T) {
		assert.Equal(t, "Tim Robbins / TIFF", FileTitle("tim-robbins--tiff-2012_7999233420_o.jpg"))
	})
	t.Run("20200102-204030-Berlin-Germany-2020-3h4.jpg", func(t *testing.T) {
		assert.Equal(t, "Berlin Germany 2020", FileTitle("20200102-204030-Berlin-Germany-2020-3h4.jpg"))
	})
	t.Run("changing-of-the-guard--buckingham-palace_7925318070_o.jpg", func(t *testing.T) {
		assert.Equal(t, "Changing of the Guard / Buckingham Palace", FileTitle("changing-of-the-guard--buckingham-palace_7925318070_o.jpg"))
	})
	/*
		Additional tests for https://github.com/photoprism/photoprism/issues/361

		-rw-r--r-- 1 root root 813009 Jun  8 23:42 えく - スカイフレア (82063926) .png
		-rw-r--r-- 1 root root 161749 Jun  6 15:48 紅シャケ＠お仕事募集中 - モスティマ (81974640) .jpg
		[root@docker Pictures]# ls -l Originals/al
		total 1276
		-rw-r--r-- 1 root root 451062 Jun 18 19:00 Cyka - swappable mag (82405706) .jpg
		-rw-r--r-- 1 root root 662922 Jun 15 21:18 dishwasher1910 - Friedrich the smol (82201574) 1ページ.jpg
		-rw-r--r-- 1 root root 185971 Jun 19 21:07 EaycddvU0AAfuUR.jpg
	*/
	t.Run("issue_361_a", func(t *testing.T) {
		assert.Equal(t, "えく スカイフレア", FileTitle("えく - スカイフレア (82063926) .png"))
	})
	t.Run("issue_361_b", func(t *testing.T) {
		assert.Equal(t, "紅シャケ お仕事募集中 モスティマ", FileTitle("紅シャケ＠お仕事募集中 - モスティマ (81974640) .jpg"))
	})
	t.Run("issue_361_c", func(t *testing.T) {
		assert.Equal(t, "Cyka Swappable Mag", FileTitle("Cyka - swappable mag (82405706) .jpg"))
	})
	t.Run("issue_361_d", func(t *testing.T) {
		assert.Equal(t, "Dishwasher1910 Friedrich the Smol", FileTitle("dishwasher1910 - Friedrich the smol (82201574) 1ページ.jpg"))
	})
	t.Run("issue_361_e", func(t *testing.T) {
		assert.Equal(t, "EaycddvU0AAfuUR", FileTitle("EaycddvU0AAfuUR.jpg"))
	})
	t.Run("Eigene Bilder 1013/2007/oldies/neumühle", func(t *testing.T) {
		// TODO: Normalize strings, see https://godoc.org/golang.org/x/text/unicode/norm
		assert.Equal(t, "Neumu", FileTitle("Eigene Bilder 1013/2007/oldies/neumühle"))
	})
	t.Run("Neumühle", func(t *testing.T) {
		assert.Equal(t, "Neumühle", FileTitle("Neumühle"))
	})
	t.Run("IQVG4929", func(t *testing.T) {
		assert.Equal(t, "IQVG4929", FileTitle("IQVG4929.jpg"))
	})
	t.Run("IMG_1234", func(t *testing.T) {
		assert.Equal(t, "", FileTitle("IMG_1234.jpg"))
	})
	t.Run("du,-ich,-er, Sie und es", func(t *testing.T) {
		assert.Equal(t, "Du, Ich, Er, Sie und Es", FileTitle("du,-ich,-er, Sie und es"))
	})
	t.Run("Title too short", func(t *testing.T) {
		assert.Equal(t, "", FileTitle("jg.jpg"))
	})
	t.Run("invalid words", func(t *testing.T) {
		assert.Equal(t, "", FileTitle("jg hg "))
	})
	t.Run("ampersand", func(t *testing.T) {
		assert.Equal(t, "Coouussinen, Du & Ich", FileTitle("coouussinen, du & ich"))
	})
	t.Run("plus", func(t *testing.T) {
		assert.Equal(t, "Foo+Bar, Du + Ich & Er", FileTitle("Foo+bar, du + ich & er +"))
	})
	t.Run("NewYears", func(t *testing.T) {
		assert.Equal(t, "Boston New Year's", FileTitle("boston new year's"))
	})
	t.Run("Screenshot", func(t *testing.T) {
		assert.Equal(t, "Screenshot 2020 05", FileTitle("Screenshot 2020-05-04 at 14:25:01.jpeg"))
	})
	t.Run("HD", func(t *testing.T) {
		assert.Equal(t, "Desktop Nebula HD Wallpapers", FileTitle("Desktop-Nebula-hd-Wallpapers.jpeg"))
	})
	t.Run("NonCommercialPics", func(t *testing.T) {
		assert.Equal(t, "Non Commercial Pics", FileTitle("Non Commercial Pics"))
	})
	t.Run("ImgNonCommercialPics", func(t *testing.T) {
		assert.Equal(t, "Non Commercial Pics", FileTitle("Img Non Commercial Pics"))
	})
	t.Run("Birthday", func(t *testing.T) {
		assert.Equal(t, "40th Birthday in Berlin", FileTitle("2024-10-23 40th Birthday in Berlin.jpg"))
	})
	t.Run("February2nd", func(t *testing.T) {
		assert.Equal(t, "February 2nd", FileTitle("2024-10-23 February 2nd.jpg"))
	})
	t.Run("Boeing737", func(t *testing.T) {
		assert.Equal(t, "Boeing", FileTitle("Boeing 737.jpg"))
	})
	t.Run("Boeing747-8F", func(t *testing.T) {
		assert.Equal(t, "Boeing 747 8F", FileTitle("Boeing 747-8F.jpg"))
	})
	t.Run("Boeing747-100SR", func(t *testing.T) {
		assert.Equal(t, "Boeing 747 100SR", FileTitle("Boeing 747-100SR.jpg"))
	})
	t.Run("Apostrophe", func(t *testing.T) {
		assert.Equal(t, "Ma'am", FileTitle("Ma'am"))
	})
	t.Run("Download", func(t *testing.T) {
		assert.Equal(t, "Tourist Attraction Berlin", FileTitle("20170812-185131-Tourist-Attraction-Berlin-2017.jpg"))
	})
	t.Run("BreakLongTitle", func(t *testing.T) {
		assert.Equal(t, "Tourist Attraction Berlin Apple Green Blue Sky Holiday Food Restaurant Germany City Vacation Friend Sun Water Drink", FileTitle("Tourist-Attraction-Berlin-Apple-Green-Blue-Sky-Holiday-Food-Restaurant-Germany-City-Vacation-Friend-Sun-Water-Drink-Fun.jpg"))
	})
}
