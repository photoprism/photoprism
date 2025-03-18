/*
Package duf provides file system usage information.

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

This code is copied and modified in part from:

  - https://github.com/muesli/duf
    MIT License, Copyright (c) 2020 Christian Muehlhaeuser
    see https://github.com/muesli/duf?tab=License-1-ov-file#readme

  - https://github.com/shirou/gopsutil
    BSD License, Copyright (c) 2014, WAKAYAMA Shirou
    see https://github.com/shirou/gopsutil?tab=License-1-ov-file#readme

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package duf

import (
	"fmt"
	"sort"
	"strings"
)

// Mounts returns all active file system mounts, along with any warnings or errors that have occurred.
func Mounts() (m []Mount, warnings []string, err error) {
	return mounts()
}

// PathInfo returns the closest file system mount for the given path, or an error if not found.
func PathInfo(dir string) (Mount, error) {
	if m, _, err := FindByPath(dir); err != nil {
		return Mount{}, err
	} else if len(m) < 1 {
		return Mount{}, fmt.Errorf("mount not found")
	} else {
		return m[0], nil
	}
}

// FindByPath returns the active file system mounts for the given path and it's parents, if any.
func FindByPath(dir string) (m []Mount, warnings []string, err error) {
	dir = strings.TrimSpace(dir)

	if dir == "" {
		return m, warnings, fmt.Errorf("empty path name")
	}

	folders := strings.Split(dir, "/")

	filter := FilterOptions{}

	if len(folders) <= 1 {
		filter.OnlyMountPoints = NewFilterValues("/")
	} else if parent := strings.TrimSpace(folders[1]); parent != "" {
		filter.OnlyMountPoints = NewFilterValues("/", "/"+parent+"*")
	} else if len(folders) > 2 {
		filter.OnlyMountPoints = NewFilterValues("/")
	}

	m, warnings, err = Find(filter)

	sort.SliceStable(m, func(i, j int) bool {
		return strings.Compare(m[i].Mountpoint, m[j].Mountpoint) >= 0
	})

	return m, warnings, err
}

// Find returns the active file system mounts that match the filter.
func Find(filters FilterOptions) (results []Mount, warnings []string, err error) {
	m, warnings, err := mounts()

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

		results = append(results, v)
	}

	return results, warnings, err
}
