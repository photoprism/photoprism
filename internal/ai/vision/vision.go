/*
Package vision provides a computer vision request handler and a client for using external APIs.

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package vision

import (
	"os"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

var log = event.Log

var ensureEnvOnce sync.Once

// ensureEnv loads environment-backed credentials once so adapters can look up
// OPENAI_API_KEY even when operators rely on OPENAI_API_KEY_FILE. Future engine
// integrations can reuse this hook to normalise additional secrets.
func ensureEnv() {
	ensureEnvOnce.Do(func() {
		if os.Getenv("OPENAI_API_KEY") != "" {
			return
		}

		if path := strings.TrimSpace(os.Getenv("OPENAI_API_KEY_FILE")); fs.FileExistsNotEmpty(path) {
			// #nosec G304 path provided via env
			if data, err := os.ReadFile(path); err == nil {
				if key := clean.Auth(string(data)); key != "" {
					_ = os.Setenv("OPENAI_API_KEY", key)
				}
			}
		}
	})
}
