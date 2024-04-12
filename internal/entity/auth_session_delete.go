package entity

import (
	"fmt"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/unix"
)

// DeleteSession permanently deletes a session.
func DeleteSession(s *Session) error {
	if s == nil {
		return nil
	} else if !rnd.IsSessionID(s.ID) {
		return fmt.Errorf("invalid session id")
	}

	// Delete any other sessions that were authenticated with the specified session.
	if n := DeleteChildSessions(s); n > 0 {
		event.AuditInfo([]string{s.IP(), "session %s", "deleted %s"}, s.RefID, english.Plural(n, "child session", "child sessions"))
	}

	// Delete session from cache.
	DeleteFromSessionCache(s.ID)

	if s.PreviewToken != "" {
		PreviewToken.Set(s.PreviewToken, s.ID)
	}

	if s.DownloadToken != "" {
		DownloadToken.Set(s.DownloadToken, s.ID)
	}

	return UnscopedDb().Delete(s).Error
}

// DeleteChildSessions deletes any other sessions that were authenticated with the specified session.
func DeleteChildSessions(s *Session) (deleted int) {
	if s == nil {
		return 0
	} else if !rnd.IsSessionID(s.ID) || s.Method().Is(authn.MethodSession) {
		return 0
	}

	found := Sessions{}

	if err := Db().Where("auth_id = ? AND auth_method = ?", s.ID, authn.MethodSession.String()).Find(&found).Error; err != nil {
		event.AuditErr([]string{"failed to find child sessions", "%s"}, err)
		return deleted
	}

	for _, sess := range found {
		if err := sess.Delete(); err != nil {
			event.AuditErr([]string{sess.IP(), "session %s", "failed to delete child session %s", "%s"}, s.RefID, sess.RefID, err)
		} else {
			deleted++
		}
	}

	return deleted
}

// DeleteClientSessions deletes client sessions above the specified limit.
func DeleteClientSessions(client *Client, authMethod authn.MethodType, limit int64) (deleted int) {
	if limit < 0 {
		return 0
	} else if client == nil {
		return 0
	}

	q := Db()

	if client.HasUID() {
		q = q.Where("client_uid = ?", client.UID())
	} else if client.HasName() {
		q = q.Where("client_name = ?", client.Name())
	} else {
		return 0
	}

	if client.HasUser() {
		q = q.Where("user_uid = ?", client.UserUID)
	}

	if !authMethod.IsUndefined() {
		q = q.Where("auth_method = ?", authMethod.String())
	}

	q = q.Order("created_at DESC").Limit(1000000000).Offset(limit)

	found := Sessions{}

	if err := q.Find(&found).Error; err != nil {
		event.AuditErr([]string{"failed to fetch client sessions", "%s"}, err)
		return deleted
	}

	for _, sess := range found {
		if err := sess.Delete(); err != nil {
			event.AuditErr([]string{sess.IP(), "session %s", "failed to delete", "%s"}, sess.RefID, err)
		} else {
			deleted++
		}
	}

	return deleted
}

// DeleteExpiredSessions deletes all expired sessions.
func DeleteExpiredSessions() (deleted int) {
	found := Sessions{}

	if err := Db().Where("sess_expires > 0 AND sess_expires < ?", unix.Now()).Find(&found).Error; err != nil {
		event.AuditErr([]string{"failed to fetch expired sessions", "%s"}, err)
		return deleted
	}

	for _, sess := range found {
		if err := sess.Delete(); err != nil {
			event.AuditErr([]string{sess.IP(), "session %s", "failed to delete", "%s"}, sess.RefID, err)
		} else {
			deleted++
		}
	}

	return deleted
}

// DeleteFromSessionCache deletes a session from the cache.
func DeleteFromSessionCache(id string) {
	if id == "" {
		return
	}

	sessionCache.Delete(id)
}
