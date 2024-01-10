package session

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
)

// Get returns an existing client session.
func (s *Session) Get(id string) (m *entity.Session, err error) {
	if id == "" {
		return &entity.Session{}, fmt.Errorf("invalid session id")
	}

	return entity.FindSession(id)
}

// Exists checks whether a client session with the specified ID exists.
func (s *Session) Exists(id string) bool {
	_, err := s.Get(id)

	return err == nil
}
