package meta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestData_AspectRatio(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
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

	t.Run("invalid", func(t *testing.T) {
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
	t.Run("true", func(t *testing.T) {
		data := Data{
			Width:  500,
			Height: 600,
		}

		assert.Equal(t, true, data.Portrait())
	})

	t.Run("false", func(t *testing.T) {
		data := Data{
			Width:  800,
			Height: 600,
		}

		assert.Equal(t, false, data.Portrait())
	})
}

func TestData_Megapixels(t *testing.T) {
	t.Run("30 MP", func(t *testing.T) {
		data := Data{
			Width:  5000,
			Height: 6000,
		}

		assert.Equal(t, 30, data.Megapixels())
	})
}

func TestData_HasDocumentID(t *testing.T) {
	t.Run("6ba7b810-9dad-11d1-80b4-00c04fd430c8", func(t *testing.T) {
		data := Data{
			DocumentID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		}

		assert.Equal(t, true, data.HasDocumentID())
	})

	t.Run("asdfg12345hjyt6", func(t *testing.T) {
		data := Data{
			DocumentID: "asdfg12345hjyt6",
		}

		assert.Equal(t, false, data.HasDocumentID())
	})

	t.Run("asdfg12345hj", func(t *testing.T) {
		data := Data{
			DocumentID: "asdfg12345hj",
		}

		assert.Equal(t, false, data.HasDocumentID())
	})
}

func TestData_HasInstanceID(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		data := Data{
			InstanceID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		}

		assert.Equal(t, true, data.HasInstanceID())
	})

	t.Run("false", func(t *testing.T) {
		data := Data{
			InstanceID: "asdfg12345hj",
		}

		assert.Equal(t, false, data.HasInstanceID())
	})
}

func TestData_HasTimeAndPlace(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		data := Data{
			Lat:     1.334,
			Lng:     4.567,
			TakenAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.Equal(t, true, data.HasTimeAndPlace())
	})

	t.Run("false", func(t *testing.T) {
		data := Data{
			Lat:     1.334,
			Lng:     0,
			TakenAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.Equal(t, false, data.HasTimeAndPlace())
	})
	t.Run("false", func(t *testing.T) {
		data := Data{
			Lat:     0,
			Lng:     4.567,
			TakenAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.Equal(t, false, data.HasTimeAndPlace())
	})

	t.Run("false", func(t *testing.T) {
		data := Data{
			Lat: 1.334,
			Lng: 4.567,
		}

		assert.Equal(t, false, data.HasTimeAndPlace())
	})
}

func TestData_CellID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := Data{
			Lat:     1.334,
			Lng:     4.567,
			TakenAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.Equal(t, "s2:100c9acde614", data.CellID())
	})
}

func TestData_IsHDR(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		data := Data{
			ImageType: 3,
			TakenAt:   time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.True(t, data.IsHDR())
	})
	t.Run("false", func(t *testing.T) {
		data := Data{
			ImageType: 2,
			TakenAt:   time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		assert.False(t, data.IsHDR())
	})
}
