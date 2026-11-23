package config

import (
	"os"
	"time"

	"github.com/photoprism/photoprism/internal/service/hub"
	"github.com/photoprism/photoprism/pkg/clean"
)

// hubRenewAfter is how long after the hub config file was last written the stored key is renewed again, and
// hubConnectAfter how long a start with PHOTOPRISM_CONNECT set waits before sending the token again.
var (
	hubRenewAfter   = 10 * time.Minute
	hubConnectAfter = time.Minute
)

// hubRenewalDue reports whether the hub config file was last written at least hubRenewAfter ago, has a
// modification time in the future, or cannot be read, so that frequent restarts do not renew the key each time.
func hubRenewalDue(fileName string, now time.Time) bool {
	return hubFileOlder(fileName, now, hubRenewAfter)
}

// hubFileOlder reports whether the hub config file was last written at least d ago, has a modification time in
// the future, or cannot be read.
func hubFileOlder(fileName string, now time.Time, d time.Duration) bool {
	info, err := os.Stat(fileName)

	if err != nil {
		return true
	}

	age := now.Sub(info.ModTime())

	return age < 0 || age >= d
}

// hubUpdate reports whether initHub requests api credentials after loading the stored ones: always without a key,
// since activation must precede the database setup, at a start with PHOTOPRISM_CONNECT set while not a sponsor
// once hubConnectAfter has passed, and otherwise once the renewal is due, at server start or, for other commands,
// if the stored session has expired or could not be decoded.
func (c *Config) hubUpdate(loadErr error) bool {
	if hub.Disabled() {
		return false
	}

	c.hub.Sanitize()

	switch {
	case c.hub.Key == "":
		return true
	case c.start && os.Getenv(EnvVar("CONNECT")) != "" && !c.hub.Sponsor() && hubFileOlder(c.hub.FileName, time.Now(), hubConnectAfter):
		return true
	case !hubRenewalDue(c.hub.FileName, time.Now()):
		return false
	case c.start:
		return true
	default:
		return loadErr != nil
	}
}

// touchHubFile sets the modification time of an existing hub config file to the current time of this instance
// after an automatic request, also when it or the save failed, so that the window starts from the attempt.
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
