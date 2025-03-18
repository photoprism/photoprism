//go:build freebsd

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
		if stat.Flags&unix.MNT_SUIDDIR != 0 {
			opts += ",suiddir"
		}
		if stat.Flags&unix.MNT_SOFTDEP != 0 {
			opts += ",softdep"
		}
		if stat.Flags&unix.MNT_NOSYMFOLLOW != 0 {
			opts += ",nosymfollow"
		}
		if stat.Flags&unix.MNT_GJOURNAL != 0 {
			opts += ",gjournal"
		}
		if stat.Flags&unix.MNT_MULTILABEL != 0 {
			opts += ",multilabel"
		}
		if stat.Flags&unix.MNT_ACLS != 0 {
			opts += ",acls"
		}
		if stat.Flags&unix.MNT_NOATIME != 0 {
			opts += ",noatime"
		}
		if stat.Flags&unix.MNT_NOCLUSTERR != 0 {
			opts += ",noclusterr"
		}
		if stat.Flags&unix.MNT_NOCLUSTERW != 0 {
			opts += ",noclusterw"
		}
		if stat.Flags&unix.MNT_NFS4ACLS != 0 {
			opts += ",nfsv4acls"
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
			Total:      (uint64(stat.Blocks) * uint64(stat.Bsize)),
			Free:       (uint64(stat.Bavail) * uint64(stat.Bsize)),
			Used:       (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(stat.Bsize),
			Inodes:     stat.Files,
			InodesFree: uint64(stat.Ffree),
			InodesUsed: stat.Files - uint64(stat.Ffree),
			Blocks:     uint64(stat.Blocks),
			BlockSize:  uint64(stat.Bsize),
		}
		d.DeviceType = deviceType(d)

		ret = append(ret, d)
	}

	return ret, warnings, nil
}
