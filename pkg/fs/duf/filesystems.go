package duf

import (
	"os"
	"path/filepath"
	"strings"
)

// findMounts returns the mounts that contain path, closest (longest mountpoint) first.
func findMounts(mounts []Mount, path string) ([]Mount, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(path)
	if err != nil {
		return nil, err
	}

	var m []Mount
	for _, v := range mounts {
		if path == v.Device {
			return []Mount{v}, nil
		}

		if mountContains(v.Mountpoint, path) {
			var nm []Mount

			// keep all entries that are as close or closer to the target
			for _, mv := range m {
				if len(mv.Mountpoint) >= len(v.Mountpoint) {
					nm = append(nm, mv)
				}
			}
			m = nm

			// add entry only if we didn't already find something closer
			if len(nm) == 0 || len(v.Mountpoint) >= len(nm[0].Mountpoint) {
				m = append(m, v)
			}
		}
	}

	return m, nil
}

// mountContains reports whether path is at or below mountpoint, comparing on path
// segment boundaries so a mountpoint such as /media does not spuriously match a
// sibling path like /media-data. The root mountpoint contains every absolute path.
func mountContains(mountpoint, path string) bool {
	if path == mountpoint {
		return true
	}

	sep := string(filepath.Separator)

	// A trailing separator (e.g. the root "/") already marks the segment boundary.
	if strings.HasSuffix(mountpoint, sep) {
		return strings.HasPrefix(path, mountpoint)
	}

	return strings.HasPrefix(path, mountpoint+sep)
}

func deviceType(m Mount) string {
	if isNetworkFs(m) {
		return NetworkDevice
	}
	if isSpecialFs(m) {
		return SpecialDevice
	}
	if isFuseFs(m) {
		return FuseDevice
	}

	return LocalDevice
}

// remote: [ "nfs", "smbfs", "cifs", "ncpfs", "afs", "coda", "ftpfs", "mfs", "sshfs", "fuse.sshfs", "nfs4" ]
// special: [ "tmpfs", "devpts", "devtmpfs", "proc", "sysfs", "usbfs", "devfs", "fdescfs", "linprocfs" ]
