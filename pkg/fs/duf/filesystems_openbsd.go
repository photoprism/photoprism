//go:build openbsd

package duf

func isFuseFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isNetworkFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isSpecialFs(m Mount) bool {
	return m.Fstype == "devfs"
}

func isHiddenFs(m Mount) bool {
	return false
}
