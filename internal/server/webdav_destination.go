package server

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// WebDAVDestinationStatus checks collection and account paths for COPY and MOVE destinations.
func WebDAVDestinationStatus(r *http.Request, prefix string, user *entity.User) int {
	if r.Method != header.MethodCopy && r.Method != header.MethodMove {
		return http.StatusOK
	}

	destination := r.Header.Get("Destination")
	if destination == "" {
		return http.StatusBadRequest
	}

	u, err := url.Parse(destination)
	switch {
	case err != nil:
		return http.StatusBadRequest
	case u.Host != "" && u.Host != r.Host:
		return http.StatusBadGateway
	case prefix != "" && !hasCollectionPath(u.Path, prefix):
		return http.StatusNotFound
	}

	rel := strings.TrimPrefix(u.Path, prefix)
	if rel == "" {
		return http.StatusBadGateway
	}

	// Match webdav.Dir's rooted slash normalization after decoding the destination URL once.
	name := path.Clean("/" + rel)
	if base := user.GetBasePath(); base == "" && user.RequiresBasePath() {
		return http.StatusForbidden
	} else if base != "" && !hasCollectionPath(name, "/"+base) {
		return http.StatusForbidden
	}

	if upload := user.GetUploadPath(); upload != "" {
		root := "/" + upload
		if name == root || !hasCollectionPath(name, root) {
			return http.StatusForbidden
		}
	}

	return http.StatusOK
}
