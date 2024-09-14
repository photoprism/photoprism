package api

import (
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// WebSocket registers the /ws endpoint for establishing websocket connections.
func WebSocket(router *gin.RouterGroup) {
	if router == nil {
		return
	}

	conf := get.Config()

	if conf == nil {
		return
	}

	router.GET("/ws", func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		ws, err := wsConnection.Upgrade(w, r, nil)

		if err != nil {
			return
		}

		var writeMutex sync.Mutex

		defer ws.Close()

		connId := rnd.UUID()

		// Init connection.
		wsAuth.mutex.Lock()

		if conf.Public() {
			wsAuth.user[connId] = entity.Admin
		} else {
			wsAuth.user[connId] = entity.UnknownUser
		}

		wsAuth.mutex.Unlock()

		// Init writer.
		go wsWriter(ws, &writeMutex, connId)

		// Init reader.
		wsReader(ws, &writeMutex, connId, conf)
	})
}
