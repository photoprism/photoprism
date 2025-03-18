package duf

import (
	"strings"
)

var (
	groups = []string{LocalDevice, NetworkDevice, FuseDevice, SpecialDevice, LoopsDevice, BindsMount}
)

type GroupedMounts map[string][]Mount

func GroupMounts(m []Mount, filters FilterOptions) GroupedMounts {
	deviceMounts := make(GroupedMounts)
	hasOnlyDevices := len(filters.OnlyDevices) != 0

	_, hideLocal := filters.HiddenDevices[LocalDevice]
	_, hideNetwork := filters.HiddenDevices[NetworkDevice]
	_, hideFuse := filters.HiddenDevices[FuseDevice]
	_, hideSpecial := filters.HiddenDevices[SpecialDevice]
	_, hideLoops := filters.HiddenDevices[LoopsDevice]
	_, hideBinds := filters.HiddenDevices[BindsMount]

	_, onlyLocal := filters.OnlyDevices[LocalDevice]
	_, onlyNetwork := filters.OnlyDevices[NetworkDevice]
	_, onlyFuse := filters.OnlyDevices[FuseDevice]
	_, onlySpecial := filters.OnlyDevices[SpecialDevice]
	_, onlyLoops := filters.OnlyDevices[LoopsDevice]
	_, onlyBinds := filters.OnlyDevices[BindsMount]

	// sort/filter devices
	for _, v := range m {
		if len(filters.OnlyFilesystems) != 0 {
			// skip not onlyFs
			if _, ok := filters.OnlyFilesystems[strings.ToLower(v.Fstype)]; !ok {
				continue
			}
		} else {
			// skip hideFs
			if _, ok := filters.HiddenFilesystems[strings.ToLower(v.Fstype)]; ok {
				continue
			}
		}

		// skip hidden devices
		if isHiddenFs(v) && !all {
			continue
		}

		// skip bind-mounts
		if strings.Contains(v.Opts, "bind") {
			if (hasOnlyDevices && !onlyBinds) || (hideBinds && !all) {
				continue
			}
		}

		// skip loop devices
		if strings.HasPrefix(v.Device, "/dev/loop") {
			if (hasOnlyDevices && !onlyLoops) || (hideLoops && !all) {
				continue
			}
		}

		// skip special devices
		if v.Blocks == 0 && !all {
			continue
		}

		// skip zero size devices
		if v.BlockSize == 0 && !all {
			continue
		}

		// skip not only mount point
		if len(filters.OnlyMountPoints) != 0 {
			if !findInKey(v.Mountpoint, filters.OnlyMountPoints) {
				continue
			}
		}

		// skip hidden mount point
		if len(filters.HiddenMountPoints) != 0 {
			if findInKey(v.Mountpoint, filters.HiddenMountPoints) {
				continue
			}
		}

		t := deviceType(v)

		if !all {
			switch {
			case hasOnlyDevices && onlyLocal && t != LocalDevice:
				continue
			case hasOnlyDevices && onlyNetwork && t != NetworkDevice:
				continue
			case hasOnlyDevices && onlyFuse && t != FuseDevice:
				continue
			case hasOnlyDevices && onlySpecial && t != SpecialDevice:
				continue
			case
				t == LocalDevice && hideLocal,
				t == NetworkDevice && hideNetwork,
				t == FuseDevice && hideFuse,
				t == SpecialDevice && hideSpecial:
				continue
			}
		}

		deviceMounts[t] = append(deviceMounts[t], v)
	}

	return deviceMounts
}
