package server

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// WebDAVMethodPermissions returns the actions required by a WebDAV method.
func WebDAVMethodPermissions(method string) acl.Permissions {
	switch method {
	case header.MethodGet, header.MethodHead, header.MethodPost:
		return acl.Permissions{acl.ActionDownload}
	case header.MethodPropfind:
		return acl.Permissions{acl.ActionView}
	case header.MethodPut, header.MethodMkcol, header.MethodMove:
		return acl.Permissions{acl.ActionUpload}
	case header.MethodCopy:
		return acl.Permissions{acl.ActionDownload, acl.ActionUpdate}
	case header.MethodDelete:
		return acl.Permissions{acl.ActionDelete}
	case header.MethodProppatch, header.MethodLock, header.MethodUnlock:
		return acl.Permissions{acl.ActionUpdate}
	default:
		return nil
	}
}

// WebDAVWriteMethod reports whether a method requires a filesystem mutation action.
func WebDAVWriteMethod(method string) bool {
	for _, perm := range WebDAVMethodPermissions(method) {
		switch perm {
		case acl.ActionUpload, acl.ActionUpdate, acl.ActionDelete:
			return true
		}
	}

	return false
}
