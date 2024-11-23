package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_EstimateCountry(t *testing.T) {
	t.Run("UnitedKingdom", func(t *testing.T) {
		m := Photo{
			CameraID:     2,
			PhotoType:    MediaImage,
			PhotoName:    "20200102_194030_9EFA9E5E",
			PhotoPath:    "2000/05",
			OriginalName: "flickr import/changing-of-the-guard--buckingham-palace_7925318070_o.jpg",
		}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "gb", m.CountryCode())
		assert.Equal(t, "United Kingdom", m.CountryName())
	})
	t.Run("Berlin", func(t *testing.T) {
		m := Photo{
			CameraID:     2,
			PhotoType:    MediaImage,
			PhotoName:    "20200102_194030_ADADADAD",
			PhotoPath:    "2020/Berlin",
			OriginalName: "Zimmermannstrasse.jpg",
		}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "de", m.CountryCode())
		assert.Equal(t, "Germany", m.CountryName())
	})
	t.Run("Munich", func(t *testing.T) {
		m := Photo{
			CameraID:     2,
			PhotoType:    MediaImage,
			PhotoName:    "Brauhaus",
			PhotoPath:    "2020/Bayern",
			OriginalName: "München.jpg",
		}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "de", m.CountryCode())
		assert.Equal(t, "Germany", m.CountryName())
	})
	t.Run("Toronto", func(t *testing.T) {
		m := Photo{
			CameraID:   2,
			PhotoType:  MediaImage,
			PhotoTitle: "Port Lands / Gardiner Expressway / Toronto",
			PhotoPath:  "2012/09", PhotoName: "20120910_231851_CA06E1AD",
			OriginalName: "demo/Toronto/port-lands--gardiner-expressway--toronto_7999515645_o.jpg",
		}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "ca", m.CountryCode())
		assert.Equal(t, "Canada", m.CountryName())
	})
	t.Run("GpsCoordinates", func(t *testing.T) {
		m := Photo{
			PhotoTitle:   "Port Lands / Gardiner Expressway / Toronto",
			PhotoLat:     13.333,
			PhotoLng:     40.998,
			PhotoCountry: UnknownID,
			CellID:       "161437aab90c",
			PhotoName:    "20120910_231851_CA06E1AD",
			OriginalName: "demo/Toronto/port-lands--gardiner-expressway--toronto_7999515645_o.jpg",
		}
		m.EstimateCountry()
		assert.Equal(t, UnknownID, m.CountryCode())
		assert.Equal(t, "Unknown", m.CountryName())
	})
	t.Run("Vector", func(t *testing.T) {
		m := Photo{
			CameraID:   2,
			PhotoType:  MediaVector,
			PhotoTitle: "Logo",
		}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "zz", m.CountryCode())
		assert.Equal(t, "Unknown", m.CountryName())
	})
	t.Run("London", func(t *testing.T) {
		m := Photo{
			CameraID:         2,
			PhotoType:        MediaImage,
			PhotoDescription: "Trip to London",
			DescriptionSrc:   "meta",
		}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "gb", m.CountryCode())
		assert.Equal(t, "United Kingdom", m.CountryName())
	})
	t.Run("Los Angeles", func(t *testing.T) {
		m := Photo{
			CameraID:  2,
			PhotoType: MediaImage,
			PhotoName: "Los-Angeles-2020",
		}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "us", m.CountryCode())
		assert.Equal(t, "United States", m.CountryName())
	})
}

func TestPhoto_EstimateLocation(t *testing.T) {
	t.Run("HasLocation", func(t *testing.T) {
		p := &Place{ID: "1000000001", PlaceCountry: "mx"}
		m := Photo{
			CameraID:     2,
			TakenSrc:     SrcMeta,
			PhotoType:    MediaImage,
			PhotoName:    "PhotoWithLocation",
			OriginalName: "demo/xyz.jpg",
			Place:        p,
			PlaceID:      "1000000001",
			PlaceSrc:     SrcManual,
			PhotoCountry: "mx",
		}
		assert.True(t, m.HasPlace())
		assert.Equal(t, "mx", m.CountryCode())
		assert.Equal(t, "Mexico", m.CountryName())
		m.EstimateLocation(true)
		assert.Equal(t, "mx", m.CountryCode())
		assert.Equal(t, "Mexico", m.CountryName())
	})
	t.Run("RecentlyEstimated", func(t *testing.T) {
		m := Photo{
			CameraID:     2,
			TakenSrc:     SrcMeta,
			PhotoType:    MediaImage,
			PhotoName:    "PhotoWithoutLocation",
			OriginalName: "demo/xyy.jpg",
			EstimatedAt:  TimeStamp(),
			TakenAt:      time.Date(2016, 11, 11, 8, 7, 18, 0, time.UTC),
		}
		assert.Equal(t, UnknownID, m.CountryCode())
		m.EstimateLocation(false)
		assert.Equal(t, "zz", m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		assert.Equal(t, SrcAuto, m.PlaceSrc)
	})
	t.Run("NotRecentlyEstimated", func(t *testing.T) {
		estimateTime := Now().Add(-1 * (MetadataEstimateInterval + 2*time.Hour))
		m := Photo{
			CameraID:     2,
			TakenSrc:     SrcMeta,
			PhotoType:    MediaImage,
			PhotoName:    "PhotoWithoutLocation",
			OriginalName: "demo/xyy.jpg",
			EstimatedAt:  &estimateTime,
			TakenAt:      time.Date(2016, 11, 11, 8, 7, 18, 0, time.UTC)}
		assert.Equal(t, UnknownID, m.CountryCode())
		m.EstimateLocation(false)
		assert.Equal(t, "mx", m.CountryCode())
		assert.Equal(t, "Mexico", m.CountryName())
		assert.Equal(t, SrcEstimate, m.PlaceSrc)
	})
	t.Run("ForceEstimate", func(t *testing.T) {
		m := Photo{
			CameraID:     2,
			TakenSrc:     SrcMeta,
			PhotoType:    MediaImage,
			PhotoName:    "PhotoWithoutLocation",
			OriginalName: "demo/xyy.jpg",
			EstimatedAt:  TimeStamp(),
			TakenAt:      time.Date(2016, 11, 11, 8, 7, 18, 0, time.UTC)}
		assert.Equal(t, UnknownID, m.CountryCode())
		m.EstimateLocation(true)
		assert.Equal(t, "mx", m.CountryCode())
		assert.Equal(t, "Mexico", m.CountryName())
		assert.Equal(t, SrcEstimate, m.PlaceSrc)
	})
	t.Run("HasPlace", func(t *testing.T) {
		m := Photo{
			CameraID:     2,
			TakenSrc:     SrcMeta,
			PhotoType:    MediaImage,
			PhotoName:    "PhotoWithoutLocation",
			OriginalName: "demo/xyy.jpg",
			TakenAt:      time.Date(2016, 11, 11, 8, 7, 18, 0, time.UTC),
		}
		assert.Equal(t, UnknownID, m.CountryCode())
		m.EstimateLocation(false)
		assert.Equal(t, "mx", m.CountryCode())
		assert.Equal(t, "Mexico", m.CountryName())
		assert.Equal(t, SrcEstimate, m.PlaceSrc)
	})
	t.Run("SrcAuto", func(t *testing.T) {
		m := Photo{
			CameraID:     2,
			TakenSrc:     SrcAuto,
			PhotoType:    MediaImage,
			PhotoName:    "PhotoWithoutLocation",
			OriginalName: "demo/xyy.jpg",
			TakenAt:      time.Date(2016, 11, 11, 8, 7, 18, 0, time.UTC),
		}
		assert.Equal(t, UnknownID, m.CountryCode())
		m.EstimateLocation(false)
		assert.Equal(t, "zz", m.CountryCode())
		assert.Equal(t, "Unknown", m.CountryName())
		assert.Equal(t, "zz", m.PlaceID)
		assert.Equal(t, SrcAuto, m.PlaceSrc)
	})
	t.Run("CannotEstimate", func(t *testing.T) {
		m := Photo{
			CameraID:     2,
			TakenSrc:     SrcMeta,
			PhotoType:    MediaImage,
			PhotoName:    "PhotoWithoutLocation",
			OriginalName: "demo/xyy.jpg",
			TakenAt:      time.Date(2016, 11, 13, 8, 7, 18, 0, time.UTC),
		}
		assert.Equal(t, UnknownID, m.CountryCode())
		m.EstimateLocation(true)
		assert.Equal(t, UnknownID, m.CountryCode())
	})
	t.Run("Vector", func(t *testing.T) {
		m := Photo{
			CameraID:    2,
			TakenSrc:    SrcMeta,
			PhotoType:   MediaVector,
			EstimatedAt: TimeStamp(),
			TakenAt:     time.Date(2016, 11, 11, 8, 7, 18, 0, time.UTC)}
		assert.Equal(t, UnknownID, m.CountryCode())
		m.EstimateLocation(true)
		assert.Equal(t, "zz", m.CountryCode())
		assert.Equal(t, "Unknown", m.CountryName())
		assert.Equal(t, "", m.PlaceSrc)
	})
	t.Run("TooMuchDifference", func(t *testing.T) {
		m := Photo{
			CameraID:    2,
			TakenSrc:    SrcMeta,
			PhotoType:   MediaImage,
			EstimatedAt: TimeStamp(),
			TakenAt:     time.Date(2016, 11, 12, 22, 7, 18, 0, time.UTC)}
		assert.Equal(t, UnknownID, m.CountryCode())
		m.EstimateLocation(true)
		assert.Equal(t, "zz", m.CountryCode())
		assert.Equal(t, "Unknown", m.CountryName())
		assert.Equal(t, "", m.PlaceSrc)
	})
	/*t.Run("HasCountry", func(t *testing.T) {
		m2 := Photo{PhotoName: "PhotoWithoutLocation", OriginalName: "demo/zzz.jpg", TakenAt:  time.Date(2001, 1, 1, 7, 20, 0, 0, time.UTC)}
		assert.Equal(t, UnknownID, m2.CountryCode())
		m2.EstimateLocation()
		assert.Equal(t, "mx", m2.CountryCode())
		assert.Equal(t, "Mexico", m2.CountryName())
		assert.Equal(t, SrcEstimate, m2.PlaceSrc)
	})*/
}
