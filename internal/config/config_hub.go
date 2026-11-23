package config

import (
	"os"
	"time"

	"github.com/sevlyar/go-daemon"

	"github.com/photoprism/photoprism/internal/service/hub"
	"github.com/photoprism/photoprism/pkg/clean"
)

// hubRenewAfter is how long after the hub config file was last written commands other than start renew an
// expired or undecodable session again.
var hubRenewAfter = 10 * time.Minute

// hubWasReborn reports whether this process is the daemon that a start with detach-server spawned; NewConfig
// records it, as the start command clears the mark before the server runs.
var hubWasReborn = daemon.WasReborn

// detachingServer reports whether this is a start with detach-server in a process other than the daemon, which
// then initializes and renews on its own.
func (c *Config) detachingServer() bool {
	return c.start && c.DetachServer() && !c.reborn
}

// logMembershipHint logs how to become a member unless this instance is a sponsor. A detach-server start outside
// the daemon skips it, since it did not renew and the daemon logs the hint based on the current membership.
func (c *Config) logMembershipHint() {
	if c.Sponsor() || c.detachingServer() {
		return
	}

	log.Info(MsgSponsor)
	log.Info(MsgSignUp)
}

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

// hubUpdate reports whether initHub requests api credentials after loading the stored ones: without a key, as the
// editions need activation before the database setup, and at every start, so membership changes apply on restart,
// except in a detach-server start outside the daemon, which leaves both to the daemon; other commands use the
// stored key and renew only an expired or undecodable session once due.
func (c *Config) hubUpdate(loadErr error) bool {
	if hub.Disabled() || c.detachingServer() {
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
