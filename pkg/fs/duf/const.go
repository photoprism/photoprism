package duf

// Supported device types.
const (
	LocalDevice   = "local"
	NetworkDevice = "network"
	FuseDevice    = "fuse"
	SpecialDevice = "special"
	LoopsDevice   = "loops"
	BindsMount    = "binds"
)

// Size constants for KByte, MByte, and GByte.
const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
)
