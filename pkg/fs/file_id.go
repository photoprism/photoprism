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

// MaybeSameFile reports whether two files on different devices have the same size, modification time and inode
// number, as one file on a network share that is mounted twice does, e.g. over SMB with the server's inode numbers.
// Times are compared in whole seconds, since mounts may report them with another precision.
func MaybeSameFile(a, b os.FileInfo) bool {
	if a == nil || b == nil || a.Size() != b.Size() || !a.ModTime().Truncate(time.Second).Equal(b.ModTime().Truncate(time.Second)) {
		return false
	}

	aDev, aIno, aOk := fileID(a)
	bDev, bIno, bOk := fileID(b)

	return aOk && bOk && aDev != bDev && aIno == bIno
}
