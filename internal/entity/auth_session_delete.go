package entity

import (
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/dustin/go-humanize/english"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

// DeleteSession permanently deletes the session and any sessions that were derived from it.
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

	// Release the session's preview token from the in-memory lookup cache. The reverse index
	// keeps a token resolvable until the last session that shares it has been removed, so no
	// token is dropped while still in use.
	PreviewToken.Unset(s.ID)

	if err := UnscopedDb().Delete(s).Error; err != nil {
		return err
	}

	invalidateSessionCache(s.ID)

	return nil
}

// invalidateSessionCache removes a deleted session from the cache and records its ID with a new cache version
// under the lock CacheSession takes, so a lookup that read the session before it was deleted does not cache it.
func invalidateSessionCache(id string) {
	authCacheState.Lock()
	defer authCacheState.Unlock()

	now := time.Now()
	authCacheState.version++
	authCacheState.deleted[id] = sessionTombstone{version: authCacheState.version, at: now}
	sessionCache.Delete(id)

	// Forget deleted IDs no lookup in progress can still have read.
	if now.Sub(authCacheState.pruned) > sessionTombstoneTTL {
		for k, t := range authCacheState.deleted {
			if now.Sub(t.at) > sessionTombstoneTTL {
				delete(authCacheState.deleted, k)
			}
		}

		authCacheState.pruned = now
	}
}

// DeleteChildSessions deletes sessions that authenticated via the provided parent session ID.
func DeleteChildSessions(s *Session) (deleted int) {
	if s == nil {
		return 0
	} else if !rnd.IsSessionID(s.ID) || s.GetMethod().Is(authn.MethodSession) {
		return 0
	}

	found := Sessions{}

	if err := Db().Where("auth_id = ? AND auth_method = ?", s.ID, authn.MethodSession.String()).Find(&found).Error; err != nil {
		event.AuditErr([]string{"failed to find child sessions", status.Error(err)})
		return deleted
	}

	for _, sess := range found {
		if err := sess.Delete(); err != nil {
			event.AuditErr([]string{sess.IP(), "session %s", "failed to delete child session %s", status.Error(err)}, s.RefID, sess.RefID)
		} else {
			deleted++
		}
	}

	return deleted
}

// DeleteClientSessions deletes sessions for the client beyond the configured retention limit,
// keeping the most recently created ones. Ordering is fully specified: within one creation
// second the session identified by keepID ranks first, then the session ID decides.
func DeleteClientSessions(client *Client, authMethod authn.MethodType, limit int64, keepID string) (deleted int) {
	if limit < 0 {
		return 0
	} else if client == nil {
		return 0
	}

	q := Db()

	switch {
	case client.HasUID():
		q = q.Where("client_uid = ?", client.GetUID())
	case client.HasName():
		q = q.Where("client_name = ?", client.Name())
	default:
		return 0
	}

	if client.HasUser() {
		q = q.Where("user_uid = ?", client.UserUID)
	}

	if !authMethod.IsUndefined() {
		q = q.Where("auth_method = ?", authMethod.String())
	}

	q = q.Order(clause.OrderBy{Expression: gorm.Expr("created_at DESC, CASE WHEN id = ? THEN 1 ELSE 0 END DESC, id DESC", keepID)})
	q = q.Limit(1000000000).Offset(int(limit))

	found := Sessions{}

	if err := q.Find(&found).Error; err != nil {
		event.AuditErr([]string{"failed to fetch client sessions", status.Error(err)})
		return deleted
	}

	for _, sess := range found {
		if err := sess.Delete(); err != nil {
			event.AuditErr([]string{sess.IP(), "session %s", "failed to delete", status.Error(err)}, sess.RefID)
		} else {
			deleted++
		}
	}

	return deleted
}

// DeleteExpiredSessions removes sessions whose expiration timestamp has passed.
func DeleteExpiredSessions() (deleted int) {
	found := Sessions{}

	if err := Db().Where("sess_expires > 0 AND sess_expires < ?", unix.Now()).Find(&found).Error; err != nil {
		event.AuditErr([]string{"failed to fetch expired sessions", status.Error(err)})
		return deleted
	}

	for _, sess := range found {
		if err := sess.Delete(); err != nil {
			event.AuditErr([]string{sess.IP(), "session %s", "failed to delete", status.Error(err)}, sess.RefID)
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
