package config

import (
	"os"
	"time"

	"github.com/photoprism/photoprism/internal/service/hub"
	"github.com/photoprism/photoprism/pkg/clean"
)

// hubRenewAfter is how long after the hub config file was last written commands other than start renew an
// expired or undecodable session again.
var hubRenewAfter = 10 * time.Minute

// hubRenewalDue reports whether the hub config file was last written at least hubRenewAfter ago, has a
// modification time in the future, or cannot be read.
func hubRenewalDue(fileName string, now time.Time) bool {
	info, err := os.Stat(fileName)

	if err != nil {
		return true
	}

	age := now.Sub(info.ModTime())

	return age < 0 || age >= hubRenewAfter
}

// hubUpdate reports whether initHub requests api credentials after loading the stored ones: always without a key,
// since activation must precede the database setup, and at every start command, so that membership changes apply
// on restart; other commands use the stored key and renew only an expired or undecodable session once due.
func (c *Config) hubUpdate(loadErr error) bool {
	if hub.Disabled() {
		return false
	}

	c.hub.Sanitize()

	switch {
	case c.hub.Key == "", c.start:
		return true
	case loadErr == nil:
		return false
	default:
		return hubRenewalDue(c.hub.FileName, time.Now())
	}
}

// touchHubFile sets the modification time of an existing hub config file to the current time of this instance
// after an automatic request, also when it or the save failed, so that the window of other commands starts from it.
// The file contents remain unchanged.
func touchHubFile(fileName string) {
	if _, err := os.Stat(fileName); err != nil {
		return
	}

	now := time.Now()

	if err := os.Chtimes(fileName, now, now); err != nil {
		log.Debugf("config: failed to update the modification time of %s (%s)", clean.Log(fileName), clean.Error(err))
	}
}
