//go:build debug
// +build debug

package server

import (
	"github.com/photoprism/photoprism/internal/api"
)

func init() {
	registerAPIDocs = api.GetDocs
}
