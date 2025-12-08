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

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

var log = event.Log

var ensureEnvOnce sync.Once

// ensureEnv loads environment-backed credentials once so adapters can look up
// OPENAI_API_KEY / OLLAMA_API_KEY even when operators rely on *_FILE fallbacks.
// Future engine integrations can reuse this hook to normalise additional
// secrets.
func ensureEnv() {
	ensureEnvOnce.Do(func() {
		loadEnvKeyFromFile(openai.APIKeyEnv, openai.APIKeyFileEnv)
		loadEnvKeyFromFile(ollama.APIKeyEnv, ollama.APIKeyFileEnv)
	})
}

// loadEnvKeyFromFile populates envVar from fileVar when the environment value
// is empty and the referenced file exists and is non-empty.
func loadEnvKeyFromFile(envVar, fileVar string) {
	if os.Getenv(envVar) != "" {
		return
	}

	filePath := strings.TrimSpace(os.Getenv(fileVar))

	if !fs.FileExistsNotEmpty(filePath) {
		return
	}

	// #nosec G304 path provided via env
	if data, err := os.ReadFile(filePath); err == nil {
		if key := clean.Auth(string(data)); key != "" {
			_ = os.Setenv(envVar, key)
		}
	}
}
