package workers

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"time"

	"github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// JWTKeySchedule is the cron expression for the portal signing key check.
// Key lifetimes are measured in days, so a daily check is close enough.
const JWTKeySchedule = "0 4 * * *"

// jwtManager resolves the portal signing key manager, replaced in tests.
var jwtManager = get.JWTManager

// RunJWTKeyRotation replaces the portal JWT signing key once it reaches the configured
// lifetime. The due-date check reads memory only, so a run that is not due touches no disk.
func RunJWTKeyRotation(conf *config.Config) {
	if conf == nil || !conf.Portal() {
		return
	}

	days := conf.JWTRotateDays()

	if days <= 0 {
		return
	}

	rotateJWTKeys(jwtManager(), days)
}

// rotateJWTKeys rotates the signing key held by manager once it reaches days, and
// otherwise reapplies a retirement that an earlier run did not persist.
func rotateJWTKeys(manager *jwt.Manager, days int) {
	if manager == nil {
		return
	}

	if !manager.NeedsRotation(time.Duration(days) * 24 * time.Hour) {
		// Pick up a key whose retirement did not reach disk on an earlier run.
		if n, err := manager.RetireSuperseded(); err != nil {
			event.SystemError([]string{"jwt", "retire superseded signing key", "%s"}, keyErrorText(err))
		} else if n > 0 {
			event.SystemInfo([]string{"jwt", "retired %d superseded signing keys"}, n)
		}

		return
	}

	key, err := manager.RotateKey()

	// A new key signs as soon as it exists, so an error after that point means the
	// replacement succeeded and only the retirement is outstanding.
	switch {
	case err != nil && key == nil:
		event.SystemError([]string{"jwt", "rotate signing key", "%s"}, keyErrorText(err))
	case err != nil:
		event.SystemWarn([]string{"jwt", "rotated signing key, retirement pending", "%s"}, keyErrorText(err))
	default:
		event.SystemInfo([]string{"jwt", "rotated signing key after %d days", "new key id %s"}, days, clean.Log(key.Kid))
	}
}

// keyErrorText renders a key I/O error without the file path it names, since the system
// channel is delivered to the web UI.
func keyErrorText(err error) string {
	var pathErr *iofs.PathError

	if errors.As(err, &pathErr) {
		return clean.Error(fmt.Errorf("%s: %w", pathErr.Op, pathErr.Err))
	}

	return clean.Error(err)
}
