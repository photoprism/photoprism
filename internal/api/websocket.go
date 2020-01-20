package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
)

var wsConnection = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var wsTimeout = 60 * time.Second

func wsReader(ws *websocket.Conn) {
	defer ws.Close()

	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(wsTimeout))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(wsTimeout)); return nil })

	for {
		_, m, err := ws.ReadMessage()
		if err != nil {
			break
		}
		log.Debugf("websocket: received %d bytes", len(m))
	}
}

func wsWriter(ws *websocket.Conn) {
	pingTicker := time.NewTicker(10 * time.Second)
	s := event.Subscribe("log.*", "notify.*", "index.*", "upload.*", "import.*", "config.*", "count.*")

	defer func() {
		pingTicker.Stop()
		event.Unsubscribe(s)
		ws.Close()
	}()

	for {
		select {
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case msg := <-s.Receiver:
			ws.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if err := ws.WriteJSON(gin.H{"event": msg.Name, "data": msg.Fields}); err != nil {
				log.Debug(err)
				return
			}
		}
	}
}

// GET /api/v1/ws
func Websocket(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/ws", func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		ws, err := wsConnection.Upgrade(w, r, nil)
		if err != nil {
			log.Error(err)
			return
		}

		defer ws.Close()

		log.Debug("websocket: connected")

		go wsWriter(ws)

		wsReader(ws)
	})
}
