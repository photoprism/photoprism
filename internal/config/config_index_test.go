package config

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestConfig_IndexWorkers(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.GreaterOrEqual(t, c.IndexWorkers(), 1)
}

// TestConfig_IndexWorkersOverride exercises every branch of the
// configured-value parsing — the auto sentinel, an empty string, a
// numeric override, a junk string, and a value above runtime.NumCPU() —
// and asserts the getter still respects the SQLite cap and returns at
// least one worker.
func TestConfig_IndexWorkersOverride(t *testing.T) {
	c := NewConfig(CliTestContext())
	original := c.options.IndexWorkers
	t.Cleanup(func() { c.options.IndexWorkers = original })

	cases := []struct {
		name     string
		value    string
		minCount int
	}{
		{"Auto", IndexWorkersAuto, 1},
		{"AutoMixedCase", "Auto", 1},
		{"Empty", "", 1},
		{"Numeric", "1", 1},
		{"Garbage", "not-a-number", 1},
		{"OverCPU", "9999", 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c.options.IndexWorkers = tc.value
			got := c.IndexWorkers()
			assert.GreaterOrEqual(t, got, tc.minCount)
			assert.LessOrEqual(t, got, runtime.NumCPU())
		})
	}
}

// TestConfig_IndexWorkersReason walks every branch of the auto-detection
// helper and pins the (count, reason) pair the report and startup log
// rely on. The DatabaseDriver and IndexWorkers options are mutated in
// place against a CliTestContext config and restored on cleanup; the
// global TotalMem is swapped via t.Cleanup to keep the suite isolated.
func TestConfig_IndexWorkersReason(t *testing.T) {
	c := NewConfig(CliTestContext())

	originalDriver := c.options.DatabaseDriver
	originalWorkers := c.options.IndexWorkers
	originalMem := TotalMem
	t.Cleanup(func() {
		c.options.DatabaseDriver = originalDriver
		c.options.IndexWorkers = originalWorkers
		TotalMem = originalMem
	})

	cpus := runtime.NumCPU()

	t.Run("LowMemory", func(t *testing.T) {
		TotalMem = RecommendedMem - 1
		t.Cleanup(func() { TotalMem = originalMem })

		c.options.DatabaseDriver = dsn.DriverMySQL
		c.options.IndexWorkers = "16"

		got, reason := c.indexWorkers()
		assert.Equal(t, 1, got)
		assert.Equal(t, "low-memory", reason)
	})

	t.Run("SqliteAutoLargeHost", func(t *testing.T) {
		if cpus < 8 {
			t.Skipf("requires at least 8 CPUs (have %d)", cpus)
		}
		TotalMem = originalMem
		c.options.DatabaseDriver = dsn.DriverSQLite3
		c.options.IndexWorkers = IndexWorkersAuto

		got, reason := c.indexWorkers()
		assert.Equal(t, 4, got)
		assert.Equal(t, "sqlite-cap", reason)
	})

	t.Run("SqliteOverrideAboveCap", func(t *testing.T) {
		TotalMem = originalMem
		c.options.DatabaseDriver = dsn.DriverSQLite3
		c.options.IndexWorkers = "12"

		got, reason := c.indexWorkers()
		assert.Equal(t, 4, got)
		assert.Equal(t, "sqlite-cap", reason)
	})

	t.Run("SqliteOverrideAtOrBelowCap", func(t *testing.T) {
		TotalMem = originalMem
		c.options.DatabaseDriver = dsn.DriverSQLite3
		c.options.IndexWorkers = "3"

		got, reason := c.indexWorkers()
		assert.Equal(t, 3, got)
		assert.Equal(t, "configured", reason)
	})

	t.Run("MysqlAuto", func(t *testing.T) {
		TotalMem = originalMem
		c.options.DatabaseDriver = dsn.DriverMySQL
		c.options.IndexWorkers = IndexWorkersAuto

		got, reason := c.indexWorkers()
		switch {
		case cpus >= 2:
			assert.Equal(t, cpus/2, got)
			assert.Equal(t, "auto", reason)
		default:
			assert.Equal(t, 1, got)
			assert.Equal(t, "single-cpu", reason)
		}
	})

	t.Run("MysqlConfiguredWithinBudget", func(t *testing.T) {
		TotalMem = originalMem
		c.options.DatabaseDriver = dsn.DriverMySQL
		c.options.IndexWorkers = "2"

		got, reason := c.indexWorkers()
		assert.Equal(t, 2, got)
		assert.Equal(t, "configured", reason)
	})

	t.Run("MysqlConfiguredAboveCpus", func(t *testing.T) {
		TotalMem = originalMem
		c.options.DatabaseDriver = dsn.DriverMySQL
		c.options.IndexWorkers = "9999"

		got, reason := c.indexWorkers()
		assert.Equal(t, cpus, got)
		assert.Equal(t, "configured-clamped", reason)
	})

	t.Run("ReportGetterMatchesHelper", func(t *testing.T) {
		TotalMem = originalMem
		c.options.DatabaseDriver = dsn.DriverMySQL
		c.options.IndexWorkers = IndexWorkersAuto

		want, wantReason := c.indexWorkers()
		assert.Equal(t, want, c.IndexWorkers())
		assert.Equal(t, wantReason, c.IndexWorkersReason())
	})
}

func TestParseIndexWorkers(t *testing.T) {
	cases := map[string]int{
		"":               0,
		" ":              0,
		IndexWorkersAuto: 0,
		"Auto":           0,
		"AUTO":           0,
		"0":              0,
		"-1":             -1,
		"4":              4,
		"  8 ":           8,
		"junk":           0,
	}

	for input, want := range cases {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, want, parseIndexWorkers(input))
		})
	}
}

func TestConfig_IndexSchedule(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultIndexSchedule, c.IndexSchedule())
}

func TestConfig_WakeupInterval(t *testing.T) {
	c := NewConfig(CliTestContext())
	i := c.WakeupInterval()

	assert.Equal(t, "1h34m9s", c.WakeupInterval().String())

	c.options.WakeupInterval = 45

	assert.Equal(t, "45s", c.WakeupInterval().String())

	c.options.WakeupInterval = 0

	assert.Equal(t, "15m0s", c.WakeupInterval().String())

	c.options.WakeupInterval = 150

	assert.Equal(t, "2m30s", c.WakeupInterval().String())

	c.options.WakeupInterval = i

	assert.Equal(t, "1h34m9s", c.WakeupInterval().String())
}

func TestConfig_AutoIndex(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, -1*time.Second, c.AutoIndex())
}

func TestConfig_AutoImport(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 2*time.Hour, c.AutoImport())
}

func TestConfig_OriginalsLimit(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, -1, c.OriginalsLimit())
	c.options.OriginalsLimit = 800
	assert.Equal(t, 800, c.OriginalsLimit())
}

func TestConfig_OriginalsLimitBytes(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, int64(-1), c.OriginalsLimitBytes())
	c.options.OriginalsLimit = 800
	assert.Equal(t, int64(838860800), c.OriginalsLimitBytes())
}

func TestConfig_ResolutionLimit(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, DefaultResolutionLimit, c.ResolutionLimit())
	c.options.ResolutionLimit = 800
	assert.Equal(t, 800, c.ResolutionLimit())
	c.options.ResolutionLimit = 950
	assert.Equal(t, 900, c.ResolutionLimit())
	c.options.ResolutionLimit = 0
	assert.Equal(t, DefaultResolutionLimit, c.ResolutionLimit())
	c.options.ResolutionLimit = -1
	assert.Equal(t, -1, c.ResolutionLimit())
	c.options.Sponsor = false
	assert.Equal(t, -1, c.ResolutionLimit())
	c.options.Sponsor = true
	assert.Equal(t, -1, c.ResolutionLimit())
}
