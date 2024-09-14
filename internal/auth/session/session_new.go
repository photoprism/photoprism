package session

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
)

// New creates a session with a context if it is specified.
func (s *Session) New(c *gin.Context) (m *entity.Session) {
	return entity.NewSession(s.conf.SessionMaxAge(), s.conf.SessionTimeout()).SetContext(c)
}
