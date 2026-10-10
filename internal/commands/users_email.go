package commands

import (
	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

// AuditSharedEmail adds an audit log entry if other accounts hold the email address of the account:
// a warning for a new account or an address verified there, and an info entry otherwise.
func AuditSharedEmail(m *entity.User, created bool) {
	if m == nil {
		return
	}

	count, verified := m.ReportSharedEmail()

	if count == 0 {
		return
	}

	holders := english.Plural(count, "other account", "other accounts")

	switch {
	case verified:
		event.AuditWarn([]string{"cli", "user %s", "email is also verified for %s"}, clean.LogQuote(m.UserName), holders)
	case created:
		event.AuditWarn([]string{"cli", "user %s", "email is also assigned to %s"}, clean.LogQuote(m.UserName), holders)
	default:
		event.AuditInfo([]string{"cli", "user %s", "email is also assigned to %s"}, clean.LogQuote(m.UserName), holders)
	}
}
