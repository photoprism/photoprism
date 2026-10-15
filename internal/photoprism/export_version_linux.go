//go:build linux

package photoprism

import (
	"fmt"
	"os"
	"syscall"
)

// exportFileVersion includes ctime so restored mtimes cannot reuse a stale content verification.
func exportFileVersion(info os.FileInfo) string {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%d:%d:%d:%d:%d:%d:%d", stat.Dev, stat.Ino, stat.Size, stat.Mtim.Sec, stat.Mtim.Nsec, stat.Ctim.Sec, stat.Ctim.Nsec)
}
