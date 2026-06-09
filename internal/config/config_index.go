package config

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// indexWorkersLogOnce gates the per-process startup log emitted by IndexWorkers,
// so the chosen worker count and reason appear once instead of on every call.
var indexWorkersLogOnce sync.Once

// IndexWorkers returns the number of indexing workers and logs the
// chosen value once per process so support sessions can see why it was
// picked (configured override, low-memory cap, SQLite cap, or auto).
func (c *Config) IndexWorkers() int {
	num, reason := c.indexWorkers()
	indexWorkersLogOnce.Do(func() {
		log.Debugf("config: %s (%s, %s, driver %s)",
			english.Plural(num, "indexing worker", "indexing workers"),
			reason,
			english.Plural(runtime.NumCPU(), "cpu", "cpus"),
			c.DatabaseDriver())
	})
	return num
}

// IndexWorkersReason returns a short tag describing how the current
// IndexWorkers() value was derived. It mirrors the internal computation
// without re-logging and is intended for the config report and tests.
func (c *Config) IndexWorkersReason() string {
	_, reason := c.indexWorkers()
	return reason
}

// indexWorkers derives the indexing worker count from the configured option,
// runtime.NumCPU(), and the database driver. Returns the count and a reason
// tag (low-memory, sqlite-cap, configured, configured-clamped, auto, single-cpu)
// so callers can surface the rationale without re-deriving it. Uses
// runtime.NumCPU() (cgroup- and affinity-aware) rather than CPUID-based
// physical-core probes, which are unreliable under virtualization.
func (c *Config) indexWorkers() (n int, reason string) {
	// Cap to one worker on systems below the recommended memory threshold.
	if TotalMem < RecommendedMem {
		return 1, "low-memory"
	}

	cpus := runtime.NumCPU()
	configured := parseIndexWorkers(c.options.IndexWorkers)

	// SQLite serializes writes, so we cap workers to avoid lock contention.
	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		switch {
		case configured > 4:
			return 4, "sqlite-cap"
		case configured > 0:
			return configured, "configured"
		case cpus >= 8:
			return 4, "sqlite-cap"
		}
		// Small SQLite installs fall through to the auto path below.
	}

	// Honor an explicit operator override, clamped to the visible CPUs.
	if configured > cpus {
		return cpus, "configured-clamped"
	}
	if configured > 0 {
		return configured, "configured"
	}

	// Half the visible CPUs leaves headroom for OS, DB, HTTP, and other workers.
	if half := cpus / 2; half >= 1 {
		return half, "auto"
	}

	return 1, "single-cpu"
}

// parseIndexWorkers normalizes the index-workers option to an int; empty,
// "auto", and unparseable values return 0 to fall through to auto-derivation.
func parseIndexWorkers(value string) int {
	value = strings.TrimSpace(value)

	if value == "" || strings.EqualFold(value, IndexWorkersAuto) {
		return 0
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return n
}

// IndexSchedule returns the indexing schedule in cron format, e.g. "0 */3 * * *" to start indexing every 3 hours.
func (c *Config) IndexSchedule() string {
	return Schedule(c.options.IndexSchedule)
}

// WakeupInterval returns the duration between background worker runs
// required for face recognition and index maintenance (1-86400s).
func (c *Config) WakeupInterval() time.Duration {
	if c.options.WakeupInterval <= 0 {
		if c.Unsafe() {
			// Worker can be disabled only in unsafe mode.
			return time.Duration(0)
		} else {
			// Default to 15 minutes if no interval is set.
			return DefaultWakeupInterval
		}
	}

	// Do not run more than once per minute.
	if c.options.WakeupInterval < MinWakeupInterval/time.Second {
		return MinWakeupInterval
	} else if c.options.WakeupInterval < MinWakeupInterval {
		c.options.WakeupInterval *= time.Second
	}

	// Do not run less than once per day.
	if c.options.WakeupInterval > MaxWakeupInterval {
		return MaxWakeupInterval
	}

	return c.options.WakeupInterval
}

// AutoIndex returns the auto index delay duration.
func (c *Config) AutoIndex() time.Duration {
	if c.options.AutoIndex < 0 {
		return -1 * time.Second
	} else if c.options.AutoIndex == 0 || c.options.AutoIndex > 604800 {
		return DefaultAutoIndexDelay * time.Second
	}

	return time.Duration(c.options.AutoIndex) * time.Second
}

// AutoImport returns the auto import delay duration.
func (c *Config) AutoImport() time.Duration {
	if c.options.AutoImport < 0 || c.ReadOnly() {
		return -1 * time.Second
	} else if c.options.AutoImport == 0 || c.options.AutoImport > 604800 {
		return DefaultAutoImportDelay * time.Second
	}

	return time.Duration(c.options.AutoImport) * time.Second
}

// OriginalsLimit returns the maximum size of originals in MB.
func (c *Config) OriginalsLimit() int {
	if c.options.OriginalsLimit <= 0 || c.options.OriginalsLimit > 100000 {
		return -1
	}

	return c.options.OriginalsLimit
}

// OriginalsLimitBytes returns the maximum size of originals in bytes.
func (c *Config) OriginalsLimitBytes() int64 {
	if result := c.OriginalsLimit(); result <= 0 {
		return -1
	} else {
		return int64(result) * 1024 * 1024
	}
}

// ResolutionLimit returns the maximum resolution of originals in megapixels (width x height).
func (c *Config) ResolutionLimit() int {
	result := c.options.ResolutionLimit

	// Disabling or increasing the limit is at your own risk.
	// Only sponsors receive support in case of problems.
	switch {
	case result == 0:
		return DefaultResolutionLimit
	case result < 0:
		return -1
	case result > 900:
		result = 900
	}

	return result
}
