package api

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
)

// Session finds the client session for the given ID or returns nil otherwise.
func Session(id string) *entity.Session {
	// Skip authentication if app is running in public mode.
	if get.Config().Public() {
		return get.Session().Public()
	} else if id == "" {
		return nil
	}

	// Find session or otherwise return nil.
	s, err := get.Session().Get(id)

	if err != nil {
		return nil
	}

	return s
}
