/*

Package session provides session storage and management.

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/
package session

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"time"

	gc "github.com/patrickmn/go-cache"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// Session represents a session store.
type Session struct {
	cacheFile string
	cache     *gc.Cache
}

// New returns a new session store with an optional cachePath.
func New(expiration time.Duration, cachePath string) *Session {
	s := &Session{}

	cleanupInterval := 15 * time.Minute

	if cachePath != "" {
		var items map[string]gc.Item

		s.cacheFile = path.Join(cachePath, "sessions.json")

		if cached, err := ioutil.ReadFile(s.cacheFile); err != nil {
			log.Infof("session: %s", err)
		} else if err := json.Unmarshal(cached, &items); err != nil {
			log.Errorf("session: %s", err)
		} else {
			s.cache = gc.NewFrom(expiration, cleanupInterval, items)
		}
	}

	if s.cache == nil {
		s.cache = gc.New(expiration, cleanupInterval)
	}

	return s
}
