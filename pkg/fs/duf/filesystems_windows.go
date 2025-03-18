//go:build windows

package duf

import (
	"golang.org/x/sys/windows/registry"
)

const (
	WindowsSandboxMountPointRegistryPath = `Software\Microsoft\Windows\CurrentVersion\Explorer\MountPoints2\CPC\LocalMOF`
)

var windowsSandboxMountPoints = loadRegisteredWindowsSandboxMountPoints()

func loadRegisteredWindowsSandboxMountPoints() (ret FilterValues) {
	ret = make(FilterValues)
	key, err := registry.OpenKey(registry.CURRENT_USER, WindowsSandboxMountPointRegistryPath, registry.READ)
	if err != nil {
		return
	}

	keyInfo, err := key.Stat()
	if err != nil {
		return
	}

	mountPoints, err := key.ReadValueNames(int(keyInfo.ValueCount))
	if err != nil {
		return
	}

	for _, val := range mountPoints {
		ret[val] = struct{}{}
	}
	return ret
}

func isFuseFs(m Mount) bool {
	//FIXME: implement
	return false
}

func isNetworkFs(m Mount) bool {
	_, ok := m.Metadata.(*NetResource)
	return ok
}

func isSpecialFs(m Mount) bool {
	_, ok := windowsSandboxMountPoints[m.Mountpoint]
	return ok
}

func isHiddenFs(m Mount) bool {
	return false
}
