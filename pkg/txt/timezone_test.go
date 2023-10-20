package txt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeZone(t *testing.T) {
	t.Run("UTC", func(t *testing.T) {
		assert.Equal(t, time.UTC.String(), TimeZone(time.UTC.String()).String())
	})
	t.Run("UTC+2", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 11:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		timeZone := UtcOffset(local, utc, "")

		assert.Equal(t, "UTC+2", timeZone)

		loc := TimeZone(timeZone)

		assert.Equal(t, "UTC+2", loc.String())
	})
}

func TestIsUtcOffset(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, true, IsUtcOffset("UTC-2"))
		assert.Equal(t, true, IsUtcOffset("UTC"))
		assert.Equal(t, true, IsUtcOffset("UTC+1"))
		assert.Equal(t, true, IsUtcOffset("UTC+2"))
		assert.Equal(t, true, IsUtcOffset("UTC+12"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, false, IsUtcOffset("UTC-15"))
		assert.Equal(t, false, IsUtcOffset("UTC-14"))
		assert.Equal(t, false, IsUtcOffset("UTC--2"))
		assert.Equal(t, false, IsUtcOffset("UTC1"))
		assert.Equal(t, false, IsUtcOffset("UTC13"))
		assert.Equal(t, false, IsUtcOffset("UTC+13"))
	})
}

func TestNormalizeUtcOffset(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("UTC-2"))
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("UTC-02:00"))
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("-02:00"))
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("-02"))
		assert.Equal(t, "UTC-2", NormalizeUtcOffset("-2"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("UTC"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("UTC+0"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("UTC-00:00"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("UTC+00:00"))
		assert.Equal(t, "UTC", NormalizeUtcOffset("Z"))
		assert.Equal(t, "UTC+1", NormalizeUtcOffset("UTC+1"))
		assert.Equal(t, "UTC+2", NormalizeUtcOffset("UTC+2"))
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

		assert.Equal(t, "", UtcOffset(local, utc, ""))
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

		assert.Equal(t, "UTC", UtcOffset(local, utc, "00:00"))
		assert.Equal(t, "UTC", UtcOffset(local, utc, "+00:00"))
		assert.Equal(t, "UTC", UtcOffset(local, utc, "Z"))
	})
	t.Run("UTC+2", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 11:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		timeZone := UtcOffset(local, utc, "")

		assert.Equal(t, "UTC+2", timeZone)

		loc := time.FixedZone("UTC+2", 2*3600)

		assert.Equal(t, "UTC+2", loc.String())
	})
	t.Run("+02:00", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "UTC+2", UtcOffset(local, utc, "02:00"))
	})
	t.Run("UTC+2.5", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:50:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 11:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(local, utc, ""))
	})
	t.Run("+02:30", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 13:50:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 11:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(local, utc, "+02:30"))
	})
	t.Run("UTC-14", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 00:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 14:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(local, utc, ""))
	})
	t.Run("UTC-15", func(t *testing.T) {
		local, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 00:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		utc, err := time.Parse("2006-01-02 15:04:05 Z07:00", "2023-10-02 15:20:17 +00:00")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", UtcOffset(local, utc, ""))
	})
}

func TestTimeOffset(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, -2*3600, TimeOffset("UTC-2"))
		assert.Equal(t, 0, TimeOffset("UTC"))
		assert.Equal(t, 3600, TimeOffset("UTC+1"))
		assert.Equal(t, 2*3600, TimeOffset("UTC+2"))
		assert.Equal(t, 12*3600, TimeOffset("UTC+12"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, 0, TimeOffset("UTC-15"))
		assert.Equal(t, 0, TimeOffset("UTC-14"))
		assert.Equal(t, 0, TimeOffset("UTC--2"))
		assert.Equal(t, 0, TimeOffset("UTC0"))
		assert.Equal(t, 0, TimeOffset("UTC1"))
		assert.Equal(t, 0, TimeOffset("UTC13"))
		assert.Equal(t, 0, TimeOffset("UTC+13"))
	})
}
