package clean

import (
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

// IP returns the sanitized and normalized network address if it is valid, or the default otherwise.
func IP(s, defaultIp string) string {
	return header.IP(s, defaultIp)
}
