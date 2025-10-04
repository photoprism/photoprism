package tz

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOffset(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		sec, err := Offset("UTC-2")
		assert.Equal(t, -2*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC")
		assert.Equal(t, 0, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+1")
		assert.Equal(t, 3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+2")
		assert.Equal(t, 2*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+3")
		assert.Equal(t, 3*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+4")
		assert.Equal(t, 4*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+5")
		assert.Equal(t, 5*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+6")
		assert.Equal(t, 6*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+7")
		assert.Equal(t, 7*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+8")
		assert.Equal(t, 8*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+9")
		assert.Equal(t, 9*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+10")
		assert.Equal(t, 10*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+11")
		assert.Equal(t, 11*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC+12")
		assert.Equal(t, 12*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-1")
		assert.Equal(t, -3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-2")
		assert.Equal(t, -2*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-3")
		assert.Equal(t, -3*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-4")
		assert.Equal(t, -4*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-5")
		assert.Equal(t, -5*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-6")
		assert.Equal(t, -6*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-7")
		assert.Equal(t, -7*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-8")
		assert.Equal(t, -8*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-9")
		assert.Equal(t, -9*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-10")
		assert.Equal(t, -10*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-11")
		assert.Equal(t, -11*3600, sec)
		assert.NoError(t, err)

		sec, err = Offset("UTC-12")
		assert.Equal(t, -12*3600, sec)
		assert.NoError(t, err)
	})
	t.Run("Invalid", func(t *testing.T) {
		sec, err := Offset("UTC-15")
		assert.Equal(t, 0, sec)
		assert.Error(t, err)

		sec, err = Offset("UTC--2")
		assert.Equal(t, 0, sec)
		assert.Error(t, err)

		sec, err = Offset("UTC0")
		assert.Equal(t, 0, sec)
		assert.Error(t, err)

		sec, err = Offset("UTC1")
		assert.Equal(t, 0, sec)
		assert.Error(t, err)

		sec, err = Offset("UTC13")
		assert.Equal(t, 0, sec)
		assert.Error(t, err)

		sec, err = Offset("UTC+13")
		assert.Equal(t, 0, sec)
		assert.Error(t, err)
	})
}

func TestNormalizeUtcOffset(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, "UTC-12", NormalizeUtcOffset("UTC-12"))
		assert.Equal(t, "UTC-11", NormalizeUtcOffset("-11"))
		assert.Equal(t, "UTC-10", NormalizeUtcOffset("-10:00"))
		assert.Equal(t, "UTC-9", NormalizeUtcOffset("UTC-09:00"))
		assert.Equal(t, "UTC-8", NormalizeUtcOffset("GMT-8"))
		assert.Equal(t, "UTC-7", NormalizeUtcOffset("-07"))
		assert.Equal(t, "UTC-6", NormalizeUtcOffset("-06:00"))
		assert.Equal(t, "UTC-5", NormalizeUtcOffset("UTC-05:00"))
		assert.Equal(t, "UTC-4", NormalizeUtcOffset("UTC-4"))
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("UTC-2"))
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("UTC-02:00"))
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("-02:00"))
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("-02"))
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("-2"))
		assert.Equal(t, "UTC-1", NormalizeUtcOffset("-1"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("UTC"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("UTC+0"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("UTC-00:00"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("UTC+00:00"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("Z"))
		assert.Equal(t, "UTC+1", NormalizeUtcOffset("UTC+1"))
		assert.Equal(t, "UTC+2", NormalizeUtcOffset("UTC+2"))
		assert.Equal(t, "UTC+3", NormalizeUtcOffset("+3"))
		assert.Equal(t, "UTC+4", NormalizeUtcOffset("+04"))
		assert.Equal(t, "UTC+5", NormalizeUtcOffset("GMT+5"))
		assert.Equal(t, "UTC+6", NormalizeUtcOffset("GMT+6"))
		assert.Equal(t, "UTC+7", NormalizeUtcOffset("Etc/GMT+07"))
		assert.Equal(t, "UTC+8", NormalizeUtcOffset("Etc/GMT+08:00"))
		assert.Equal(t, "UTC+9", NormalizeUtcOffset("+09:00"))
		assert.Equal(t, "UTC+10", NormalizeUtcOffset("UTC+10:00"))
		assert.Equal(t, "UTC+11", NormalizeUtcOffset("GMT+11"))
		assert.Equal(t, "UTC+12", NormalizeUtcOffset("UTC+12"))
		assert.Equal(t, "UTC+12", NormalizeUtcOffset("+12"))
		assert.Equal(t, "UTC+12", NormalizeUtcOffset("+12:00"))
		assert.Equal(t, "UTC+12", NormalizeUtcOffset("12:00"))
		assert.Equal(t, "UTC+12", NormalizeUtcOffset("UTC+12:00"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "", NormalizeUtcOffset("UTC-15"))
		assert.Equal(t, "", NormalizeUtcOffset("UTC-14:00"))
		assert.Equal(t, "", NormalizeUtcOffset("UTC-14"))
		assert.Equal(t, "", NormalizeUtcOffset("UTC--2"))
		assert.Equal(t, "", NormalizeUtcOffset("UTC1"))
		assert.Equal(t, "", NormalizeUtcOffset("UTC13"))
		assert.Equal(t, "", NormalizeUtcOffset("UTC+13"))
	})
}

func TestUtcOffset(t *testing.T) {
	t.Run("GMT", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(utc, local, ""))
	})
	t.Run("UTC", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(utc, local, "00:00"))
		assert.Equal(t, "", UtcOffset(utc, local, "+00:00"))
		assert.Equal(t, "UTC", UtcOffset(utc, local, "Z"))
	})
	t.Run("UtcTwo", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 11:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		timeZone := UtcOffset(utc, local, "")

		assert.Equal(t, "UTC+2", timeZone)

		loc := time.FixedZone("UTC+2", 2*3600)

		assert.Equal(t, "UTC+2", loc.String())
	})
	t.Run("Num02Num00", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "UTC+2", UtcOffset(utc, local, "02:00"))
	})
	t.Run("UtcTwoFive", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:50:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 11:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(utc, local, ""))
	})
	t.Run("Num02Num30", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:50:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 11:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(utc, local, "+02:30"))
	})
	t.Run("UtcFourteen", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 00:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 14:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(utc, local, ""))
	})
	t.Run("UtcFifteen", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 00:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 15:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(utc, local, ""))
	})
	t.Run("UtcNum02Num00", func(t *testing.T) {
		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:50:17 +02:00")

		if err != nil {
			t.Fatal(err)
		}

		result := UtcOffset(utc, time.Time{}, "")

		assert.Equal(t, "UTC+2", result)
	})

}
