//go:build windows

package duf

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Local devices
const (
	guidBufLen       = windows.MAX_PATH + 1
	volumeNameBufLen = windows.MAX_PATH + 1
	rootPathBufLen   = windows.MAX_PATH + 1
	fileSystemBufLen = windows.MAX_PATH + 1
)

func getMountPoint(guidBuf []uint16) (mountPoint string, err error) {
	var rootPathLen uint32
	rootPathBuf := make([]uint16, rootPathBufLen)

	err = windows.GetVolumePathNamesForVolumeName(&guidBuf[0], &rootPathBuf[0], rootPathBufLen*2, &rootPathLen)
	if err != nil && err.(windows.Errno) == windows.ERROR_MORE_DATA {
		// Retry if buffer size is too small
		rootPathBuf = make([]uint16, (rootPathLen+1)/2)
		err = windows.GetVolumePathNamesForVolumeName(
			&guidBuf[0], &rootPathBuf[0], rootPathLen, &rootPathLen)
	}
	return windows.UTF16ToString(rootPathBuf), err
}

func getVolumeInfo(guidOrMountPointBuf []uint16) (volumeName string, fsType string, err error) {
	volumeNameBuf := make([]uint16, volumeNameBufLen)
	fsTypeBuf := make([]uint16, fileSystemBufLen)

	err = windows.GetVolumeInformation(&guidOrMountPointBuf[0], &volumeNameBuf[0], volumeNameBufLen*2,
		nil, nil, nil,
		&fsTypeBuf[0], fileSystemBufLen*2)

	return windows.UTF16ToString(volumeNameBuf), windows.UTF16ToString(fsTypeBuf), err
}

func getSpaceInfo(guidOrMountPointBuf []uint16) (totalBytes uint64, freeBytes uint64, err error) {
	err = windows.GetDiskFreeSpaceEx(&guidOrMountPointBuf[0], nil, &totalBytes, &freeBytes)
	return
}

func getClusterInfo(guidOrMountPointBuf []uint16) (totalClusters uint32, clusterSize uint32, err error) {
	var sectorsPerCluster uint32
	var bytesPerSector uint32
	err = GetDiskFreeSpace(&guidOrMountPointBuf[0], &sectorsPerCluster, &bytesPerSector, nil, &totalClusters)
	clusterSize = bytesPerSector * sectorsPerCluster
	return
}

func getMount(guidOrMountPointBuf []uint16, isGUID bool) (m Mount, skip bool, warnings []string) {
	var err error
	guidOrMountPoint := windows.UTF16ToString(guidOrMountPointBuf)

	mountPoint := guidOrMountPoint
	if isGUID {
		mountPoint, err = getMountPoint(guidOrMountPointBuf)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %s", guidOrMountPoint, err))
		}
		// Skip unmounted volumes
		if len(mountPoint) == 0 {
			skip = true
			return
		}
	}

	// Get volume name & filesystem type
	volumeName, fsType, err := getVolumeInfo(guidOrMountPointBuf)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("%s: %s", guidOrMountPoint, err))
	}

	// Get space info
	totalBytes, freeBytes, err := getSpaceInfo(guidOrMountPointBuf)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("%s: %s", guidOrMountPoint, err))
	}

	// Get cluster info
	totalClusters, clusterSize, err := getClusterInfo(guidOrMountPointBuf)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("%s: %s", guidOrMountPoint, err))
	}

	m = Mount{
		Device:     volumeName,
		Mountpoint: mountPoint,
		Fstype:     fsType,
		Type:       fsType,
		Opts:       "",
		Total:      totalBytes,
		Free:       freeBytes,
		Used:       totalBytes - freeBytes,
		Blocks:     uint64(totalClusters),
		BlockSize:  uint64(clusterSize),
	}
	m.DeviceType = deviceType(m)
	return
}

func getMountFromGUID(guidBuf []uint16) (m Mount, skip bool, warnings []string) {
	m, skip, warnings = getMount(guidBuf, true)

	// Use GUID as volume name if no label was set
	if len(m.Device) == 0 {
		m.Device = windows.UTF16ToString(guidBuf)
	}

	return
}

func getMountFromMountPoint(mountPointBuf []uint16) (m Mount, warnings []string) {
	m, _, warnings = getMount(mountPointBuf, false)

	// Use mount point as volume name if no label was set
	if len(m.Device) == 0 {
		m.Device = windows.UTF16ToString(mountPointBuf)
	}

	return m, warnings
}

func appendLocalMounts(mounts []Mount, warnings []string) ([]Mount, []string, error) {
	guidBuf := make([]uint16, guidBufLen)

	hFindVolume, err := windows.FindFirstVolume(&guidBuf[0], guidBufLen*2)
	if err != nil {
		return mounts, warnings, err
	}

VolumeLoop:
	for ; ; err = windows.FindNextVolume(hFindVolume, &guidBuf[0], guidBufLen*2) {
		if err != nil {
			switch err.(windows.Errno) {
			case windows.ERROR_NO_MORE_FILES:
				break VolumeLoop
			default:
				warnings = append(warnings, fmt.Sprintf("%s: %s", windows.UTF16ToString(guidBuf), err))
				continue VolumeLoop
			}
		}

		if m, skip, w := getMountFromGUID(guidBuf); !skip {
			mounts = append(mounts, m)
			warnings = append(warnings, w...)
		}
	}

	if err = windows.FindVolumeClose(hFindVolume); err != nil {
		warnings = append(warnings, fmt.Sprintf("%s", err))
	}
	return mounts, warnings, nil
}

// Network devices
func getMountFromNetResource(netResource NetResource) (m Mount, warnings []string) {
	mountPoint := windows.UTF16PtrToString(netResource.LocalName)
	if !strings.HasSuffix(mountPoint, string(filepath.Separator)) {
		mountPoint += string(filepath.Separator)
	}
	mountPointBuf := windows.StringToUTF16(mountPoint)

	m, _, warnings = getMount(mountPointBuf, false)

	// Use remote name as volume name if no label was set
	if len(m.Device) == 0 {
		m.Device = windows.UTF16PtrToString(netResource.RemoteName)
	}

	return
}

func appendNetworkMounts(mounts []Mount, warnings []string) ([]Mount, []string, error) {
	hEnumResource, err := WNetOpenEnum(RESOURCE_CONNECTED, RESOURCETYPE_DISK, RESOURCEUSAGE_CONNECTABLE, nil)
	if err != nil {
		return mounts, warnings, err
	}

EnumLoop:
	for {
		// Reference: https://docs.microsoft.com/en-us/windows/win32/wnet/enumerating-network-resources
		var nrBuf [16384]byte
		count := uint32(math.MaxUint32)
		size := uint32(len(nrBuf))
		if err := WNetEnumResource(hEnumResource, &count, &nrBuf[0], &size); err != nil {
			switch err.(windows.Errno) {
			case windows.ERROR_NO_MORE_ITEMS:
				break EnumLoop
			default:
				warnings = append(warnings, err.Error())
				break EnumLoop
			}
		}

		for i := uint32(0); i < count; i++ {
			nr := (*NetResource)(unsafe.Pointer(&nrBuf[uintptr(i)*NetResourceSize]))
			m, w := getMountFromNetResource(*nr)
			mounts = append(mounts, m)
			warnings = append(warnings, w...)
		}
	}

	if err = WNetCloseEnum(hEnumResource); err != nil {
		warnings = append(warnings, fmt.Sprintf("%s", err))
	}
	return mounts, warnings, nil
}

func mountPointAlreadyPresent(mounts []Mount, mountPoint string) bool {
	for _, m := range mounts {
		if m.Mountpoint == mountPoint {
			return true
		}
	}

	return false
}

func appendLogicalDrives(mounts []Mount, warnings []string) ([]Mount, []string) {
	driveBitmap, err := windows.GetLogicalDrives()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("GetLogicalDrives(): %s", err))
		return mounts, warnings
	}

	for drive := 'A'; drive <= 'Z'; drive, driveBitmap = drive+1, driveBitmap>>1 {
		if driveBitmap&0x1 == 0 {
			continue
		}

		mountPoint := fmt.Sprintf("%c:\\", drive)
		if mountPointAlreadyPresent(mounts, mountPoint) {
			continue
		}

		mountPointBuf := windows.StringToUTF16(mountPoint)
		m, w := getMountFromMountPoint(mountPointBuf)
		mounts = append(mounts, m)
		warnings = append(warnings, w...)
	}

	return mounts, warnings
}

func mounts() (ret []Mount, warnings []string, err error) {
	ret = make([]Mount, 0)

	// Local devices
	if ret, warnings, err = appendLocalMounts(ret, warnings); err != nil {
		return
	}

	// Network devices
	if ret, warnings, err = appendNetworkMounts(ret, warnings); err != nil {
		return
	}

	// Logical devices (from GetLogicalDrives bitflag)
	// Check any possible logical drives, in case of some special virtual devices, such as RAM disk
	ret, warnings = appendLogicalDrives(ret, warnings)

	return ret, warnings, nil
}

// Windows API
const (
	// Windows Networking const
	// Reference: https://docs.microsoft.com/en-us/windows/win32/api/winnetwk/nf-winnetwk-wnetopenenumw
	RESOURCE_CONNECTED  = 0x00000001
	RESOURCE_GLOBALNET  = 0x00000002
	RESOURCE_REMEMBERED = 0x00000003
	RESOURCE_RECENT     = 0x00000004
	RESOURCE_CONTEXT    = 0x00000005

	RESOURCETYPE_ANY      = 0x00000000
	RESOURCETYPE_DISK     = 0x00000001
	RESOURCETYPE_PRINT    = 0x00000002
	RESOURCETYPE_RESERVED = 0x00000008
	RESOURCETYPE_UNKNOWN  = 0xFFFFFFFF

	RESOURCEUSAGE_CONNECTABLE   = 0x00000001
	RESOURCEUSAGE_CONTAINER     = 0x00000002
	RESOURCEUSAGE_NOLOCALDEVICE = 0x00000004
	RESOURCEUSAGE_SIBLING       = 0x00000008
	RESOURCEUSAGE_ATTACHED      = 0x00000010
	RESOURCEUSAGE_ALL           = RESOURCEUSAGE_CONNECTABLE | RESOURCEUSAGE_CONTAINER | RESOURCEUSAGE_ATTACHED
	RESOURCEUSAGE_RESERVED      = 0x80000000
)

var (
	// Windows syscall
	modmpr      = windows.NewLazySystemDLL("mpr.dll")
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procWNetOpenEnumW     = modmpr.NewProc("WNetOpenEnumW")
	procWNetCloseEnum     = modmpr.NewProc("WNetCloseEnum")
	procWNetEnumResourceW = modmpr.NewProc("WNetEnumResourceW")
	procGetDiskFreeSpaceW = modkernel32.NewProc("GetDiskFreeSpaceW")

	NetResourceSize = unsafe.Sizeof(NetResource{})
)

// Reference: https://docs.microsoft.com/en-us/windows/win32/api/winnetwk/ns-winnetwk-netresourcew
type NetResource struct {
	Scope       uint32
	Type        uint32
	DisplayType uint32
	Usage       uint32
	LocalName   *uint16
	RemoteName  *uint16
	Comment     *uint16
	Provider    *uint16
}

// Reference: https://docs.microsoft.com/en-us/windows/win32/api/winnetwk/nf-winnetwk-wnetopenenumw
func WNetOpenEnum(scope uint32, resourceType uint32, usage uint32, resource *NetResource) (handle windows.Handle, err error) {
	r1, _, e1 := syscall.Syscall6(procWNetOpenEnumW.Addr(), 5, uintptr(scope), uintptr(resourceType), uintptr(usage), uintptr(unsafe.Pointer(resource)), uintptr(unsafe.Pointer(&handle)), 0)
	if r1 != windows.NO_ERROR {
		if e1 != 0 {
			err = e1
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// Reference: https://docs.microsoft.com/en-us/windows/win32/api/winnetwk/nf-winnetwk-wnetenumresourcew
func WNetEnumResource(enumResource windows.Handle, count *uint32, buffer *byte, bufferSize *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procWNetEnumResourceW.Addr(), 4, uintptr(enumResource), uintptr(unsafe.Pointer(count)), uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(bufferSize)), 0, 0)
	if r1 != windows.NO_ERROR {
		if e1 != 0 {
			err = e1
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// Reference: https://docs.microsoft.com/en-us/windows/win32/api/winnetwk/nf-winnetwk-wnetcloseenum
func WNetCloseEnum(enumResource windows.Handle) (err error) {
	r1, _, e1 := syscall.Syscall(procWNetCloseEnum.Addr(), 1, uintptr(enumResource), 0, 0)
	if r1 != windows.NO_ERROR {
		if e1 != 0 {
			err = e1
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// Reference: https://docs.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getdiskfreespacew
func GetDiskFreeSpace(directoryName *uint16, sectorsPerCluster *uint32, bytesPerSector *uint32, numberOfFreeClusters *uint32, totalNumberOfClusters *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procGetDiskFreeSpaceW.Addr(), 5, uintptr(unsafe.Pointer(directoryName)), uintptr(unsafe.Pointer(sectorsPerCluster)), uintptr(unsafe.Pointer(bytesPerSector)), uintptr(unsafe.Pointer(numberOfFreeClusters)), uintptr(unsafe.Pointer(totalNumberOfClusters)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = e1
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
