package session

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
)

// Save updates the client session or creates a new one if needed.
func (s *Session) Save(m *entity.Session) (*entity.Session, error) {
	if m == nil {
		return nil, fmt.Errorf("session is nil")
	}

	// Save session.
	err := m.Save()

	// Return session.
	return m, err
}

// Create initializes a new client session and returns it.
func (s *Session) Create(u *entity.User, c *gin.Context, data *entity.SessionData) (m *entity.Session, err error) {
	// New session with context, user, and data.
	m = s.New(c).SetUser(u).SetData(data)

	// Create session.
	err = m.Create()

	return m, err
}
