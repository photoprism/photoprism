package config

import (
	"math"
	"time"

	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/duf"
)

var usageCache = gc.New(5*time.Minute, 5*time.Minute)

// FlushUsageCache resets the usage information cache.
func FlushUsageCache() {
	usageCache.Flush()
}

// Usage represents storage usage information.
// Percent fields remain for UI/backward compatibility, but metrics should
// derive their values from the ratio helpers to align with Prometheus
// conventions and avoid rounding artifacts.
type Usage struct {
	// File storage usage and quota (total).
	FilesUsed    uint64 `json:"filesUsed"`
	FilesUsedPct int    `json:"filesUsedPct"`
	FilesFree    uint64 `json:"filesFree"`
	FilesFreePct int    `json:"filesFreePct"`
	FilesTotal   uint64 `json:"filesTotal"`

	// UsersQuota is the configured account quota; kept internal because the
	// public client config should not expose capacity limits directly.
	UsersQuota int `json:"-"`
	// UsersActive and GuestsActive report how many registered accounts are
	// currently enabled; kept internal so metrics can use them without leaking
	// into client-visible JSON.
	UsersActive  int `json:"-"`
	GuestsActive int `json:"-"`
	UsersUsedPct int `json:"usersUsedPct"`
	UsersFreePct int `json:"usersFreePct"`
}

// FilesUsedRatio calculates the file storage usage ratio.
func (info *Usage) FilesUsedRatio() float64 {
	if info.FilesUsed == 0 {
		return 0
	}

	if info.FilesTotal == 0 {
		// Return a tiny non-zero value to avoid emitting NaN in metrics when
		// totals are unknown (e.g. quota disabled or filesystem probe failed).
		return 0.01
	}

	return float64(info.FilesUsed) / float64(info.FilesTotal)
}

// UsersUsedRatio calculates the user account usage ratio.
func (info *Usage) UsersUsedRatio() float64 {
	if info.UsersActive == 0 || info.UsersQuota == 0 {
		return 0
	}

	return float64(info.UsersActive) / float64(info.UsersQuota)
}

// Usage returns the used, free and total storage size in bytes and caches the result.
func (c *Config) Usage() Usage {
	// Return nil if feature is not enabled.
	if !c.UsageInfo() {
		return Usage{}
	}

	originalsPath := c.OriginalsPath()

	if cached, hit := usageCache.Get(originalsPath); hit && cached != nil {
		return cached.(Usage)
	}

	info := Usage{}

	if err := c.Db().Unscoped().
		Table("files").
		Select("SUM(file_size) AS files_used").
		Where("deleted_at IS NULL").
		Take(&info).Error; err != nil {
		log.Warnf("config: failed to calculate indexed file usage (%s)", err.Error())
	}

	quotaTotal := c.FilesQuotaBytes()

	if m, err := duf.PathInfo(originalsPath); err == nil {
		info.FilesFree = m.Free
		info.FilesTotal = info.FilesUsed + m.Free
	} else {
		log.Debugf("config: failed to detect filesystem usage (%s)", err.Error())
	}

	if quotaTotal > 0 && quotaTotal < info.FilesTotal {
		info.FilesTotal = quotaTotal
	}

	if info.FilesTotal > 0 {
		info.FilesUsedPct = int(math.RoundToEven(info.FilesUsedRatio() * 100))
	}

	if info.FilesUsed > 0 && info.FilesUsedPct <= 0 {
		info.FilesUsedPct = 1
	}

	info.FilesFreePct = max(100-info.FilesUsedPct, 0)

	info.UsersActive = query.CountUsers(true, true, nil, []string{"guest"})
	info.GuestsActive = query.CountUsers(true, true, []string{"guest"}, nil)

	if info.UsersQuota = c.UsersQuota(); info.UsersQuota > 0 {
		info.UsersUsedPct = int(math.Floor(info.UsersUsedRatio() * 100))
		info.UsersFreePct = max(100-info.UsersUsedPct, 0)
	} else {
		info.UsersUsedPct = 0
		info.UsersFreePct = 0
	}

	usageCache.SetDefault(originalsPath, info)

	return info
}

// UsageInfo returns true if resource usage information should be displayed in the user interface.
func (c *Config) UsageInfo() bool {
	return c.options.UsageInfo || c.options.FilesQuota > 0 || c.options.UsersQuota > 0
}

// FilesQuota returns the maximum aggregated size of all indexed files in gigabytes, or 0 if no limit exists.
func (c *Config) FilesQuota() uint64 {
	if c.options.FilesQuota <= 0 {
		return 0
	}

	return c.options.FilesQuota
}

// FilesQuotaBytes returns the maximum aggregated size of all indexed files in bytes, or 0 if unlimited.
func (c *Config) FilesQuotaBytes() uint64 {
	if c.options.FilesQuota <= 0 {
		return 0
	}

	return c.options.FilesQuota * fs.GB
}

// FilesQuotaReached checks whether the filesystem usage has been reached or exceeded.
func (c *Config) FilesQuotaReached() bool {
	return c.FilesQuotaExceeded(99)
}

// FilesQuotaExceeded checks if the filesystem quota specified in percent has been exceeded.
func (c *Config) FilesQuotaExceeded(usedPct int) bool {
	if c.options.FilesQuota <= 0 {
		return false
	}

	return c.Usage().FilesUsedPct > usedPct
}

// UsersQuota returns the maximum number of user accounts without guests, or 0 if unlimited.
func (c *Config) UsersQuota() int {
	if c.options.UsersQuota <= 0 {
		return 0
	}

	return c.options.UsersQuota
}

// UsersQuotaReached checks whether the maximum number of user accounts has been reached or exceeded.
func (c *Config) UsersQuotaReached(role acl.Role) bool {
	return c.UsersQuotaExceeded(99, role)
}

// UsersQuotaExceeded checks whether the number of user accounts specified in percent has been exceeded.
func (c *Config) UsersQuotaExceeded(usedPct int, role acl.Role) bool {
	if c.options.UsersQuota <= 0 {
		return false
	}

	switch role {
	case acl.RoleNone, acl.RoleGuest, acl.RoleVisitor, acl.RoleClient:
		return false
	default:
		return c.Usage().UsersUsedPct > usedPct
	}
}
