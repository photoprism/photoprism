package fs

import (
	"os"
	"syscall"
	"time"
)

// fileID returns the device and inode number of a file, and can be replaced in tests.
var fileID = func(info os.FileInfo) (dev, ino uint64, ok bool) {
	if info == nil {
		return 0, 0, false
	} else if st, isStat := info.Sys().(*syscall.Stat_t); isStat {
		return uint64(st.Dev), st.Ino, true //nolint:gosec,unconvert // Dev is int32 on some platforms.
	}

	return 0, 0, false
}

// FileID returns the device and inode number of a file, if the file system reports them.
func FileID(info os.FileInfo) (dev, ino uint64, ok bool) {
	return fileID(info)
}

// maybeSameTime is how far apart the modification times of one file may be when read through different mounts.
const maybeSameTime = 2 * time.Second

// MaybeSameFile reports whether two files on different devices have the same size, modification time and inode
// number, as one file on a network share that is mounted twice does, e.g. over SMB with the server's inode numbers.
// Times may differ by up to two seconds, as mounts may round them or keep another precision. The result only means
// something for two files already known to have the same content.
func MaybeSameFile(a, b os.FileInfo) bool {
	if a == nil || b == nil || a.Size() != b.Size() {
		return false
	} else if d := a.ModTime().Sub(b.ModTime()); d > maybeSameTime || d < -maybeSameTime {
		return false
	}

	aDev, aIno, aOk := fileID(a)
	bDev, bIno, bOk := fileID(b)

	return aOk && bOk && aDev != bDev && aIno == bIno
}
