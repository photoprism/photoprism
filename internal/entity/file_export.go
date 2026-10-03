package entity

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Exportable reports whether the session may export the file after row and download admission.
func (m *File) Exportable(sess *Session) bool {
	if m == nil {
		return false
	}
	if fs.FileType(m.FileName) != fs.SidecarYaml && fs.FileType("file."+m.FileType) != fs.SidecarYaml {
		return true
	}
	if sess == nil || !sess.IsRegistered() && !sess.IsClient() {
		return false
	}
	return sess.SeesAnyDetail(acl.ResourcePhotos) || sess.SeesAnyDetail(acl.ResourceFiles)
}
