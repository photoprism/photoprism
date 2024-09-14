//go:build debug
// +build debug

package server

import (
	"github.com/photoprism/photoprism/internal/api"
)

func init() {
	registerApiDocs = api.GetDocs
}
