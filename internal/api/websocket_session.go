package api

import (
	"errors"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// wsSessionActive reports whether the session a connection authenticated with still exists and has not
// expired. Connections without a session, connections in public mode, and connections that have meanwhile
// authenticated with another session count as active.
func wsSessionActive(connId string) bool {
	sid := wsSessionID(connId)

	if sid == "" || get.Config().Public() {
		return true
	}

	active, err := entity.SessionActive(sid)

	if err != nil && !errors.Is(err, entity.ErrSessionIdInvalid) {
		log.Debugf("websocket: %s", clean.Error(err))
		return true
	}

	return active || wsSessionID(connId) != sid
}

// wsSessionID returns the ID of the session a connection authenticated with, if any.
func wsSessionID(connId string) string {
	wsAuth.mutex.RLock()
	defer wsAuth.mutex.RUnlock()

	return wsAuth.sid[connId]
}
