package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	gc "github.com/patrickmn/go-cache"
)

func (s *Session) Create(data interface{}) string {
	token := Token()
	s.cache.Set(token, data, gc.DefaultExpiration)
	log.Debugf("session: created")

	if err := s.Save(); err != nil {
		log.Errorf("session: %s (create)", err)
	}

	return token
}

func (s *Session) Update(token string, data interface{}) error {
	if _, found := s.cache.Get(token); !found {
		return fmt.Errorf("session: %s not found (update)", token)
	}

	s.cache.Set(token, data, gc.DefaultExpiration)

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

func (s *Session) Get(token string) (data interface{}, exists bool) {
	return s.cache.Get(token)
}

func (s *Session) Exists(token string) bool {
	_, found := s.cache.Get(token)

	return found
}

func (s *Session) Save() error {
	if s.cacheFile == "" {
		return nil
	} else if serialized, err := json.MarshalIndent(s.cache.Items(), "", " "); err != nil {
		return err
	} else if err = ioutil.WriteFile(s.cacheFile, serialized, 0600); err != nil {
		return err
	}

	return nil
}
