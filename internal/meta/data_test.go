package meta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestData_AspectRatio(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		data := Data{
			DocumentID:   "123",
			InstanceID:   "456",
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TimeZone:     "UTC",
			Codec:        "avc1",
			Lat:          1.334,
			Lng:          44.567,
			Altitude:     5.0,
			Width:        500,
			Height:       600,
			Error:        nil,
			exif:         nil,
		}

		assert.Equal(t, float32(0.83), data.AspectRatio())
	})
	t.Run("Invalid", func(t *testing.T) {
		data := Data{
			DocumentID:   "123",
			InstanceID:   "456",
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TimeZone:     "UTC",
			Codec:        "avc1",
			Lat:          1.334,
			Lng:          44.567,
			Altitude:     5.0,
			Width:        0,
			Height:       600,
			Error:        nil,
			exif:         nil,
		}

		assert.Equal(t, float32(0), data.AspectRatio())
	})
}

func TestData_Portrait(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		data := Data{
			Width:  500,
			Height: 600,
		}

		assert.Equal(t, true, data.Portrait())
	})
	t.Run("False", func(t *testing.T) {
		data := Data{
			Width:  800,
			Height: 600,
		}

		assert.Equal(t, false, data.Portrait())
	})
}

func TestData_Megapixels(t *testing.T) {
	t.Run("Num30Mp", func(t *testing.T) {
		data := Data{
			Width:  5000,
			Height: 6000,
		}

		assert.Equal(t, 30, data.Megapixels())
	})
}

func TestData_HasDocumentID(t *testing.T) {
	t.Run("SixBa7b810NineDadElevenD1Num80B4Num00C04fd430c8", func(t *testing.T) {
		data := Data{
			DocumentID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		}

		assert.Equal(t, true, data.HasDocumentID())
	})
	t.Run("Asdfg12345hjyt6", func(t *testing.T) {
		data := Data{
			DocumentID: "asdfg12345hjyt6",
		}

		assert.Equal(t, false, data.HasDocumentID())
	})
	t.Run("Asdfg12345hj", func(t *testing.T) {
		data := Data{
			DocumentID: "asdfg12345hj",
		}

		assert.Equal(t, false, data.HasDocumentID())
	})
}

func TestData_HasInstanceID(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		data := Data{
			InstanceID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		}

		assert.Equal(t, true, data.HasInstanceID())
	})
	t.Run("False", func(t *testing.T) {
		data := Data{
			InstanceID: "asdfg12345hj",
		}

		assert.Equal(t, false, data.HasInstanceID())
	})
}

func TestData_HasTimeAndPlace(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		data := Data{
			Lat:     1.334,
			Lng:     4.567,
			TakenAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.Equal(t, true, data.HasTimeAndPlace())
	})
	t.Run("False", func(t *testing.T) {
		data := Data{
			Lat:     1.334,
			Lng:     0,
			TakenAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.Equal(t, false, data.HasTimeAndPlace())
	})
	t.Run("False", func(t *testing.T) {
		data := Data{
			Lat:     0,
			Lng:     4.567,
			TakenAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.Equal(t, false, data.HasTimeAndPlace())
	})
	t.Run("False", func(t *testing.T) {
		data := Data{
			Lat: 1.334,
			Lng: 4.567,
		}

		assert.Equal(t, false, data.HasTimeAndPlace())
	})
}

func TestData_CellID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		data := Data{
			Lat:     1.334,
			Lng:     4.567,
			TakenAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.Equal(t, "s2:100c9acde614", data.CellID())
	})
}

func TestData_IsHDR(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		data := Data{
			ImageType: 3,
			TakenAt:   time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.True(t, data.IsHDR())
	})
	t.Run("False", func(t *testing.T) {
		data := Data{
			ImageType: 2,
			TakenAt:   time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.False(t, data.IsHDR())
	})
}

// TestData_TakenOrModified verifies that the modify time is only used, and resolved like TakenAt, without a capture time.
func TestData_TakenOrModified(t *testing.T) {
	modifiedAt := time.Date(2020, 10, 26, 15, 46, 31, 0, time.UTC)

	t.Run("TakenAt", func(t *testing.T) {
		data := Data{TakenAt: time.Date(2020, 10, 26, 13, 46, 29, 0, time.UTC), TakenAtLocal: time.Date(2020, 10, 26, 15, 46, 29, 0, time.UTC), TimeZone: "Europe/Berlin", ModifiedAt: modifiedAt}
		utc, local, zone, modified := data.TakenOrModified()
		assert.False(t, modified)
		assert.Equal(t, data.TakenAt, utc)
		assert.Equal(t, data.TakenAtLocal, local)
		assert.Equal(t, "Europe/Berlin", zone)
	})
	t.Run("ModifiedAt", func(t *testing.T) {
		data := Data{ModifiedAt: modifiedAt}
		utc, local, _, modified := data.TakenOrModified()
		assert.True(t, modified)
		assert.Equal(t, modifiedAt, utc)
		assert.Equal(t, "2020-10-26 15:46:31", local.Format(time.DateTime))
	})
	t.Run("ModifiedAtWithOffset", func(t *testing.T) {
		data := Data{ModifiedAt: time.Date(2020, 10, 26, 15, 46, 31, 0, time.FixedZone("", 2*3600)), TimeOffset: "+02:00"}
		utc, local, zone, modified := data.TakenOrModified()
		assert.True(t, modified)
		assert.Equal(t, "UTC+2", zone)
		assert.Equal(t, "2020-10-26 15:46:31", local.Format(time.DateTime))
		assert.Equal(t, "2020-10-26T13:46:31Z", utc.Format(time.RFC3339))
	})
	t.Run("ModifiedAtVideo", func(t *testing.T) {
		// QuickTime times are in UTC, so the local time follows the time zone of the position.
		data := Data{ModifiedAt: modifiedAt, MimeType: MimeVideoMp4, Lat: 52.52, Lng: 13.405}
		utc, local, zone, modified := data.TakenOrModified()
		assert.True(t, modified)
		assert.Equal(t, "Europe/Berlin", zone)
		assert.Equal(t, "2020-10-26T15:46:31Z", utc.Format(time.RFC3339))
		assert.Equal(t, "2020-10-26 16:46:31", local.Format(time.DateTime))
	})
	t.Run("ModifiedAtSubSec", func(t *testing.T) {
		// Sub-seconds belong to the capture time.
		utc, _, _, _ := Data{ModifiedAt: modifiedAt, TakenNs: 500000000}.TakenOrModified()
		assert.Equal(t, 0, utc.Nanosecond())
	})
	t.Run("ModifiedAtWithPosition", func(t *testing.T) {
		data := Data{ModifiedAt: modifiedAt, Lat: 52.52, Lng: 13.405}
		utc, local, zone, modified := data.TakenOrModified()
		assert.True(t, modified)
		assert.Equal(t, "Europe/Berlin", zone)
		assert.Equal(t, "2020-10-26 15:46:31", local.Format(time.DateTime))
		assert.Equal(t, "2020-10-26T14:46:31Z", utc.Format(time.RFC3339))

		// Stacking by time and place only uses the time the picture was taken.
		assert.False(t, data.HasTimeAndPlace())
	})
	t.Run("None", func(t *testing.T) {
		utc, _, _, modified := Data{}.TakenOrModified()
		assert.False(t, modified)
		assert.True(t, utc.IsZero())
		assert.False(t, Data{Lat: 52.52, Lng: 13.405}.HasTimeAndPlace())
	})
}
