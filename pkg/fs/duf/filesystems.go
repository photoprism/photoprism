package duf

import (
	"os"
	"path/filepath"
	"strings"
)

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

		if strings.HasPrefix(path, v.Mountpoint) {
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
