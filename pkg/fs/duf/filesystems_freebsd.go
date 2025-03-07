//go:build freebsd

package duf

import "slices"

func isFuseFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isNetworkFs(m Mount) bool {
	fs := []string{"nfs", "smbfs"}

	return slices.Contains(fs, m.Fstype)
}

func isSpecialFs(m Mount) bool {
	fs := []string{"devfs", "tmpfs", "linprocfs", "linsysfs", "fdescfs", "procfs"}

	return slices.Contains(fs, m.Fstype)
}

func isHiddenFs(m Mount) bool {
	return false
}
