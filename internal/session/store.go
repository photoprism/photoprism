package session

import (
	"encoding/json"
	"io/ioutil"

	gc "github.com/patrickmn/go-cache"
)

func (s *Session) Create(data interface{}) string {
	token := Token()
	s.cache.Set(token, data, gc.DefaultExpiration)
	log.Debugf("session: created")

	if err := s.Save(); err != nil {
		log.Errorf("session: %s", err)
	}

	return token
}

func (s *Session) Delete(token string) {
	s.cache.Delete(token)
	log.Debugf("session: deleted")

	if err := s.Save(); err != nil {
		log.Errorf("session: %s", err)
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
