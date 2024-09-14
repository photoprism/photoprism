package session

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// Delete removes a client session by id.
func (s *Session) Delete(id string) error {
	if id == "" {
		return nil
	}

	if m, err := entity.FindSession(id); err != nil {
		return err
	} else {
		return m.Delete()
	}
}
