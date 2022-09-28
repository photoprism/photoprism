package session

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Save updates the client session or creates a new one if needed.
func (s *Session) Save(id string, u *entity.User, c *gin.Context, data *entity.SessionData) (m *entity.Session, err error) {
	// Try to find existing session.
	if cached, err := entity.FindSession(id); err != nil {
		m = entity.NewSession(s.expiresAfter)
	} else {
		m = &cached
	}

	// Save session.
	err = m.SetContext(c).SetUser(u).SetData(data).Save()

	// Return session.
	return m, err
}

// Create initializes a new client session and returns it.
func (s *Session) Create(u *entity.User, c *gin.Context, data *entity.SessionData) (m *entity.Session, err error) {
	// Create entity.
	m = entity.NewSession(s.expiresAfter)

	// Create session.
	err = m.SetContext(c).SetUser(u).SetData(data).Create()

	// Return session.
	return m, err
}

// Update updates session data.
func (s *Session) Update(id string, u *entity.User, c *gin.Context, data *entity.SessionData) (m *entity.Session, err error) {
	// Valid session id?
	if !rnd.IsSessionID(id) {
		return m, fmt.Errorf("invalid session id")
	}

	// Fetch cached entity.
	if cached, err := entity.FindSession(id); err != nil {
		return m, err
	} else {
		m = &cached
	}

	// Update session.
	err = m.SetContext(c).SetUser(u).SetData(data).Save()

	// Return session.
	return m, err
}
