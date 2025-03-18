//go:build darwin

package duf

import (
	"golang.org/x/sys/unix"
)

func (m *Mount) Stat() unix.Statfs_t {
	return m.Metadata.(unix.Statfs_t)
}

func mounts() ([]Mount, []string, error) {
	var ret []Mount
	var warnings []string

	count, err := unix.Getfsstat(nil, unix.MNT_WAIT)
	if err != nil {
		return nil, nil, err
	}
	fs := make([]unix.Statfs_t, count)
	if _, err = unix.Getfsstat(fs, unix.MNT_WAIT); err != nil {
		return nil, nil, err
	}

	for _, stat := range fs {
		opts := "rw"
		if stat.Flags&unix.MNT_RDONLY != 0 {
			opts = "ro"
		}
		if stat.Flags&unix.MNT_SYNCHRONOUS != 0 {
			opts += ",sync"
		}
		if stat.Flags&unix.MNT_NOEXEC != 0 {
			opts += ",noexec"
		}
		if stat.Flags&unix.MNT_NOSUID != 0 {
			opts += ",nosuid"
		}
		if stat.Flags&unix.MNT_UNION != 0 {
			opts += ",union"
		}
		if stat.Flags&unix.MNT_ASYNC != 0 {
			opts += ",async"
		}
		if stat.Flags&unix.MNT_DONTBROWSE != 0 {
			opts += ",nobrowse"
		}
		if stat.Flags&unix.MNT_AUTOMOUNTED != 0 {
			opts += ",automounted"
		}
		if stat.Flags&unix.MNT_JOURNALED != 0 {
			opts += ",journaled"
		}
		if stat.Flags&unix.MNT_MULTILABEL != 0 {
			opts += ",multilabel"
		}
		if stat.Flags&unix.MNT_NOATIME != 0 {
			opts += ",noatime"
		}
		if stat.Flags&unix.MNT_NODEV != 0 {
			opts += ",nodev"
		}

		device := byteToString(stat.Mntfromname[:])
		mountPoint := byteToString(stat.Mntonname[:])
		fsType := byteToString(stat.Fstypename[:])

		if len(device) == 0 {
			continue
		}

		d := Mount{
			Device:     device,
			Mountpoint: mountPoint,
			Fstype:     fsType,
			Type:       fsType,
			Opts:       opts,
			Metadata:   stat,
			Total:      stat.Blocks * uint64(stat.Bsize),
			Free:       stat.Bavail * uint64(stat.Bsize),
			Used:       (stat.Blocks - stat.Bfree) * uint64(stat.Bsize),
			Inodes:     stat.Files,
			InodesFree: stat.Ffree,
			InodesUsed: stat.Files - stat.Ffree,
			Blocks:     stat.Blocks,
			BlockSize:  uint64(stat.Bsize),
		}
		d.DeviceType = deviceType(d)

		ret = append(ret, d)
	}

	return ret, warnings, nil
}
