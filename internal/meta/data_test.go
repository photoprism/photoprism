package meta

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
			Altitude:     5,
			Width:        500,
			Height:       600,
			Error:        nil,
			All:          nil,
		}

		assert.Equal(t, float32(0.8333333), data.AspectRatio())
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
			Altitude:     5,
			Width:        0,
			Height:       600,
			Error:        nil,
			All:          nil,
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
	t.Run("true", func(t *testing.T) {
		data := Data{
			DocumentID: "asdfg12345hjyt6",
		}

		assert.Equal(t, true, data.HasDocumentID())
	})

	t.Run("false", func(t *testing.T) {
		data := Data{
			DocumentID: "asdfg12345hj",
		}

		assert.Equal(t, false, data.HasDocumentID())
	})
}

func TestData_HasInstanceID(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		data := Data{
			InstanceID: "asdfg12345hjyt6",
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
