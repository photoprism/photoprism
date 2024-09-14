package entity

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestUTC(t *testing.T) {
	t.Run("Zone", func(t *testing.T) {
		utc := UTC()

		if zone, offset := utc.Zone(); zone != time.UTC.String() {
			t.Error("should be UTC")
		} else if offset != 0 {
			t.Error("offset should be 0")
		}
	})
	t.Run("RunGorm", func(t *testing.T) {
		utc := UTC()
		utcGorm := gorm.NowFunc()

		t.Logf("NOW: %s, %s", utc.String(), utcGorm.String())

		assert.True(t, utcGorm.After(utc))

		if zone, offset := utcGorm.Zone(); zone != time.UTC.String() {
			t.Error("gorm time should be UTC")
		} else if offset != 0 {
			t.Error("gorm time offset should be 0")
		}

		assert.InEpsilon(t, utc.Unix(), utcGorm.Unix(), 2)
	})
}

func TestNow(t *testing.T) {
	t.Run("UTC", func(t *testing.T) {
		if Now().Location() != time.UTC {
			t.Fatal("timestamp zone must be UTC")
		}
	})
	t.Run("Past", func(t *testing.T) {
		if Now().After(time.Now().Add(time.Second)) {
			t.Fatal("timestamp should be in the past from now")
		}
	})
	t.Run("JSON", func(t *testing.T) {
		t1 := Now().Add(time.Nanosecond * 123456)

		if b, err := t1.MarshalJSON(); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("JSON: %s", b)
		}
	})
	t.Run("UnixMicro", func(t *testing.T) {
		t1 := time.Date(-3000, 1, 1, 1, 1, 1, 0, time.UTC)
		t2 := Now().Add(time.Nanosecond * 123456)
		t3 := time.Date(3000, 1, 1, 1, 1, 1, 0, time.UTC)

		ms1 := t1.UnixMilli()
		ms2 := t2.UnixMilli()
		ms3 := t3.UnixMilli()

		m1 := t1.UnixMicro()
		m2 := t2.UnixMicro()
		m3 := t3.UnixMicro()

		t.Logf("MS1: %20d", ms1)
		t.Logf("MS2: %20d", ms2)
		t.Logf("MS3: %20d", ms3)

		t.Logf("U1: %20d", m1)
		t.Logf("U2: %20d", m2)
		t.Logf("U3: %20d", m3)

		i1, i2, i3 := 1e18-m1, 1e18-m2, 1e18-m3

		t.Logf("ZZ: %20d", 9223372036854775807)
		t.Logf("I1: %20d", i1)
		t.Logf("I2: %20d", i2)
		t.Logf("I3: %20d", i3)

		t.Logf("T1: %20d", 1e18-i1)
		t.Logf("T2: %20d", 1e18-i2)
		t.Logf("T3: %20d", 1e18-i3)

		t.Logf("D1: %s", time.UnixMicro(1e18-i1).String())
		t.Logf("D2: %s", time.UnixMicro(1e18-i2).String())
		t.Logf("D3: %s", time.UnixMicro(1e18-i3).String())
	})
}

func TestTimeStamp(t *testing.T) {
	result := TimeStamp()

	if result == nil {
		t.Fatal("result must not be nil")
	}

	if result.Location() != time.UTC {
		t.Fatal("timestamp zone must be UTC")
	}

	if result.After(time.Now().Add(time.Second)) {
		t.Fatal("timestamp should be in the past from now")
	}
}

func TestTime(t *testing.T) {
	result := Time("2022-01-02T13:04:05+01:00")

	if result == nil {
		t.Fatal("result must not be nil")
	}

	assert.Equal(t, "2022-01-02T12:04:05Z", result.Format("2006-01-02T15:04:05Z07:00"))
}

func TestSeconds(t *testing.T) {
	result := Seconds(23)

	if result != 23*time.Second {
		t.Error("must be 23 seconds")
	}
}

func TestYesterday(t *testing.T) {
	now := time.Now()
	result := Yesterday()

	t.Logf("yesterday: %s", result)

	if result.After(now) {
		t.Error("yesterday is not before now")
	}

	if !result.Before(now) {
		t.Error("yesterday is before now")
	}
}
