//go:build openbsd

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
		if stat.F_flags&unix.MNT_RDONLY != 0 {
			opts = "ro"
		}
		if stat.F_flags&unix.MNT_SYNCHRONOUS != 0 {
			opts += ",sync"
		}
		if stat.F_flags&unix.MNT_NOEXEC != 0 {
			opts += ",noexec"
		}
		if stat.F_flags&unix.MNT_NOSUID != 0 {
			opts += ",nosuid"
		}
		if stat.F_flags&unix.MNT_NODEV != 0 {
			opts += ",nodev"
		}
		if stat.F_flags&unix.MNT_ASYNC != 0 {
			opts += ",async"
		}
		if stat.F_flags&unix.MNT_SOFTDEP != 0 {
			opts += ",softdep"
		}
		if stat.F_flags&unix.MNT_NOATIME != 0 {
			opts += ",noatime"
		}
		if stat.F_flags&unix.MNT_WXALLOWED != 0 {
			opts += ",wxallowed"
		}

		device := byteToString(stat.F_mntfromname[:])
		mountPoint := byteToString(stat.F_mntonname[:])
		fsType := byteToString(stat.F_fstypename[:])

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
			Total:      (uint64(stat.F_blocks) * uint64(stat.F_bsize)),
			Free:       (uint64(stat.F_bavail) * uint64(stat.F_bsize)),
			Used:       (uint64(stat.F_blocks) - uint64(stat.F_bfree)) * uint64(stat.F_bsize),
			Inodes:     stat.F_files,
			InodesFree: uint64(stat.F_ffree),
			InodesUsed: stat.F_files - uint64(stat.F_ffree),
			Blocks:     uint64(stat.F_blocks),
			BlockSize:  uint64(stat.F_bsize),
		}
		d.DeviceType = deviceType(d)

		ret = append(ret, d)
	}

	return ret, warnings, nil
}
