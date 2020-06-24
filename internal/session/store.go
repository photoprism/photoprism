package session

import (
	"fmt"

	gc "github.com/patrickmn/go-cache"
)

func (s *Session) Create(data Data) string {
	token := Token()
	s.cache.Set(token, &data, gc.DefaultExpiration)
	log.Debugf("session: created")

	if err := s.Save(); err != nil {
		log.Errorf("session: %s (create)", err)
	}

	return token
}

func (s *Session) Update(token string, data Data) error {
	if _, found := s.cache.Get(token); !found {
		return fmt.Errorf("session: %s not found (update)", token)
	}

	s.cache.Set(token, &data, gc.DefaultExpiration)

	log.Debugf("session: updated")

	if err := s.Save(); err != nil {
		return fmt.Errorf("session: %s (update)", err.Error())
	}

	return nil
}

func (s *Session) Delete(token string) {
	s.cache.Delete(token)
	log.Debugf("session: deleted")

	if err := s.Save(); err != nil {
		log.Errorf("session: %s (delete)", err)
	}
}

func (s *Session) Get(token string) (data *Data) {
	if hit, ok := s.cache.Get(token); ok {
		return hit.(*Data)
	}

	return nil
}

func (s *Session) Exists(token string) bool {
	_, found := s.cache.Get(token)

	return found
}
