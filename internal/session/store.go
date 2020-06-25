package session

import (
	"fmt"

	gc "github.com/patrickmn/go-cache"
)

func (s *Session) Create(data Data) string {
	id := NewID()
	s.cache.Set(id, data, gc.DefaultExpiration)
	log.Debugf("session: created")

	if err := s.Save(); err != nil {
		log.Errorf("session: %s (create)", err)
	}

	return id
}

func (s *Session) Update(id string, data Data) error {
	if id == "" {
		return fmt.Errorf("session: empty id")
	}

	if _, found := s.cache.Get(id); !found {
		return fmt.Errorf("session: %s not found (update)", id)
	}

	s.cache.Set(id, data, gc.DefaultExpiration)

	log.Debugf("session: updated")

	if err := s.Save(); err != nil {
		return fmt.Errorf("session: %s (update)", err.Error())
	}

	return nil
}

func (s *Session) Delete(id string) {
	s.cache.Delete(id)
	log.Debugf("session: deleted")

	if err := s.Save(); err != nil {
		log.Errorf("session: %s (delete)", err)
	}
}

func (s *Session) Get(id string) Data {
	if id == "" {
		return Data{}
	}

	if hit, ok := s.cache.Get(id); ok {
		return hit.(Data)
	}

	return Data{}
}

func (s *Session) Exists(id string) bool {
	_, found := s.cache.Get(id)

	return found
}
