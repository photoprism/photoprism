package api

import (
	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

// AuditSharedEmail adds an audit log entry if other accounts hold the email address of the account:
// a warning for a new account or an address verified there, and an info entry otherwise.
func AuditSharedEmail(c *gin.Context, s *entity.Session, m *entity.User, created bool) {
	if s == nil || m == nil {
		return
	}

	count, verified := m.ReportSharedEmail()

	if count == 0 {
		return
	}

	holders := english.Plural(count, "other account", "other accounts")

	switch {
	case verified:
		event.AuditWarn([]string{ClientIP(c), "session %s", "user", "%s", "email is also verified for %s"}, s.RefID, clean.LogQuote(m.UserName), holders)
	case created:
		event.AuditWarn([]string{ClientIP(c), "session %s", "user", "%s", "email is also assigned to %s"}, s.RefID, clean.LogQuote(m.UserName), holders)
	default:
		event.AuditInfo([]string{ClientIP(c), "session %s", "user", "%s", "email is also assigned to %s"}, s.RefID, clean.LogQuote(m.UserName), holders)
	}
}
